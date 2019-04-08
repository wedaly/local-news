package feed

import (
	"reflect"
	"testing"
	"time"
)

func TestParseRssFeed(t *testing.T) {
	rssXml := `
		<?xml version="1.0" encoding="UTF-8"?>
		<rss>
			<channel>
				<title>Blog &#8211; My RSS Feed</title>
				<link>https://example.com</link>
			</channel>
			<item>
				<title>First post!</title>
				<link>https://example.com/first</link>
				<guid>abcd1234</guid>
				<pubDate>Sat, 06 Apr 2019 02:00:22 +0000</pubDate>
			</item>
		</rss>`

	rawFeed, err := rssFeedFromXml([]byte(rssXml))
	if err != nil {
		t.Fatalf("Could not parse RSS XML: %v", err)
	}

	if err := rawFeed.validate(); err != nil {
		t.Errorf("RSS feed failed validation check: %v", err)
	}

	feed, err := rawFeed.convertToFeed()
	if err != nil {
		t.Errorf("Could not convert RSS feed to standard format: %v", err)
	}

	expectedFeed := Feed{
		Url:  "https://example.com",
		Name: "Blog – My RSS Feed",
		Items: []FeedItem{
			FeedItem{
				Title: "First post!",
				Date:  time.Unix(1554516022, 0).UTC(),
				Url:   "https://example.com/first",
				Guid:  "abcd1234",
			},
		},
	}
	if !reflect.DeepEqual(feed, expectedFeed) {
		t.Errorf(
			"Incorrect values for feed, expected %v but got %v",
			expectedFeed, feed)
	}
}
