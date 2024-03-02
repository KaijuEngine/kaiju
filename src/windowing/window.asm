//go:build amd64

#include "textflag.h"

#define WIN_SHADOW_CALL(fn) \
	MOVQ	SP, BX          \
	ANDQ	$~15, SP	    \ // alignment
	MOVQ	BX, 8(SP)       \
	ADJSP	$32             \
	CALL fn(SB)             \
	BYTE	$0x90	        \ // NOP
	ADJSP	$-32            \
	MOVQ	8(SP), BX       \
	MOVQ	BX, SP

// func cWindowPollController(handle unsafe.Pointer) uint32
TEXT   ·cWindowPollController(SB),NOSPLIT,$0
	MOVQ handle+0(FP), CX
	WIN_SHADOW_CALL(window_poll_controller)
	MOVQ AX, ret+8(FP)
	RET

// func cWindowPoll(handle unsafe.Pointer) uint32
TEXT   ·cWindowPoll(SB),NOSPLIT,$0-16
	MOVQ handle+0(FP), CX
	WIN_SHADOW_CALL(window_poll)
	MOVQ AX, ret+8(FP)
	RET

// func cWindowCursorStandard(handle unsafe.Pointer)
TEXT   ·cWindowCursorStandard(SB),NOSPLIT,$0-8
	MOVQ handle+0(FP), CX
	WIN_SHADOW_CALL(window_cursor_standard)
	RET

// func cWindowCursorIbeam(handle unsafe.Pointer)
TEXT   ·cWindowCursorIbeam(SB),NOSPLIT,$0-8
	MOVQ handle+0(FP), CX
	WIN_SHADOW_CALL(window_cursor_ibeam)
	RET

// func cWindowFocus(handle unsafe.Pointer)
TEXT   ·cWindowFocus(SB),NOSPLIT,$0-8
	MOVQ handle+0(FP), CX
	WIN_SHADOW_CALL(window_focus)
	RET

// func cRemoveBorder(handle unsafe.Pointer)
TEXT   ·cRemoveBorder(SB),NOSPLIT,$0-8
	MOVQ handle+0(FP), CX
	WIN_SHADOW_CALL(remove_border)
	RET

// func cAddBorder(handle unsafe.Pointer)
TEXT   ·cAddBorder(SB),NOSPLIT,$0-8
	MOVQ handle+0(FP), CX
	WIN_SHADOW_CALL(add_border)
	RET

// func cSetWindowPosition(handle unsafe.Pointer, x, y int32)
TEXT   ·cSetWindowPosition(SB),NOSPLIT,$0-24
	MOVQ x+0(FP), CX
	MOVQ y+8(FP), DX
	MOVQ y+16(FP), R8
	WIN_SHADOW_CALL(set_window_position)
	RET

// func cSetWindowSize(handle unsafe.Pointer, width, height int32)
TEXT   ·cSetWindowSize(SB),NOSPLIT,$0-24
	MOVQ x+0(FP), CX
	MOVQ y+8(FP), DX
	MOVQ y+16(FP), R8
	WIN_SHADOW_CALL(set_window_size)
	RET

// func cWindowPosition(handle unsafe.Pointer, x, y *int32)
TEXT   ·cWindowPosition(SB),NOSPLIT,$0-24
	MOVQ x+0(FP), CX
	MOVQ y+8(FP), DX
	MOVQ y+16(FP), R8
	WIN_SHADOW_CALL(window_position)
	RET
