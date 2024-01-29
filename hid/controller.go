package hid

import "errors"

// Based off XBOX controller
const (
	ControllerMaxDevices = 8
)

const (
	ControllerButtonA = iota
	ControllerButtonB
	ControllerButtonX
	ControllerButtonY
	ControllerButtonBack
	ControllerButtonPause
	ControllerButtonStickLeft
	ControllerButtonStickRight
	ControllerButtonBumperLeft
	ControllerButtonBumperRight
	ControllerButtonDPadUp
	ControllerButtonDPadDown
	ControllerButtonDPadLeft
	ControllerButtonDPadRight
	ControllerButtonMax
)

const (
	ControllerAxisLeftVertical = iota
	ControllerAxisLeftHorizontal
	ControllerAxisRightVertical
	ControllerAxisRightHorizontal
	ControllerAxisLeftTrigger
	ControllerAxisRightTrigger
	ControllerAxisMax
)

const (
	controllerButtonStateIdle = iota
	controllerButtonStateDown
	controllerButtonStateHeld
	controllerButtonStateUp
)

type ControllerDevice struct {
	buttons [ControllerButtonMax]int
	axis    [ControllerAxisMax]float32
	id      int
}

type Controller struct {
	devices [ControllerMaxDevices]ControllerDevice
}

func validateJoystick(id int) error {
	if id < 0 || id >= ControllerMaxDevices {
		return errors.New("invalid joystick id")
	} else {
		return nil
	}
}

func NewController() Controller {
	c := Controller{}
	for i := 0; i < ControllerMaxDevices; i++ {
		c.devices[i].id = -1
	}
	return c
}

func (c *Controller) Available(id int) bool {
	err := validateJoystick(id)
	return err == nil && c.devices[id].id >= 0
}

func (c *Controller) Connected(id int) {
	err := validateJoystick(id)
	if err == nil {
		c.devices[id].id = id
	}
}

func (c *Controller) Disconnected(id int) {
	err := validateJoystick(id)
	if err == nil {
		c.devices[id].id = -1
	}
}

func (device *ControllerDevice) EndUpdate() {
	if device.id >= 0 {
		for i := 0; i < ControllerButtonMax; i++ {
			if device.buttons[i] == controllerButtonStateDown {
				device.buttons[i] = controllerButtonStateHeld
			} else if device.buttons[i] == controllerButtonStateUp {
				device.buttons[i] = controllerButtonStateIdle
			}
		}
	}
}

func (c *Controller) EndUpdate() {
	for i := 0; i < ControllerMaxDevices; i++ {
		c.devices[i].EndUpdate()
	}
}

func (c *Controller) SetButtonDown(id, button int) {
	c.devices[id].buttons[button] = controllerButtonStateDown
}

func (c *Controller) SetButtonUp(id, button int) {
	c.devices[id].buttons[button] = controllerButtonStateUp
}

func (c *Controller) SetAxis(id, stick int, axis float32) {
	c.devices[id].axis[stick] = axis
}

func (c *Controller) Axis(id, stick int) float32 {
	return c.devices[id].axis[stick]
}

func (c *Controller) IsButtonUp(id, button int) bool {
	return c.devices[id].buttons[button] == controllerButtonStateUp
}

func (c *Controller) IsButtonDown(id, button int) bool {
	return c.devices[id].buttons[button] == controllerButtonStateDown
}

func (c *Controller) IsButtonHeld(id, button int) bool {
	return c.devices[id].buttons[button] == controllerButtonStateHeld
}
