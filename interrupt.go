package fs

import (
	"runtime"
	"syscall"

	"golang.org/x/net/context"
)

// interrupt starts a background task to send the current goroutine a SIGINT
// when ctx is done.
func interrupt(ctx context.Context) (cleanup func()) {
	runtime.LockOSThread()
	done := make(chan struct{})
	id := threadID()

	go func() {
		select {
		case <-ctx.Done():
			// Ignore error, nothing useful we can do with it.
			threadKill(id, int(syscall.SIGINT))
		case <-done:
		}
	}()

	return func() {
		runtime.UnlockOSThread()
		done <- struct{}{} // don't leak goroutine
	}
}
