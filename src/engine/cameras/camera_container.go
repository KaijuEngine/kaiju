/******************************************************************************/
/* camera_container.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package cameras

import "kaijuengine.com/engine/graviton"

type Container struct {
	Camera Camera
}

func NewContainer(camera Camera) Container      { return Container{camera} }
func (c *Container) ChangeCamera(camera Camera) { c.Camera = camera }
func (c *Container) IsValid() bool              { return c.Camera != nil }
func (c *Container) Equal(other Camera) bool    { return c.Camera == other }
func (c *Container) ViewChanged() bool          { return c.Camera.IsDirty() }

func (c *Container) IsInView(box graviton.AABB) bool {
	return box.IntersectsFrustum(c.Camera.Frustum())
}
