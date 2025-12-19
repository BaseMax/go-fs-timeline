package watcher

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BaseMax/go-fs-timeline/pkg/database"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fsWatcher     *fsnotify.Watcher
	db            *database.DB
	eventBuffer   []*database.Event
	bufferMu      sync.Mutex
	flushInterval time.Duration
	maxBufferSize int
}

func New(db *database.DB, flushInterval time.Duration, maxBufferSize int) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fs watcher: %w", err)
	}

	return &Watcher{
		fsWatcher:     fsWatcher,
		db:            db,
		eventBuffer:   make([]*database.Event, 0, maxBufferSize),
		flushInterval: flushInterval,
		maxBufferSize: maxBufferSize,
	}, nil
}

func (w *Watcher) AddPath(path string) error {
	return w.fsWatcher.Add(path)
}

func (w *Watcher) Watch(ctx context.Context) error {
	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Flush remaining events
			w.flush()
			return nil

		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return nil
			}
			w.handleEvent(event)

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("[ERROR] File system watcher error: %v\n", err)

		case <-ticker.C:
			w.flush()
		}
	}
}

func (w *Watcher) handleEvent(fsEvent fsnotify.Event) {
	eventType := w.getEventType(fsEvent.Op)
	fileName := filepath.Base(fsEvent.Name)
	directory := filepath.Dir(fsEvent.Name)
	fileType := w.getFileType(fileName)

	event := &database.Event{
		Timestamp: time.Now(),
		EventType: eventType,
		FilePath:  fsEvent.Name,
		FileName:  fileName,
		FileType:  fileType,
		Directory: directory,
	}

	w.bufferMu.Lock()
	w.eventBuffer = append(w.eventBuffer, event)
	shouldFlush := len(w.eventBuffer) >= w.maxBufferSize
	w.bufferMu.Unlock()

	if shouldFlush {
		w.flush()
	}
}

func (w *Watcher) flush() {
	w.bufferMu.Lock()
	if len(w.eventBuffer) == 0 {
		w.bufferMu.Unlock()
		return
	}

	events := make([]*database.Event, len(w.eventBuffer))
	copy(events, w.eventBuffer)
	w.eventBuffer = w.eventBuffer[:0]
	w.bufferMu.Unlock()

	if err := w.db.InsertEvents(events); err != nil {
		fmt.Printf("[ERROR] Failed to flush %d events to database: %v\n", len(events), err)
	} else {
		fmt.Printf("[INFO] Flushed %d events to database\n", len(events))
	}
}

func (w *Watcher) getEventType(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "CREATE"
	case op&fsnotify.Write == fsnotify.Write:
		return "WRITE"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "REMOVE"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "RENAME"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "CHMOD"
	default:
		return "UNKNOWN"
	}
}

func (w *Watcher) getFileType(fileName string) string {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return "no-extension"
	}
	return strings.TrimPrefix(ext, ".")
}

func (w *Watcher) Close() error {
	w.flush()
	return w.fsWatcher.Close()
}
