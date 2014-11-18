package fs

// +build darwin

import (
	"syscall"
	"unsafe"
)

const (
	SYS_SIGACTION       = 46
	SYS_PTHREAD_KILL    = 328
	SYS_PTHREAD_SIGMASK = 329
)

func threadKill(tid uintptr) error {
	_, _, errno := syscall.Syscall(SYS_PTHREAD_KILL, tid, uintptr(intrSig), 0)
	if errno != 0 {
		return errno
	}
	return nil
}

func setnonblock(fd uintptr) {
	syscall.Syscall(syscall.SYS_FCNTL, fd, syscall.F_SETFL, syscall.O_NONBLOCK)
}

func setsighandler() {
	const SA_ONSTACK = 0x1
	var sa sigactiont
	sa.sa_flags = SA_ONSTACK
	sa.sa_mask = ^uint32(0)
	sa.sa_tramp = unsafe.Pointer(funcPC(sigtramp))
	*(*uintptr)(unsafe.Pointer(&sa.__sigaction_u)) = funcPC(mysighandler)
	_, _, errno := syscall.Syscall(SYS_SIGACTION, uintptr(intrSig), uintptr(unsafe.Pointer(&sa)), 0)
	if errno != 0 {
		panic(errno.Error())
	}
}

// TODO: do we need to block/unblock the signal? it appears not
const SIG_SETMASK = 3

var oset uint32

func blocksig() {
	osetptr := uintptr(unsafe.Pointer(&oset))
	syscall.Syscall(SYS_PTHREAD_SIGMASK, SIG_SETMASK, osetptr, 0)
}

func unblocksig() {
	// TODO: tigher mask
	sigsetnone := uint32(0)
	sigsetnoneptr := uintptr(unsafe.Pointer(&sigsetnone))
	osetptr := uintptr(unsafe.Pointer(&oset))
	syscall.Syscall(SYS_PTHREAD_SIGMASK, SIG_SETMASK, sigsetnoneptr, osetptr)
}
