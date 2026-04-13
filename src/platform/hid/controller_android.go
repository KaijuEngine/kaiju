package hid

func ToControllerButton(nativeButton int) (ControllerButton, error) {
	return ControllerButton(nativeButton), nil
}
