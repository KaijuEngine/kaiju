/******************************************************************************/
/* controller.go                                                              */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package hid

import "errors"

// Based off XBOX controller
const (
	ControllerMaxDevices = 8
)

const (
	ControllerButtonUp = iota
	ControllerButtonDown
	ControllerButtonLeft
	ControllerButtonRight
	ControllerButtonStart
	ControllerButtonSelect
	ControllerButtonLeftStick
	ControllerButtonRightStick
	ControllerButtonLeftBumper
	ControllerButtonRightBumper
	ControllerButtonEx1 // TODO:  Name this correctly
	ControllerButtonEx2 // TODO:  Name this correctly
	ControllerButtonA
	ControllerButtonB
	ControllerButtonX
	ControllerButtonY
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

// NewController creates a new controller. This is called
// automatically by the system and should not be called by the end-developer
func NewController() Controller {
	c := Controller{}
	for i := 0; i < ControllerMaxDevices; i++ {
		c.devices[i].id = -1
	}
	return c
}

// Available returns true if the controller is available. This is called
// automatically by the system and should not be called by the end-developer
func (c *Controller) Available(id int) bool {
	err := validateJoystick(id)
	return err == nil && c.devices[id].id >= 0
}

// Connected returns true if the controller is connected. This is called
// automatically by the system and should not be called by the end-developer
func (c *Controller) Connected(id int) {
	err := validateJoystick(id)
	if err == nil && c.devices[id].id < 0 {
		c.devices[id].id = id
	}
}

// Disconnected returns true if the controller is disconnected. This is called
// automatically by the system and should not be called by the end-developer
func (c *Controller) Disconnected(id int) {
	err := validateJoystick(id)
	if err == nil && c.devices[id].id >= 0 {
		c.devices[id].id = -1
	}
}

func (device *ControllerDevice) endUpdate() {
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

// EndUpdate is called at the end of the frame. It updates the state of each
// controller for the next frame. This is called automatically by the system
// and should not be called by the end-developer
func (c *Controller) EndUpdate() {
	for i := 0; i < ControllerMaxDevices; i++ {
		c.devices[i].endUpdate()
	}
}

// SetButtonDown sets the button down on the given controller. This is called
// automatically by the system and should not be called by the end-developer
func (c *Controller) SetButtonDown(id, button int) {
	if c.devices[id].buttons[button] == controllerButtonStateIdle {
		c.devices[id].buttons[button] = controllerButtonStateDown
	}
}

// SetButtonUp sets the button up on the given controller. This is called
// automatically by the system and should not be called by the end-developer
func (c *Controller) SetButtonUp(id, button int) {
	if c.devices[id].buttons[button] != controllerButtonStateIdle {
		c.devices[id].buttons[button] = controllerButtonStateUp
	}
}

// SetAxis sets the axis on the given controller. This is called
// automatically by the system and should not be called by the end-developer
func (c *Controller) SetAxis(id, stick int, axis float32) {
	c.devices[id].axis[stick] = axis
}

// Axis returns the axis value for the given controller and stick
func (c *Controller) Axis(id, stick int) float32 {
	return c.devices[id].axis[stick]
}

// IsButtonUp returns true if the button is up
func (c *Controller) IsButtonUp(id, button int) bool {
	return c.devices[id].buttons[button] == controllerButtonStateUp
}

// IsButtonDown returns true if the button is down
func (c *Controller) IsButtonDown(id, button int) bool {
	return c.devices[id].buttons[button] == controllerButtonStateDown
}

// IsButtonHeld returns true if the button is held
func (c *Controller) IsButtonHeld(id, button int) bool {
	return c.devices[id].buttons[button] == controllerButtonStateHeld
}

// Reset will completely wipe the state of all controllers
func (c *Controller) Reset() {
	for i := 0; i < ControllerMaxDevices; i++ {
		for j := 0; j < ControllerButtonMax; j++ {
			if c.devices[i].buttons[j] == controllerButtonStateDown ||
				c.devices[i].buttons[j] == controllerButtonStateHeld {
				c.devices[i].buttons[j] = controllerButtonStateUp
			}
		}
		c.devices[i].axis[ControllerAxisLeftVertical] = 0.0
		c.devices[i].axis[ControllerAxisLeftHorizontal] = 0.0
		c.devices[i].axis[ControllerAxisRightVertical] = 0.0
		c.devices[i].axis[ControllerAxisRightHorizontal] = 0.0
		c.devices[i].axis[ControllerAxisLeftTrigger] = 0.0
		c.devices[i].axis[ControllerAxisRightTrigger] = 0.0
	}
}
