/******************************************************************************/
/* easing_functions.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package tweening

import (
	"math"

	"kaijuengine.com/matrix"
)

func easeLinear(t matrix.Float) matrix.Float { return t }

func easeInSine(t matrix.Float) matrix.Float { return 1 - matrix.Cos((t*math.Pi)/2) }

func easeOutSine(t matrix.Float) matrix.Float { return matrix.Sin((t * math.Pi) / 2) }

func easeInOutSine(t matrix.Float) matrix.Float { return -(matrix.Cos(math.Pi*t) - 1) / 2 }

func easeInQuad(t matrix.Float) matrix.Float { return t * t }

func easeOutQuad(t matrix.Float) matrix.Float { return 1 - (1-t)*(1-t) }

func easeInOutQuad(t matrix.Float) matrix.Float {
	if t < 0.5 {
		return 2 * t * t
	} else {
		return 1 - matrix.Pow(-2*t+2, 2)/2
	}
}

func easeInCubic(t matrix.Float) matrix.Float { return t * t * t }

func easeOutCubic(t matrix.Float) matrix.Float { return 1 - matrix.Pow(1-t, 3) }

func easeInOutCubic(t matrix.Float) matrix.Float {
	if t < 0.5 {
		return 4 * t * t * t
	} else {
		return 1 - matrix.Pow(-2*t+2, 3)/2
	}
}

func easeInQuart(t matrix.Float) matrix.Float { return t * t * t * t }

func easeOutQuart(t matrix.Float) matrix.Float { return 1 - matrix.Pow(1-t, 4) }

func easeInOutQuart(t matrix.Float) matrix.Float {
	if t < 0.5 {
		return 8 * t * t * t * t
	} else {
		return 1 - matrix.Pow(-2*t+2, 4)/2
	}
}

func easeInQuint(t matrix.Float) matrix.Float { return t * t * t * t * t }

func easeOutQuint(t matrix.Float) matrix.Float { return 1 - matrix.Pow(1-t, 5) }

func easeInOutQuint(t matrix.Float) matrix.Float {
	if t < 0.5 {
		return 16 * t * t * t * t * t
	} else {
		return 1 - matrix.Pow(-2*t+2, 5)/2
	}
}

func easeInExpo(t matrix.Float) matrix.Float {
	if matrix.Approx(t, 0) {
		return 0
	} else {
		return matrix.Pow(2, 10*t-10)
	}
}

func easeOutExpo(t matrix.Float) matrix.Float {
	if matrix.Approx(t, 1) {
		return 1
	} else {
		return 1 - matrix.Pow(2, -10*t)
	}
}

func easeInOutExpo(t matrix.Float) matrix.Float {
	if matrix.Approx(t, 0) {
		return 0
	} else if matrix.Approx(t, 1) {
		return 1
	} else if t < 0.5 {
		return matrix.Pow(2, 20*t-10) / 2
	} else {
		return (2 - matrix.Pow(2, -20*t+10)) / 2
	}
}

func easeInCirc(t matrix.Float) matrix.Float { return 1 - matrix.Sqrt(1-matrix.Pow(t, 2)) }

func easeOutCirc(t matrix.Float) matrix.Float { return matrix.Sqrt(1 - matrix.Pow(t-1, 2)) }

func easeInOutCirc(t matrix.Float) matrix.Float {
	if t < 0.5 {
		return (1 - matrix.Sqrt(1-matrix.Pow(2*t, 2))) / 2
	} else {
		return (matrix.Sqrt(1-matrix.Pow(-2*t+2, 2)) + 1) / 2
	}
}

func easeInBack(t matrix.Float) matrix.Float {
	const c1 = 1.70158
	const c3 = c1 + 1
	return c3*t*t*t - c1*t*t
}

func easeOutBack(t matrix.Float) matrix.Float {
	const c1 = 1.70158
	const c3 = c1 + 1
	return 1 + c3*matrix.Pow(t-1, 3) + c1*matrix.Pow(t-1, 2)
}

func easeInOutBack(t matrix.Float) matrix.Float {
	const c1 = 1.70158
	const c2 = c1 * 1.525
	if t < 0.5 {
		return (matrix.Pow(2*t, 2) * ((c2+1)*2*t - c2)) / 2
	} else {
		return (matrix.Pow(2*t-2, 2)*((c2+1)*(t*2-2)+c2) + 2) / 2
	}
}

func easeInElastic(t matrix.Float) matrix.Float {
	const c4 = (2 * math.Pi) / 3
	if matrix.Approx(t, 0) {
		return 0
	} else if matrix.Approx(t, 1) {
		return 1
	} else {
		return -matrix.Pow(2, 10*t-10) * matrix.Sin((t*10-10.75)*c4)
	}
}

func easeOutElastic(t matrix.Float) matrix.Float {
	const c4 = (2 * math.Pi) / 3
	if matrix.Approx(t, 0) {
		return 0
	} else if matrix.Approx(t, 1) {
		return 1
	} else {
		return matrix.Pow(2, -10*t)*matrix.Sin((t*10-0.75)*c4) + 1
	}
}

func easeInOutElastic(t matrix.Float) matrix.Float {
	const c5 = (2 * math.Pi) / 4.5
	if matrix.Approx(t, 0) {
		return 0
	} else if matrix.Approx(t, 1) {
		return 1
	} else if t < 0.5 {
		return -(matrix.Pow(2, 20*t-10) * matrix.Sin((20*t-11.125)*c5)) / 2
	} else {
		return (matrix.Pow(2, -20*t+10)*matrix.Sin((20*t-11.125)*c5))/2 + 1
	}
}

func easeOutBounce(t matrix.Float) matrix.Float {
	const n1 = 7.5625
	const d1 = 2.75
	if t < (1 / d1) {
		return n1 * t * t
	} else if t < (2 / d1) {
		t = t - 1.5/d1
		return n1*t*t + 0.75
	} else if t < (2.5 / d1) {
		t = t - 2.25/d1
		return n1*t*t + 0.9375
	} else {
		t = t - 2.625/d1
		return n1*t*t + 0.984375
	}
}

func easeInBounce(t matrix.Float) matrix.Float {
	return 1 - easeOutBounce(1-t)
}

func easeInOutBounce(t matrix.Float) matrix.Float {
	if t < 0.5 {
		return (1 - easeOutBounce(1-2*t)) / 2
	} else {
		return (1 + easeOutBounce(2*t-1)) / 2
	}
}
