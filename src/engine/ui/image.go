/******************************************************************************/
/* image.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"log/slog"
	"slices"

	"kaijuengine.com/engine/systems/visual2d/sprite"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

type Image Panel

type imageData struct {
	panelData
	flipBook                 []*rendering.Texture
	spriteSheet              sprite.SpriteSheet
	frameDelay, fps          matrix.Float
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

func (s *Image) InitFlipbook(framesPerSecond matrix.Float, textures []*rendering.Texture) {
	s.elmData = &imageData{}
	p := s.Base().ToPanel()
	p.Init(nil, ElementTypeImage)
	s.SetFlipBookAnimation(framesPerSecond, textures...)
	if p.shaderData != nil {
		p.shaderData.BorderLen = matrix.Vec2Zero()
	}
}

func (s *Image) InitSpriteSheet(framesPerSecond matrix.Float, texture *rendering.Texture, jsonStr string) {
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
		data.frameDelay -= matrix.Float(deltaTime)
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

func (img *Image) SetFlipBookAnimation(framesPerSecond matrix.Float, textures ...*rendering.Texture) {
	data := img.ImageData()
	count := len(textures)
	data.flipBook = slices.Clone(textures)
	data.frameCount = count
	data.fps = framesPerSecond
	data.currentFrame = 0
	img.resetDelay()
	img.SetTexture(data.flipBook[data.currentFrame])
}

func (img *Image) SetSpriteSheet(framesPerSecond matrix.Float, texture *rendering.Texture, jsonStr string) {
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

func (img *Image) SetFrameRate(framesPerSecond matrix.Float) {
	img.ImageData().fps = framesPerSecond
}

func (img *Image) PlayAnimation() {
	img.ImageData().paused = false
}

func (img *Image) StopAnimation() {
	img.ImageData().paused = true
}
