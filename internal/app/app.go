package app

import "github.com/wedaly/local-news/internal/store"

func Run(config Config) error {
	feedStore := store.NewFeedStore(config.DBPath)
	if err := feedStore.Initialize(); err != nil {
		return err
	}
	defer feedStore.Close()
	return nil
}
