/******************************************************************************/
/* sprite_sheet.go                                                            */
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

package sprite

import (
	"encoding/json"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"strings"
)

type SpriteSheetClip struct {
	Name   string             `json:"name"`
	Frames []SpriteSheetFrame `json:"frames"`
}

type SpriteSheetFrame struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
	W int `json:"w"`
}

type SpriteSheet struct {
	clips    map[string]*SpriteSheetClip
	clipList []SpriteSheetClip
}

func (s *SpriteSheet) FirstClip() SpriteSheetClip {
	return s.clipList[0]
}

func (s *SpriteSheet) IsValid() bool {
	return len(s.clipList) > 0
}

func (f SpriteSheetFrame) Left() int   { return f.X }
func (f SpriteSheetFrame) Top() int    { return f.Y }
func (f SpriteSheetFrame) Width() int  { return f.Z }
func (f SpriteSheetFrame) Height() int { return f.W }

func (f SpriteSheetFrame) UVs(texSize matrix.Vec2) matrix.Vec4 {
	w := texSize.Width()
	h := texSize.Height()
	bh := matrix.Float(f.Height()) / h
	return matrix.NewVec4(matrix.Float(f.X)/w, 1.0-matrix.Float(f.Y)/h-bh,
		matrix.Float(f.Width())/w, bh)
}

func ReadSpriteSheetData(jsonStr string) (SpriteSheet, error) {
	defer tracing.NewRegion("sprite.ReadSpriteSheetData").End()
	sheet := SpriteSheet{
		clips: make(map[string]*SpriteSheetClip),
	}
	err := klib.JsonDecode(json.NewDecoder(strings.NewReader(jsonStr)), &sheet.clipList)
	if err != nil {
		return sheet, err
	}
	for i := range sheet.clipList {
		sheet.clips[sheet.clipList[i].Name] = &sheet.clipList[i]
	}
	return sheet, err
}
