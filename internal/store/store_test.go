package store

import (
	"errors"
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

	err = store.SyncFeed(feedId, f)
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

func assertFeedSyncStatus(t *testing.T, store *FeedStore, feedId FeedId, expectFound, expectSuccess bool, expectedError error) {
	found, status, err := store.RetrieveFeedSyncStatus(feedId)
	if err != nil {
		t.Errorf("Could not retrieve feed sync status: %v", err)
	}

	if found != expectFound {
		t.Errorf("Sync status found was %v, expected %v", found, expectFound)
	}

	if status.Success != expectSuccess {
		t.Errorf("Sync status success was %v, expected %v", status.Success, expectSuccess)
	}

	if fmt.Sprintf("%v", status.Error) != fmt.Sprintf("%v", expectedError) {
		t.Errorf("Sync status error was %v, expected %v", status.Error, expectedError)
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
			Id:   feedId,
			Url:  "http://foo.com",
			Name: "Foo Feed",
		}
		assertFeed(t, store, feedId, expected)
	})
}

func TestSyncFeedNewItems(t *testing.T) {
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
		assertFeedSyncStatus(t, store, feedId, true, true, nil)
	})
}

func TestSyncFeedExistingItems(t *testing.T) {
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
		if err := store.SyncFeed(feedId, updatedFeed); err != nil {
			t.Fatalf("Could not update feed: %v", err)
		}

		expectedFeed := FeedRecord{
			Id:   feedId,
			Url:  "http://foo.com",
			Name: updatedFeed.Name,
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
		assertFeedSyncStatus(t, store, feedId, true, true, nil)
	})
}

func TestRetrieveSyncStatusBeforeSync(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		feedId, err := store.GetOrCreateFeedWithUrl("http://foo.com")
		if err != nil {
			t.Fatalf("Could not insert new feed: %v", err)
		}
		assertFeedSyncStatus(t, store, feedId, false, false, nil)
	})
}

func TestSetFeedSyncStatusError(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		feedId, err := store.GetOrCreateFeedWithUrl("http://foo.com")
		if err != nil {
			t.Fatalf("Could not insert new feed: %v", err)
		}

		syncErr := errors.New("KABOOM!")
		err = store.SetFeedSyncStatusError(feedId, syncErr)
		if err != nil {
			t.Fatalf("Could not set feed sync status: %v", err)
		}

		assertFeedSyncStatus(t, store, feedId, true, false, syncErr)
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
