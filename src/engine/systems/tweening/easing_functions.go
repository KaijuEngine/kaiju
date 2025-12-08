/******************************************************************************/
/* easing_functions.go                                                        */
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
	"kaiju/matrix"
	"math"
)

func easeLinear(t float32) float32 { return t }

func easeInSine(t float32) float32 { return 1 - matrix.Cos((t*math.Pi)/2) }

func easeOutSine(t float32) float32 { return matrix.Sin((t * math.Pi) / 2) }

func easeInOutSine(t float32) float32 { return -(matrix.Cos(math.Pi*t) - 1) / 2 }

func easeInQuad(t float32) float32 { return t * t }

func easeOutQuad(t float32) float32 { return 1 - (1-t)*(1-t) }

func easeInOutQuad(t float32) float32 {
	if t < 0.5 {
		return 2 * t * t
	} else {
		return 1 - matrix.Pow(-2*t+2, 2)/2
	}
}

func easeInCubic(t float32) float32 { return t * t * t }

func easeOutCubic(t float32) float32 { return 1 - matrix.Pow(1-t, 3) }

func easeInOutCubic(t float32) float32 {
	if t < 0.5 {
		return 4 * t * t * t
	} else {
		return 1 - matrix.Pow(-2*t+2, 3)/2
	}
}

func easeInQuart(t float32) float32 { return t * t * t * t }

func easeOutQuart(t float32) float32 { return 1 - matrix.Pow(1-t, 4) }

func easeInOutQuart(t float32) float32 {
	if t < 0.5 {
		return 8 * t * t * t * t
	} else {
		return 1 - matrix.Pow(-2*t+2, 4)/2
	}
}

func easeInQuint(t float32) float32 { return t * t * t * t * t }

func easeOutQuint(t float32) float32 { return 1 - matrix.Pow(1-t, 5) }

func easeInOutQuint(t float32) float32 {
	if t < 0.5 {
		return 16 * t * t * t * t * t
	} else {
		return 1 - matrix.Pow(-2*t+2, 5)/2
	}
}

func easeInExpo(t float32) float32 {
	if matrix.Approx(t, 0) {
		return 0
	} else {
		return matrix.Pow(2, 10*t-10)
	}
}

func easeOutExpo(t float32) float32 {
	if matrix.Approx(t, 1) {
		return 1
	} else {
		return 1 - matrix.Pow(2, -10*t)
	}
}

func easeInOutExpo(t float32) float32 {
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

func easeInCirc(t float32) float32 { return 1 - matrix.Sqrt(1-matrix.Pow(t, 2)) }

func easeOutCirc(t float32) float32 { return matrix.Sqrt(1 - matrix.Pow(t-1, 2)) }

func easeInOutCirc(t float32) float32 {
	if t < 0.5 {
		return (1 - matrix.Sqrt(1-matrix.Pow(2*t, 2))) / 2
	} else {
		return (matrix.Sqrt(1-matrix.Pow(-2*t+2, 2)) + 1) / 2
	}
}

func easeInBack(t float32) float32 {
	const c1 = 1.70158
	const c3 = c1 + 1
	return c3*t*t*t - c1*t*t
}

func easeOutBack(t float32) float32 {
	const c1 = 1.70158
	const c3 = c1 + 1
	return 1 + c3*matrix.Pow(t-1, 3) + c1*matrix.Pow(t-1, 2)
}

func easeInOutBack(t float32) float32 {
	const c1 = 1.70158
	const c2 = c1 * 1.525
	if t < 0.5 {
		return (matrix.Pow(2*t, 2) * ((c2+1)*2*t - c2)) / 2
	} else {
		return (matrix.Pow(2*t-2, 2)*((c2+1)*(t*2-2)+c2) + 2) / 2
	}
}

func easeInElastic(t float32) float32 {
	const c4 = (2 * math.Pi) / 3
	if matrix.Approx(t, 0) {
		return 0
	} else if matrix.Approx(t, 1) {
		return 1
	} else {
		return -matrix.Pow(2, 10*t-10) * matrix.Sin((t*10-10.75)*c4)
	}
}

func easeOutElastic(t float32) float32 {
	const c4 = (2 * math.Pi) / 3
	if matrix.Approx(t, 0) {
		return 0
	} else if matrix.Approx(t, 1) {
		return 1
	} else {
		return matrix.Pow(2, -10*t)*matrix.Sin((t*10-0.75)*c4) + 1
	}
}

func easeInOutElastic(t float32) float32 {
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

func easeOutBounce(t float32) float32 {
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

func easeInBounce(t float32) float32 {
	return 1 - easeOutBounce(1-t)
}

func easeInOutBounce(t float32) float32 {
	if t < 0.5 {
		return (1 - easeOutBounce(1-2*t)) / 2
	} else {
		return (1 + easeOutBounce(2*t-1)) / 2
	}
}
