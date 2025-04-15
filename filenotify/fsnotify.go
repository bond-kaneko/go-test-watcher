package filenotify

import (
	"github.com/fsnotify/fsnotify"
)

// EventWatcher is an implementation of FileWatcher using fsnotify
type EventWatcher struct {
	watcher *fsnotify.Watcher
	events  chan fsnotify.Event
	errors  chan error
	stopped bool
}

// NewEventWatcher returns a new EventWatcher
func NewEventWatcher() (FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	eventWatcher := &EventWatcher{
		watcher: watcher,
		events:  make(chan fsnotify.Event),
		errors:  make(chan error),
		stopped: false,
	}

	go eventWatcher.watch()

	return eventWatcher, nil
}

// Events returns the event channel
func (w *EventWatcher) Events() <-chan fsnotify.Event {
	return w.events
}

// Errors returns the error channel
func (w *EventWatcher) Errors() <-chan error {
	return w.errors
}

// Add adds a file or directory to the watch list
func (w *EventWatcher) Add(name string) error {
	return w.watcher.Add(name)
}

// Remove removes a file or directory from the watch list
func (w *EventWatcher) Remove(name string) error {
	return w.watcher.Remove(name)
}

// Close closes the watcher
func (w *EventWatcher) Close() error {
	if w.stopped {
		return nil
	}
	w.stopped = true

	// Close the fsnotify watcher
	err := w.watcher.Close()

	// Close the event and error channels
	close(w.events)
	close(w.errors)

	return err
}

// watch forwards events from the fsnotify watcher to the event channel
func (w *EventWatcher) watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.events <- event
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.errors <- err
		}
	}
}
