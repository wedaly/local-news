package store

import (
	"github.com/wedaly/local-news/internal/feed"
)

// FeedId is a unique identifier for each feed stored in the database
type FeedId int64

// FeedItemId is a unique identifier for each feed item stored in the database
type FeedItemId int64

// FeedRecord is the data associated with a feed in the database
type FeedRecord struct {
	Id FeedId

	// Feed contains the data retrieved from the feed source (e.g. XML)
	Feed feed.Feed

	// Count of number of items that are unread in this feed
	NumUnread uint
}

// FeedItemRecord is the data associated with a feed item in the database
type FeedItemRecord struct {
	Id FeedItemId

	// FeedItem contains the data retrieved from the feed source (e.g. XML)
	FeedItem feed.FeedItem

	// Read indicates whether the user has read this item
	// The default value is false
	Read bool
}

// FeedStore provides thread-safe CRUD operations for feeds and feed items
type FeedStore interface {
	// Initialize opens the database and (if necessary) creates tables,
	// indices, and prepared statements.
	Initialize() error

	// Close gracefully shuts down the database
	// This must be called before the application exits
	Close()

	// UpsertFeed transactionally inserts-or-updates the specified feed
	// Feeds are uniquely identified by their URLs
	// Once inserted, the feed is assigned a unique primary key
	UpsertFeed(feed feed.Feed) (FeedId, error)

	// DeleteFeed transactionally deletes the specified feed and all its items
	DeleteFeed(feedId FeedId) error

	// RetrieveFeeds retrieves a record for every feed in the database
	RetrieveFeeds() ([]FeedRecord, error)

	// UpsertFeedItem transactionally inserts-or-updates the specified feed item.
	// Feed items are uniquely identified by the "guid" attribute
	// Once inserted, the item is assigned a unique primary key
	// By default, the feed item is marked as unread.
	UpsertFeedItem(feedId FeedId, item feed.FeedItem) (FeedItemId, error)

	// RetrieveFeedItems retrieves a record for every feed item for a given feed
	RetrieveFeedItems(feedId FeedId) ([]FeedItemRecord, error)

	// MarkRead marks a particular feed item as read
	// (meaning the user has seen it)
	MarkRead(feedItemId FeedItemId) error
}
