#include "textflag.h"

// func CAdd(a, b int) int
TEXT   ·CAdd(SB),NOSPLIT,$0-24
	MOVQ a+0(FP), CX
	MOVQ b+8(FP), DX
	CALL add(SB)
	MOVQ AX, ret+16(FP)
	RET
