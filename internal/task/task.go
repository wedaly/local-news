package task

import (
	"github.com/wedaly/local-news/internal/feed"
	"github.com/wedaly/local-news/internal/store"
)

type TaskResult struct {
	feedId store.FeedId
	err    error
}

type TaskSubscriber interface {
	// TODO: explain thread safety
	HandleTaskCompletion(TaskResult)
}

type TaskManager struct {
	feedStore   *store.FeedStore
	subscribers []TaskSubscriber
	loaderChan  chan *feed.FeedLoader
}

func NewTaskManager(feedStore *store.FeedStore, subscribers []TaskSubscriber) *TaskManager {
	const numFeedLoaders int = 10
	loaderChan := make(chan *feed.FeedLoader, numFeedLoaders)
	for i := 0; i < numFeedLoaders; i++ {
		loaderChan <- feed.NewFeedLoader()
	}

	return &TaskManager{feedStore, subscribers, loaderChan}
}

func (m *TaskManager) ScheduleLoadFeedTask(url string) {
	go func() {
		// Block until loader is available
		loader := <-m.loaderChan
		defer func() { m.loaderChan <- loader }()

		// Retrieve and parse the feed from a URL
		feed, err := loader.LoadFeedFromUrl(url)
		if err != nil {
			m.notifySubscribers(TaskResult{err: err})
			return
		}

		// Update the database
		feedId, err := m.feedStore.UpsertFeed(feed)
		if err != nil {
			m.notifySubscribers(TaskResult{err: err})
			return
		}

		// Notify subscribers that the task completed successfully
		m.notifySubscribers(TaskResult{feedId: feedId})
	}()
}

func (m *TaskManager) notifySubscribers(r TaskResult) {
	// This is thread-safe because the subscribers list is immutable
	for _, s := range m.subscribers {
		// The subscriber is responsible for ensuring that
		// this method is thread-safe
		s.HandleTaskCompletion(r)
	}
}
