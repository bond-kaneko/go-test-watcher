# Go Test Watcher

A simple library to watch Go files for changes and automatically run tests when changes are detected.

## Features

- Watches Go files for changes
- Automatically runs tests when files are modified
- Debounces test runs to prevent multiple runs for rapid changes
- Customizable file filtering

## Installation

```bash
go get github.com/bond-kaneko/go-test-watcher
```

## Usage

### Basic Usage

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bond-kaneko/go-test-watcher/watcher"
)

func main() {
	// Create a new test watcher for the current directory
	testWatcher, err := watcher.NewTestWatcher("")
	if err != nil {
		fmt.Printf("Error creating test watcher: %v\n", err)
		os.Exit(1)
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
```

### Advanced Usage

You can customize the behavior of the test watcher:

```go
// Create a watcher for a specific directory
testWatcher, err := watcher.NewTestWatcher("/path/to/your/project")

// Set a custom debounce delay
testWatcher.SetDebounceDelay(1 * time.Second)

// Set a custom file filter (only watch test files)
testWatcher.SetFileFilter(func(path string) bool {
    return strings.HasSuffix(path, "_test.go")
})
```

## License

MIT 