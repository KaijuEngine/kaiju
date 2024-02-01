package klib

func MM2PX[T Number](pixels, mm, targetMM T) T {
	return targetMM * (pixels / mm)
}
