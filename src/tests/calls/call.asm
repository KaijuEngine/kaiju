#include "textflag.h"

#define WIN_SHADOW_CALL(fn) \
	ANDQ	$~8, SP         \ // alignment
	ADJSP	$32             \
	CALL    fn(SB)          \
	ADJSP	$-32

// func CAdd(stack *byte, a, b int) int
TEXT   ·CAdd(SB),NOSPLIT,$0-32
	MOVQ    a+8(FP), CX
	MOVQ    b+16(FP), DX
	PUSHQ   (SP)
	MOVQ	stack+0(FP), BX
	MOVQ	BX, (SP)
	WIN_SHADOW_CALL(add)
	MOVQ AX, ret+24(FP)
	POPQ    (SP)
	RET
