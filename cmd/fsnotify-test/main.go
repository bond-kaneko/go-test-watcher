package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// Check command line arguments
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <directory_to_watch>\n", os.Args[0])
		os.Exit(1)
	}

	dirToWatch := os.Args[1]
	fmt.Printf("Starting fsnotify test on directory: %s\n", dirToWatch)

	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to create watcher:", err)
	}
	defer watcher.Close()

	// Setup done channel for program termination
	done := make(chan bool)

	// Start goroutine to process events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fmt.Printf("Event: %s - %s\n", event.Op, event.Name)

				// Check event types
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Printf("Modified file: %s\n", event.Name)
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Printf("Created file: %s\n", event.Name)
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					fmt.Printf("Removed file: %s\n", event.Name)
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					fmt.Printf("Renamed file: %s\n", event.Name)
				}
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					fmt.Printf("Changed permissions: %s\n", event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	// Add directory to watch
	err = watcher.Add(dirToWatch)
	if err != nil {
		log.Fatal("Failed to add directory to watcher:", err)
	}
	fmt.Printf("Successfully watching directory: %s\n", dirToWatch)

	// Demo: Create a test file every few seconds
	go func() {
		counter := 1
		for {
			// Create test file
			filename := filepath.Join(dirToWatch, fmt.Sprintf("test-file-%d.txt", counter))
			fmt.Printf("Creating file: %s\n", filename)

			file, err := os.Create(filename)
			if err != nil {
				log.Printf("Error creating file: %v\n", err)
				continue
			}

			// Write some content
			_, err = file.WriteString(fmt.Sprintf("Test content %d\n", counter))
			if err != nil {
				log.Printf("Error writing to file: %v\n", err)
			}
			file.Close()

			// Wait before next operation
			time.Sleep(3 * time.Second)

			// Modify the file
			file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("Error opening file for modification: %v\n", err)
				continue
			}
			_, err = file.WriteString(fmt.Sprintf("Modified content %d\n", counter))
			if err != nil {
				log.Printf("Error modifying file: %v\n", err)
			}
			file.Close()

			time.Sleep(3 * time.Second)

			// Delete the file
			fmt.Printf("Removing file: %s\n", filename)
			err = os.Remove(filename)
			if err != nil {
				log.Printf("Error removing file: %v\n", err)
			}

			time.Sleep(3 * time.Second)
			counter++
		}
	}()

	fmt.Println("Test program running. Press Ctrl+C to exit.")
	<-done // Block until done channel receives a value
}
