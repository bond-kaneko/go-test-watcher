package watcher

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gosuri/uilive"
)

// TestWatcher watches for file changes and runs tests
type TestWatcher struct {
	watchDir      string
	debounceDelay time.Duration
	fileFilter    func(string) bool
	watcher       *fsnotify.Watcher
	stopChan      chan struct{}
	withCoverage  bool
	writer        *uilive.Writer
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

	writer := uilive.New()
	writer.RefreshInterval = time.Millisecond * 100

	return &TestWatcher{
		watchDir:      watchDir,
		debounceDelay: 500 * time.Millisecond,
		fileFilter: func(path string) bool {
			return filepath.Ext(path) == ".go"
		},
		watcher:      watcher,
		stopChan:     make(chan struct{}),
		withCoverage: false,
		writer:       writer,
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

// EnableCoverage enables test coverage reporting
func (tw *TestWatcher) EnableCoverage(enabled bool) {
	tw.withCoverage = enabled
}

// RunTests runs the go tests in the watch directory
func (tw *TestWatcher) RunTests() error {
	fmt.Fprintf(tw.writer, "Running tests...\n")
	tw.writer.Flush()

	args := []string{"test", "./...", "-v=true"} // Enable verbosity for more detailed output
	if tw.withCoverage {
		args = append(args, "-cover")
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = tw.watchDir

	// Capture all output
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	// Run the command
	err := cmd.Run()

	// Parse the output to get a summary
	outputStr := output.String()
	if err != nil {
		// Get the number of failing tests
		failCount := strings.Count(outputStr, "--- FAIL")
		packageName := "unknown"

		// Extract test failures
		var failures []string
		lines := strings.Split(outputStr, "\n")
		inFailBlock := false

		for _, line := range lines {
			// Get package name
			if strings.HasPrefix(line, "FAIL\t") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					packageName = parts[1]
				}
				continue
			}

			// Collect failure details
			if strings.HasPrefix(line, "--- FAIL") {
				inFailBlock = true
				failures = append(failures, line)
				continue
			}

			if inFailBlock {
				if strings.HasPrefix(line, "    ") { // Test failure details indented with spaces
					failures = append(failures, line)
				} else if line == "" {
					// Empty line after failure block
					inFailBlock = false
				}
			}
		}

		// Show detailed failure summary
		fmt.Fprintf(tw.writer, "TEST FAILED: %d tests in %s\n", failCount, packageName)

		// Display failure details
		if len(failures) > 0 {
			fmt.Fprintf(tw.writer, "\nFailure Details:\n")
			for _, failure := range failures {
				fmt.Fprintf(tw.writer, "%s\n", failure)
			}
		}

		tw.writer.Flush()
		fmt.Print("\a") // Play bell sound
		return err
	}

	// For successful tests
	packageName := "unknown"
	duration := "unknown"
	coverage := ""

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		// Check for test result line
		if strings.HasPrefix(line, "ok") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				packageName = parts[1]
				duration = parts[2]

				// Look for coverage information
				if tw.withCoverage && len(parts) >= 4 {
					for i, part := range parts {
						if strings.Contains(part, "coverage") || strings.HasSuffix(part, "%") {
							// Coverage information found
							coverage = strings.Join(parts[i:], " ")
							break
						}
					}
				}
				break
			}
		}
	}

	if tw.withCoverage && coverage == "" {
		// Try to find coverage information in another line
		for _, line := range lines {
			if strings.Contains(line, "coverage") {
				coverage = strings.TrimSpace(line)
				break
			}
		}
	}

	// Format the success message with coverage information if available
	testResult := fmt.Sprintf("TEST PASSED: %s (%s)", packageName, duration)
	if coverage != "" {
		testResult += fmt.Sprintf(" - %s", coverage)
	}

	fmt.Fprintf(tw.writer, "%s\n", testResult)
	tw.writer.Flush()
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

	// Start the live writer
	tw.writer.Start()
	defer tw.writer.Stop()

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
						// Show which file changed
						fmt.Fprintf(tw.writer, "%s changed. Running tests again.\n", event.Name)
						tw.writer.Flush()
						tw.RunTests()
					})
				}
			}

		case err, ok := <-tw.watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(tw.writer, "Watch error: %v\n", err)
			tw.writer.Flush()
		}
	}
}

// Stop stops the test watcher
func (tw *TestWatcher) Stop() {
	close(tw.stopChan)
	tw.watcher.Close()
	tw.writer.Stop()
}
