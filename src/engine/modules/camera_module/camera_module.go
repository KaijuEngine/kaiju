package camera_module

import (
	"kaiju/cameras"
	"kaiju/engine"
)

type CameraModule struct {
	entity   *engine.Entity
	host     *engine.Host
	updateId int
	isBound  bool
	camera   cameras.Camera
}

type CameraModuleBinding struct {
	FOV       float32 `clamp:"60,45,120"` //default,min,max
	NearPlane float32 `default:"0.01"`
	FarPlane  float32 `default:"500.0"`
	IsPrimary bool
}

func (c *CameraModuleBinding) Init(e *engine.Entity, host *engine.Host) {
	cm := &CameraModule{}
	e.AddNamedData("CameraModule", cm)
	cm.entity = e
	cm.host = host
	cm.camera = cameras.NewStandardCamera(float32(host.Window.Width()),
		float32(host.Window.Height()), e.Transform.Position())
	cm.updateId = host.Updater.AddUpdate(cm.update)
	cm.entity.OnDestroy.Add(func() { host.Updater.RemoveUpdate(cm.updateId) })
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
