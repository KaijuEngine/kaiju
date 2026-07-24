/******************************************************************************/
/* tween.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package tweening

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type Tween struct {
	val         *matrix.Float
	totalUpdate float64
	time        float64
	initial     matrix.Float
	target      matrix.Float
	easing      func(t matrix.Float) matrix.Float
	onChange    func(val matrix.Float)
	onDone      func()
	scale       matrix.Float
}

func (t *Tween) update(deltaTime float64) bool {
	defer tracing.NewRegion("Tween.update").End()
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

func (t *Tween) calculate() matrix.Float {
	time := matrix.Float(1 - ((t.time - t.totalUpdate) / t.time))
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
