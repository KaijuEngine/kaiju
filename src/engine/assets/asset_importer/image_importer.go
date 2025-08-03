/******************************************************************************/
/* image_importer.go                                                          */
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

package asset_importer

import (
	"kaiju/rendering"
	"log/slog"
)

var (
	// Registered in meta_options_export.go
	imageFilterOptions = map[string]rendering.TextureFilter{
		"Linear":  rendering.TextureFilterLinear,
		"Nearest": rendering.TextureFilterNearest,
	}
	imagePivot = map[string]rendering.QuadPivot{
		"Center":       rendering.QuadPivotCenter,
		"Left":         rendering.QuadPivotLeft,
		"Top":          rendering.QuadPivotTop,
		"Right":        rendering.QuadPivotRight,
		"Bottom":       rendering.QuadPivotBottom,
		"Bottom left":  rendering.QuadPivotBottomLeft,
		"Bottom right": rendering.QuadPivotBottomRight,
		"Top left":     rendering.QuadPivotTopLeft,
		"Top right":    rendering.QuadPivotTopRight,
	}
	imageMaxSize = map[string]int32{
		"16":   16,
		"32":   32,
		"64":   64,
		"128":  128,
		"256":  256,
		"512":  512,
		"1024": 1024,
		"2048": 2048,
		"4096": 4096,
		"8192": 8192,
	}
)

type ImageMetadata struct {
	Filter        string `options:"imageFilterOptions"`
	Pivot         string `options:"imagePivot"`
	PixelsPerUnit int32
	Mipmaps       int32

	// TODO:  This needs to be used for packaging the content
	MaxSize string `options:"imageMaxSize"`
}

func defaultImageMetadata() *ImageMetadata {
	return &ImageMetadata{
		Filter:        "Linear",
		Pivot:         "Center",
		PixelsPerUnit: 128,
		Mipmaps:       1,
		MaxSize:       "8192",
	}
}

func (m *ImageMetadata) ImageFilterMeta() rendering.TextureFilter {
	if f, ok := imageFilterOptions[m.Filter]; ok {
		return f
	}
	slog.Warn("tried to read image filter metadata but has invalid key",
		"key", m.Filter)
	return rendering.TextureFilterLinear
}

func (m *ImageMetadata) ImagePivotMeta() rendering.QuadPivot {
	if f, ok := imagePivot[m.Pivot]; ok {
		return f
	}
	slog.Warn("tried to read image pivot metadata but has invalid key",
		"key", m.Pivot)
	return rendering.QuadPivotCenter
}

func (m *ImageMetadata) MaxSizeMeta() int32 {
	if f, ok := imageMaxSize[m.MaxSize]; ok {
		return f
	}
	slog.Warn("tried to read image max size metadata but has invalid key",
		"key", m.MaxSize)
	return 8192
}
