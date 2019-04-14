package main

import (
	"fmt"
	"github.com/wedaly/local-news/internal/controller"
	"github.com/wedaly/local-news/internal/i18n"
	"github.com/wedaly/local-news/internal/store"
	"github.com/wedaly/local-news/internal/task"
	"os"
	"os/user"
	"path"
)

func main() {
	// Command line arg to set the DB path (optional)
	dbPath := getDefaultDBPath()
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	// Load i18n translations
	i18n.InitTranslations("localnews")

	// Open connection to the SQLite database
	feedStore := store.NewFeedStore(dbPath)
	if err := feedStore.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open database at '%v': %v", dbPath, err)
		os.Exit(1)
	}
	defer feedStore.Close()

	// Set up task manager
	taskManager := task.NewTaskManager(feedStore)

	// Set up TUI and run event loop
	ac := controller.NewAppController(
		feedStore,
		taskManager)
	if err := ac.App.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running event loop: %v", err)
		os.Exit(1)
	}
}

func getDefaultDBPath() string {
	const dbName string = ".localnews.db"
	if usr, err := user.Current(); err != nil {
		return dbName
	} else {
		return path.Join(usr.HomeDir, dbName)
	}
}
