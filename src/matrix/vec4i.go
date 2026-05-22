/******************************************************************************/
/* vec4i.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

type Vec4i [4]int32

func (v Vec4i) X() int32      { return v[0] }
func (v Vec4i) Y() int32      { return v[1] }
func (v Vec4i) Z() int32      { return v[2] }
func (v Vec4i) W() int32      { return v[3] }
func (v Vec4i) Width() int32  { return v[2] }
func (v Vec4i) Height() int32 { return v[3] }
