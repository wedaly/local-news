package store

import (
	"fmt"
	"github.com/wedaly/local-news/internal/feed"
	"reflect"
	"testing"
	"time"
)

func execWithStore(f func(*FeedStore)) {
	store := NewFeedStore("file::memory:")

	if err := store.Initialize(); err != nil {
		panic(err)
	}
	defer store.Close()

	f(store)
}

func createFeedAndItems(t *testing.T, store *FeedStore, numItems int) FeedId {
	f := feed.Feed{
		Name:  "Foo Feed",
		Items: make([]feed.FeedItem, 0, numItems),
	}

	for i := 0; i < numItems; i++ {
		item := feed.FeedItem{
			Title: fmt.Sprintf("Item %v", i),
			Date:  time.Unix(int64(i), 0),
			Url:   fmt.Sprintf("http://foo.com/%v", i),
			Guid:  fmt.Sprintf("guid.%v", i),
		}
		f.Items = append(f.Items, item)
	}

	feedId, err := store.GetOrCreateFeedWithUrl("http://foo.com")
	if err != nil {
		t.Fatalf("Could not insert new feed: %v", err)
	}

	err = store.UpdateFeed(feedId, f)
	if err != nil {
		t.Fatalf("Could not upsert new feed: %v", err)
	}

	return feedId
}

func assertFeed(t *testing.T, store *FeedStore, id FeedId, expected FeedRecord) {
	feed, err := store.RetrieveFeed(id)
	if err != nil {
		t.Errorf("Could not retrieve feed: %v", err)
	}

	if !reflect.DeepEqual(feed, expected) {
		t.Errorf("Incorrect feed, expected %v but got %v", expected, feed)
	}
}

func assertFeeds(t *testing.T, store *FeedStore, expected []FeedRecord) {
	feeds, err := store.RetrieveFeeds()
	if err != nil {
		t.Errorf("Could not retrieve feeds: %v", err)
	}

	if !reflect.DeepEqual(feeds, expected) {
		t.Errorf("Incorrect feeds, expected %v but got %v", expected, feeds)
	}
}

func assertFeedItems(t *testing.T, store *FeedStore, feedId FeedId, expected []FeedItemRecord) {
	items, err := store.RetrieveFeedItems(feedId)
	if err != nil {
		t.Errorf("Could not retrieve feed items: %v", err)
	}

	if !reflect.DeepEqual(items, expected) {
		t.Errorf("Incorrect feed items, expected %v but got %v", expected, items)
	}
}

func TestRetrieveFeedsEmpty(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		expected := []FeedRecord{}
		assertFeeds(t, store, expected)
	})
}

func TestInsertFeed(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		numItems := 1
		feedId := createFeedAndItems(t, store, numItems)
		expected := FeedRecord{
			Id:        feedId,
			Url:       "http://foo.com",
			Name:      "Foo Feed",
			NumUnread: 1,
		}
		assertFeed(t, store, feedId, expected)
	})
}

func TestUpdateFeedNewItems(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		numItems := 2
		feedId := createFeedAndItems(t, store, numItems)
		expected := []FeedItemRecord{
			FeedItemRecord{
				Id:    2,
				Title: "Item 1",
				Date:  time.Unix(1, 0),
				Url:   "http://foo.com/1",
				Guid:  "guid.1",
			},
			FeedItemRecord{
				Id:    1,
				Title: "Item 0",
				Date:  time.Unix(0, 0),
				Url:   "http://foo.com/0",
				Guid:  "guid.0",
			},
		}
		assertFeedItems(t, store, feedId, expected)
	})
}

func TestUpdateFeedExistingItems(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		numItems := 2
		feedId := createFeedAndItems(t, store, numItems)

		updatedFeed := feed.Feed{
			Name: "Updated feed",
			Items: []feed.FeedItem{
				feed.FeedItem{
					Title: "Updated 0",
					Date:  time.Unix(3, 0),
					Url:   "http://updated/0",
					Guid:  "guid.0",
				},
			},
		}
		if err := store.UpdateFeed(feedId, updatedFeed); err != nil {
			t.Fatalf("Could not update feed: %v", err)
		}

		expectedFeed := FeedRecord{
			Id:        feedId,
			Url:       "http://foo.com",
			Name:      updatedFeed.Name,
			NumUnread: 2,
		}
		assertFeed(t, store, feedId, expectedFeed)

		expectedItems := []FeedItemRecord{
			FeedItemRecord{
				Id:    1,
				Title: "Updated 0",
				Date:  time.Unix(3, 0),
				Url:   "http://updated/0",
				Guid:  "guid.0",
			},
			FeedItemRecord{
				Id:    2,
				Title: "Item 1",
				Date:  time.Unix(1, 0),
				Url:   "http://foo.com/1",
				Guid:  "guid.1",
			},
		}
		assertFeedItems(t, store, feedId, expectedItems)
	})
}

func TestNumUnread(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		const numFeedItems int = 10
		createFeedAndItems(t, store, numFeedItems)

		// Initially, all items should be marked unread
		if retrieved, err := store.RetrieveFeeds(); err != nil {
			t.Fatalf("Could not retrieve feeds: %v", err)
		} else if len(retrieved) != 1 {
			t.Fatalf("Expected one feed, but got %v", retrieved)
		} else if retrieved[0].NumUnread != uint(numFeedItems) {
			t.Fatalf("Expected all %v feed items unread, but got %v",
				numFeedItems, retrieved[0].NumUnread)
		}

		// Mark an item as read
		if err := store.MarkRead(FeedItemId(numFeedItems - 1)); err != nil {
			t.Fatalf("Could not mark feed item as read: %v", err)
		}

		// Check that the number is updated
		if retrieved, err := store.RetrieveFeeds(); err != nil {
			t.Fatalf("Could not retrieve feeds: %v", err)
		} else if len(retrieved) != 1 {
			t.Fatalf("Expected one feed, but got %v", retrieved)
		} else if retrieved[0].NumUnread != uint(numFeedItems-1) {
			t.Fatalf("Expected all %v feed items unread, but got %v",
				numFeedItems-1, retrieved[0].NumUnread)
		}
	})
}

func TestDeleteFeed(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		feedId := createFeedAndItems(t, store, 10)

		// Delete the feed and items
		if err := store.DeleteFeed(feedId); err != nil {
			t.Fatalf("Could not delete feed: %v", err)
		}

		// Check that no feeds exist
		if retrieved, err := store.RetrieveFeeds(); err != nil {
			t.Fatalf("Could not retrieve feeds: %v", err)
		} else if len(retrieved) > 0 {
			t.Fatalf("Expected all feeds to be deleted")
		}
	})
}
