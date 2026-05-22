/******************************************************************************/
/* vec3i.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

type Vec3i [3]int32

func (v Vec3i) X() int32      { return v[0] }
func (v Vec3i) Y() int32      { return v[1] }
func (v Vec3i) Z() int32      { return v[2] }
func (v Vec3i) Width() int32  { return v[0] }
func (v Vec3i) Height() int32 { return v[1] }
func (v Vec3i) Depth() int32  { return v[2] }

func (v *Vec3i) SetX(x int32)           { v[0] = x }
func (v *Vec3i) SetY(y int32)           { v[1] = y }
func (v *Vec3i) SetZ(z int32)           { v[2] = z }
func (v *Vec3i) SetWidth(width int32)   { v[0] = width }
func (v *Vec3i) SetHeight(height int32) { v[1] = height }
func (v *Vec3i) SetDepth(depth int32)   { v[2] = depth }
