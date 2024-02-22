package transform_tools

type ToolState = uint8

const (
	ToolStateNone ToolState = iota
	ToolStateMove
	ToolStateRotate
	ToolStateScale
)
