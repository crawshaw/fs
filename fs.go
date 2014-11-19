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

// Pipe returns a connected pair of Files; reads from r return bytes written to w.
func Pipe(ctx context.Context) (r, w *File, err error) {
	osr, osw, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return newFile(osr), newFile(osw), nil
}

type fio struct {
	f   *os.File
	ctx context.Context
}

func (fio fio) Seek(offset int64, whence int) (int64, error) {
	defer interrupt(fio.ctx)()
	return fio.f.Seek(offset, whence)
}

func (fio fio) Write(p []byte) (int, error) {
	defer interrupt(fio.ctx)()
	n := 0
	for len(p) > 0 {
		wn, err := fio.f.Write(p)
		n += wn
		p = p[wn:]
		if err != nil {
			eagain := false
			if perr, ok := err.(*os.PathError); ok {
				eagain = perr.Err == syscall.EAGAIN
				if perr.Err == syscall.EINTR {
					perr.Err = context.Canceled
				}
			}
			if !eagain {
				return n, err
			}
		}
		select {
		case <-fio.ctx.Done():
			return n, &os.PathError{
				Op:   "write",
				Path: fio.f.Name(),
				Err:  context.Canceled,
			}
		default:
		}
	}
	return n, nil
}

func (fio fio) Read(data []byte) (int, error) {
	defer interrupt(fio.ctx)()

	// The io.Reader contract encourages us not to return zero bytes,
	// so we spin on EAGAIN until we are canceled or bytes appear.
	for {
		n, err := fio.f.Read(data)
		if err != nil {
			eagain := false
			if perr, ok := err.(*os.PathError); ok {
				eagain = perr.Err == syscall.EAGAIN
				if perr.Err == syscall.EINTR {
					perr.Err = context.Canceled
				}
			}
			if !eagain {
				return n, err
			}
		}
		if n > 0 {
			return n, err
		}
		select {
		case <-fio.ctx.Done():
			return n, &os.PathError{
				Op:   "write",
				Path: fio.f.Name(),
				Err:  context.Canceled,
			}
		default:
		}
	}
	return len(data), nil
}

func (fio fio) ReadAt(p []byte, off int64) (n int, err error) {
	defer interrupt(fio.ctx)()
	// TODO: oh dear O_NONBLOCK woes.
	return fio.f.ReadAt(p, off)
}

func (fio fio) Close() error {
	defer interrupt(fio.ctx)()
	return fio.f.Close()
	// TODO recover OpenLimit
}
