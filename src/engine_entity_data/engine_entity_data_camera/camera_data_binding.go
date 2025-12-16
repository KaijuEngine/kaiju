/******************************************************************************/
/* camera_data_binding.go                                                     */
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

package engine_entity_data_camera

import (
	"kaiju/engine"
	"kaiju/engine/cameras"
)

const BindingKey = "kaiju.CameraDataBinding"

type CameraType int

const (
	CameraTypePerspective CameraType = iota
	CameraTypeOrthographic
	CameraTypeTurntable
)

func init() {
	engine.RegisterEntityData(BindingKey, CameraDataBinding{})
}

type CameraDataBinding struct {
	Width        float32 `default:"0" tip:"0 = viewport width"`
	Height       float32 `default:"0" tip:"0 = viewport height"`
	FOV          float32 `clamp:"60,45,120"` //default,min,max
	NearPlane    float32 `default:"0.01"`
	FarPlane     float32 `default:"500.0"`
	Type         CameraType
	IsMainCamera bool
}

type CameraModule struct {
	entity   *engine.Entity
	host     *engine.Host
	updateId engine.UpdateId
	camera   cameras.Camera
}

func NewCameraDataBinding() CameraDataBinding {
	return CameraDataBinding{
		FOV:          60,
		NearPlane:    0.01,
		FarPlane:     500,
		IsMainCamera: false,
	}
}

func (c CameraDataBinding) Init(e *engine.Entity, host *engine.Host) {
	cm := &CameraModule{}
	e.AddNamedData("CameraModule", cm)
	cm.entity = e
	cm.host = host
	w := c.Width
	h := c.Height
	if w <= 0 {
		w = float32(host.Window.Width())
	}
	if h <= 0 {
		h = float32(host.Window.Height())
	}
	switch c.Type {
	case CameraTypeOrthographic:
		cm.camera = cameras.NewStandardCameraOrthographic(w, h, w, h, e.Transform.Position())
	case CameraTypeTurntable:
		cm.camera = cameras.ToTurntable(cameras.NewStandardCamera(w, h, w, h, e.Transform.Position()))
	case CameraTypePerspective:
		fallthrough
	default:
		cm.camera = cameras.NewStandardCamera(w, h, w, h, e.Transform.Position())
	}
	cm.camera.SetProperties(c.FOV, c.NearPlane, c.FarPlane, w, h)
	cm.updateId = host.Updater.AddUpdate(cm.update)
	cm.entity.OnDestroy.Add(func() { host.Updater.RemoveUpdate(&cm.updateId) })
	if c.IsMainCamera {
		cm.SetAsActive()
	}
}

func (c *CameraModule) SetAsActive() {
	c.host.Cameras.Primary.ChangeCamera(c.camera)
}

func (c *CameraModule) update(deltaTime float64) {
	p := &c.host.Cameras.Primary
	if !c.entity.IsActive() || !p.Equal(c.camera) {
		return
	}
	t := &c.entity.Transform
	lookAt := t.Position().Add(t.Forward())
	p.Camera.SetPositionAndLookAt(t.Position(), lookAt)
}
