package feed

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// FeedLoader retrieves a feed from a URL and parses it into
// a standardized format.
type FeedLoader struct {
	client *http.Client
}

func NewFeedLoader() *FeedLoader {
	transport := http.Transport{
		IdleConnTimeout: 30 * time.Second,
	}
	client := http.Client{Transport: &transport}
	return &FeedLoader{&client}
}

// LoadFeedFromUrl retrieves an RSS feed and parses it into the standardized format.
// It does not attempt to handle malformed feeds.
// In addition, it *requires* the GUID field to be set on each feed item,
// even though the RSS 2.0 spec makes that field optional.
func (f *FeedLoader) LoadFeedFromUrl(url string) (Feed, error) {
	bytes, err := f.loadFeedData(url)
	if err != nil {
		return Feed{}, err
	}

	rawFeed, err := rssFeedFromXml(bytes)
	if err != nil {
		return Feed{}, err
	}

	if err := rawFeed.validate(); err != nil {
		return Feed{}, err
	}

	return rawFeed.convertToFeed()
}

func (f *FeedLoader) loadFeedData(url string) ([]byte, error) {
	resp, err := f.client.Get(url)
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
