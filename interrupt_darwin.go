package fs

// +build darwin

import (
	"log"
	"syscall"
)

func threadID() uintptr {
	// TODO syscall.Gettid() on linux
	id, _, _ := syscall.Syscall(372 /* darwin thread_selfid */, 0, 0, 0)
	return id
}

func threadKill(id uintptr, signal int) error {
	// darwin: __pthread_kill=328
	_, _, errno := syscall.Syscall(328 /* __pthread_kill */, id, uintptr(signal), 0)
	if errno != 0 {
		log.Printf("threadKill error: %v", errno)
		return errno
	}
	return nil
}
