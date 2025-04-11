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
