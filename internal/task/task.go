package task

import (
	"github.com/wedaly/local-news/internal/feed"
	"github.com/wedaly/local-news/internal/store"
	"sync"
)

// TaskResult describes the outcome of a task to load a feed
type TaskResult struct {
	FeedId store.FeedId
	Err    error
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
	feedStore        *store.FeedStore
	subscribersMutex sync.Mutex
	subscribers      []TaskSubscriber
	loaderChan       chan *feed.FeedLoader
}

func NewTaskManager(feedStore *store.FeedStore) *TaskManager {
	const numFeedLoaders int = 10
	loaderChan := make(chan *feed.FeedLoader, numFeedLoaders)
	for i := 0; i < numFeedLoaders; i++ {
		loaderChan <- feed.NewFeedLoader()
	}

	return &TaskManager{
		feedStore:   feedStore,
		subscribers: make([]TaskSubscriber, 0),
		loaderChan:  loaderChan,
	}
}

func (m *TaskManager) Subscribe(s TaskSubscriber) {
	m.subscribersMutex.Lock()
	defer m.subscribersMutex.Unlock()
	{
		m.subscribers = append(m.subscribers, s)
	}
}

// ScheduleLoadFeedTask enqueues a new task to load a feed from a URL.
// If successfully loaded, the feed data is written to the database.
// Subscribers are notified when the task is scheduled and completed.
func (m *TaskManager) ScheduleLoadFeedTask(feedId store.FeedId) {
	m.notifyTaskScheduled()
	go func() {
		// Block until loader is available
		loader := <-m.loaderChan
		defer func() { m.loaderChan <- loader }()

		// Retrieve the feed record
		// This implicitly validates that the feed has not been deleted
		feedRecord, err := m.feedStore.RetrieveFeed(feedId)
		if err != nil {
			m.notifyTaskCompleted(TaskResult{Err: err})
			return
		}

		// Retrieve and parse the feed from a URL
		feed, err := loader.LoadFeedFromUrl(feedRecord.Url)
		if err != nil {
			m.notifyTaskCompleted(TaskResult{Err: err})
			return
		}

		// Update the database
		err = m.feedStore.UpdateFeed(feedId, feed)
		if err != nil {
			m.notifyTaskCompleted(TaskResult{Err: err})
			return
		}

		// Notify subscribers that the task completed successfully
		m.notifyTaskCompleted(TaskResult{FeedId: feedId})
	}()
}

func (m *TaskManager) notifyTaskScheduled() {
	m.subscribersMutex.Lock()
	defer m.subscribersMutex.Unlock()
	{
		for _, s := range m.subscribers {
			// The subscriber is responsible for ensuring that
			// this method is thread-safe
			s.HandleTaskScheduled()
		}
	}
}

func (m *TaskManager) notifyTaskCompleted(r TaskResult) {
	m.subscribersMutex.Lock()
	defer m.subscribersMutex.Unlock()
	{
		for _, s := range m.subscribers {
			// The subscriber is responsible for ensuring that
			// this method is thread-safe
			s.HandleTaskCompleted(r)
		}
	}
}
