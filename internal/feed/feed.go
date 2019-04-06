package feed

import "time"

type Feed struct {
	Url  string
	Name string
}

type FeedItem struct {
	Title string
	Date  time.Time
	Url   string
	Guid  string
}
