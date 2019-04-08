package store

import (
	"fmt"
	"github.com/wedaly/local-news/internal/feed"
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
		Url:   "http://foo.com",
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

	feedId, err := store.UpsertFeed(f)
	if err != nil {
		t.Fatalf("Could not upsert new feed: %v", err)
	}

	return feedId
}

func TestRetrieveFeedsEmpty(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		if feeds, err := store.RetrieveFeeds(); err != nil {
			t.Errorf("Could not retrieve feeds: %v", err)
		} else if len(feeds) != 0 {
			t.Errorf("Expected feeds list to be empty, but got %v", feeds)
		}
	})
}

func TestUpsertNewFeed(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		feed := feed.Feed{Url: "http://foo.com", Name: "Foo Feed"}
		if _, err := store.UpsertFeed(feed); err != nil {
			t.Fatalf("Could not upsert new feed: %v", err)
		}

		if retrieved, err := store.RetrieveFeeds(); err != nil {
			t.Fatalf("Could not retrieve feeds: %v", err)
		} else if len(retrieved) != 1 {
			t.Fatalf("Expected one feed, but got %v", retrieved)
		} else {
			record := retrieved[0]
			expected := FeedRecord{
				Id:        1,
				Url:       feed.Url,
				Name:      feed.Name,
				NumUnread: 0,
			}
			if record != expected {
				t.Errorf("Expected %v but got %v", expected, record)
			}
		}
	})
}

func TestUpsertExistingFeed(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		url := "http://foo.com"
		initialFeed := feed.Feed{Url: url, Name: "Foo Feed"}
		insertedId, err := store.UpsertFeed(initialFeed)
		if err != nil {
			t.Fatalf("Could not upsert new feed: %v", err)
		}

		updatedFeed := feed.Feed{Url: url, Name: "Bar Feed"}
		updatedId, err := store.UpsertFeed(updatedFeed)
		if err != nil {
			t.Fatalf("Could not upsert new feed: %v", err)
		}

		if insertedId != updatedId {
			t.Errorf("Inserted new record instead of updating existing record")
		}

		if retrieved, err := store.RetrieveFeeds(); err != nil {
			t.Fatalf("Could not retrieve feeds: %v", err)
		} else if len(retrieved) != 1 {
			t.Fatalf("Expected one feed, but got %v", retrieved)
		} else {
			record := retrieved[0]
			expected := FeedRecord{
				Id:        1,
				Url:       updatedFeed.Url,
				Name:      updatedFeed.Name,
				NumUnread: 0,
			}
			if record != expected {
				t.Errorf("Expected %v but got %v", expected, record)
			}
		}
	})
}

func TestUpsertMultipleFeeds(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		const numFeeds int = 10
		for i := 0; i < numFeeds; i++ {
			url := fmt.Sprintf("http://feed/%v", i)
			name := fmt.Sprintf("feed.%v", i)
			feed := feed.Feed{Url: url, Name: name}
			if _, err := store.UpsertFeed(feed); err != nil {
				t.Fatalf("Could not upsert new feed: %v", err)
			}
		}

		if retrieved, err := store.RetrieveFeeds(); err != nil {
			t.Fatalf("Could not retrieve feeds: %v", err)
		} else if len(retrieved) != numFeeds {
			t.Fatalf("Expected %v feeds, but got %v", numFeeds, retrieved)
		}
	})
}

func TestUpsertNewFeedItem(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		item := feed.FeedItem{
			Title: "Foo feed item",
			Date:  time.Unix(10, 0),
			Url:   "http://foo.com/item",
			Guid:  "abcd1234",
		}
		f := feed.Feed{
			Url:   "http://foo.com",
			Name:  "Foo Feed",
			Items: []feed.FeedItem{item},
		}
		feedId, err := store.UpsertFeed(f)
		if err != nil {
			t.Fatalf("Could not upsert new feed: %v", err)
		}

		if retrieved, err := store.RetrieveFeedItems(feedId); err != nil {
			t.Fatalf("Could not retrieve feeds: %v", err)
		} else if len(retrieved) != 1 {
			t.Fatalf("Expected one feed item, but got %v", retrieved)
		} else {
			record := retrieved[0]
			expected := FeedItemRecord{
				Id:    1,
				Title: item.Title,
				Date:  item.Date,
				Url:   item.Url,
				Guid:  item.Guid,
				Read:  false,
			}
			if record != expected {
				t.Fatalf("Expected %v but got %v", expected, record)
			}
		}
	})
}

func TestUpsertExistingFeedItem(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		initialItem := feed.FeedItem{
			Title: "Initial feed item",
			Date:  time.Unix(10, 0),
			Url:   "http://foo.com/initial",
			Guid:  "abcd1234",
		}

		updatedItem := feed.FeedItem{
			Title: "Updated feed item",
			Date:  time.Unix(20, 0),
			Url:   "http://foo.com/updated",
			Guid:  "abcd1234",
		}

		// Upsert feed with initial version of the item
		f := feed.Feed{
			Url:   "http://foo.com",
			Name:  "Foo Feed",
			Items: []feed.FeedItem{initialItem},
		}
		feedId, err := store.UpsertFeed(f)
		if err != nil {
			t.Fatalf("Could not upsert new feed: %v", err)
		}

		// Upsert feed with the updated version of the item
		f.Items = []feed.FeedItem{updatedItem}
		_, err = store.UpsertFeed(f)
		if err != nil {
			t.Fatalf("Could not upsert updated feed: %v", err)
		}

		// Check that the item was updated
		if retrieved, err := store.RetrieveFeedItems(feedId); err != nil {
			t.Fatalf("Could not retrieve feeds: %v", err)
		} else if len(retrieved) != 1 {
			t.Fatalf("Expected one feed item, but got %v", retrieved)
		} else {
			record := retrieved[0]
			expected := FeedItemRecord{
				Id:    1,
				Title: updatedItem.Title,
				Date:  updatedItem.Date,
				Url:   updatedItem.Url,
				Guid:  updatedItem.Guid,
				Read:  false,
			}
			if record != expected {
				t.Fatalf("Expected %v but got %v", expected, record)
			}
		}
	})
}

func TestUpsertMultipleFeedItems(t *testing.T) {
	execWithStore(func(store *FeedStore) {
		const numFeedItems int = 10
		feedId := createFeedAndItems(t, store, numFeedItems)

		if retrieved, err := store.RetrieveFeedItems(feedId); err != nil {
			t.Fatalf("Could not retrieve feed items: %v", err)
		} else if len(retrieved) != numFeedItems {
			t.Fatalf("Expected %v feed items, but got %v", numFeedItems, retrieved)
		}
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
