/******************************************************************************/
/* host.go                                                                    */
/******************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/******************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/******************************************************************************/

package engine

import (
	"context"
	"kaiju/assets"
	"kaiju/cameras"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/systems/events"
	"kaiju/systems/logging"
	"kaiju/windowing"
	"math"
	"time"
)

type FrameId = uint64

const InvalidFrameId = math.MaxUint64

type Host struct {
	name           string
	editorEntities EditorEntities
	entities       []*Entity
	Window         *windowing.Window
	LogStream      *logging.LogStream
	Camera         cameras.Camera
	UICamera       cameras.Camera
	shaderCache    rendering.ShaderCache
	textureCache   rendering.TextureCache
	meshCache      rendering.MeshCache
	fontCache      rendering.FontCache
	Drawings       rendering.Drawings
	frame          FrameId
	frameTime      float64
	Closing        bool
	Updater        Updater
	LateUpdater    Updater
	assetDatabase  assets.Database
	OnClose        events.Event
	CloseSignal    chan struct{}
	frameRateLimit *time.Ticker
	inEditorEntity int
}

func NewHost(name string, logStream *logging.LogStream) *Host {
	w := float32(DefaultWindowWidth)
	h := float32(DefaultWindowHeight)
	host := &Host{
		name:           name,
		editorEntities: newEditorEntities(),
		entities:       make([]*Entity, 0),
		frameTime:      0,
		Closing:        false,
		Updater:        NewUpdater(),
		LateUpdater:    NewUpdater(),
		assetDatabase:  assets.NewDatabase(),
		Drawings:       rendering.NewDrawings(),
		OnClose:        events.New(),
		CloseSignal:    make(chan struct{}),
		Camera:         cameras.NewStandardCamera(w, h, matrix.Vec3{0, 0, 1}),
		UICamera:       cameras.NewStandardCameraOrthographic(w, h, matrix.Vec3{0, 0, 250}),
		LogStream:      logStream,
	}
	return host
}

func (host *Host) Initialize(width, height int) error {
	if width <= 0 {
		width = DefaultWindowWidth
	}
	if height <= 0 {
		height = DefaultWindowHeight
	}
	win, err := windowing.New(host.name, width, height)
	if err != nil {
		return err
	}
	host.Window = win
	host.Camera.ViewportChanged(float32(width), float32(height))
	host.UICamera.ViewportChanged(float32(width), float32(height))
	host.shaderCache = rendering.NewShaderCache(host.Window.Renderer, &host.assetDatabase)
	host.textureCache = rendering.NewTextureCache(host.Window.Renderer, &host.assetDatabase)
	host.meshCache = rendering.NewMeshCache(host.Window.Renderer, &host.assetDatabase)
	host.fontCache = rendering.NewFontCache(host.Window.Renderer, &host.assetDatabase)
	host.Window.OnResize.Add(host.resized)
	return nil
}

func (host *Host) Name() string { return host.name }

func (host *Host) resized() {
	w, h := float32(host.Window.Width()), float32(host.Window.Height())
	host.Camera.ViewportChanged(w, h)
	host.UICamera.ViewportChanged(w, h)
}

func (host *Host) CreatingEditorEntities() {
	host.inEditorEntity++
}

func (host *Host) DoneCreatingEditorEntities() {
	host.inEditorEntity--
}

func (host *Host) ShaderCache() *rendering.ShaderCache   { return &host.shaderCache }
func (host *Host) TextureCache() *rendering.TextureCache { return &host.textureCache }
func (host *Host) MeshCache() *rendering.MeshCache       { return &host.meshCache }
func (host *Host) FontCache() *rendering.FontCache       { return &host.fontCache }
func (host *Host) AssetDatabase() *assets.Database       { return &host.assetDatabase }

func (host *Host) AddEntity(entity *Entity) {
	host.addEntity(entity)
}

func (host *Host) RemoveEntity(entity *Entity) {
	if host.editorEntities.contains(entity) {
		host.editorEntities.remove(entity)
	} else {
		for i, e := range host.entities {
			if e == entity {
				host.entities = klib.RemoveUnordered(host.entities, i)
				break
			}
		}
	}
}

func (host *Host) AddEntities(entities ...*Entity) {
	host.addEntities(entities...)
}

func (host *Host) Entities() []*Entity { return host.entities }

func (host *Host) NewEntity() *Entity {
	entity := NewEntity()
	host.AddEntity(entity)
	return entity
}

func (host *Host) Update(deltaTime float64) {
	host.frame++
	host.frameTime += deltaTime
	host.Window.Poll()
	host.Updater.Update(deltaTime)
	host.LateUpdater.Update(deltaTime)
	if host.Window.IsClosed() || host.Window.IsCrashed() {
		host.Closing = true
	}
	back := len(host.entities)
	for i, e := range host.entities {
		if e.TickCleanup() {
			host.entities[i] = host.entities[back-1]
			back--
		}
	}
	host.entities = host.entities[:back]
	host.editorEntities.tickCleanup()
	host.Window.EndUpdate()
}

func (host *Host) Render() {
	host.Drawings.PreparePending()
	host.shaderCache.CreatePending()
	host.textureCache.CreatePending()
	host.meshCache.CreatePending()
	host.Window.Renderer.ReadyFrame(host.Camera, host.UICamera, float32(host.Runtime()))
	host.Drawings.Render(host.Window.Renderer)
	host.Window.SwapBuffers()
	// TODO:  Thread this or make the dirty on demand, and have a flag for the dirty frame
	for _, e := range host.entities {
		e.Transform.ResetDirty()
	}
	host.editorEntities.resetDirty()
}

func (host *Host) Frame() FrameId   { return host.frame }
func (host *Host) Runtime() float64 { return host.frameTime }

func (host *Host) Teardown() {
	host.OnClose.Execute()
	host.Updater.Destroy()
	host.LateUpdater.Destroy()
	host.Drawings.Destroy(host.Window.Renderer)
	host.textureCache.Destroy()
	host.meshCache.Destroy()
	host.shaderCache.Destroy()
	host.fontCache.Destroy()
	host.assetDatabase.Destroy()
	host.Window.Destroy()
	host.CloseSignal <- struct{}{}
}

/* context.Context implementation */

func (h *Host) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (h *Host) Done() <-chan struct{} {
	return h.CloseSignal
}

func (h *Host) Err() error {
	if h.Closing {
		return context.Canceled
	}
	return nil
}

func (h *Host) Value(key any) any {
	return nil
}

func (h *Host) WaitForFrameRate() {
	if h.frameRateLimit != nil {
		<-h.frameRateLimit.C
	}
}

func (h *Host) SetFrameRateLimit(fps int64) {
	if fps == 0 {
		h.frameRateLimit.Stop()
		h.frameRateLimit = nil
	} else {
		h.frameRateLimit = time.NewTicker(time.Second / time.Duration(fps))
	}
}

func (host *Host) Close() {
	host.Closing = true
}
