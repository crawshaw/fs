#include "textflag.h"

// sigtramp is borrowed from the runtime.
TEXT ·sigtramp(SB),NOSPLIT,$0
	JMP runtime·sigtramp(SB)
