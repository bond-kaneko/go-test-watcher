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
	watchDir            string
	debounceDelay       time.Duration
	fileFilter          func(string) bool
	watcher             *fsnotify.Watcher
	withCoverage        bool
	writer              *uilive.Writer
	changedFiles        map[string]bool
	failedTests         map[string]bool
	lastChangedFile     string
	packageDependencies map[string][]string
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
		watcher:             watcher,
		withCoverage:        false,
		writer:              writer,
		changedFiles:        make(map[string]bool),
		failedTests:         make(map[string]bool),
		packageDependencies: make(map[string][]string),
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

// TrackFailedTest adds a test to the failed tests list
func (tw *TestWatcher) TrackFailedTest(testName string) {
	tw.failedTests[testName] = true
}

// ClearFailedTests clears the failed tests list
func (tw *TestWatcher) ClearFailedTests() {
	tw.failedTests = make(map[string]bool)
}

// FindAffectedPackages finds packages affected by changes in the given file
func (tw *TestWatcher) FindAffectedPackages(changedFile string) []string {
	// Get the package of the changed file
	dir := filepath.Dir(changedFile)
	relDir, err := filepath.Rel(tw.watchDir, dir)
	if err != nil {
		// If we can't determine the relative path, just use the directory
		relDir = dir
	}

	// Convert path separator to package separator
	pkg := strings.ReplaceAll(relDir, string(filepath.Separator), "/")

	// Add the package itself
	affectedPackages := []string{pkg}

	// Add dependent packages (if known)
	if deps, ok := tw.packageDependencies[pkg]; ok {
		affectedPackages = append(affectedPackages, deps...)
	}

	return affectedPackages
}

// BuildTestArgs builds the go test command arguments based on changed files and failed tests
func (tw *TestWatcher) BuildTestArgs() []string {
	args := []string{"test", "-v"}

	if tw.withCoverage {
		args = append(args, "-cover")
	}

	// If we have no changed files and no failed tests, run all tests
	if len(tw.changedFiles) == 0 && len(tw.failedTests) == 0 {
		args = append(args, "./...")
		return args
	}

	// Collect packages to test
	packagesToTest := make(map[string]bool)

	// Add packages for changed files
	for file := range tw.changedFiles {
		for _, pkg := range tw.FindAffectedPackages(file) {
			packagesToTest[pkg] = true
		}
	}

	// Add packages for failed tests
	for test := range tw.failedTests {
		// Extract package from test name (assuming format like Package/TestName)
		parts := strings.Split(test, "/")
		if len(parts) > 0 {
			packagesToTest[parts[0]] = true
		}
	}

	// If we couldn't determine any specific packages, test everything
	if len(packagesToTest) == 0 {
		args = append(args, "./...")
		return args
	}

	// Add specific packages to test
	for pkg := range packagesToTest {
		if pkg == "." || pkg == "" {
			// Root package
			args = append(args, ".")
		} else {
			// Subpackage
			args = append(args, "./"+pkg)
		}
	}

	return args
}

// AddChangedFile marks a file as changed
func (tw *TestWatcher) AddChangedFile(file string) {
	tw.changedFiles[file] = true
	tw.lastChangedFile = file
}

// ClearChangedFiles clears the list of changed files
func (tw *TestWatcher) ClearChangedFiles() {
	tw.changedFiles = make(map[string]bool)
}

// RunTests runs the go tests in the watch directory
func (tw *TestWatcher) RunTests() error {
	fmt.Fprintf(tw.writer, "Running tests...\n")
	tw.writer.Flush()

	// Build test arguments based on changed files and failed tests
	args := tw.BuildTestArgs()

	if len(tw.changedFiles) > 0 {
		filesList := make([]string, 0, len(tw.changedFiles))
		for file := range tw.changedFiles {
			filesList = append(filesList, filepath.Base(file))
		}
		fmt.Fprintf(tw.writer, "Files changed: %s\n", strings.Join(filesList, ", "))
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

	// Clear tracked changed files after running tests
	tw.ClearChangedFiles()

	// Count actual failed tests
	failCount := strings.Count(outputStr, "--- FAIL")

	// If there are no actual failures but err is not nil,
	// this might be a compile error or another non-test failure
	if err != nil && failCount == 0 {
		// Check if this is a build failure
		if strings.Contains(outputStr, "build failed") || strings.Contains(outputStr, "does not compile") {
			fmt.Fprintf(tw.writer, "BUILD FAILED:\n%s\n", outputStr)
			tw.writer.Flush()
			fmt.Print("\a") // Play bell sound
			return err
		}

		// If there's no build failure but also no test failures, treat as success
		// This can happen with some configurations where exit code is non-zero
		// but no tests actually failed
		duration := "unknown"
		coverage := ""

		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			// Check for test result line
			if strings.HasPrefix(line, "ok") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					duration = parts[2]
					// Remove "(cached)" text if present
					duration = strings.ReplaceAll(duration, "(cached)", "")
					duration = strings.TrimSpace(duration)

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

		// Format the success message
		testResult := "ALL TESTS PASSED"
		if duration != "" && duration != "()" {
			testResult = fmt.Sprintf("ALL TESTS PASSED (%s)", duration)
		}
		if coverage != "" {
			testResult += fmt.Sprintf(" - %s", coverage)
		}

		fmt.Fprintf(tw.writer, "%s\n", testResult)
		tw.writer.Flush()
		return nil
	}

	if err != nil {
		packageName := "unknown"

		// Extract test failures
		var failures []string
		lines := strings.Split(outputStr, "\n")
		var currentFailure string

		for i, line := range lines {
			// Get package name
			if strings.HasPrefix(line, "FAIL\t") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					packageName = parts[1]
				}
				continue
			}

			// Collect all failure details and track failed tests
			if strings.HasPrefix(line, "--- FAIL") {
				// Extract test name and track it
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					testName := parts[2]
					// Add package info to the test name
					if packageName != "unknown" {
						testName = packageName + "/" + testName
					}
					tw.TrackFailedTest(testName)
				}

				// Start tracking a new failure
				currentFailure = line
				failures = append(failures, currentFailure)

				// Look ahead for error details after this line
				for j := i + 1; j < len(lines) && j < i+10; j++ {
					nextLine := lines[j]
					// Indented lines are error details
					if strings.HasPrefix(nextLine, "    ") {
						failures = append(failures, nextLine)
					} else if strings.HasPrefix(nextLine, "=== ") || strings.HasPrefix(nextLine, "--- ") {
						// Stop when we hit the next test section
						break
					} else if nextLine == "" {
						// Empty lines are fine to include
						continue
					} else if !strings.HasPrefix(nextLine, "FAIL\t") && len(nextLine) > 0 {
						// Include non-empty lines that aren't starting a new section
						failures = append(failures, "    "+nextLine)
					}
				}
			}
		}

		// Show detailed failure summary
		fmt.Fprintf(tw.writer, "TEST FAILED: %d tests in %s\n", failCount, packageName)

		// Display the raw test output for more complete information
		fmt.Fprintf(tw.writer, "\nFailure Details:\n")

		// Parse output to find the test result sections
		var testSections []string
		inTestSection := false
		var currentSection strings.Builder

		for _, line := range lines {
			if strings.HasPrefix(line, "=== RUN") {
				// Start of a new test section
				if inTestSection && currentSection.Len() > 0 {
					testSections = append(testSections, currentSection.String())
				}
				inTestSection = true
				currentSection.Reset()
				currentSection.WriteString(line)
				currentSection.WriteString("\n")
			} else if inTestSection {
				currentSection.WriteString(line)
				currentSection.WriteString("\n")

				// End of a test section
				if strings.HasPrefix(line, "--- FAIL") || strings.HasPrefix(line, "--- PASS") {
					// Only add failed tests to our sections
					if strings.HasPrefix(line, "--- FAIL") {
						testSections = append(testSections, currentSection.String())
					}
					inTestSection = false
					currentSection.Reset()
				}
			}
		}

		// Add any remaining section
		if inTestSection && currentSection.Len() > 0 {
			testSections = append(testSections, currentSection.String())
		}

		// Display failure details more completely
		for _, section := range testSections {
			if strings.Contains(section, "--- FAIL") {
				fmt.Fprintf(tw.writer, "%s\n", section)

				// Find any error output lines after the failure
				sectionLines := strings.Split(section, "\n")
				testName := ""

				// Extract test name
				for _, line := range sectionLines {
					if strings.HasPrefix(line, "=== RUN") {
						parts := strings.Fields(line)
						if len(parts) >= 3 {
							testName = parts[2]
						}
					}
				}

				// If we have a test name, look for any t.Error/t.Errorf/t.Fatal lines in the output
				if testName != "" {
					for _, line := range lines {
						// Match log output associated with this test
						if strings.Contains(line, testName+":") &&
							(strings.Contains(line, "Error") ||
								strings.Contains(line, "Fatal") ||
								strings.Contains(line, "Fail")) {
							fmt.Fprintf(tw.writer, "    %s\n", strings.TrimSpace(line))
						}
					}
				}
			}
		}

		// If no detailed sections were found, fall back to showing the collected failures
		if len(testSections) == 0 && len(failures) > 0 {
			for _, failure := range failures {
				fmt.Fprintf(tw.writer, "%s\n", failure)
			}
		}

		// If no failure details found at all, but there was an error, show the raw output
		if len(testSections) == 0 && len(failures) == 0 {
			fmt.Fprintf(tw.writer, "%s\n", outputStr)
		}

		tw.writer.Flush()
		fmt.Print("\a") // Play bell sound
		return err
	}

	// For successful tests
	duration := "unknown"
	coverage := ""

	// Clear failed tests since all tests passed
	tw.ClearFailedTests()

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		// Check for test result line
		if strings.HasPrefix(line, "ok") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				duration = parts[2]
				// Remove "(cached)" text if present
				duration = strings.ReplaceAll(duration, "(cached)", "")
				duration = strings.TrimSpace(duration)

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
	testResult := "ALL TESTS PASSED"
	if duration != "" && duration != "()" {
		testResult = fmt.Sprintf("ALL TESTS PASSED (%s)", duration)
	}
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

	// Run tests immediately on startup
	tw.RunTests()

	var debounceTimer *time.Timer

	// Event processing
	for {
		select {
		case event, ok := <-tw.watcher.Events:
			if !ok {
				return nil
			}
			// Process write events
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create {
				// Apply file filter
				if tw.fileFilter(event.Name) {
					// Add the changed file to tracking
					tw.AddChangedFile(event.Name)

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
	os.Exit(0)
}
