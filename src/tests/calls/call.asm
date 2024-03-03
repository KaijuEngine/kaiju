#include "textflag.h"

#define WIN_SHADOW_CALL(fn) \
	PUSHQ   SP              \
	ANDQ	$~8, SP         \ // alignment
	ADJSP	$32             \
	CALL    fn(SB)          \
	ADJSP	$-32            \
	POPQ    SP

// func CAdd(a, b int) int
TEXT   ·CAdd(SB),NOSPLIT,$0-24
	MOVQ    a+0(FP), CX
	MOVQ    b+8(FP), DX
	WIN_SHADOW_CALL(add)
	MOVQ AX, ret+16(FP)
	RET
