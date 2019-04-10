package feed

import "time"

// Feed represents a content syndication feed (e.g. RSS or Atom)
type Feed struct {
	Name  string
	Items []FeedItem
}

// FeedItem represents an item in a feed (e.g. a blog post)
type FeedItem struct {
	Title string
	Date  time.Time
	Url   string
	Guid  string
}
