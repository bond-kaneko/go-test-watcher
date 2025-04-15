// Package filenotify provides a mechanism for watching file(s) for changes.
// It abstracts fsnotify, and provides a poll-based notifier which fsnotify does not support.
// These are wrapped up in a common interface so that either can be used interchangeably
// in your code.
package filenotify

import (
	"github.com/fsnotify/fsnotify"
)

// FileWatcher is an interface for implementing file notification watchers
type FileWatcher interface {
	// Events returns the channel for watching events
	Events() <-chan fsnotify.Event
	// Errors returns the channel for watching errors
	Errors() <-chan error
	// Add starts watching the named file or directory
	Add(name string) error
	// Remove stops watching the named file or directory
	Remove(name string) error
	// Close stops watching and closes the channels
	Close() error
}

// New tries to use an fs-event watcher, and falls back to the poller if there is an error
func New() (FileWatcher, error) {
	watcher, err := NewEventWatcher()
	if err != nil {
		return NewPollingWatcher(), nil
	}
	return watcher, nil
}
