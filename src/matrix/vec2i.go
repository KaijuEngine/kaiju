/******************************************************************************/
/* vec2i.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

type Vec2i [2]int32

func (v Vec2i) X() int32      { return v[0] }
func (v Vec2i) Y() int32      { return v[1] }
func (v Vec2i) Width() int32  { return v[0] }
func (v Vec2i) Height() int32 { return v[1] }

func (v *Vec2i) SetX(x int32)           { v[0] = x }
func (v *Vec2i) SetY(y int32)           { v[1] = y }
func (v *Vec2i) SetWidth(width int32)   { v[0] = width }
func (v *Vec2i) SetHeight(height int32) { v[1] = height }
