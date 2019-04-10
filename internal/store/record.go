package store

import "time"

// FeedId is a unique identifier for each feed stored in the database
type FeedId int64

// FeedItemId is a unique identifier for each feed item stored in the database
type FeedItemId int64

// FeedRecord is the data associated with a feed in the database
type FeedRecord struct {
	Id FeedId

	// Url for the feed, must be unique
	Url string

	// Name of the feed
	Name string

	// Count of number of items that are unread in this feed
	NumUnread uint
}

// FeedItemRecord is the data associated with a feed item in the database
type FeedItemRecord struct {
	Id FeedItemId

	// Title of the item (retrieved)
	Title string

	// Date the item was published
	Date time.Time

	// Url of the item
	Url string

	// Globally unique identifier for the item, retrieved
	// from the feed source.
	Guid string

	// Read indicates whether the user has read this item
	// The default value is false
	Read bool
}

// FeedSyncStatus represents the most recent attempt to synchronize
// the feed with its external source.
type FeedSyncStatus struct {
	// The datetime of the last sync attempt
	Date time.Time

	// Whether the last sync attempt was successful
	Success bool

	// The error, if any, that occurred during the sync
	// This isn't displayed in the UI, but we store it anyway
	// for debugging purposes.
	Error error
}
