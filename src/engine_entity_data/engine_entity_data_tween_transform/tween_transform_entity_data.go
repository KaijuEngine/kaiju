/******************************************************************************/
/* tween_transform_entity_data.go                                             */
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

package engine_entity_data_physics

import (
	"kaiju/engine"
	"kaiju/engine/systems/tweening"
	"kaiju/matrix"
	"time"
	"weak"
)

const BindingKey = "kaiju.TweenTransformEntityData"

type Easing int
type Repeat int

const (
	EasingLinear Easing = iota
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

const (
	RepeatNone Repeat = iota
	RepeatLoop
	RepeatPingPong
)

func init() {
	engine.RegisterEntityData(BindingKey, TweenTransformEntityData{})
}

type TweenTransformEntityData struct {
	Position    matrix.Vec3
	Rotation    matrix.Vec3
	Scale       matrix.Vec3
	Repeat      Repeat
	Easing      Easing
	Time        float64
	Delay       float64
	RepeatDelay float64
	RepeatCount int
	IsAbsolute  bool
}

type tweenData struct {
	originalPosition matrix.Vec3
	originalRotation matrix.Vec3
	originalScale    matrix.Vec3
	targetPosition   matrix.Vec3
	targetRotation   matrix.Vec3
	targetScale      matrix.Vec3
	repeat           Repeat
	delay            float64
	repeatCount      int
	t                float32
	e                weak.Pointer[engine.Entity]
	isGoingTo        bool
}

func (d TweenTransformEntityData) Init(e *engine.Entity, host *engine.Host) {
	wp, wr, ws := e.Transform.WorldTransform()
	tween := tweenData{
		originalPosition: wp,
		originalRotation: wr,
		originalScale:    ws,
		targetPosition:   d.Position,
		targetRotation:   d.Rotation,
		targetScale:      d.Scale,
		repeat:           d.Repeat,
		repeatCount:      d.RepeatCount,
		delay:            d.RepeatDelay,
		e:                weak.Make(e),
		isGoingTo:        true,
	}
	if !d.IsAbsolute {
		tween.targetPosition.AddAssign(wp)
		tween.targetRotation.AddAssign(wr)
		tween.targetScale.AddAssign(ws)
	}
	tweenTo := func(float32) {
		se := tween.e.Value()
		if se == nil {
			return
		}
		p := matrix.Vec3Lerp(tween.originalPosition, tween.targetPosition, tween.t)
		r := matrix.Vec3Lerp(tween.originalRotation, tween.targetRotation, tween.t)
		s := matrix.Vec3Lerp(tween.originalScale, tween.targetScale, tween.t)
		se.Transform.SetWorldPosition(p)
		se.Transform.SetWorldRotation(r)
		se.Transform.SetWorldScale(s)
	}
	var runTween func(to float32)
	tweenDone := func() {
		switch tween.repeat {
		case RepeatLoop:
			if tween.repeatCount > 0 {
				if tween.repeatCount == 1 {
					break
				}
				tween.repeatCount--
			}
			runTween(1)
		case RepeatPingPong:
			tween.isGoingTo = !tween.isGoingTo
			if tween.isGoingTo && tween.repeatCount > 0 {
				if tween.repeatCount == 1 {
					break
				}
				tween.repeatCount--
			}
			if tween.isGoingTo {
				runTween(1)
			} else {
				runTween(0)
			}
		}
	}
	wh := weak.Make(host)
	runTween = func(to float32) {
		h := wh.Value()
		if h == nil {
			return
		}
		run := func() {
			tween.t = 1 - to
			tweening.DoTweenExt(&tween.t, to, d.Time, tweening.Easing(d.Easing), tweenTo, tweenDone)
		}
		if d.Delay > 0 {
			h.RunAfterTime(time.Millisecond*(time.Duration(d.Delay*1000)), run)
		} else {
			h.RunNextFrame(run)
		}
	}
	runTween(1)
}
