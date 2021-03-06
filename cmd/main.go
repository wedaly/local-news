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

	// Set locale from environment variables (e.g. LC_ALL)
	if err := i18n.SetLocaleFromEnv(); err != nil {
		// Fallback to "C", which should be available everywhere
		if err := i18n.SetLocale("C"); err != nil {
			panic(err)
		}
	}

	// Initialize i18n modules based on current locale
	i18n.InitTranslations("localnews", []string{
		"./configs/locale",
		"/usr/share/locale",
	})
	i18n.InitDateFormats()

	// Load localized app configuration
	config := i18n.LoadConfig([]string{
		"./configs/etc",
		"/etc/localnews",
	})

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
		config,
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
