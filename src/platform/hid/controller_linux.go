//go:build !android

package hid

import "errors"

func ToControllerButton(nativeButton int) (ControllerButton, error) {
	switch nativeButton {
	case 0:
		return ControllerButtonA, nil
	case 1:
		return ControllerButtonB, nil
	case 2:
		return ControllerButtonX, nil
	case 3:
		return ControllerButtonY, nil
	case 4:
		return ControllerButtonLeftBumper, nil
	case 5:
		return ControllerButtonRightBumper, nil
	case 6:
		return ControllerButtonSelect, nil
	case 7:
		return ControllerButtonStart, nil
	case 8:
		return ControllerButtonEx1, nil
	case 9:
		return ControllerButtonLeftStick, nil
	case 10:
		return ControllerButtonRightStick, nil
	case 11:
		return ControllerButtonEx2, nil
	case 12:
		return ControllerButtonUp, nil
	case 13:
		return ControllerButtonDown, nil
	case 14:
		return ControllerButtonLeft, nil
	case 15:
		return ControllerButtonRight, nil
	default:
		return 0, errors.New("invalid controller button")
	}
}
