/******************************************************************************/
/* tween.go                                                                   */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package tweening

import "github.com/KaijuEngine/kaiju/platform/profiler/tracing"

type Tween struct {
	val         *float32
	delayTime   float64
	totalUpdate float64
	time        float64
	initial     float32
	target      float32
	easing      func(t float32) float32
	onChange    func(val float32)
	onDone      func()
	scale       float32
}

func (t *Tween) update(deltaTime float64) bool {
	defer tracing.NewRegion("Tween.update").End()
	if t.delayTime > 0.0 {
		t.delayTime -= deltaTime
		if t.delayTime <= 0.0 {
			t.initial = *t.val
		}
		return false
	}
	*t.val = t.calculate()
	if t.onChange != nil {
		t.onChange(*t.val)
	}
	t.totalUpdate += deltaTime
	if t.totalUpdate >= t.time {
		*t.val = t.target
		t.totalUpdate = 0.0
		return true
	}
	return false
}

func (t *Tween) calculate() float32 {
	time := float32(1 - ((t.time - t.totalUpdate) / t.time))
	return t.initial + (t.target-t.initial)*t.easing(time)
}

func (t *Tween) stop(jumpToEnd, skipCallback bool) {
	defer tracing.NewRegion("Tween.stop").End()
	if jumpToEnd {
		*t.val = t.target
		t.totalUpdate = 0.0
	}
	if !skipCallback {
		if t.onDone != nil {
			t.onDone()
		}
	}
}
