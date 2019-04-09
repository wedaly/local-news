package task

import (
	"github.com/wedaly/local-news/internal/feed"
	"github.com/wedaly/local-news/internal/store"
)

// TaskResult describes the outcome of a task to load a feed
type TaskResult struct {
	feedId store.FeedId
	err    error
}

// TaskSubscriber receives notifications about tasks
type TaskSubscriber interface {
	// HandleTaskScheduled is invoked when a new task is scheduled.
	// This must be thread-safe.
	HandleTaskScheduled()

	// HandleTaskCompleted is invoked when a task has been completed.
	// This must be thread-safe.
	HandleTaskCompleted(TaskResult)
}

// TaskManager schedules async, concurrent tasks to load feeds
// It notifies all subscribers when tasks are scheduled and completed.
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

// ScheduleLoadFeedTask enqueues a new task to load a feed from a URL.
// If successfully loaded, the feed data is written to the database.
// Subscribers are notified when the task is scheduled and completed.
func (m *TaskManager) ScheduleLoadFeedTask(url string) {
	m.notifyTaskScheduled()
	go func() {
		// Block until loader is available
		loader := <-m.loaderChan
		defer func() { m.loaderChan <- loader }()

		// Retrieve and parse the feed from a URL
		feed, err := loader.LoadFeedFromUrl(url)
		if err != nil {
			m.notifyTaskCompleted(TaskResult{err: err})
			return
		}

		// Update the database
		feedId, err := m.feedStore.UpsertFeed(feed)
		if err != nil {
			m.notifyTaskCompleted(TaskResult{err: err})
			return
		}

		// Notify subscribers that the task completed successfully
		m.notifyTaskCompleted(TaskResult{feedId: feedId})
	}()
}

func (m *TaskManager) notifyTaskScheduled() {
	// This is thread-safe because the subscribers list is immutable
	for _, s := range m.subscribers {
		// The subscriber is responsible for ensuring that
		// this method is thread-safe
		s.HandleTaskScheduled()
	}
}

func (m *TaskManager) notifyTaskCompleted(r TaskResult) {
	// This is thread-safe because the subscribers list is immutable
	for _, s := range m.subscribers {
		// The subscriber is responsible for ensuring that
		// this method is thread-safe
		s.HandleTaskCompleted(r)
	}
}
