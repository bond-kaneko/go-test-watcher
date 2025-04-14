package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bond-kaneko/go-test-watcher/watcher"
)

// Version information - will be set by the build process
var (
	Version = "dev"
)

func main() {
	// Configure command line arguments
	versionFlag := flag.Bool("v", false, "Display version information")
	coverageFlag := flag.Bool("c", false, "Enable test coverage reporting")
	dirFlag := flag.String("r", "", "Directory to watch (default: current directory)")
	delayFlag := flag.Duration("d", 500*time.Millisecond, "Debounce delay for running tests after changes")
	filterFlag := flag.String("f", "*.go", "File filter pattern (e.g., \"*.go\", \"*_test.go\")")
	flag.Parse()

	// Display version if requested
	if *versionFlag {
		fmt.Printf("go-test-watcher version %s\n", Version)
		return
	}

	// Create a new test watcher for the current directory
	testWatcher, err := watcher.NewTestWatcher(*dirFlag)
	if err != nil {
		fmt.Printf("Error creating test watcher: %v\n", err)
		os.Exit(1)
	}

	// Set debounce delay
	testWatcher.SetDebounceDelay(*delayFlag)

	// Set file filter if provided
	if *filterFlag != "" {
		testWatcher.SetFileFilter(func(path string) bool {
			matched, err := filepath.Match(*filterFlag, filepath.Base(path))
			if err != nil {
				fmt.Printf("Error in file filter pattern: %v\n", err)
				return false // Or handle error appropriately
			}
			return matched
		})
	}

	// Set coverage option
	if *coverageFlag {
		testWatcher.EnableCoverage(true)
		fmt.Println("Test coverage reporting enabled")
	}

	go func() {
		if err := testWatcher.Watch(); err != nil {
			fmt.Printf("Error watching: %v\n", err)
			os.Exit(1)
		}
	}()
}
