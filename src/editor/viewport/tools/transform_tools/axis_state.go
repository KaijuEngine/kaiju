package transform_tools

type AxisState uint8

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
