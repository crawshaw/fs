// Package fs provides Context-aware file system access.
//
// TODO: implement.
package fs

import (
	"io"
	"os"
	"runtime"
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
	return fio{f.f, ctx}
}

func (f *File) Name() string {
	return f.f.Name()
}

func newFile(osf *os.File) *File {
	if osf == nil {
		return nil
	}
	f := &File{osf}
	runtime.SetFinalizer(osf, func(osf *os.File) {
		osf.Close()
		// TODO recover OpenLimit
	})
	return f
}

// Open opens the named file for reading.
//
// If the number of opened files exceeds OpenLimit, Open will block until
// another file is closed.
//
// If there is an error, it will be of type *PathError.
func Open(ctx context.Context, name string) (file *File, err error) {
	// TODO: is O_NONBLOCK a bad idea?
	return OpenFile(ctx, name, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}

// OpenFile is the generalized open call; most users will use Open
// or Create instead.
//
// If the number of open files exceeds OpenLimit, Open will block until
// another file is closed.
//
// If there is an error, it will be of type *os.PathError.
func OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (file *File, err error) {
	defer interrupt(ctx)()
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return newFile(f), nil
}

type fio struct {
	f   *os.File
	ctx context.Context
}

func (fio fio) Seek(offset int64, whence int) (int64, error) {
	defer interrupt(fio.ctx)()
	return fio.f.Seek(offset, whence)
}

func (fio fio) Write(p []byte) (n int, err error) {
	defer interrupt(fio.ctx)()
	return fio.f.Write(p)
}

func (fio fio) Read(data []byte) (n int, err error) {
	defer interrupt(fio.ctx)()
	return fio.f.Read(data)
}

func (fio fio) ReadAt(p []byte, off int64) (n int, err error) {
	defer interrupt(fio.ctx)()
	return fio.f.ReadAt(p, off)
}

func (fio fio) Close() error {
	defer interrupt(fio.ctx)()
	return fio.f.Close()
	// TODO recover OpenLimit
}
