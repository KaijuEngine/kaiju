/*****************************************************************************/
/* camera.go                                                                 */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
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

package cameras

import (
	"kaiju/collision"
	"kaiju/matrix"
)

type Camera interface {
	SetPosition(position matrix.Vec3)
	SetFOV(fov float32)
	SetNearPlane(near float32)
	SetFarPlane(far float32)
	SetWidth(width float32)
	SetHeight(height float32)
	ViewportChanged(width, height float32)
	SetProperties(fov, nearPlane, farPlane, width, height float32)
	SetYaw(yaw float32)
	SetPitch(pitch float32)
	SetYawAndPitch(yaw, pitch float32)
	Forward() matrix.Vec3
	Right() matrix.Vec3
	Up() matrix.Vec3
	SetLookAt(position matrix.Vec3)
	LookAt(point, up matrix.Vec3)
	SetPositionAndLookAt(position, lookAt matrix.Vec3)
	Raycast(screenPos matrix.Vec2) collision.Ray
	TryPlaneHit(screenPos matrix.Vec2, planePos, planeNml matrix.Vec3) (hit matrix.Vec3, success bool)
	ForwardPlaneHit(screenPos matrix.Vec2, planePos matrix.Vec3) (matrix.Vec3, bool)
	Position() matrix.Vec3
	Width() float32
	Height() float32
	View() matrix.Mat4
	Projection() matrix.Mat4
	Center() matrix.Vec3
	Yaw() float32
	Pitch() float32
	NearPlane() float32
	FarPlane() float32
	Zoom() float32
}
