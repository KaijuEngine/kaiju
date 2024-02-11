/*****************************************************************************/
/* controller.go                                                             */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

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
	if err == nil && c.devices[id].id < 0 {
		c.devices[id].id = id
	}
}

func (c *Controller) Disconnected(id int) {
	err := validateJoystick(id)
	if err == nil && c.devices[id].id >= 0 {
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
	if c.devices[id].buttons[button] == controllerButtonStateIdle {
		c.devices[id].buttons[button] = controllerButtonStateDown
	}
}

func (c *Controller) SetButtonUp(id, button int) {
	if c.devices[id].buttons[button] != controllerButtonStateIdle {
		c.devices[id].buttons[button] = controllerButtonStateUp
	}
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
