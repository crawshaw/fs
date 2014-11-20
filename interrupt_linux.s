#include "textflag.h"

// sigreturn is borrowed from the runtime.
TEXT ·sigreturn(SB),NOSPLIT,$0
	JMP runtime·sigreturn(SB)
