/*****************************************************************************/
/* image.go                                                                  */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

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

func NewImage(host *engine.Host, texture *rendering.Texture, anchor Anchor) *Image {
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
