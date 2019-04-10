package feed

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadRssFromUrl(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
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
				</channel>
			</rss>`
		fmt.Fprintln(w, rssXml)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	loader := NewFeedLoader()
	feed, err := loader.LoadFeedFromUrl(server.URL)
	if err != nil {
		t.Fatalf("Error loading feed from test server: %v", err)
	}

	if len(feed.Items) != 1 {
		t.Errorf("Incorrect data from loaded feed (num feed items)")
	}
}
