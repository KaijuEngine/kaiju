/******************************************************************************/
/* css_vertical_align.go                                                      */
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

package properties

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/rendering"
)

func (p VerticalAlign) Sort() int { return 1 }

func directChildLabels(elm *document.Element) []*ui.Label {
	labels := make([]*ui.Label, 0)
	for _, child := range elm.Children {
		if child.IsText() {
			labels = append(labels, child.UI.ToLabel())
		}
	}
	return labels
}

func verticalAlignOffset(value string, elm *document.Element) (float32, bool, error) {
	switch value {
	case "auto", "baseline", "top", "text-top", "initial":
		return 0, false, nil
	case "middle":
		return 0.5, false, nil
	case "bottom", "text-bottom":
		return 1, false, nil
	case "sub":
		return 0, true, nil
	case "super":
		return 0, true, nil
	case "inherit":
		if parent := elm.Parent.Value(); parent != nil {
			if labels := directChildLabels(parent); len(labels) > 0 {
				return labelVerticalAlignOffset(labels[0]), false, nil
			}
		}
		return 0, false, nil
	default:
		return 0, false, fmt.Errorf("unsupported vertical-align value: %s", value)
	}
}

func labelVerticalAlignOffset(label *ui.Label) float32 {
	parent := label.Base().Entity().Parent
	if parent == nil {
		return 0
	}
	parentLayout := ui.FirstOnEntity(parent).Layout()
	available := parentLayout.ContentSize().Y() - label.Measure().Y()
	if available <= 0 {
		return 0
	}
	return -label.Base().Layout().InnerOffset().Top() / available
}

func setLabelVerticalAlign(label *ui.Label, align float32, shifted bool, value string) {
	label.SetBaseline(rendering.FontBaselineTop)
	alignOffset := float32(0)
	if parent := label.Base().Entity().Parent; parent != nil {
		parentLayout := ui.FirstOnEntity(parent).Layout()
		contentHeight := parentLayout.ContentSize().Y()
		labelHeight := label.Measure().Y()
		if contentHeight > labelHeight {
			alignOffset = (contentHeight - labelHeight) * align
		}
	}
	layout := label.Base().Layout()
	layout.SetInnerOffsetTop(-alignOffset)
	local := layout.LocalInnerOffset()
	offset := float32(0)
	if shifted {
		offset = label.FontSize() * 0.35
		if value == "sub" {
			offset = -offset
		}
	}
	layout.SetLocalInnerOffset(local.Left(), offset, local.Right(), local.Bottom())
}

// auto|baseline|bottom|middle|sub|super|text-bottom|text-top|top|initial|inherit
func (p VerticalAlign) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	}
	align, shifted, err := verticalAlignOffset(values[0].Str, elm)
	if err != nil {
		return err
	}
	labels := directChildLabels(elm)
	for _, l := range labels {
		setLabelVerticalAlign(l, align, shifted, values[0].Str)
	}
	return nil
}
