package fs

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"unsafe"
)

func threadID() uintptr { return uintptr(syscall.Gettid()) }

func threadKill(tid uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_TKILL, tid, uintptr(intrSig), 0)
	if errno != 0 {
		return errno
	}
	return nil
}

func setnonblock(fd uintptr) {
	syscall.Syscall(syscall.SYS_FCNTL, fd, syscall.F_SETFL, syscall.O_NONBLOCK)
}

func sigaction(sig uintptr, new, old *sigactiont, size uintptr) int32 {
	newptr := uintptr(unsafe.Pointer(new))
	_, _, errno := syscall.Syscall6(syscall.SYS_RT_SIGACTION, sig, newptr, 0, size, 0, 0)
	return int32(errno)
}

func sigreturn()

func setsighandler() {
	// If we used our own custom signal handler like on darwin, we
	// could avoid the os/signal package. However that requires
	// implementing our own sigtramp, so to cut down on gnarly code,
	// we reuse this.
	signal.Notify(make(chan os.Signal), intrSig)

	const (
		SA_ONSTACK  = 0x08000000
		SA_RESTORER = 0x4000000
		SA_SIGINFO  = 0x4
	)
	var sa sigactiont
	sa.sa_flags = SA_SIGINFO | SA_ONSTACK | SA_RESTORER
	sa.sa_mask = ^uint64(0)
	sa.sa_handler = funcPC(sigtramp)
	if runtime.GOARCH == "386" || runtime.GOARCH == "amd64" {
		sa.sa_restorer = funcPC(sigreturn)
	}
	errno := sigaction(uintptr(intrSig), &sa, nil, unsafe.Sizeof(sa.sa_mask))
	if errno != 0 {
		panic(syscall.Errno(errno).Error())
	}
}
