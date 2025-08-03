/******************************************************************************/
/* details_tag_parser.go                                                      */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package details_window

import (
	"kaiju/klib"
	"log/slog"
	"strings"
)

var (
	tagParsers = map[string]func(f *entityDataField, value string){
		"default": tagDefault,
		"clamp":   tagClamp,
	}
)

func tagDefault(f *entityDataField, value string) {
	f.Value = klib.StringToTypeValue(f.Type, value)
}

func tagClamp(f *entityDataField, value string) {
	if !f.IsNumber() {
		slog.Warn("cannot use the clamp tag on non-numeric field", "field", f.Name)
		return
	}
	parts := strings.Split(value, ",")
	if len(parts) == 2 {
		parts = append([]string{"0"}, parts...)
	}
	if len(parts) == 3 {
		values := make([]any, len(parts))
		for i := range parts {
			values[i] = klib.StringToTypeValue(f.Type, parts[i])
		}
		f.Value = values[0]
		f.Min = values[1]
		f.Max = values[2]
	} else {
		slog.Warn("invalid format for the 'clamp' tag on field", "field", f.Name)
	}
}
