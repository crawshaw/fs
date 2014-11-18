package fs

import (
	"log"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/net/context"
)

const intrSig = syscall.SIGUSR1

const SIG_DFL = 0
const SIG_IGN = 1

func funcPC(f interface{}) uintptr {
	const ptrSize = unsafe.Sizeof(uintptr(0))
	pc := uintptr(unsafe.Pointer(&f)) + ptrSize
	return **(**uintptr)(unsafe.Pointer(pc))
}

func threadID() uintptr
func sigtramp()

func mysighandler(sig int32) {
	// TODO replace with no-op we can swap out when testing
	println("mysighandler", threadID())
}

func init() {
	setsighandler()
}

// interrupt starts a background task to send the current goroutine a SIGUSR1
// when ctx is done.
func interrupt(ctx context.Context) (cleanup func()) {
	runtime.LockOSThread()
	done := make(chan struct{}, 1)
	tid := threadID()

	//unblocksig()

	go func() {
		select {
		case <-ctx.Done():
			threadKill(tid)
		case <-done:
		}
	}()

	return func() {
		log.Printf("in cleanup")
		//blocksig()
		runtime.UnlockOSThread()
		done <- struct{}{} // don't leak goroutine
	}
}
