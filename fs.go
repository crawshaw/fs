// Package fs provides Context-aware file system access.
//
// TODO: implement.
package fs

import (
	"io"
	"os"
	"syscall"

	"golang.org/x/net/context"
)

// OpenLimit is the maximum number of file descriptors that can be open
// simultaneously by the fs package.
//
// Initial value is 10% less than RLIMIT_NOFILE at process initialization.
var OpenLimit int

// IO is the interface provided to read and write files.
type IO interface {
	io.ReadWriteSeeker
	io.ReaderAt
	io.Closer
}

type File struct {
	f *os.File
}

func (f *File) IO(ctx context.Context) IO {
	return nil
	//return fio{f, ctx}
}

func (f *File) Name() string {
	return f.f.Name()
}

// Open opens the named file for reading.
//
// If the number of opened files exceeds OpenLimit, Open will block until
func Open(ctx context.Context, name string) (*File, error) {
	defer interrupt(ctx)()

	f, err := os.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}

	return &File{f}, nil
}

type fio struct {
	f   *File
	ctx context.Context
}
