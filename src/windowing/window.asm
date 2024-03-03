//go:build amd64

#include "textflag.h"

#define WIN_SHADOW_CALL(fn) \
	MOVQ	SP, BX          \
	ANDQ	$~8, SP         \ // alignment
	MOVQ	BX, 8(SP)       \
	ADJSP	$32             \
	CALL fn(SB)             \
	NOPL 0(AX)              \ // NOP
	ADJSP	$-32            \
	MOVQ	8(SP), BX       \
	MOVQ	BX, SP

// func cWindowPollController(handle unsafe.Pointer) uint32
TEXT   ·cWindowPollController(SB),NOSPLIT,$0
	MOVQ handle+0(FP), CX
	CALL window_poll_controller(SB)
	MOVQ AX, ret+8(FP)
	RET

// func cWindowPoll(handle unsafe.Pointer) uint32
TEXT   ·cWindowPoll(SB),NOSPLIT,$0-16
	MOVQ handle+0(FP), CX
	CALL window_poll(SB)
	MOVQ AX, ret+8(FP)
	RET

// func cShowWindow(handle unsafe.Pointer)
TEXT   ·cShowWindow(SB),DUPOK|NOSPLIT,$0-8
	MOVQ (TLS), CX
	CMPQ SP, 16(CX)
	JGT 4(PC)
	NOPL 0(AX)
	CALL runtime·morestack_noctxt(SB)
	JMP -6(PC)
	MOVQ handle+0(FP), CX
	PUSHQ BP
	MOVQ SP, BP
	CALL window_show(SB)
	//WIN_SHADOW_CALL(window_show)
	POPQ BP
	RET

// func cWindowCursorStandard(handle unsafe.Pointer)
TEXT   ·cWindowCursorStandard(SB),NOSPLIT,$0-8
	//MOVQ (TLS), CX
	//CMPQ SP, 16(CX)
	//CMPQ SP, 16(CX)
	//JLE 3(PC)
	//NOPL 0(AX)
	//CALL runtime·morestack_noctxt(SB)
	MOVQ handle+0(FP), CX
	CALL window_cursor_standard(SB)
	//CALL window_cursor_standard)
	//MOVQ window_cursor_standard, AX
	RET

// func cWindowCursorIbeam(handle unsafe.Pointer)
TEXT   ·cWindowCursorIbeam(SB),NOSPLIT,$0-8
	MOVQ handle+0(FP), CX
	CALL window_cursor_ibeam(SB)
	RET

// func cWindowFocus(handle unsafe.Pointer)
TEXT   ·cWindowFocus(SB),NOSPLIT,$0-8
	MOVQ handle+0(FP), CX
	CALL window_focus(SB)
	RET

// func cRemoveBorder(handle unsafe.Pointer)
TEXT   ·cRemoveBorder(SB),NOSPLIT,$0-8
	MOVQ handle+0(FP), CX
	CALL remove_border(SB)
	RET

// func cAddBorder(handle unsafe.Pointer)
TEXT   ·cAddBorder(SB),NOSPLIT,$0-8
	MOVQ handle+0(FP), CX
	CALL add_border(SB)
	RET

// func cSetWindowPosition(handle unsafe.Pointer, x, y int32)
TEXT   ·cSetWindowPosition(SB),NOSPLIT,$0-24
	MOVQ x+0(FP), CX
	MOVQ y+8(FP), DX
	MOVQ y+16(FP), R8
	CALL set_window_position(SB)
	RET

// func cSetWindowSize(handle unsafe.Pointer, width, height int32)
TEXT   ·cSetWindowSize(SB),NOSPLIT,$0-24
	MOVQ x+0(FP), CX
	MOVQ y+8(FP), DX
	MOVQ y+16(FP), R8
	CALL set_window_size(SB)
	RET

// func cWindowPosition(handle unsafe.Pointer, x, y *int32)
TEXT   ·cWindowPosition(SB),NOSPLIT,$0-24
	MOVQ x+0(FP), CX
	MOVQ y+8(FP), DX
	MOVQ y+16(FP), R8
	CALL window_position(SB)
	RET
