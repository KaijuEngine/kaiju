/******************************************************************************/
/* tweener.go                                                                 */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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

import (
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
)

type Easing int

const (
	EasingLinear = Easing(iota)
	EasingIn
	EasingOut
	EasingInAndOut
	EasingInSine
	EasingOutSine
	EasingInAndOutSine
	EasingInQuad
	EasingOutQuad
	EasingInAndOutQuad
	EasingInCubic
	EasingOutCubic
	EasingInAndOutCubic
	EasingInQuart
	EasingOutQuart
	EasingInAndOutQuart
	EasingInQuint
	EasingOutQuint
	EasingInAndOutQuint
	EasingInExpo
	EasingOutExpo
	EasingInAndOutExpo
	EasingInCirc
	EasingOutCirc
	EasingInAndOutCirc
	EasingInBack
	EasingOutBack
	EasingInAndOutBack
	EasingInElastic
	EasingOutElastic
	EasingInAndOutElastic
	EasingInBounce
	EasingOutBounce
	EasingInAndOutBounce
)

var (
	tweens = make([]Tween, 0, 16)
)

func DoTween(val *float32, target float32, time float64, easing Easing) {
	defer tracing.NewRegion("Tweener.DoTween").End()
	DoTweenExt(val, target, time, easing, nil, nil)
}

func DoTweenExt(val *float32, target float32, time float64, easing Easing,
	onChange func(val float32), onDone func()) {
	defer tracing.NewRegion("Tweener.DoTweenExt").End()
	tween := Tween{
		val:         val,
		initial:     *val,
		target:      target,
		time:        time,
		onChange:    onChange,
		onDone:      onDone,
		scale:       (target - *val) / max(float32(time), 0.00001),
		totalUpdate: 0,
	}
	switch easing {
	case EasingLinear:
		tween.easing = easeLinear
	case EasingIn:
		tween.easing = easeInQuad
	case EasingOut:
		tween.easing = easeOutQuad
	case EasingInAndOut:
		tween.easing = easeInOutQuad
	case EasingInSine:
		tween.easing = easeInSine
	case EasingOutSine:
		tween.easing = easeOutSine
	case EasingInAndOutSine:
		tween.easing = easeInOutSine
	case EasingInQuad:
		tween.easing = easeInQuad
	case EasingOutQuad:
		tween.easing = easeOutQuad
	case EasingInAndOutQuad:
		tween.easing = easeInOutQuad
	case EasingInCubic:
		tween.easing = easeInCubic
	case EasingOutCubic:
		tween.easing = easeOutCubic
	case EasingInAndOutCubic:
		tween.easing = easeInOutCubic
	case EasingInQuart:
		tween.easing = easeInQuart
	case EasingOutQuart:
		tween.easing = easeOutQuart
	case EasingInAndOutQuart:
		tween.easing = easeInOutQuart
	case EasingInQuint:
		tween.easing = easeInQuint
	case EasingOutQuint:
		tween.easing = easeOutQuint
	case EasingInAndOutQuint:
		tween.easing = easeInOutQuint
	case EasingInExpo:
		tween.easing = easeInExpo
	case EasingOutExpo:
		tween.easing = easeOutExpo
	case EasingInAndOutExpo:
		tween.easing = easeInOutExpo
	case EasingInCirc:
		tween.easing = easeInCirc
	case EasingOutCirc:
		tween.easing = easeOutCirc
	case EasingInAndOutCirc:
		tween.easing = easeInOutCirc
	case EasingInBack:
		tween.easing = easeInBack
	case EasingOutBack:
		tween.easing = easeOutBack
	case EasingInAndOutBack:
		tween.easing = easeInOutBack
	case EasingInElastic:
		tween.easing = easeInElastic
	case EasingOutElastic:
		tween.easing = easeOutElastic
	case EasingInAndOutElastic:
		tween.easing = easeInOutElastic
	case EasingInBounce:
		tween.easing = easeInBounce
	case EasingOutBounce:
		tween.easing = easeOutBounce
	case EasingInAndOutBounce:
		tween.easing = easeInOutBounce
	}
	// Stop the tweener for the same value if one exists
	Stop(val, true, false)
	tweens = append(tweens, tween)
}

func Update(deltaTime float64) {
	defer tracing.NewRegion("Tweener.Update").End()
	for i := 0; i < len(tweens); i++ {
		tween := &tweens[i]
		if tween.update(deltaTime) {
			// TODO:  Put this tween into the cached list of tweens that are
			// being removed that way we can do their callback without worrying
			// about it being added to this list and then being freed
			if tween.onChange != nil {
				tween.onChange(*tween.val)
			}
			if tween.onDone != nil {
				tween.onDone()
			}
			tweens = klib.RemoveUnordered(tweens, i)
			i--
		}
	}
}

func Clear() {
	defer tracing.NewRegion("Tweener.Clear").End()
	for i := range tweens {
		t := &tweens[i]
		t.stop(true, false)
	}
	tweens = klib.WipeSlice(tweens)
}

func Stop(val *float32, jumpToEnd, skipCallback bool) {
	defer tracing.NewRegion("Tweener.Stop").End()
	for i := range tweens {
		t := &tweens[i]
		if t.val == val {
			t.stop(jumpToEnd, skipCallback)
			tweens = klib.RemoveUnordered(tweens, i)
			break
		}
	}
}
