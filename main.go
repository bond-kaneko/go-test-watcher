package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bond-kaneko/go-test-watcher/watcher"
)

func main() {
	// Configure command line arguments
	coverageFlag := flag.Bool("c", false, "Enable test coverage reporting")
	dirFlag := flag.String("r", "", "Directory to watch (default: current directory)")
	delayFlag := flag.Duration("d", 500*time.Millisecond, "Debounce delay for running tests after changes")
	filterFlag := flag.String("f", "*.go", "File filter pattern (e.g., \"*.go\", \"*_test.go\")")
	flag.Parse()

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

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start watching in a goroutine
	go func() {
		if err := testWatcher.Watch(); err != nil {
			fmt.Printf("Error watching: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-signalChan
	fmt.Println("\nShutting down...")
	testWatcher.Stop()
}
