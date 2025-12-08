/******************************************************************************/
/* sprite_sheet.go                                                            */
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

package sprite

import (
	"bytes"
	"encoding/json"
	"kaiju/klib"
	"kaiju/klib/streaming"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"strings"
)

type SpriteSheetFrame struct {
	Hold      int32
	Rectangle matrix.Vec4
}

type SpriteSheetClip struct {
	Name   string
	Frames []SpriteSheetFrame
}

type SpriteSheet struct {
	Clips map[string]SpriteSheetClip
}

func (s *SpriteSheet) IsValid() bool { return len(s.Clips) > 0 }
func (s *SpriteSheet) FirstClip() SpriteSheetClip {
	for k := range s.Clips {
		return s.Clips[k]
	}
	return SpriteSheetClip{}
}

func NewSheetFromClips(clips []SpriteSheetClip) SpriteSheet {
	defer tracing.NewRegion("sprite.NewSheetFromClips").End()
	sheet := SpriteSheet{}
	for i := range clips {
		sheet.AddClip(clips[i])
	}
	return sheet
}

func NewSheetFromJson(jsonStr string) (SpriteSheet, error) {
	defer tracing.NewRegion("sprite.ReadSpriteSheetData").End()
	sheet := SpriteSheet{
		Clips: make(map[string]SpriteSheetClip),
	}
	clipList := []SpriteSheetClip{}
	err := klib.JsonDecode(json.NewDecoder(strings.NewReader(jsonStr)), &clipList)
	if err != nil {
		return sheet, err
	}
	for i := range clipList {
		sheet.Clips[clipList[i].Name] = clipList[i]
	}
	return sheet, err
}

func NewSheetFromBin(bin []byte) (SpriteSheet, error) {
	defer tracing.NewRegion("sprite.NewSheetFromBin").End()
	sheet := SpriteSheet{}
	stream := bytes.NewReader(bin)
	clipCount := int32(0)
	err := streaming.StreamRead(stream, &clipCount)
	sheet.Clips = make(map[string]SpriteSheetClip)
	for i := 0; i < int(clipCount) && err == nil; i++ {
		clip := SpriteSheetClip{}
		frameCount := int32(0)
		err = streaming.StreamRead(stream, &clip.Name, &frameCount)
		clip.Frames = make([]SpriteSheetFrame, frameCount)
		for j := 0; j < int(frameCount) && err == nil; j++ {
			frame := SpriteSheetFrame{}
			err = streaming.StreamRead(stream,
				frame.Hold,
				frame.Rectangle[matrix.Vx],
				frame.Rectangle[matrix.Vy],
				frame.Rectangle[matrix.Vz],
				frame.Rectangle[matrix.Vw],
			)
		}
		sheet.Clips[clip.Name] = clip
	}
	return sheet, err
}

func (s *SpriteSheet) AddClip(clip SpriteSheetClip) {
	s.Clips[clip.Name] = clip
}

func (s *SpriteSheet) ToJson() (string, error) {
	defer tracing.NewRegion("SpriteSheet.ToJson").End()
	clips := s.clipsSlice()
	out, err := json.Marshal(clips)
	return string(out), err
}

func (s *SpriteSheet) ToBin() ([]byte, error) {
	defer tracing.NewRegion("SpriteSheet.ToBin").End()
	clips := s.clipsSlice()
	stream := bytes.NewBuffer(nil)
	err := streaming.StreamWrite(stream, int32(len(clips)))
	for i := 0; i < len(clips) && err == nil; i++ {
		err = streaming.StreamWrite(stream,
			clips[i].Name,
			int32(len(clips[i].Frames)),
		)
		for j := 0; j < len(clips[i].Frames) && err == nil; j++ {
			f := &clips[i].Frames[j]
			err = streaming.StreamWrite(stream,
				f.Hold,
				f.Rectangle.X(),
				f.Rectangle.Y(),
				f.Rectangle.Z(),
				f.Rectangle.W(),
			)
		}
	}
	return stream.Bytes(), err
}

func (f SpriteSheetFrame) UVs(texSize matrix.Vec2) matrix.Vec4 {
	defer tracing.NewRegion("SpriteSheetFrame.UVs").End()
	w := texSize.Width()
	h := texSize.Height()
	bh := matrix.Float(f.Rectangle.Height()) / h
	return matrix.NewVec4(matrix.Float(f.Rectangle.X())/w, 1.0-matrix.Float(f.Rectangle.Y())/h-bh,
		matrix.Float(f.Rectangle.Width())/w, bh)
}

func (s *SpriteSheet) clipsSlice() []SpriteSheetClip {
	defer tracing.NewRegion("SpriteSheet.clipsSlice").End()
	clips := make([]SpriteSheetClip, 0, len(s.Clips))
	for _, v := range s.Clips {
		clips = append(clips, v)
	}
	return clips
}
