# Go Test Watcher

A simple command-line tool that watches Go files for changes and automatically runs tests when changes are detected.

## Features

- Watches Go files for changes
- Automatically runs tests when files are modified
- Debounces test runs to prevent multiple runs for rapid changes
- Customizable file filtering
- Audio notification (bell) when tests fail
- Optional test coverage reporting

## Installation

```bash
go install github.com/bond-kaneko/go-test-watcher@latest
```

## Usage

### Basic Usage

Simply run the tool in your directory to start watching for file changes:

```bash
go-test-watcher
```

This will:
1. Monitor all `.go` files in the current directory
2. Automatically run `go test ./...` when files change
3. Exit with Ctrl+C

### Command Line Options

```bash
go-test-watcher [options]

Options:
  -r string
        Directory to watch (default: current directory)
  -d duration
        Debounce delay for running tests after changes (default: 500ms)
  -f string
        File filter pattern (e.g., "*.go", "*_test.go") (default: "*.go")
  -c
        Enable test coverage reporting
  -v
        Display version information
```

### Examples

Watch a specific directory:
```bash
go-test-watcher -r /path/to/your/project
```

Use a longer debounce delay (for projects with frequent changes):
```bash
go-test-watcher -d 2s
```

Only watch test files:
```bash
go-test-watcher -f "*_test.go"
```

Run with test coverage reporting:
```bash
go-test-watcher -c
```

Display version:
```bash
go-test-watcher -v
```

## For Developers

You can also use this tool as a library in your own Go projects:

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bond-kaneko/go-test-watcher/watcher"
)

func main() {
	// Create a new test watcher for the current directory
	testWatcher, err := watcher.NewTestWatcher("")
	if err != nil {
		fmt.Printf("Error creating test watcher: %v\n", err)
		os.Exit(1)
	}

	// Optional custom configuration
	 testWatcher.SetDebounceDelay(1 * time.Second) // Change debounce delay
	 testWatcher.SetFileFilter(func(path string) bool { // Example custom filter
	     // Only watch files ending with _test.go using glob pattern matching
	     matched, _ := filepath.Match("*_test.go", filepath.Base(path))
	     return matched
	 })
	// testWatcher.EnableCoverage(true) // Enable test coverage reporting

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
```

## Building from Source

To build the tool with the current Git tag as the version:

```bash
git_tag=$(git describe --tags --always --dirty)
go build -ldflags="-X 'main.Version=${git_tag}'" -o go-test-watcher
```

This will embed the Git tag (e.g., v1.0.0) into the binary as the version information, which will be displayed when using the `-v` flag.

## License

MIT 