package watcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// TestWatcher watches for file changes and runs tests
type TestWatcher struct {
	watchDir      string
	debounceDelay time.Duration
	fileFilter    func(string) bool
	watcher       *fsnotify.Watcher
	stopChan      chan struct{}
}

// NewTestWatcher creates a new test watcher for the specified directory
func NewTestWatcher(watchDir string) (*TestWatcher, error) {
	if watchDir == "" {
		var err error
		watchDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}

	return &TestWatcher{
		watchDir:      watchDir,
		debounceDelay: 500 * time.Millisecond,
		fileFilter: func(path string) bool {
			return filepath.Ext(path) == ".go"
		},
		watcher:  watcher,
		stopChan: make(chan struct{}),
	}, nil
}

// SetDebounceDelay sets the debounce delay for test runs
func (tw *TestWatcher) SetDebounceDelay(delay time.Duration) {
	tw.debounceDelay = delay
}

// SetFileFilter sets a custom file filter function
func (tw *TestWatcher) SetFileFilter(filter func(string) bool) {
	tw.fileFilter = filter
}

// RunTests runs the go tests in the watch directory
func (tw *TestWatcher) RunTests() error {
	fmt.Println("Running tests...")
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = tw.watchDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running tests: %v\n", err)
		// Play a bell sound to notify user of test failure
		fmt.Print("\a")
		return err
	}
	fmt.Println("Tests completed")
	return nil
}

// Watch starts watching for file changes and running tests
func (tw *TestWatcher) Watch() error {
	// Add directories to watch (non-recursive)
	if err := filepath.Walk(tw.watchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return tw.watcher.Add(path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error setting up directory watch: %w", err)
	}

	fmt.Println("Watching for file changes. Press Ctrl+C to exit.")

	// Run tests immediately on startup
	tw.RunTests()

	var debounceTimer *time.Timer

	// Event processing
	for {
		select {
		case <-tw.stopChan:
			return nil

		case event, ok := <-tw.watcher.Events:
			if !ok {
				return nil
			}
			// Process write events
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create {
				// Apply file filter
				if tw.fileFilter(event.Name) {
					// Reset timer if already set
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					// Debounce to run tests only once for multiple changes
					debounceTimer = time.AfterFunc(tw.debounceDelay, func() {
						fmt.Printf("\n%s changed. Running tests again.\n", event.Name)
						tw.RunTests()
					})
				}
			}

		case err, ok := <-tw.watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Watch error: %v\n", err)
		}
	}
}

// Stop stops the test watcher
func (tw *TestWatcher) Stop() {
	close(tw.stopChan)
	tw.watcher.Close()
}
