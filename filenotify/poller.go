package filenotify

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// PollingWatcher is an implementation of FileWatcher based on polling
type PollingWatcher struct {
	// interval is the time between polling for file changes
	interval time.Duration
	// files is the list of files being watched
	files map[string]fileInfo
	// events is the channel where events are reported
	events chan fsnotify.Event
	// errors is the channel where errors are reported
	errors chan error
	// stop is used to stop the polling
	stop chan struct{}
	// mutex guards access to files map
	mutex sync.Mutex
	// done is closed when polling has stopped
	done chan struct{}
}

type fileInfo struct {
	ModTime time.Time
	Size    int64
	IsDir   bool
}

// NewPollingWatcher returns a new polling watcher with the default interval of 200ms
func NewPollingWatcher() FileWatcher {
	return NewPollingWatcherWithInterval(200 * time.Millisecond)
}

// NewPollingWatcherWithInterval returns a new polling watcher with the specified interval
func NewPollingWatcherWithInterval(interval time.Duration) FileWatcher {
	watcher := &PollingWatcher{
		interval: interval,
		files:    make(map[string]fileInfo),
		events:   make(chan fsnotify.Event),
		errors:   make(chan error),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	go watcher.poll()
	return watcher
}

// Add adds a file or directory to the watch list
func (w *PollingWatcher) Add(name string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Get initial file info
	f, err := os.Stat(name)
	if err != nil {
		return err
	}

	// Convert to our internal fileInfo type
	info := fileInfo{
		ModTime: f.ModTime(),
		Size:    f.Size(),
		IsDir:   f.IsDir(),
	}

	// Add to the watched files
	w.files[name] = info

	return nil
}

// Remove removes a file or directory from the watch list
func (w *PollingWatcher) Remove(name string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if _, exists := w.files[name]; !exists {
		return errors.New("file or directory is not being watched")
	}

	delete(w.files, name)
	return nil
}

// Events returns the event channel
func (w *PollingWatcher) Events() <-chan fsnotify.Event {
	return w.events
}

// Errors returns the error channel
func (w *PollingWatcher) Errors() <-chan error {
	return w.errors
}

// Close stops the polling watcher
func (w *PollingWatcher) Close() error {
	close(w.stop)
	<-w.done
	close(w.events)
	close(w.errors)
	return nil
}

// poll checks for changes to the watched files at the specified interval
func (w *PollingWatcher) poll() {
	defer close(w.done)

	// Use a ticker to poll at the specified interval
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkFiles()
		case <-w.stop:
			return
		}
	}
}

// checkFiles checks all watched files for changes
func (w *PollingWatcher) checkFiles() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for name, oldInfo := range w.files {
		// Get current file info
		currentFileInfo, err := os.Stat(name)
		if err != nil {
			// Check if the file was deleted
			if os.IsNotExist(err) {
				// Fire a delete event
				w.events <- fsnotify.Event{
					Name: name,
					Op:   fsnotify.Remove,
				}
				// Remove the file from our tracking
				delete(w.files, name)
			} else {
				// Some other error
				w.errors <- err
			}
			continue
		}

		// Get file details
		currentInfo := fileInfo{
			ModTime: currentFileInfo.ModTime(),
			Size:    currentFileInfo.Size(),
			IsDir:   currentFileInfo.IsDir(),
		}

		// Check if the file was modified
		if currentInfo.ModTime != oldInfo.ModTime || currentInfo.Size != oldInfo.Size {
			// Fire a modify event
			w.events <- fsnotify.Event{
				Name: name,
				Op:   fsnotify.Write,
			}
			// Update the file info
			w.files[name] = currentInfo
		}
	}
}
