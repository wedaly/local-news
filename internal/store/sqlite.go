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

type sqliteFeedStore struct {
	dbPath     string
	db         *sql.DB
	statements []*sql.Stmt
}

// NewSqliteFeedStore creates a new store in SQLite
// See the SQLite driver documentation for the `dbPath` format
func NewSQLiteFeedStore(dbPath string) FeedStore {
	return &sqliteFeedStore{dbPath: dbPath}
}

func (s *sqliteFeedStore) Initialize() error {
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

func (s *sqliteFeedStore) Close() {
	for _, stmt := range s.statements {
		stmt.Close()
	}

	s.db.Close()
}

func (s *sqliteFeedStore) UpsertFeed(feed feed.Feed) (FeedId, error) {
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

func (s *sqliteFeedStore) DeleteFeed(feedId FeedId) error {
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

func (s *sqliteFeedStore) RetrieveFeeds() ([]FeedRecord, error) {
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

func (s *sqliteFeedStore) UpsertFeedItem(feedId FeedId, item feed.FeedItem) (FeedItemId, error) {
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

func (s *sqliteFeedStore) RetrieveFeedItems(feedId FeedId) ([]FeedItemRecord, error) {
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

func (s *sqliteFeedStore) MarkRead(id FeedItemId) error {
	stmt := s.statements[markFeedItemReadStmt]
	_, err := stmt.Exec(id)
	return err
}

func (s *sqliteFeedStore) installSchema() error {
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

func (s *sqliteFeedStore) prepareStatements() error {
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

func (s *sqliteFeedStore) wrapInTx(f func(tx *sql.Tx) error) error {
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

func (s *sqliteFeedStore) wrapInTxReturnId(f func(tx *sql.Tx) (int64, error)) (int64, error) {
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

func (s *sqliteFeedStore) insertFeed(tx *sql.Tx, feed feed.Feed) (int64, error) {
	stmt := tx.Stmt(s.statements[insertFeedStmt])
	result, err := stmt.Exec(feed.Url, feed.Name)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (s *sqliteFeedStore) updateFeed(tx *sql.Tx, id int64, feed feed.Feed) error {
	stmt := tx.Stmt(s.statements[updateFeedStmt])
	_, err := stmt.Exec(feed.Name, id)
	return err
}

func (s *sqliteFeedStore) deleteFeed(tx *sql.Tx, id FeedId) error {
	stmt := tx.Stmt(s.statements[deleteFeedStmt])
	_, err := stmt.Exec(id)
	return err
}

func (s *sqliteFeedStore) insertFeedItem(tx *sql.Tx, feedId FeedId, item feed.FeedItem) (int64, error) {
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

func (s *sqliteFeedStore) updateFeedItem(tx *sql.Tx, id int64, item feed.FeedItem) error {
	stmt := tx.Stmt(s.statements[updateFeedItemStmt])
	_, err := stmt.Exec(item.Url, item.Title, item.Date.Unix(), id)
	return err
}

func (s *sqliteFeedStore) deleteItemsInFeed(tx *sql.Tx, id FeedId) error {
	stmt := tx.Stmt(s.statements[deleteItemsInFeedStmt])
	_, err := stmt.Exec(id)
	return err
}
