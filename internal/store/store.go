package store

import (
	"database/sql"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/wedaly/local-news/internal/feed"
	"log"
	"time"
)

const numStatements int = 11

const (
	selectEveryFeedStmt = iota
	selectFeedStmt      = iota
	insertFeedStmt
	updateFeedStmt
	deleteFeedStmt
	selectFeedItemsForFeedStmt
	selectFeedItemIdByGuidStmt
	insertFeedItemStmt
	updateFeedItemStmt
	deleteItemsInFeedStmt
	markFeedItemReadStmt
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
func (s *FeedStore) InsertFeedWithUrl(url string) (FeedId, error) {
	placeholder := feed.Feed{
		Url:  url,
		Name: url,
	}
	id, err := s.insertFeedRecord(placeholder)
	return FeedId(id), err
}

// UpdateFeed atomically updates a feed record and its items.
// The feed name (but not URL or ID) is overwritten with the new name.
// The feed items are upserted, using the record GUID as the record's identity.
// Existing items NOT included in the new feed are retained (not deleted)
func (s *FeedStore) UpdateFeed(id FeedId, feed feed.Feed) error {
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
		var numUnread uint

		if err := rows.Scan(&id, &url, &name, &numUnread); err != nil {
			return nil, err
		}

		records = append(records, FeedRecord{
			Id:        FeedId(id),
			Url:       url,
			Name:      name,
			NumUnread: numUnread,
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
	var numUnread uint

	stmt := s.statements[selectFeedStmt]
	err := stmt.QueryRow(id).Scan(&url, &name, &numUnread)
	if err != nil {
		return FeedRecord{}, err
	}

	record := FeedRecord{
		Id:        id,
		Url:       url,
		Name:      name,
		NumUnread: numUnread,
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
		var read bool

		if err := rows.Scan(&id, &guid, &url, &title, &date, &read); err != nil {
			return nil, err
		}

		records = append(records, FeedItemRecord{
			Id:    FeedItemId(id),
			Title: title,
			Date:  time.Unix(date, 0),
			Url:   url,
			Guid:  guid,
			Read:  read,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// MarkRead marks a particular feed item as read
// (meaning the user has seen it)
func (s *FeedStore) MarkRead(id FeedItemId) error {
	stmt := s.statements[markFeedItemReadStmt]
	_, err := stmt.Exec(id)
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
		guid VARCHAR UNIQUE NOT NULL,
		url VARCHAR NOT NULL,
		title VARCHAR NOT NULL,
		date INTEGER NOT NULL,
		read INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (feed_id) REFERENCES feed(id)
	);

	CREATE INDEX IF NOT EXISTS feed_item_date_idx ON feed_item(date);
	`
	_, err := s.db.Exec(sql)
	return err
}

func (s *FeedStore) prepareStatements() error {
	s.statements = make([]*sql.Stmt, numStatements)

	selectEveryFeedSql := `
		SELECT id, url, name,
			(SELECT COUNT(id) FROM feed_item
				WHERE feed_id = feed.id AND read = 0) AS num_unread
		FROM feed`
	if stmt, err := s.db.Prepare(selectEveryFeedSql); err != nil {
		return err
	} else {
		s.statements[selectEveryFeedStmt] = stmt
	}

	selectFeedSql := `
		SELECT url, name,
			(SELECT COUNT(id) FROM feed_item
				WHERE feed_id = feed.id AND read = 0) AS num_unread
		FROM feed
		WHERE id = ?`
	if stmt, err := s.db.Prepare(selectFeedSql); err != nil {
		return err
	} else {
		s.statements[selectFeedStmt] = stmt
	}

	insertFeedSql := "INSERT INTO feed (url, name) VALUES (?, ?)"
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
		SELECT id, guid, url, title, date, read
		FROM feed_item
		WHERE feed_id = ?
		ORDER BY date DESC`
	if stmt, err := s.db.Prepare(selectFeedItemsForFeedSql); err != nil {
		return err
	} else {
		s.statements[selectFeedItemsForFeedStmt] = stmt
	}

	selectFeedItemIdByGuidSql := `
		SELECT id FROM feed_item
		WHERE feed_id = ? AND guid = ?`
	if stmt, err := s.db.Prepare(selectFeedItemIdByGuidSql); err != nil {
		return err
	} else {
		s.statements[selectFeedItemIdByGuidStmt] = stmt
	}

	insertFeedItemSql := `
		INSERT INTO feed_item (feed_id, guid, url, title, date)
		VALUES (?, ?, ?, ?, ?)`
	if stmt, err := s.db.Prepare(insertFeedItemSql); err != nil {
		return err
	} else {
		s.statements[insertFeedItemStmt] = stmt
	}

	updateFeedItemSql := `
		UPDATE feed_item
		SET url = ?, title = ?, date = ?
		WHERE id = ?`
	if stmt, err := s.db.Prepare(updateFeedItemSql); err != nil {
		return err
	} else {
		s.statements[updateFeedItemStmt] = stmt
	}

	deleteItemsInFeedSql := "DELETE FROM feed_item WHERE feed_id = ?"
	if stmt, err := s.db.Prepare(deleteItemsInFeedSql); err != nil {
		return err
	} else {
		s.statements[deleteItemsInFeedStmt] = stmt
	}

	markFeedItemReadSql := "UPDATE feed_item SET read = 1 WHERE id = ?"
	if stmt, err := s.db.Prepare(markFeedItemReadSql); err != nil {
		return err
	} else {
		s.statements[markFeedItemReadStmt] = stmt
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

func (s *FeedStore) insertFeedRecord(feed feed.Feed) (int64, error) {
	stmt := s.statements[insertFeedStmt]
	result, err := stmt.Exec(feed.Url, feed.Name)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
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
	var existingId int64
	stmt := tx.Stmt(s.statements[selectFeedItemIdByGuidStmt])
	err := stmt.QueryRow(feedId, item.Guid).Scan(&existingId)
	if err == nil {
		return s.updateFeedItemRecord(tx, existingId, item)
	} else if err == sql.ErrNoRows {
		return s.insertFeedItemRecord(tx, feedId, item)
	} else {
		return err
	}
}

func (s *FeedStore) insertFeedItemRecord(tx *sql.Tx, feedId FeedId, item feed.FeedItem) error {
	stmt := tx.Stmt(s.statements[insertFeedItemStmt])
	_, err := stmt.Exec(
		feedId,
		item.Guid,
		item.Url,
		item.Title,
		item.Date.Unix())
	return err
}

func (s *FeedStore) updateFeedItemRecord(tx *sql.Tx, id int64, item feed.FeedItem) error {
	stmt := tx.Stmt(s.statements[updateFeedItemStmt])
	_, err := stmt.Exec(item.Url, item.Title, item.Date.Unix(), id)
	return err
}

func (s *FeedStore) deleteItemsInFeed(tx *sql.Tx, id FeedId) error {
	stmt := tx.Stmt(s.statements[deleteItemsInFeedStmt])
	_, err := stmt.Exec(id)
	return err
}
