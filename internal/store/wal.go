package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type WALEntry struct {
	Ops       string    `json:"operation"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Version   int64     `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

type WAL struct {
	file     *os.File
	filePath string
	writer   *bufio.Writer

	mu sync.Mutex
}

// NewWAL creates a new WAL instance
func NewWAL(dir string) (*WAL, error) {
	// Create WAL directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WAL directory: %w", err)
	}

	// Create or open WAL file
	filePath := filepath.Join(dir, "wal.log")
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	return &WAL{
		file:     file,
		filePath: filePath,
		writer:   bufio.NewWriter(file),
	}, nil
}

func (w *WAL) Write(entry WALEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal WAL entry: %w", err)
	}
	// Write entry to WAL
	if _, err := w.writer.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to WAL: %w", err)
	}

	// Flush to disk
	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush WAL: %w", err)
	}

	return nil
}

// Recover replays the WAL to restore the store state
func (w *WAL) Recover(store *Store) error {
	// Open WAL file for reading
	file, err := os.OpenFile(w.filePath, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open WAL file for recovery: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry WALEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return fmt.Errorf("failed to unmarshal WAL entry: %w", err)
		}

		// Apply the operation to the store
		switch entry.Ops {
		case "SET":
			object, err := store.Set(entry.Key, entry.Value)
			if err != nil {
				return fmt.Errorf("failed to apply SET operation: %w", err)
			}
			// Ensure the version matches
			if len(object.Values) > 0 {
				object.Values[len(object.Values)-1].Version = entry.Version
				object.Values[len(object.Values)-1].Timestamp = entry.Timestamp
			}
		default:
			return fmt.Errorf("unknown operation in WAL: %s", entry.Ops)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading WAL file: %w", err)
	}

	return nil
}

func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.file.Close(); err != nil {
		return fmt.Errorf("failed to close WAL file: %w", err)
	}

	return nil
}
