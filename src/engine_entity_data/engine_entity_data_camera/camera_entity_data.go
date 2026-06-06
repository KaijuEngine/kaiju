/******************************************************************************/
/* camera_data_binding.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_camera

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/encoding/pod"
)

var bindingKey = ""

type CameraType int

const (
	CameraTypePerspective CameraType = iota
	CameraTypeOrthographic
	CameraTypeTurntable
)

func init() {
	engine.RegisterEntityData(CameraEntityData{})
}

func BindingKey() string {
	if bindingKey == "" {
		bindingKey = pod.QualifiedNameForLayout(CameraEntityData{})
	}
	return bindingKey
}

type CameraEntityData struct {
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

func NewCameraDataBinding() CameraEntityData {
	return CameraEntityData{
		FOV:          60,
		NearPlane:    0.01,
		FarPlane:     500,
		IsMainCamera: false,
	}
}

func (c CameraEntityData) Init(e *engine.Entity, host *engine.Host) {
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
	lookAt := t.Position().Subtract(t.Forward())
	p.Camera.SetPositionAndLookAt(t.Position(), lookAt)
}
