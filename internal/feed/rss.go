package feed

import (
	"encoding/xml"
	"errors"
	"io"
	"time"
)

type rssFeed struct {
	Title string    `xml:"channel>title"`
	Items []rssItem `xml:"channel>item"`
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
	Guid    string `xml:"guid"`
}

func rssFeedFromXml(r io.Reader) (rssFeed, error) {
	result := rssFeed{}
	decoder := xml.NewDecoder(r)
	decoder.Strict = false
	err := decoder.Decode(&result)
	return result, err
}

func (rawFeed rssFeed) validate() error {
	if len(rawFeed.Title) == 0 {
		return errors.New("Missing channel title")
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

func (rawFeed rssFeed) convertToFeed() (Feed, error) {
	feed := Feed{
		Name:  rawFeed.Title,
		Items: make([]FeedItem, 0, len(rawFeed.Items)),
	}

	for _, rawItem := range rawFeed.Items {
		date, err := time.ParseInLocation(time.RFC1123Z, rawItem.PubDate, time.UTC)
		if err != nil {
			return Feed{}, err
		}

		item := FeedItem{
			Title: rawItem.Title,
			Date:  date,
			Url:   rawItem.Link,
			Guid:  rawItem.Guid,
		}
		feed.Items = append(feed.Items, item)
	}

	return feed, nil
}
