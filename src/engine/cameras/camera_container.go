package cameras

import "kaiju/engine/collision"

type Container struct {
	Camera Camera
}

func NewContainer(camera Camera) Container      { return Container{camera} }
func (c *Container) ChangeCamera(camera Camera) { c.Camera = camera }
func (c *Container) IsValid() bool              { return c.Camera != nil }
func (c *Container) Equal(other Camera) bool    { return c.Camera == other }
func (c *Container) ViewChanged() bool          { return c.Camera.IsDirty() }

func (c *Container) IsInView(box collision.AABB) bool {
	return box.InFrustum(c.Camera.Frustum())
}
