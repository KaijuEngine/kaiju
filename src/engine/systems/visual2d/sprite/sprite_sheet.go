/******************************************************************************/
/* sprite_sheet.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package sprite

import (
	"bytes"
	"encoding/json"
	"strings"

	"kaijuengine.com/klib"
	"kaijuengine.com/klib/streaming"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
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
