/******************************************************************************/
/* image.go                                                                   */
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

package ui

import (
	"log/slog"
	"slices"

	"github.com/KaijuEngine/kaiju/engine/systems/visual2d/sprite"
	"github.com/KaijuEngine/kaiju/matrix"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
	"github.com/KaijuEngine/kaiju/rendering"
)

type Image Panel

type imageData struct {
	panelData
	flipBook                 []*rendering.Texture
	spriteSheet              sprite.SpriteSheet
	frameDelay, fps          float32
	frameCount, currentFrame int
	paused                   bool
}

func (i *imageData) innerPanelData() *panelData { return &i.panelData }

func (u *UI) ToImage() *Image { return (*Image)(u) }
func (s *Image) Base() *UI    { return (*UI)(s) }

func (s *Image) ImageData() *imageData {
	return s.elmData.(*imageData)
}

func (s *Image) Init(texture *rendering.Texture) {
	s.elmData = &imageData{
		flipBook: []*rendering.Texture{texture},
	}
	p := s.Base().ToPanel()
	p.Init(texture, ElementTypeImage)
	if p.shaderData != nil {
		p.shaderData.BorderLen = matrix.Vec2Zero()
	}
}

func (s *Image) InitFlipbook(framesPerSecond float32, textures []*rendering.Texture) {
	s.elmData = &imageData{}
	p := s.Base().ToPanel()
	p.Init(nil, ElementTypeImage)
	s.SetFlipBookAnimation(framesPerSecond, textures...)
	if p.shaderData != nil {
		p.shaderData.BorderLen = matrix.Vec2Zero()
	}
}

func (s *Image) InitSpriteSheet(framesPerSecond float32, texture *rendering.Texture, jsonStr string) {
	s.elmData = &imageData{}
	p := s.Base().ToPanel()
	p.Init(nil, ElementTypeImage)
	s.SetSpriteSheet(framesPerSecond, texture, jsonStr)
	if p.shaderData != nil {
		p.shaderData.BorderLen = matrix.Vec2Zero()
	}
}

func (img *Image) resetDelay() {
	data := img.ImageData()
	data.frameDelay = 1.0 / data.fps
}

func (img *Image) update(deltaTime float64) {
	defer tracing.NewRegion("Image.update").End()
	img.Base().ToPanel().update(deltaTime)
	data := img.ImageData()
	if !data.paused {
		data.frameDelay -= float32(deltaTime)
	}
	if data.frameCount > 0 && data.frameDelay <= 0.0 {
		next := data.currentFrame + 1
		if next == data.frameCount {
			next = 0
		}
		if data.spriteSheet.IsValid() || len(data.flipBook) > 0 {
			img.SetFrame(next)
		}
		// TODO:  Else Atlas animation
		img.resetDelay()
	}
}

func (img *Image) SetTexture(texture *rendering.Texture) {
	(*Panel)(img).SetBackground(texture)
}

func (img *Image) SetFlipBookAnimation(framesPerSecond float32, textures ...*rendering.Texture) {
	data := img.ImageData()
	count := len(textures)
	data.flipBook = slices.Clone(textures)
	data.frameCount = count
	data.fps = framesPerSecond
	data.currentFrame = 0
	img.resetDelay()
	img.SetTexture(data.flipBook[data.currentFrame])
}

func (img *Image) SetSpriteSheet(framesPerSecond float32, texture *rendering.Texture, jsonStr string) {
	data := img.ImageData()
	var err error
	data.spriteSheet, err = sprite.NewSheetFromJson(jsonStr)
	if err != nil {
		slog.Error("failed to load the UI sprite sheet", "error", err)
		return
	}
	data.frameCount = len(data.spriteSheet.FirstClip().Frames)
	data.fps = framesPerSecond
	data.currentFrame = -1
	img.resetDelay()
	img.SetTexture(texture)
	img.SetFrame(0)
}

func (img *Image) Frame() int { return img.ImageData().currentFrame }

func (img *Image) SetFrame(index int) {
	data := img.ImageData()
	if data.currentFrame == index {
		return
	}
	data.currentFrame = index
	if data.spriteSheet.IsValid() {
		clip := data.spriteSheet.FirstClip()
		img.shaderData.UVs = clip.Frames[data.currentFrame].UVs(img.textureSize)
	} else {
		img.SetTexture(data.flipBook[data.currentFrame])
	}
}

func (img *Image) SetFrameRate(framesPerSecond float32) {
	img.ImageData().fps = framesPerSecond
}

func (img *Image) PlayAnimation() {
	img.ImageData().paused = false
}

func (img *Image) StopAnimation() {
	img.ImageData().paused = true
}
