package alert

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileWriter is an io.Writer that appends alert lines to a log file.
type FileWriter struct {
	path string
	file *os.File
}

// NewFileWriter opens (or creates) the file at path for appending.
func NewFileWriter(path string) (*FileWriter, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("alert file dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open alert file: %w", err)
	}
	return &FileWriter{path: path, file: f}, nil
}

// Write implements io.Writer.
func (fw *FileWriter) Write(p []byte) (int, error) {
	return fw.file.Write(p)
}

// Close closes the underlying file.
func (fw *FileWriter) Close() error {
	return fw.file.Close()
}
