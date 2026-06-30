/******************************************************************************/
/* tweener.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package tweening

import (
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
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
	tween.easing = easingFunc(easing)
	// Stop the tweener for the same value if one exists
	Stop(val, true, false)
	tweens = append(tweens, tween)
}

// Apply samples this easing curve at normalized time t (typically in [0, 1]) and
// returns the eased progress. Use it to shape a progress value you compute yourself
// (e.g. a server-timestamp-derived fraction) without driving a stateful Tween/Tweener.
func (e Easing) Apply(t float32) float32 { return easingFunc(e)(t) }

// easingFunc maps an Easing to its curve function. An unrecognized value falls back to
// linear so the result is always a usable (non-nil) function.
func easingFunc(easing Easing) func(float32) float32 {
	switch easing {
	case EasingLinear:
		return easeLinear
	case EasingIn:
		return easeInQuad
	case EasingOut:
		return easeOutQuad
	case EasingInAndOut:
		return easeInOutQuad
	case EasingInSine:
		return easeInSine
	case EasingOutSine:
		return easeOutSine
	case EasingInAndOutSine:
		return easeInOutSine
	case EasingInQuad:
		return easeInQuad
	case EasingOutQuad:
		return easeOutQuad
	case EasingInAndOutQuad:
		return easeInOutQuad
	case EasingInCubic:
		return easeInCubic
	case EasingOutCubic:
		return easeOutCubic
	case EasingInAndOutCubic:
		return easeInOutCubic
	case EasingInQuart:
		return easeInQuart
	case EasingOutQuart:
		return easeOutQuart
	case EasingInAndOutQuart:
		return easeInOutQuart
	case EasingInQuint:
		return easeInQuint
	case EasingOutQuint:
		return easeOutQuint
	case EasingInAndOutQuint:
		return easeInOutQuint
	case EasingInExpo:
		return easeInExpo
	case EasingOutExpo:
		return easeOutExpo
	case EasingInAndOutExpo:
		return easeInOutExpo
	case EasingInCirc:
		return easeInCirc
	case EasingOutCirc:
		return easeOutCirc
	case EasingInAndOutCirc:
		return easeInOutCirc
	case EasingInBack:
		return easeInBack
	case EasingOutBack:
		return easeOutBack
	case EasingInAndOutBack:
		return easeInOutBack
	case EasingInElastic:
		return easeInElastic
	case EasingOutElastic:
		return easeOutElastic
	case EasingInAndOutElastic:
		return easeInOutElastic
	case EasingInBounce:
		return easeInBounce
	case EasingOutBounce:
		return easeOutBounce
	case EasingInAndOutBounce:
		return easeInOutBounce
	default:
		return easeLinear
	}
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
