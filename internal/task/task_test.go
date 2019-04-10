package task

import (
	"fmt"
	"github.com/wedaly/local-news/internal/store"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
)

type StubSubscriber struct {
	resultChan chan TaskResult
}

func (s *StubSubscriber) HandleTaskCompleted(r TaskResult) {
	s.resultChan <- r
}

func TestScheduleLoadFeedTasks(t *testing.T) {
	// Create an on-disk feed store
	// Can't use an in-memory DB because of concurrency issues with SQLite
	dbPath := path.Join(os.TempDir(), "test-task.db")
	defer func() { os.Remove(dbPath) }()
	store := store.NewFeedStore(dbPath)
	if err := store.Initialize(); err != nil {
		t.Fatalf("Could not initialize store: %v", err)
	}
	defer store.Close()

	// Set up testing HTTP server
	handler := func(w http.ResponseWriter, r *http.Request) {
		rssXml := `
			<?xml version="1.0" encoding="UTF-8"?>
			<rss>
				<channel>
					<title>Blog &#8211; My RSS Feed</title>
					<link>https://example.com</link>
				</channel>
				<item>
					<title>First post!</title>
					<link>https://example.com/first</link>
					<guid>abcd1234</guid>
					<pubDate>Sat, 06 Apr 2019 02:00:22 +0000</pubDate>
				</item>
			</rss>`
		fmt.Fprintln(w, rssXml)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// Create a stub subscriber
	subscriber := &StubSubscriber{
		resultChan: make(chan TaskResult, 100),
	}

	// Set up the task manager
	tm := NewTaskManager(store)
	tm.Subscribe(subscriber)

	// Insert a new feed
	feedId, err := store.GetOrCreateFeedWithUrl(server.URL)
	if err != nil {
		t.Fatalf("Could not insert feed record: %v", err)
	}

	// Kick off some load feed tasks
	const numTasks int = 50
	for i := 0; i < numTasks; i++ {
		tm.ScheduleLoadFeedTask(feedId)
	}

	// Block until all tasks processed
	for i := 0; i < numTasks; i++ {
		r := <-subscriber.resultChan

		if r.Err != nil {
			t.Errorf("Unexpected error processing task: %v", r.Err)
		} else if r.FeedId != 1 {
			t.Errorf("Unexpected feed id in task result: %v", r.FeedId)
		}
	}
}
