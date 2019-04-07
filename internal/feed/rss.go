package feed

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type rssFeedLoader struct {
	client *http.Client
}

// NewRssFeedLoader instantiates a feed loader for an RSS feed
func NewRssFeedLoader() FeedLoader {
	transport := http.Transport{
		IdleConnTimeout: 30 * time.Second,
	}
	client := http.Client{Transport: &transport}
	return &rssFeedLoader{&client}
}

// LoadFeedFromUrl retrieves an RSS feed and parses it into the standardized format.
// It does not attempt to handle malformed feeds.
// In addition, it *requires* the GUID field to be set on each feed item,
// even though the RSS 2.0 spec makes that field optional.
func (rss *rssFeedLoader) LoadFeedFromUrl(url string) (Feed, []FeedItem, error) {
	bytes, err := rss.loadFeedData(url)
	if err != nil {
		return Feed{}, nil, err
	}

	rawFeed, err := rssFeedFromXml(bytes)
	if err != nil {
		return Feed{}, nil, err
	}

	if err := rawFeed.validate(); err != nil {
		return Feed{}, nil, err
	}

	return rawFeed.convertToFeedAndFeedItems()
}

func (rss *rssFeedLoader) loadFeedData(url string) ([]byte, error) {
	resp, err := rss.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errMsg := fmt.Sprintf(
			"Received HTTP status %v from url %v",
			resp.StatusCode, url)
		return nil, errors.New(errMsg)
	}

	return ioutil.ReadAll(resp.Body)
}

type rssFeed struct {
	Channel rssChannel `xml:"channel"`
	Items   []rssItem  `xml:"item"`
}

type rssChannel struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
	Guid    string `xml:"guid"`
}

func rssFeedFromXml(bytes []byte) (rssFeed, error) {
	result := rssFeed{}
	if err := xml.Unmarshal(bytes, &result); err != nil {
		return rssFeed{}, err
	}
	return result, nil
}

func (rawFeed rssFeed) validate() error {
	rawChannel := rawFeed.Channel

	if len(rawChannel.Title) == 0 {
		return errors.New("Missing channel title")
	}

	if len(rawChannel.Link) == 0 {
		return errors.New("Missing channel link")
	}

	for _, rawItem := range rawFeed.Items {
		if len(rawItem.Title) == 0 {
			return errors.New("Missing item title")
		}

		if len(rawItem.Link) == 0 {
			return errors.New("Missing item link")
		}

		if len(rawItem.PubDate) == 0 {
			return errors.New("Missing item pubDate")
		}

		if len(rawItem.Guid) == 0 {
			return errors.New("Missing item guid")
		}
	}

	return nil
}

func (rawFeed rssFeed) convertToFeedAndFeedItems() (Feed, []FeedItem, error) {
	rawChannel := rawFeed.Channel
	feed := Feed{
		Url:  rawChannel.Link,
		Name: rawChannel.Title,
	}

	feedItems := make([]FeedItem, 0, len(rawFeed.Items))
	for _, rawItem := range rawFeed.Items {
		date, err := time.ParseInLocation(time.RFC1123Z, rawItem.PubDate, time.UTC)
		if err != nil {
			return Feed{}, nil, err
		}

		item := FeedItem{
			Title: rawItem.Title,
			Date:  date,
			Url:   rawItem.Link,
			Guid:  rawItem.Guid,
		}
		feedItems = append(feedItems, item)
	}

	return feed, feedItems, nil
}
