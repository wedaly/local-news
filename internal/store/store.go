package store

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wedaly/local-news/internal/feed"
	"log"
	"time"
)

const numStatements int = 11

const (
	selectEveryFeedStmt = iota
	selectFeedIdByUrlStmt
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

// UpsertFeed transactionally inserts-or-updates the specified feed
// Feeds are uniquely identified by their URLs
// Once inserted, the feed is assigned a unique primary key
func (s *FeedStore) UpsertFeed(feed feed.Feed) (FeedId, error) {
	id, err := s.wrapInTxReturnId(func(tx *sql.Tx) (int64, error) {
		var existingId int64
		stmt := tx.Stmt(s.statements[selectFeedIdByUrlStmt])
		err := stmt.QueryRow(feed.Url).Scan(&existingId)
		if err == nil {
			return existingId, s.updateFeed(tx, existingId, feed)
		} else if err == sql.ErrNoRows {
			return s.insertFeed(tx, feed)
		} else {
			return 0, err
		}
	})

	if err != nil {
		return 0, err
	} else {
		return FeedId(id), nil
	}
}

// DeleteFeed transactionally deletes the specified feed and all its items
func (s *FeedStore) DeleteFeed(feedId FeedId) error {
	return s.wrapInTx(func(tx *sql.Tx) error {
		if err := s.deleteItemsInFeed(tx, feedId); err != nil {
			return err
		}

		if err := s.deleteFeed(tx, feedId); err != nil {
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

// UpsertFeedItem transactionally inserts-or-updates the specified feed item.
// Feed items are uniquely identified by the "guid" attribute
// Once inserted, the item is assigned a unique primary key
// By default, the feed item is marked as unread.
func (s *FeedStore) UpsertFeedItem(feedId FeedId, item feed.FeedItem) (FeedItemId, error) {
	id, err := s.wrapInTxReturnId(func(tx *sql.Tx) (int64, error) {
		var existingId int64
		stmt := tx.Stmt(s.statements[selectFeedItemIdByGuidStmt])
		err := stmt.QueryRow(feedId, item.Guid).Scan(&existingId)
		if err == nil {
			return existingId, s.updateFeedItem(tx, existingId, item)
		} else if err == sql.ErrNoRows {
			return s.insertFeedItem(tx, feedId, item)
		} else {
			return 0, err
		}
	})

	if err != nil {
		return 0, err
	} else {
		return FeedItemId(id), nil
	}
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

	selectFeedIdByUrlSql := "SELECT id FROM feed WHERE url = ?"
	if stmt, err := s.db.Prepare(selectFeedIdByUrlSql); err != nil {
		return err
	} else {
		s.statements[selectFeedIdByUrlStmt] = stmt
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

func (s *FeedStore) wrapInTx(f func(tx *sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := f(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Fatalf("Unable to rollback: %v", err)
		}
		return err
	}

	return tx.Commit()
}

func (s *FeedStore) wrapInTxReturnId(f func(tx *sql.Tx) (int64, error)) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}

	id, err := f(tx)
	if err != nil {
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

func (s *FeedStore) insertFeed(tx *sql.Tx, feed feed.Feed) (int64, error) {
	stmt := tx.Stmt(s.statements[insertFeedStmt])
	result, err := stmt.Exec(feed.Url, feed.Name)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (s *FeedStore) updateFeed(tx *sql.Tx, id int64, feed feed.Feed) error {
	stmt := tx.Stmt(s.statements[updateFeedStmt])
	_, err := stmt.Exec(feed.Name, id)
	return err
}

func (s *FeedStore) deleteFeed(tx *sql.Tx, id FeedId) error {
	stmt := tx.Stmt(s.statements[deleteFeedStmt])
	_, err := stmt.Exec(id)
	return err
}

func (s *FeedStore) insertFeedItem(tx *sql.Tx, feedId FeedId, item feed.FeedItem) (int64, error) {
	stmt := tx.Stmt(s.statements[insertFeedItemStmt])
	result, err := stmt.Exec(
		feedId,
		item.Guid,
		item.Url,
		item.Title,
		item.Date.Unix())
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (s *FeedStore) updateFeedItem(tx *sql.Tx, id int64, item feed.FeedItem) error {
	stmt := tx.Stmt(s.statements[updateFeedItemStmt])
	_, err := stmt.Exec(item.Url, item.Title, item.Date.Unix(), id)
	return err
}

func (s *FeedStore) deleteItemsInFeed(tx *sql.Tx, id FeedId) error {
	stmt := tx.Stmt(s.statements[deleteItemsInFeedStmt])
	_, err := stmt.Exec(id)
	return err
}