package windowing

type CursorKind int

const (
	CursorKindAuto CursorKind = iota
	CursorKindDefault
	CursorKindNone
	CursorKindContextMenu
	CursorKindText
	CursorKindVerticalText
	CursorKindPointer
	CursorKindHelp
	CursorKindWait
	CursorKindProgress
	CursorKindCrosshair
	CursorKindCell
	CursorKindAlias
	CursorKindCopy
	CursorKindMove
	CursorKindNoDrop
	CursorKindNotAllowed
	CursorKindGrab
	CursorKindGrabbing
	CursorKindResizeN
	CursorKindResizeE
	CursorKindResizeS
	CursorKindResizeW
	CursorKindResizeNE
	CursorKindResizeNW
	CursorKindResizeSE
	CursorKindResizeSW
	CursorKindResizeNS
	CursorKindResizeEW
	CursorKindResizeNWSE
	CursorKindResizeNESW
	CursorKindResizeCol
	CursorKindResizeRow
	CursorKindResizeAll
	CursorKindZoomIn
	CursorKindZoomOut
)
