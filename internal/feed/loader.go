package feed

import (
	"errors"
	"fmt"
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

// LoadFeedFromUrl retrieves a feed and parses it into the standardized format.
func (f *FeedLoader) LoadFeedFromUrl(url string) (Feed, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return Feed{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errMsg := fmt.Sprintf(
			"Received HTTP status %v from url %v",
			resp.StatusCode, url)
		return Feed{}, errors.New(errMsg)
	}

	return ParseExternalFeed(resp.Body)
}
