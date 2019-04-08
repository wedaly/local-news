package feed

import (
	"encoding/xml"
	"errors"
	"time"
)

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

func (rawFeed rssFeed) convertToFeed() (Feed, error) {
	rawChannel := rawFeed.Channel
	feed := Feed{
		Url:   rawChannel.Link,
		Name:  rawChannel.Title,
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
