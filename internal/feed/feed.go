package feed

import "time"

// Feed represents a content syndication feed (e.g. RSS or Atom)
type Feed struct {
	Url  string
	Name string
}

// FeedItem represents an item in a feed (e.g. a blog post)
type FeedItem struct {
	Title string
	Date  time.Time
	Url   string
	Guid  string
}

// FeedLoader retrieves a feed from a URL and parses it into
// a standardized format.
type FeedLoader interface {
	LoadFeedFromUrl(url string) (Feed, []FeedItem, error)
}
