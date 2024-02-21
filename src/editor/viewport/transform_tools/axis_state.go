package transform_tools

type AxisState int

const (
	AxisStateNone AxisState = iota
	AxisStateX
	AxisStateY
	AxisStateZ
)

func (a *AxisState) Toggle(axis AxisState) {
	if *a == axis {
		*a = AxisStateNone
	} else {
		*a = axis
	}
}
