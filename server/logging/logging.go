package logging

import (
	"context"
	"errors"
	"io"
)

// ErrNotFound is returned when the log does not exist.
var ErrNotFound = errors.New("stream: not found")

// Entry defines a log entry.
type Entry struct {
	// ID identifies this message.
	ID string `json:"id,omitempty"`

	// Data is the actual data in the entry.
	Data []byte `json:"data"`

	// Tags represents the key-value pairs the
	// entry is tagged with.
	Tags map[string]string `json:"tags,omitempty"`
}

// Handler defines a callback function for handling log entries.
type Handler func(...*Entry)

// Log defines a log multiplexer.
type Log interface {
	// Open opens the log.
	Open(c context.Context, path string) error

	// Write writes the entry to the log.
	Write(c context.Context, path string, entry *Entry) error

	// Tail tails the log.
	Tail(c context.Context, path string, handler Handler) error

	// Close closes the log.
	Close(c context.Context, path string) error

	// Snapshot snapshots the stream to Writer w.
	Snapshot(c context.Context, path string, w io.Writer) error

	// Info returns runtime information about the multiplexer.
	// Info(c context.Context) (interface{}, error)
}
