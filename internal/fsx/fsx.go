// Package fsx contains file system extension
package fsx

import (
	"fmt"
	"os"
	"syscall"
)

// File is a generic file. This interface is taken from the draft
// iofs golang design. We'll use fs.File when available.
type File interface {
	Stat() (os.FileInfo, error)
	Read([]byte) (int, error)
	Close() error
}

// FS is a generic file system. Like File, it's adapted from
// the draft iofs golang design document.
type FS interface {
	Open(name string) (File, error)
}

// Open is a wrapper for os.Open
func Open(pathname string) (File, error) {
	return OpenWithFS(filesystem{}, pathname)
}

// OpenWithFS is to ensure that we're not attempting to open a directory.
func OpenWithFS(fs FS, pathname string) (File, error) {
	file, err := fs.Open(pathname)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if info.IsDir() {
		file.Close()
		return nil, fmt.Errorf("input path points to a directory: %w", syscall.EISDIR)
	}
	return file, nil
}

type filesystem struct{}

// Open with filesystem uses os.Open directly
func (filesystem) Open(pathname string) (File, error) {
	return os.Open(pathname)
}
