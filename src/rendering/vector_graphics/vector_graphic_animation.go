/******************************************************************************/
/* vector_graphic_animation.go                                                */
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

package vector_graphics

import (
	"kaiju/rendering/vector_graphics/svg"
	"strconv"
	"strings"
)

// SharedAnimation contains fields common to both <animate> and <animateTransform> elements
// for the vector graphics animation representation used in this package.
type SharedAnimation struct {
	From        string
	To          string
	By          string
	Values      string
	KeyTimes    string
	KeySplines  string
	CalcMode    CalcMode
	Duration    float64
	Begin       string
	End         string
	RepeatCount float64
	RepeatDur   string
	Fill        FillMode
	Restart     RestartMode
	Additive    Additive
	Accumulate  Accumulate
}

type CalcMode int8
type FillMode int8
type RestartMode int8
type Additive int8
type Accumulate int8

const (
	CalcModeDiscrete CalcMode = iota
	CalcModeLinear
	CalcModePaced
	CalcModeSpline
)

const (
	FillFreeze FillMode = iota
	FillRemove
)

const (
	RestartAlways RestartMode = iota
	RestartWhenNotActive
	RestartNever
)
const (
	AdditiveReplace Additive = iota
	AdditiveSum
)

const (
	AccumulateNone Accumulate = iota
	AccumulateSum
)

type Animate struct {
	AttributeName string
	AttributeType string
	SharedAnimation
	Min string
	Max string
}

// AnimateTransform represents <animateTransform> elements for animating transformations.
// It contains the type of transformation (translate, rotate, scale, skewX, skewY)
// and embeds SharedAnimation for the common animation fields.
type AnimateTransform struct {
	Type string
	SharedAnimation
}

func (a *Animate) LoopsIndefinetely() bool {
	return a.RepeatCount < 0
}

func AnimateFromSvg(anim svg.Animate) Animate {
	// Parse duration (e.g., "0.416667s")
	var dur float64
	if strings.HasSuffix(anim.Duration, "s") {
		if v, err := strconv.ParseFloat(strings.TrimSuffix(anim.Duration, "s"), 64); err == nil {
			dur = v
		}
	}
	// Parse repeatCount – "indefinite" maps to a negative value.
	var repeat float64
	if anim.RepeatCount == "indefinite" {
		repeat = -1
	} else if anim.RepeatCount != "" {
		if v, err := strconv.ParseFloat(anim.RepeatCount, 64); err == nil {
			repeat = v
		}
	}
	return Animate{
		AttributeName: string(anim.AttributeName),
		AttributeType: string(anim.AttributeType),
		SharedAnimation: SharedAnimation{
			From:        anim.From,
			To:          anim.To,
			By:          anim.By,
			Values:      anim.Values,
			KeyTimes:    anim.KeyTimes,
			KeySplines:  anim.KeySplines,
			CalcMode:    mapCalcMode(anim.CalcMode),
			Duration:    dur,
			Begin:       anim.Begin,
			End:         anim.End,
			RepeatCount: repeat,
			RepeatDur:   anim.RepeatDur,
			Fill:        mapFillMode(anim.Fill),
			Restart:     mapRestartMode(anim.Restart),
			Additive:    mapAdditive(anim.Additive),
			Accumulate:  mapAccumulate(anim.Accumulate),
		},
		Min: anim.Min,
		Max: anim.Max,
	}
}

func AnimateTransformFromSvg(anim svg.AnimateTransform) AnimateTransform {
	// Parse duration (e.g., "0.416667s")
	var dur float64
	if strings.HasSuffix(anim.Duration, "s") {
		if v, err := strconv.ParseFloat(strings.TrimSuffix(anim.Duration, "s"), 64); err == nil {
			dur = v
		}
	}
	// Parse repeatCount – "indefinite" maps to a negative value.
	var repeat float64
	if anim.RepeatCount == "indefinite" {
		repeat = -1
	} else if anim.RepeatCount != "" {
		if v, err := strconv.ParseFloat(anim.RepeatCount, 64); err == nil {
			repeat = v
		}
	}
	return AnimateTransform{
		Type: anim.Type,
		SharedAnimation: SharedAnimation{
			From:        anim.From,
			To:          anim.To,
			By:          anim.By,
			Values:      anim.Values,
			KeyTimes:    anim.KeyTimes,
			KeySplines:  anim.KeySplines,
			CalcMode:    mapCalcMode(anim.CalcMode),
			Duration:    dur,
			Begin:       anim.Begin,
			End:         anim.End,
			RepeatCount: repeat,
			RepeatDur:   anim.RepeatDur,
			Fill:        mapFillMode(anim.Fill),
			Restart:     mapRestartMode(anim.Restart),
			Additive:    mapAdditive(anim.Additive),
			Accumulate:  mapAccumulate(anim.Accumulate),
		},
	}
}

func mapCalcMode(s svg.CalcMode) CalcMode {
	switch s {
	case svg.CalcModeDiscrete:
		return CalcModeDiscrete
	case svg.CalcModeLinear:
		return CalcModeLinear
	case svg.CalcModePaced:
		return CalcModePaced
	case svg.CalcModeSpline:
		return CalcModeSpline
	default:
		return CalcModeLinear // default per SVG spec
	}
}

func mapFillMode(s svg.FillMode) FillMode {
	switch s {
	case svg.FillFreeze:
		return FillFreeze
	case svg.FillRemove:
		return FillRemove
	default:
		return FillRemove
	}
}

func mapRestartMode(s svg.RestartMode) RestartMode {
	switch s {
	case svg.RestartAlways:
		return RestartAlways
	case svg.RestartWhenNotActive:
		return RestartWhenNotActive
	case svg.RestartNever:
		return RestartNever
	default:
		return RestartAlways
	}
}

func mapAdditive(s svg.Additive) Additive {
	switch s {
	case svg.AdditiveReplace:
		return AdditiveReplace
	case svg.AdditiveSum:
		return AdditiveSum
	default:
		return AdditiveReplace
	}
}

func mapAccumulate(s svg.Accumulate) Accumulate {
	switch s {
	case svg.AccumulateNone:
		return AccumulateNone
	case svg.AccumulateSum:
		return AccumulateSum
	default:
		return AccumulateNone
	}
}
