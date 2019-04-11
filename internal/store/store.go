package store

import (
	"database/sql"
	"errors"
	"fmt"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/wedaly/local-news/internal/feed"
	"log"
	"time"
)

const numStatements int = 11

const (
	selectEveryFeedStmt = iota
	selectFeedStmt
	selectFeedIdByUrlStmt
	insertFeedStmt
	updateFeedStmt
	deleteFeedStmt
	selectFeedItemsForFeedStmt
	upsertFeedItemStmt
	deleteItemsInFeedStmt
	upsertFeedSyncStatusStmt
	selectFeedSyncStatusStmt
)

// FeedStore provides thread-safe CRUD operations for feeds and feed items
type FeedStore struct {
	dbPath     string
	db         *sql.DB
	statements []*sql.Stmt
}

// NewFeedStore configures a feed store instance.
// This does NOT open a database connection.
// See the SQLite driver documentation for the `dbPath` format
func NewFeedStore(dbPath string) *FeedStore {
	return &FeedStore{dbPath: dbPath}
}

// Initialize opens the database and (if necessary) creates tables,
// indices, and prepared statements.
func (s *FeedStore) Initialize() error {
	if s.db != nil {
		panic("Store already initialized")
	}

	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return err
	}

	s.db = db

	if err := s.enableForeignKeyConstraints(); err != nil {
		return err
	}

	if err := s.installSchema(); err != nil {
		return err
	}

	if err := s.prepareStatements(); err != nil {
		return err
	}

	return nil
}

// Close gracefully shuts down the database
// This must be called before the application exits
func (s *FeedStore) Close() {
	for _, stmt := range s.statements {
		stmt.Close()
	}

	s.db.Close()
}

// InsertFeedWithUrl creates a feed record for the specified URL
// This is a placeholder with no feed items, meant to be updated
// once the feed is loaded from an external source (e.g. RSS XML)
func (s *FeedStore) GetOrCreateFeedWithUrl(url string) (FeedId, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}

	// Insert a placeholder record, ignoring conflicts
	// This ensures that a record exists for the specified feed URL
	stmt := tx.Stmt(s.statements[insertFeedStmt])
	placeholderName := url
	if _, err := stmt.Exec(url, placeholderName); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Fatalf("Unable to rollback: %v", err)
		}
		return 0, err
	}

	// Retrieve the feed ID
	stmt = tx.Stmt(s.statements[selectFeedIdByUrlStmt])
	var id FeedId
	if err := stmt.QueryRow(url).Scan(&id); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Fatalf("Unable to rollback: %v", err)
		}
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

// SyncFeed atomically updates a feed record and its items.
// The feed name (but not ID) is overwritten with the new name.
// The feed items are upserted, using the record GUID as the record's identity.
// Existing items NOT included in the new feed are retained (not deleted)
func (s *FeedStore) SyncFeed(id FeedId, feed feed.Feed) error {
	return s.wrapInTx(func(tx *sql.Tx) error {
		err := s.updateFeedRecord(tx, id, feed)
		if err != nil {
			return err
		}

		for _, item := range feed.Items {
			err := s.upsertFeedItemRecord(tx, id, item)
			if err != nil {
				return err
			}
		}

		err = s.setFeedSyncStatusSuccess(tx, id)
		if err != nil {
			return err
		}

		return nil
	})
}

// DeleteFeed transactionally deletes the specified feed and all its items
func (s *FeedStore) DeleteFeed(feedId FeedId) error {
	return s.wrapInTx(func(tx *sql.Tx) error {
		if err := s.deleteItemsInFeed(tx, feedId); err != nil {
			return err
		}

		if err := s.deleteFeedRecord(tx, feedId); err != nil {
			return err
		}

		return nil
	})
}

// RetrieveFeeds retrieves a record for every feed in the database
func (s *FeedStore) RetrieveFeeds() ([]FeedRecord, error) {
	stmt := s.statements[selectEveryFeedStmt]
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]FeedRecord, 0)
	for rows.Next() {
		var id int64
		var url string
		var name string

		if err := rows.Scan(&id, &url, &name); err != nil {
			return nil, err
		}

		records = append(records, FeedRecord{
			Id:   FeedId(id),
			Url:  url,
			Name: name,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// RetrieveFeed retrieves a single feed record by its id.
func (s *FeedStore) RetrieveFeed(id FeedId) (FeedRecord, error) {
	var url, name string

	stmt := s.statements[selectFeedStmt]
	err := stmt.QueryRow(id).Scan(&url, &name)
	if err != nil {
		return FeedRecord{}, err
	}

	record := FeedRecord{
		Id:   id,
		Url:  url,
		Name: name,
	}
	return record, nil
}

// RetrieveFeedItems retrieves a record for every feed item for a given feed
func (s *FeedStore) RetrieveFeedItems(feedId FeedId) ([]FeedItemRecord, error) {
	stmt := s.statements[selectFeedItemsForFeedStmt]
	rows, err := stmt.Query(feedId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]FeedItemRecord, 0)
	for rows.Next() {
		var id int64
		var guid string
		var url string
		var title string
		var date int64

		if err := rows.Scan(&id, &guid, &url, &title, &date); err != nil {
			return nil, err
		}

		records = append(records, FeedItemRecord{
			Id:    FeedItemId(id),
			Title: title,
			Date:  time.Unix(date, 0),
			Url:   url,
			Guid:  guid,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// SetFeedSyncStatusError sets the most recent sync attempt to "error" status
func (s *FeedStore) SetFeedSyncStatusError(id FeedId, syncErr error) error {
	stmt := s.statements[upsertFeedSyncStatusStmt]
	syncErrStr := fmt.Sprintf("%v", syncErr)
	_, err := stmt.Exec(id, false, syncErrStr)
	return err
}

// RetrieveFeedSyncStatus retrieves the status of
// the most recently completed sync attempt
// If no attempt has yet completed because the feed was just added,
// then the first return value will be false.
func (s *FeedStore) RetrieveFeedSyncStatus(id FeedId) (bool, FeedSyncStatus, error) {
	var date int64
	var success bool
	var syncErrVal sql.NullString

	stmt := s.statements[selectFeedSyncStatusStmt]
	err := stmt.QueryRow(id).Scan(&date, &success, &syncErrVal)
	if err == sql.ErrNoRows {
		return false, FeedSyncStatus{}, nil
	} else if err != nil {
		return false, FeedSyncStatus{}, err
	}

	var syncErr error
	if syncErrVal.Valid {
		syncErr = errors.New(syncErrVal.String)
	}

	status := FeedSyncStatus{
		Date:    time.Unix(date, 0),
		Success: success,
		Error:   syncErr,
	}
	return true, status, nil
}

func (s *FeedStore) enableForeignKeyConstraints() error {
	sql := "PRAGMA foreign_keys = ON;"
	_, err := s.db.Exec(sql)
	return err
}

func (s *FeedStore) installSchema() error {
	sql := `
	CREATE TABLE IF NOT EXISTS feed (
		id INTEGER NOT NULL PRIMARY KEY,
		url VARCHAR UNIQUE NOT NULL,
		name VARCHAR NOT NULL
	);

	CREATE TABLE IF NOT EXISTS feed_item (
		id INTEGER NOT NULL PRIMARY KEY,
		feed_id INTEGER NOT NULL,
		guid VARCHAR NOT NULL,
		url VARCHAR NOT NULL,
		title VARCHAR NOT NULL,
		date INTEGER NOT NULL,
		FOREIGN KEY (feed_id) REFERENCES feed(id)
	);

	CREATE UNIQUE INDEX IF NOT EXISTS feed_item_feed_guid_idx
		ON feed_item(feed_id, guid);

	CREATE INDEX IF NOT EXISTS feed_item_date_idx
		ON feed_item(date);

	CREATE TABLE IF NOT EXISTS feed_sync_status (
		feed_id INTEGER NOT NULL PRIMARY KEY,
		date INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
		success INTEGER NOT NULL,
		error TEXT,
		FOREIGN KEY (feed_id)
			REFERENCES feed(id)
			ON DELETE CASCADE
	);
	`
	_, err := s.db.Exec(sql)
	return err
}

func (s *FeedStore) prepareStatements() error {
	s.statements = make([]*sql.Stmt, numStatements)

	selectEveryFeedSql := `
		SELECT id, url, name FROM feed
		ORDER BY name ASC`
	if stmt, err := s.db.Prepare(selectEveryFeedSql); err != nil {
		return err
	} else {
		s.statements[selectEveryFeedStmt] = stmt
	}

	selectFeedSql := "SELECT url, name FROM feed WHERE id = ?"
	if stmt, err := s.db.Prepare(selectFeedSql); err != nil {
		return err
	} else {
		s.statements[selectFeedStmt] = stmt
	}

	selectFeedIdByUrlSql := "SELECT id FROM feed WHERE url = ?"
	if stmt, err := s.db.Prepare(selectFeedIdByUrlSql); err != nil {
		return err
	} else {
		s.statements[selectFeedIdByUrlStmt] = stmt
	}

	insertFeedSql := `
		INSERT INTO feed (url, name) VALUES (?, ?)
		ON CONFLICT(url) DO NOTHING`
	if stmt, err := s.db.Prepare(insertFeedSql); err != nil {
		return err
	} else {
		s.statements[insertFeedStmt] = stmt
	}

	updateFeedSql := "UPDATE feed SET name = ? WHERE id = ?"
	if stmt, err := s.db.Prepare(updateFeedSql); err != nil {
		return err
	} else {
		s.statements[updateFeedStmt] = stmt
	}

	deleteFeedSql := "DELETE FROM feed WHERE id = ?"
	if stmt, err := s.db.Prepare(deleteFeedSql); err != nil {
		return err
	} else {
		s.statements[deleteFeedStmt] = stmt
	}

	selectFeedItemsForFeedSql := `
		SELECT id, guid, url, title, date
		FROM feed_item
		WHERE feed_id = ?
		ORDER BY date DESC, title ASC`
	if stmt, err := s.db.Prepare(selectFeedItemsForFeedSql); err != nil {
		return err
	} else {
		s.statements[selectFeedItemsForFeedStmt] = stmt
	}

	upsertFeedItemSql := `
		INSERT INTO feed_item (feed_id, guid, url, title, date)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(feed_id, guid)
		DO UPDATE SET
			url=excluded.url,
			title=excluded.title,
			date=excluded.date
	`
	if stmt, err := s.db.Prepare(upsertFeedItemSql); err != nil {
		return err
	} else {
		s.statements[upsertFeedItemStmt] = stmt
	}

	deleteItemsInFeedSql := "DELETE FROM feed_item WHERE feed_id = ?"
	if stmt, err := s.db.Prepare(deleteItemsInFeedSql); err != nil {
		return err
	} else {
		s.statements[deleteItemsInFeedStmt] = stmt
	}

	upsertFeedSyncStatusSql := `
		INSERT INTO feed_sync_status (feed_id, success, error)
		VALUES (?, ?, ?)
		ON CONFLICT(feed_id)
		DO UPDATE SET
			date = strftime('%s', 'now'),
			success = excluded.success,
			error = excluded.error
	`
	if stmt, err := s.db.Prepare(upsertFeedSyncStatusSql); err != nil {
		return err
	} else {
		s.statements[upsertFeedSyncStatusStmt] = stmt
	}

	selectFeedSyncStatusSql := `
		SELECT date, success, error
		FROM feed_sync_status
		WHERE feed_id = ?
	`
	if stmt, err := s.db.Prepare(selectFeedSyncStatusSql); err != nil {
		return err
	} else {
		s.statements[selectFeedSyncStatusStmt] = stmt
	}

	return nil
}

const maxRetries int = 10

func (s *FeedStore) wrapInTx(f func(tx *sql.Tx) error) error {
	numRetries := 0
	for {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}

		if err := f(tx); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Fatalf("Unable to rollback: %v", err)
			}

			if shouldRetryOnError(err) && numRetries < maxRetries {
				numRetries++
				continue
			}

			return err
		}

		return tx.Commit()
	}
}

func shouldRetryOnError(err error) bool {
	if sqliteErr, ok := err.(sqlite3.Error); ok {
		// This error code indicates a conflict between two threads.
		// This can be resolved by aborting and restarting the txn.
		if sqliteErr.Code == sqlite3.ErrBusy {
			return true
		}
	}
	return false
}

func (s *FeedStore) updateFeedRecord(tx *sql.Tx, id FeedId, feed feed.Feed) error {
	stmt := tx.Stmt(s.statements[updateFeedStmt])
	_, err := stmt.Exec(feed.Name, id)
	return err
}

func (s *FeedStore) deleteFeedRecord(tx *sql.Tx, id FeedId) error {
	stmt := tx.Stmt(s.statements[deleteFeedStmt])
	_, err := stmt.Exec(id)
	return err
}

func (s *FeedStore) upsertFeedItemRecord(tx *sql.Tx, feedId FeedId, item feed.FeedItem) error {
	stmt := tx.Stmt(s.statements[upsertFeedItemStmt])
	_, err := stmt.Exec(feedId, item.Guid, item.Url, item.Title, item.Date.Unix())
	return err
}

func (s *FeedStore) deleteItemsInFeed(tx *sql.Tx, id FeedId) error {
	stmt := tx.Stmt(s.statements[deleteItemsInFeedStmt])
	_, err := stmt.Exec(id)
	return err
}

func (s *FeedStore) setFeedSyncStatusSuccess(tx *sql.Tx, id FeedId) error {
	stmt := tx.Stmt(s.statements[upsertFeedSyncStatusStmt])
	_, err := stmt.Exec(id, true, nil)
	return err
}
