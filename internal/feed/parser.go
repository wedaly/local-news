package feed

import (
	"errors"
	"github.com/mmcdole/gofeed"
	"io"
)

func ParseExternalFeed(r io.Reader) (Feed, error) {
	parser := gofeed.NewParser()
	rawFeed, err := parser.Parse(r)
	if err != nil {
		return Feed{}, err
	}

	if err := validateFeed(rawFeed); err != nil {
		return Feed{}, err
	}

	feed := Feed{
		Name:  rawFeed.Title,
		Items: make([]FeedItem, 0, len(rawFeed.Items)),
	}

	for _, rawItem := range rawFeed.Items {
		// GUID isn't required by the RSS specification,
		// so fallback to using the item URL.
		guid := rawItem.GUID
		if len(guid) == 0 {
			guid = rawItem.Link
		}

		item := FeedItem{
			Title: rawItem.Title,
			Date:  *rawItem.PublishedParsed,
			Url:   rawItem.Link,
			Guid:  guid,
		}
		feed.Items = append(feed.Items, item)
	}

	return feed, nil
}

func validateFeed(rawFeed *gofeed.Feed) error {
	if len(rawFeed.Title) == 0 {
		return errors.New("Missing feed name")
	}

	for _, rawItem := range rawFeed.Items {
		if len(rawItem.Title) == 0 {
			return errors.New("Missing item title")
		}

		if rawItem.PublishedParsed == nil {
			return errors.New("Missing published date")
		}

		if len(rawItem.Link) == 0 {
			return errors.New("Missing item link")
		}
	}

	return nil
}
