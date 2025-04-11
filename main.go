package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func runTests(dir string) {
	fmt.Println("Running tests...")
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running tests: %v\n", err)
		return
	}
	fmt.Println("Tests completed")
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Run tests once at startup
	runTests(dir)

	// Setup file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error initializing watcher: %v\n", err)
		os.Exit(1)
	}
	defer watcher.Close()

	// Add directories to watch (non-recursive)
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories like .git
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		fmt.Printf("Error setting up directory watch: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Watching for file changes. Press Ctrl+C to exit.")

	// Debounce handling
	var debounceTimer *time.Timer
	const debounceDelay = 500 * time.Millisecond

	// Event processing
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Process write events
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create {
				// Only target .go files
				if filepath.Ext(event.Name) == ".go" {
					// Reset timer if already set
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					// Debounce to run tests only once for multiple changes
					debounceTimer = time.AfterFunc(debounceDelay, func() {
						fmt.Printf("\n%s changed. Running tests again.\n", event.Name)
						runTests(dir)
					})
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watch error: %v\n", err)
		}
	}
}
