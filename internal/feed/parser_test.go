package feed

import (
	"bytes"
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
				<item>
					<title>First post!</title>
					<link>https://example.com/first</link>
					<guid>abcd1234</guid>
					<pubDate>Sat, 06 Apr 2019 02:00:22 +0000</pubDate>
				</item>
				<item>
					<title>Second post!</title>
					<link>https://example.com/second</link>
					<guid>xyz567</guid>
					<pubDate>Sat, 09 Apr 2020 03:00:34 +0000</pubDate>
				</item>
			</channel>
		</rss>`

	r := bytes.NewReader([]byte(rssXml))
	feed, err := ParseExternalFeed(r)
	if err != nil {
		t.Fatalf("Could not parse feed xml: %v", err)
	}

	expectedFeed := Feed{
		Name: "Blog – My RSS Feed",
		Items: []FeedItem{
			FeedItem{
				Title: "First post!",
				Date:  time.Unix(1554516022, 0).UTC(),
				Url:   "https://example.com/first",
				Guid:  "abcd1234",
			},
			FeedItem{
				Title: "Second post!",
				Date:  time.Unix(1586401234, 0).UTC(),
				Url:   "https://example.com/second",
				Guid:  "xyz567",
			},
		},
	}
	if !reflect.DeepEqual(feed, expectedFeed) {
		t.Errorf(
			"Incorrect values for feed, expected %v but got %v",
			expectedFeed, feed)
	}
}

func TestParseFeedIgnoreAtomNamespace(t *testing.T) {
	rssXml := `
		<?xml version="1.0" encoding="UTF-8"?>
		<rss>
			<channel>
				<title>Atom</title>
				<link>https://atom.com</link>
				<item>
					<title>Atom post!</title>
					<link>https://example.com/first</link>
					<atom:link rel="standout" href="https://atom.com/first"/>
					<guid>abcd1234</guid>
					<pubDate>Sat, 06 Apr 2019 02:00:22 +0000</pubDate>
				</item>
			</channel>
		</rss>`

	r := bytes.NewReader([]byte(rssXml))
	feed, err := ParseExternalFeed(r)
	if err != nil {
		t.Fatalf("Could not parse feed xml: %v", err)
	}

	if feed.Items[0].Url != "https://example.com/first" {
		t.Errorf("Incorrect URL for item link")
	}
}
