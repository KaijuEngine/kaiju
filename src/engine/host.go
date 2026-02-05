/******************************************************************************/
/* host.go                                                                    */
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

package engine

import (
	"kaiju/build"
	"kaiju/debug"
	"kaiju/engine/assets"
	"kaiju/engine/cameras"
	"kaiju/engine/collision_system"
	"kaiju/engine/lighting"
	"kaiju/engine/systems/events"
	"kaiju/engine/systems/logging"
	"kaiju/engine/systems/tweening"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/audio"
	"kaiju/platform/concurrent"
	"kaiju/platform/profiler/tracing"
	"kaiju/platform/windowing"
	"kaiju/plugins"
	"kaiju/rendering"
	"log/slog"
	"math"
	"runtime"
	"sync"
	"time"
	"weak"
)

// FrameId is a unique identifier for a frame
type FrameId = uint64

// InvalidFrameId can be used to indicate that a frame id is invalid
const InvalidFrameId = math.MaxUint64

type frameRun struct {
	frame FrameId
	call  func()
}

type timeRun struct {
	end  time.Time
	call func()
}

type hostCameras struct {
	Primary cameras.Container
	UI      cameras.Container
}

func (c *hostCameras) NewFrame() {
	c.Primary.Camera.NewFrame()
	c.UI.Camera.NewFrame()
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
	name              string
	game              any
	destroyedEntities []*Entity
	lighting          lighting.LightingInformation
	timeRunner        []timeRun
	frameRunner       []frameRun
	preRenderRunner   []func()
	plugins           []*plugins.LuaVM
	Window            *windowing.Window
	LogStream         *logging.LogStream
	workGroup         concurrent.WorkGroup
	threads           concurrent.Threads
	updateThreads     concurrent.Threads
	uiThreads         concurrent.Threads
	Cameras           hostCameras
	collisionManager  collision_system.Manager
	audio             *audio.Audio
	shaderCache       rendering.ShaderCache
	textureCache      rendering.TextureCache
	meshCache         rendering.MeshCache
	fontCache         rendering.FontCache
	materialCache     rendering.MaterialCache
	Drawings          rendering.Drawings
	frame             FrameId
	frameTime         float64
	Closing           bool
	UIUpdater         Updater
	UILateUpdater     Updater
	Updater           Updater
	LateUpdater       Updater
	assetDatabase     assets.Database
	physics           StagePhysics
	OnClose           events.Event
	CloseSignal       chan struct{}
	frameRateLimit    *time.Ticker
	runnerMutex       sync.Mutex
}

// NewHost creates a new host with the given name and log stream. The log stream
// is the log handler that is used by the slog package functions. A Host that
// is created through NewHost has no function until #Host.Initialize is called.
//
// This is primarily called from #host_container/New
func NewHost(name string, logStream *logging.LogStream, assetDb assets.Database) *Host {
	w := float32(DefaultWindowWidth)
	h := float32(DefaultWindowHeight)
	host := &Host{
		name:          name,
		frameTime:     0,
		Closing:       false,
		assetDatabase: assetDb,
		Drawings:      rendering.NewDrawings(),
		CloseSignal:   make(chan struct{}, 1),
		LogStream:     logStream,
		lighting:      lighting.NewLightingInformation(rendering.MaxLocalLights),
		Cameras: hostCameras{
			Primary: cameras.NewContainer(cameras.NewStandardCamera(w, h, w, h, matrix.Vec3Backward())),
			UI:      cameras.NewContainer(cameras.NewStandardCameraOrthographic(w, h, w, h, matrix.Vec3{0, 0, 250})),
		},
	}
	host.workGroup.Init()
	host.threads.Initialize()
	host.updateThreads.Initialize()
	host.uiThreads.Initialize()
	host.Updater = NewConcurrentUpdater(&host.updateThreads)
	host.LateUpdater = NewConcurrentUpdater(&host.updateThreads)
	host.UIUpdater = NewConcurrentUpdater(&host.updateThreads)
	host.UILateUpdater = NewConcurrentUpdater(&host.updateThreads)
	return host
}

// Game will return the primary game mediator for the running application. In
// the editor, this would be *[editor.Editor], in the running game, this will
// be the *game_host.GameHost structure that is generated by the editor and
// filled out by the developer.
func (host *Host) Game() any { return host.game }

// SetGame is to be called by the engine in most cases. It is called by the
// editor when it first starts up to setup the editor game binding. For a game
// generated by the editor, it will be called when the game is bootstrapped
// and provide the *game_host.GameHost structure. You can call this function at
// any time you want, but you really only should need to for special cases.
func (host *Host) SetGame(game any) { host.game = game }

// Initializes the various systems and caches that are mediated through the
// host. This includes the window, the shader cache, the texture cache, the mesh
// cache, and the font cache, and the camera systems.
func (host *Host) Initialize(width, height, x, y int, platformState any) error {
	if width <= 0 {
		width = DefaultWindowWidth
	}
	if height <= 0 {
		height = DefaultWindowHeight
	}
	win, err := windowing.New(host.name, width, height,
		x, y, host.assetDatabase, platformState)
	if err != nil {
		return err
	}
	host.Window = win
	host.threads.Start()
	host.updateThreads.Start()
	host.uiThreads.Start()
	host.Cameras.Primary.Camera.ViewportChanged(float32(width), float32(height))
	host.Cameras.UI.Camera.ViewportChanged(float32(width), float32(height))
	host.shaderCache = rendering.NewShaderCache(host.Window.Renderer, host.assetDatabase)
	host.textureCache = rendering.NewTextureCache(host.Window.Renderer, host.assetDatabase)
	host.meshCache = rendering.NewMeshCache(host.Window.Renderer, host.assetDatabase)
	host.fontCache = rendering.NewFontCache(host.Window.Renderer, host.assetDatabase)
	host.materialCache = rendering.NewMaterialCache(host.Window.Renderer, host.assetDatabase)
	w := weak.Make(host)
	host.Window.OnResize.Add(func() { w.Value().resized() })
	// TODO:  This is tempoarary for testing, it should only be started if a
	// stage has rigidbodies requested to be spawned (issue: #513)
	if !build.Editor {
		host.physics.Start()
	}
	return nil
}

func (host *Host) InitializeRenderer() error {
	w, h := int32(host.Window.Width()), int32(host.Window.Height())
	if err := host.Window.Renderer.Initialize(host, w, h); err != nil {
		slog.Error("failed to initialize the renderer", "error", err)
		return err
	}
	if err := host.FontCache().Init(host.Window.Renderer, host.AssetDatabase(), host); err != nil {
		slog.Error("failed to initialize the font cache", "error", err)
		return err
	}
	if err := rendering.SetupLightMaterials(host.MaterialCache()); err != nil {
		slog.Error("failed to setup the light materials", "error", err)
		return err
	}
	return nil
}

func (host *Host) InitializeAudio() (err error) {
	host.audio, err = audio.New()
	return err
}

// WorkGroup returns the work group for this instance of host
func (host *Host) WorkGroup() *concurrent.WorkGroup { return &host.workGroup }

// Threads returns the long-running threads for this instance of host
func (host *Host) Threads() *concurrent.Threads {
	return &host.threads
}

// Physics returns the stage physics system
func (host *Host) Physics() *StagePhysics { return &host.physics }

// UIThreads returns the long-running threads for the UI
func (host *Host) UIThreads() *concurrent.Threads {
	return &host.uiThreads
}

// Name returns the name of the host
func (host *Host) Name() string { return host.name }

func (host *Host) PrimaryCamera() cameras.Camera {
	return host.Cameras.Primary.Camera
}

func (host *Host) UICamera() cameras.Camera {
	return host.Cameras.UI.Camera
}

// CollisionManager returns the collision manager for this host
func (host *Host) CollisionManager() *collision_system.Manager {
	return &host.collisionManager
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

// MaterialCache returns the font cache for the host
func (host *Host) MaterialCache() *rendering.MaterialCache {
	return &host.materialCache
}

// AssetDatabase returns the asset database for the host
func (host *Host) AssetDatabase() assets.Database {
	return host.assetDatabase
}

// Plugins returns all of the loaded plugins for the host
func (host *Host) Plugins() []*plugins.LuaVM {
	return host.plugins
}

// Audio returns the audio system for the host
func (host *Host) Audio() *audio.Audio {
	return host.audio
}

// Lighting returns a pointer to the internal lighting information
func (host *Host) Lighting() *lighting.LightingInformation {
	return &host.lighting
}

// Update is the main update loop for the host. This will poll the window for
// events, update the entities, and render the scene. This will also check if
// the window has been closed or crashed and set the closing flag accordingly.
//
// The update order is FrameRunner -> Update -> LateUpdate -> EndUpdate:
//
// [-] FrameRunner: Functions added to RunAfterFrames
// [-] UIUpdate: Functions added to UIUpdater
// [-] UILateUpdate: Functions added to UILateUpdater
// [-] Update: Functions added to Updater
// [-] LateUpdate: Functions added to LateUpdater
// [-] EndUpdate: Internal functions for preparing for the next frame
//
// Any destroyed entities will also be ticked for their cleanup. This will also
// tick the editor entities for cleanup.
func (host *Host) Update(deltaTime float64) {
	defer tracing.NewRegion("Host.Update").End()
	host.frame++
	host.Cameras.NewFrame()
	debug.Ensure(deltaTime >= 0)
	host.frameTime += max(0.0, deltaTime)
	host.processDestroyedEntities()
	host.Window.Poll()
	for i := 0; i < len(host.frameRunner); i++ {
		if host.frameRunner[i].frame <= host.frame {
			// TODO:  This shouldn't be needed, see why the [0] sometimes
			// is nil
			if host.frameRunner[i].call != nil {
				host.frameRunner[i].call()
			}
			host.frameRunner = klib.RemoveUnordered(host.frameRunner, i)
			i--
		}
	}
	if len(host.timeRunner) > 0 {
		now := time.Now()
		for i := 0; i < len(host.timeRunner); i++ {
			if host.timeRunner[i].end.Before(now) {
				host.timeRunner[i].call()
				host.timeRunner = klib.RemoveUnordered(host.timeRunner, i)
				i--
			}
		}
	}
	host.UIUpdater.Update(deltaTime)
	host.UILateUpdater.Update(deltaTime)
	tweening.Update(deltaTime)
	host.Updater.Update(deltaTime)
	if !build.Editor {
		if host.physics.IsActive() {
			host.physics.Update(&host.threads, deltaTime)
		}
	}
	host.LateUpdater.Update(deltaTime)
	host.collisionManager.Update(deltaTime)
	if host.Window.IsClosed() || host.Window.IsCrashed() {
		host.Closing = true
	}
	host.Window.EndUpdate()
}

// Render will render the scene. This starts by preparing any drawings that are
// pending. It also creates any pending shaders, textures, and meshes before
// the start of the render. The frame is then readied, buffers swapped, and any
// transformations that are dirty on entities are then cleaned.
func (host *Host) Render() {
	defer tracing.NewRegion("Host.Render").End()
	host.workGroup.Execute(matrix.TransformWorkGroup, &host.threads)
	for _, p := range host.preRenderRunner {
		p()
	}
	host.preRenderRunner = host.preRenderRunner[:0]
	host.Drawings.PreparePending(host.PrimaryCamera().NumCSMCascades())
	host.shaderCache.CreatePending()
	host.textureCache.CreatePending()
	host.meshCache.CreatePending()
	if host.Drawings.HasDrawings() {
		lights := rendering.LightsForRender{
			Lights:     host.lighting.Lights.Cache,
			HasChanges: host.lighting.Lights.HasChanges(),
		}
		for i := 0; i < len(lights.Lights) && !lights.HasChanges; i++ {
			lights.HasChanges = lights.Lights[i].ResetFrameDirty()
		}
		host.lighting.Update(host.PrimaryCamera().Position())
		if host.Window.Renderer.ReadyFrame(host.Window,
			host.Cameras.Primary.Camera, host.Cameras.UI.Camera,
			lights, float32(host.Runtime())) {
			host.Drawings.Render(host.Window.Renderer, lights)
		}
	}
	host.Window.SwapBuffers()
	host.workGroup.Execute(matrix.TransformResetWorkGroup, &host.threads)
}

// Frame will return the current frame id
func (host *Host) Frame() FrameId { return host.frame }

// Runtime will return how long the host has been running in seconds
func (host *Host) Runtime() float64 { return host.frameTime }

// RunAfterFrames will call the given function after the given number of frames
// have passed from the current frame
func (host *Host) RunAfterFrames(wait int, call func()) {
	if call == nil {
		return
	}
	host.runnerMutex.Lock()
	host.frameRunner = append(host.frameRunner, frameRun{
		frame: host.frame + uint64(wait),
		call:  call,
	})
	host.runnerMutex.Unlock()
}

// RunNextFrame will run the given function on the next frame. This is the same
// as calling RunAfterFrames(0, func(){})
func (host *Host) RunNextFrame(call func()) { host.RunAfterFrames(0, call) }

// RunAfterNextUIClean will run the given function on the next frame.
func (host *Host) RunAfterNextUIClean(call func()) {
	// Run after frames happens before the UI update, so doing the same thing
	// as RunNextFrame or RunAfterFrames(0, func(){}) would cause the function
	// to be ran before a UI clean, so we need to effectively wait 2 frames.

	// TODO:  This may change in the future to have something special that runs
	// after the UI update, but this is good enough for now
	host.RunAfterFrames(1, call)
}

func (host *Host) RunBeforeRender(call func()) {
	if call == nil {
		return
	}
	host.runnerMutex.Lock()
	host.preRenderRunner = append(host.preRenderRunner, call)
	host.runnerMutex.Unlock()
}

func (host *Host) RunOnMainThread(call func()) {
	host.RunAfterFrames(0, call)
}

// RunAfterTime will call the given function after the given number of time
// has passed from the current frame
func (host *Host) RunAfterTime(wait time.Duration, call func()) {
	if wait == 0 {
		call()
		return
	}
	host.runnerMutex.Lock()
	host.timeRunner = append(host.timeRunner, timeRun{
		end:  time.Now().Add(wait),
		call: call,
	})
	host.runnerMutex.Unlock()
}

// Teardown will destroy the host and all of its resources. This will also
// execute the OnClose event. This will also signal the CloseSignal channel.
func (host *Host) Teardown() {
	host.Window.Renderer.WaitForRender()
	host.OnClose.Execute()
	host.processDestroyedEntities()
	host.UIUpdater.Destroy()
	host.UILateUpdater.Destroy()
	host.Updater.Destroy()
	host.LateUpdater.Destroy()
	host.Drawings.Destroy(host.Window.Renderer)
	host.textureCache.Destroy()
	host.meshCache.Destroy()
	host.shaderCache.Destroy()
	host.fontCache.Destroy()
	host.materialCache.Destroy()
	host.assetDatabase.Close()
	host.Window.Destroy()
	host.threads.Stop()
	host.updateThreads.Stop()
	host.uiThreads.Stop()
	host.CloseSignal <- struct{}{}
	*host = Host{}
	runtime.GC()
}

// WaitForFrameRate will block until the desired frame rate limit is reached
func (h *Host) WaitForFrameRate() {
	defer tracing.NewRegion("Host.WaitForFrameRate").End()
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
	defer tracing.NewRegion("Host.SetFrameRateLimit").End()
	if fps == 0 {
		if h.frameRateLimit != nil {
			h.frameRateLimit.Stop()
			h.frameRateLimit = nil
		}
	} else {
		h.frameRateLimit = time.NewTicker(time.Second / time.Duration(fps))
	}
}

// Close will set the closing flag to true and signal the host to clean up
// resources and close the window.
func (host *Host) Close() {
	host.Closing = true
}

// DestroyEntity marks the given entity for destruction. The entity will be
// cleaned up at the beginning of the next frame.
func (host *Host) DestroyEntity(entity *Entity) {
	entity.destroy(host)
	host.destroyedEntities = append(host.destroyedEntities, entity)
}

// ImportPlugins will read all of the plugins that are in the specified folder
// and prepare them for execution.
func (host *Host) ImportPlugins(path string) error {
	defer tracing.NewRegion("Host.ImportPlugins").End()
	plugs, err := plugins.LaunchPlugins(host.AssetDatabase(), path)
	if err != nil {
		return err
	}
	host.plugins = append(host.plugins, plugs...)
	return nil
}

func (host *Host) resized() {
	w, h := float32(host.Window.Width()), float32(host.Window.Height())
	host.Cameras.Primary.Camera.ViewportChanged(w, h)
	host.Cameras.UI.Camera.ViewportChanged(w, h)
}

func (host *Host) processDestroyedEntities() {
	for i := range host.destroyedEntities {
		host.destroyedEntities[i].ForceCleanup()
	}
	host.destroyedEntities = klib.WipeSlice(host.destroyedEntities)
}
