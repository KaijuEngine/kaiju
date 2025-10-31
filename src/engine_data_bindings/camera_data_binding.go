package engine_data_bindings

import (
	"kaiju/engine"
	"kaiju/engine/cameras"
)

const CameraDataBindingKey = "kaiju.CameraDataBinding"

func init() {
	engine.RegisterEntityData(CameraDataBindingKey, CameraDataBinding{})
}

type CameraDataBinding struct {
	FOV       float32 `clamp:"60,45,120"` //default,min,max
	NearPlane float32 `default:"0.01"`
	FarPlane  float32 `default:"500.0"`
	IsPrimary bool
	// TODO:  Work out the orthographic camera stuff
}

type CameraModule struct {
	entity   *engine.Entity
	host     *engine.Host
	updateId engine.UpdateId
	camera   cameras.Camera
}

func NewCameraDataBinding() CameraDataBinding {
	return CameraDataBinding{
		FOV:       60,
		NearPlane: 0.01,
		FarPlane:  500,
		IsPrimary: false,
	}
}

func (c CameraDataBinding) Init(e *engine.Entity, host *engine.Host) {
	cm := &CameraModule{}
	e.AddNamedData("CameraModule", cm)
	cm.entity = e
	cm.host = host
	w := float32(host.Window.Width())
	h := float32(host.Window.Height())
	cm.camera = cameras.NewStandardCamera(w, h, w, h, e.Transform.Position())
	cm.updateId = host.Updater.AddUpdate(cm.update)
	cm.entity.OnDestroy.Add(func() { host.Updater.RemoveUpdate(&cm.updateId) })
	if c.IsPrimary {
		cm.SetAsActive()
	}
}

func (c *CameraModule) SetAsActive() {
	c.host.Camera = c.camera
}

func (c *CameraModule) update(deltaTime float64) {
	if !c.entity.IsActive() || c.camera != c.host.Camera {
		return
	}
	t := &c.entity.Transform
	lookAt := t.Position().Add(t.Forward())
	c.host.Camera.SetPositionAndLookAt(t.Position(), lookAt)
}
