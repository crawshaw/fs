package fs

type sigactiont struct {
	sa_handler  uintptr
	sa_flags    uint32
	sa_restorer uintptr
	sa_mask     uint64
}
