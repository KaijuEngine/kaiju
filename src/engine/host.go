/******************************************************************************/
/* host.go                                                                    */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package engine

import (
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

// FrameId is a unique identifier for a frame
type FrameId = uint64

// InvalidFrameId can be used to indicate that a frame id is invalid
const InvalidFrameId = math.MaxUint64

type frameRun struct {
	frame FrameId
	call  func()
}

// Host is the mediator to the entire runtime for the game/editor. It is the
// main entry point for the game loop and is responsible for managing all
// entities, the window, and the rendering context. The host can be used to
// create and manage entities, call update functions on the main thread, and
// access various caches and resources.
//
// The host is expected to be passed around quite often throughout the program.
// It is designed to remove things like service locators, singletons, and other
// global state. You can have multiple hosts in a program to isolate things like
// windows and game state.
type Host struct {
	name           string
	editorEntities editorEntities
	entities       []*Entity
	entityLookup   map[string]*Entity
	frameRunner    []frameRun
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

// NewHost creates a new host with the given name and log stream. The log stream
// is the log handler that is used by the slog package functions. A Host that
// is created through NewHost has no function until #Host.Initialize is called.
//
// This is primarily called from #host_container/New
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
		CloseSignal:    make(chan struct{}, 1),
		Camera:         cameras.NewStandardCamera(w, h, matrix.Vec3Backward()),
		UICamera:       cameras.NewStandardCameraOrthographic(w, h, matrix.Vec3{0, 0, 250}),
		LogStream:      logStream,
		frameRunner:    make([]frameRun, 0),
		entityLookup:   make(map[string]*Entity),
	}
	return host
}

// Initializes the various systems and caches that are mediated through the
// host. This includes the window, the shader cache, the texture cache, the mesh
// cache, and the font cache, and the camera systems.
func (host *Host) Initialize(width, height, x, y int) error {
	if width <= 0 {
		width = DefaultWindowWidth
	}
	if height <= 0 {
		height = DefaultWindowHeight
	}
	win, err := windowing.New(host.name, width, height, x, y)
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

// Name returns the name of the host
func (host *Host) Name() string { return host.name }

// CreatingEditorEntities is used exclusively for the editor to know that the
// entities that are being created are for the editor. This is used to logically
// separate editor entities from game entities.
//
// This will increment so it can be called many times, however it is expected
// that #Host.DoneCreatingEditorEntities is be called the same number of times.
func (host *Host) CreatingEditorEntities() {
	host.inEditorEntity++
}

// DoneCreatingEditorEntities is used to signal that the editor is done creating
// entities. This should be called the same number of times as
// #Host.CreatingEditorEntities. When the internal counter reaches 0, then any
// entity created on the host will go to the standard entity pool.
func (host *Host) DoneCreatingEditorEntities() {
	host.inEditorEntity--
}

// ShaderCache returns the shader cache for the host
func (host *Host) ShaderCache() *rendering.ShaderCache {
	return &host.shaderCache
}

// TextureCache returns the texture cache for the host
func (host *Host) TextureCache() *rendering.TextureCache {
	return &host.textureCache
}

// MeshCache returns the mesh cache for the host
func (host *Host) MeshCache() *rendering.MeshCache {
	return &host.meshCache
}

// FontCache returns the font cache for the host
func (host *Host) FontCache() *rendering.FontCache {
	return &host.fontCache
}

// AssetDatabase returns the asset database for the host
func (host *Host) AssetDatabase() *assets.Database {
	return &host.assetDatabase
}

// AddEntity adds an entity to the host. This will add the entity to the
// standard entity pool. If the host is in the process of creating editor
// entities, then the entity will be added to the editor entity pool.
func (host *Host) AddEntity(entity *Entity) {
	host.addEntity(entity)
}

// ClearEntities will remove all entities from the host. This will remove all
// entities from the standard entity pool only. The entities will be destroyed
// using the standard destroy method, so they will take not be fully removed
// during the frame that this function was called.
func (host *Host) ClearEntities() {
	for _, e := range host.entities {
		e.Destroy()
	}
}

// RemoveEntity removes an entity from the host. This will remove the entity
// from the standard entity pool. This will determine if the entity is in the
// editor entity pool and remove it from there if so, otherwise it will be
// removed from the standard entity pool. Entities are not ordered, so they are
// removed in O(n) time. Do not assume the entities are ordered at any time.
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

// AddEntities adds multiple entities to the host. This will add the entities
// using the same rules as AddEntity. If the host is in the process of creating
// editor entities, then the entities will be added to the editor entity pool.
func (host *Host) AddEntities(entities ...*Entity) {
	host.addEntities(entities...)
}

// FindEntity will search for an entity contained in this host by its id. If the
// entity is found, then it will return the entity and true, otherwise it will
// return nil and false.
func (host *Host) FindEntity(id string) (*Entity, bool) {
	e, ok := host.entityLookup[id]
	return e, ok
}

// Entities returns all the entities that are currently in the host. This will
// return all entities in the standard entity pool only.
func (host *Host) Entities() []*Entity { return host.entities }

// NewEntity creates a new entity and adds it to the host. This will add the
// entity to the standard entity pool. If the host is in the process of creating
// editor entities, then the entity will be added to the editor entity pool.
func (host *Host) NewEntity() *Entity {
	entity := NewEntity()
	host.AddEntity(entity)
	return entity
}

// Update is the main update loop for the host. This will poll the window for
// events, update the entities, and render the scene. This will also check if
// the window has been closed or crashed and set the closing flag accordingly.
//
// The update order is FrameRunner -> Update -> LateUpdate -> EndUpdate:
//
// [-] FrameRunner: Functions added to RunAfterFrames
// [-] Update: Functions added to Updater
// [-] LateUpdate: Functions added to LateUpdater
// [-] EndUpdate: Internal functions for preparing for the next frame
//
// Any destroyed entities will also be ticked for their cleanup. This will also
// tick the editor entities for cleanup.
func (host *Host) Update(deltaTime float64) {
	host.frame++
	host.frameTime += deltaTime
	host.Window.Poll()
	for i := 0; i < len(host.frameRunner); i++ {
		if host.frameRunner[i].frame == host.frame {
			host.frameRunner[i].call()
			host.frameRunner = klib.RemoveUnordered(host.frameRunner, i)
			i--
		}
	}
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

// Render will render the scene. This starts by preparing any drawings that are
// pending. It also creates any pending shaders, textures, and meshes before
// the start of the render. The frame is then readied, buffers swapped, and any
// transformations that are dirty on entities are then cleaned.
func (host *Host) Render() {
	host.Drawings.PreparePending()
	host.shaderCache.CreatePending()
	host.textureCache.CreatePending()
	host.meshCache.CreatePending()
	if host.Drawings.HasDrawings() {
		host.Window.Renderer.ReadyFrame(host.Camera,
			host.UICamera, float32(host.Runtime()))
		host.Drawings.Render(host.Window.Renderer)
	}
	host.Window.SwapBuffers()
	// TODO:  Thread this or make the dirty on demand, and have a flag for the dirty frame
	for _, e := range host.entities {
		e.Transform.ResetDirty()
	}
	host.editorEntities.resetDirty()
}

// Frame will return the current frame id
func (host *Host) Frame() FrameId { return host.frame }

// Runtime will return how long the host has been running in seconds
func (host *Host) Runtime() float64 { return host.frameTime }

// RunAfterFrames will call the given function after the given number of frames
// have passed from the current frame
func (host *Host) RunAfterFrames(wait int, call func()) {
	host.frameRunner = append(host.frameRunner, frameRun{
		frame: host.frame + uint64(wait),
		call:  call,
	})
}

// Teardown will destroy the host and all of its resources. This will also
// execute the OnClose event. This will also signal the CloseSignal channel.
func (host *Host) Teardown() {
	host.Window.Renderer.WaitForRender()
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

// WaitForFrameRate will block until the desired frame rate limit is reached
func (h *Host) WaitForFrameRate() {
	if h.frameRateLimit != nil {
		<-h.frameRateLimit.C
	}
}

// SetFrameRateLimit will set the frame rate limit for the host. If the frame
// rate is set to 0, then the frame rate limit will be removed.
//
// If a frame rate is set, then the host will block until the desired frame rate
// is reached before continuing the update loop.
func (h *Host) SetFrameRateLimit(fps int64) {
	if fps == 0 {
		h.frameRateLimit.Stop()
		h.frameRateLimit = nil
	} else {
		h.frameRateLimit = time.NewTicker(time.Second / time.Duration(fps))
	}
}

// Close will set the closing flag to true and signal the host to clean up
// resources and close the window.
func (host *Host) Close() {
	host.Closing = true
}

func (host *Host) resized() {
	w, h := float32(host.Window.Width()), float32(host.Window.Height())
	host.Camera.ViewportChanged(w, h)
	host.UICamera.ViewportChanged(w, h)
}
