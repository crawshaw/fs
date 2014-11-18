package fs

import "unsafe"

type sigactiont struct {
	__sigaction_u [8]byte
	sa_tramp      unsafe.Pointer
	sa_mask       uint32
	sa_flags      int32
}
