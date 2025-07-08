package tweening

import "kaiju/platform/profiler/tracing"

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
