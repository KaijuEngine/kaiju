package ui

import (
	"kaiju/engine"
	"kaiju/rendering"
)

type Image Panel

type imageData struct {
	flipBook                 []*rendering.Texture
	frameDelay, fps          float32
	frameCount, currentFrame int
	paused                   bool
}

func (s *Image) data() *imageData {
	return s.localData.(*imageData)
}

func NewSprite(host *engine.Host, texture *rendering.Texture, text string, anchor Anchor) *Image {
	panel := NewPanel(host, texture, anchor)
	img := (*Image)(panel)
	img.localData = &imageData{
		flipBook: []*rendering.Texture{texture},
	}
	panel.innerUpdate = img.update
	return img
}

func (img *Image) resetDelay() {
	data := img.data()
	data.frameDelay = 1.0 / data.fps
}

func (img *Image) update(deltaTime float64) {
	data := img.data()
	data.frameDelay -= float32(deltaTime)
	if data.frameCount > 0 && data.frameDelay <= 0.0 {
		data.currentFrame++
		if data.currentFrame == data.frameCount {
			data.currentFrame = 0
		}
		if len(data.flipBook) > 0 {
			img.SetTexture(data.flipBook[data.currentFrame])
		}
		// TODO:  Else Atlas animation
		img.resetDelay()
	}
}

func (img *Image) SetTexture(texture *rendering.Texture) {
	(*Panel)(img).SetBackground(texture)
}

func (img *Image) SetFlipBookAnimation(framesPerSecond float32, textures ...*rendering.Texture) {
	data := img.data()
	count := len(textures)
	data.flipBook = make([]*rendering.Texture, 0, count)
	for i := 0; i < count; i++ {
		data.flipBook = append(data.flipBook, textures[i])
	}
	data.frameCount = count
	data.fps = framesPerSecond
	data.currentFrame = 0
	img.resetDelay()
	img.SetTexture(data.flipBook[data.currentFrame])
}

func (img *Image) SetFrameRate(framesPerSecond float32) {
	img.data().fps = framesPerSecond
}

func (img *Image) PlayAnimation() {
	img.data().paused = false
}

func (img *Image) StopAnimation() {
	img.data().paused = true
}
