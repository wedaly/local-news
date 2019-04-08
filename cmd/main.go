package main

import (
	"fmt"
	"github.com/wedaly/local-news/internal/app"
	"os"
	"os/user"
	"path"
)

func main() {
	config := app.Config{DBPath: getDefaultDBPath()}
	if len(os.Args) > 1 {
		config.DBPath = os.Args[1]
	}

	err := app.Run(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %v\n", err)
		os.Exit(1)
	}
}

func getDefaultDBPath() string {
	const dbName string = ".localnews.db"
	if usr, err := user.Current(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Warning: could not find user's home directory: %v\n",
			err)
		return dbName
    } else {
		return path.Join(usr.HomeDir, dbName)
	}
}
