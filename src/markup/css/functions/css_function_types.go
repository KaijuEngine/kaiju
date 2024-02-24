/******************************************************************************/
/* css_function_types.go                                                      */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                      */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package functions

// Returns the value of an attribute of the selected element
type Attr struct{}

func (f Attr) Key() string { return "attr" }

// Allows you to perform calculations to determine CSS property values
type Calc struct{}

func (f Calc) Key() string { return "calc" }

// Creates a conic gradient
type ConicGradient struct{}

func (f ConicGradient) Key() string { return "conic-gradient" }

// Returns the current value of the named counter
type Counter struct{}

func (f Counter) Key() string { return "counter" }

// Defines a Cubic Bezier curve
type CubicBezier struct{}

func (f CubicBezier) Key() string { return "cubic-bezier" }

// Defines colors using the Hue-Saturation-Lightness model (HSL)
type Hsl struct{}

func (f Hsl) Key() string { return "hsl" }

// Defines colors using the Hue-Saturation-Lightness-Alpha model (HSLA)
type Hsla struct{}

func (f Hsla) Key() string { return "hsla" }

// Creates a linear gradient
type LinearGradient struct{}

func (f LinearGradient) Key() string { return "linear-gradient" }

// Uses the largest value, from a comma-separated list of values, as the property value
type Max struct{}

func (f Max) Key() string { return "max" }

// Uses the smallest value, from a comma-separated list of values, as the property value
type Min struct{}

func (f Min) Key() string { return "min" }

// Creates a radial gradient
type RadialGradient struct{}

func (f RadialGradient) Key() string { return "radial-gradient" }

// Repeats a conic gradient
type RepeatingConicGradient struct{}

func (f RepeatingConicGradient) Key() string { return "repeating-conic-gradient" }

// Repeats a linear gradient
type RepeatingLinearGradient struct{}

func (f RepeatingLinearGradient) Key() string { return "repeating-linear-gradient" }

// Repeats a radial gradient
type RepeatingRadialGradient struct{}

func (f RepeatingRadialGradient) Key() string { return "repeating-radial-gradient" }

// Defines colors using the Red-Green-Blue model (RGB)
type Rgb struct{}

func (f Rgb) Key() string { return "rgb" }

// Defines colors using the Red-Green-Blue-Alpha model (RGBA)
type Rgba struct{}

func (f Rgba) Key() string { return "rgba" }

// Inserts the value of a custom property
type Var struct{}

func (f Var) Key() string { return "var" }
