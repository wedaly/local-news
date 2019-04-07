package feed

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

const rssXml string = `
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

func TestParseRssFeed(t *testing.T) {
	rawFeed, err := rssFeedFromXml([]byte(rssXml))
	if err != nil {
		t.Fatalf("Could not parse RSS XML: %v", err)
	}

	if err := rawFeed.validate(); err != nil {
		t.Errorf("RSS feed failed validation check: %v", err)
	}

	feed, feedItems, err := rawFeed.convertToFeedAndFeedItems()
	if err != nil {
		t.Errorf("Could not convert RSS feed to standard format: %v", err)
	}

	expectedFeed := Feed{
		Url:  "https://example.com",
		Name: "Blog – My RSS Feed",
	}
	if feed != expectedFeed {
		t.Errorf(
			"Incorrect values for feed, expected %v but got %v",
			expectedFeed, feed)
	}

	expectedItems := []FeedItem{
		FeedItem{
			Title: "First post!",
			Date:  time.Unix(1554516022, 0).UTC(),
			Url:   "https://example.com/first",
			Guid:  "abcd1234",
		},
	}
	if !reflect.DeepEqual(feedItems, expectedItems) {
		t.Errorf(
			"Incorrect values for feed items, expected %v but got %v",
			expectedItems, feedItems)
	}
}

func TestLoadRssFromUrl(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, rssXml)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	loader := NewRssFeedLoader()
	feed, feedItems, err := loader.LoadFeedFromUrl(server.URL)
	if err != nil {
		t.Fatalf("Error loading feed from test server: %v", err)
	}

	if feed.Url != "https://example.com" {
		t.Errorf("Incorrect data from loaded feed (url)")
	}

	if len(feedItems) != 1 {
		t.Errorf("Incorrect data from loaded feed (num feed items)")
	}
}
