//go:build !editor

/******************************************************************************/
/* main.test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"kaijuengine.com/bootstrap"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/textures"
)

const rawContentPath = `editor/editor_embedded_content/editor_content`
const gameContentPath = `game_content`

type GameState int

const (
	StateReady GameState = iota
	StatePlaying
	StateGameOver
)

const (
	groundY          = 0.5
	playerZ          = 2.5
	jumpForce        = 12.5
	gravity          = -28.0
	hurdleLength     = 4.6
	hurdleThickness  = 0.32
	hurdleHeightY    = 0.38
	hitThresholdY    = 1.08
	collisionWindow  = 0.28
)

type Game struct {
	host *engine.Host

	// UI
	ui          *ui.Manager
	scoreLabel  *ui.Label
	statusLabel *ui.Label

	// Shared Assets
	matBasic  *rendering.Material
	texSquare *rendering.Texture

	// 3D Entities
	playerBody   *engine.Entity
	playerEyes   [2]*engine.Entity
	playerPupils [2]*engine.Entity
	playerShadow *engine.Entity

	arenaFloor  *engine.Entity
	arenaRing   *engine.Entity
	centerHub   *engine.Entity
	hurdleBar   *engine.Entity
	hurdleBar2  *engine.Entity
	pillars     [4]*engine.Entity

	// Shader Data
	playerSD  *shader_data_registry.ShaderDataStandard
	hurdleSD  *shader_data_registry.ShaderDataStandard
	hurdle2SD *shader_data_registry.ShaderDataStandard

	// Game State
	state        GameState
	score        int
	highScore    int
	jumpCount    int
	gameTime     float32
	statusTimer  float32
	statusText   string

	// Player Physics
	posY        matrix.Float
	velY        matrix.Float
	isGrounded  bool
	squashTimer float32
	deathRot    matrix.Float

	// Hurdle Physics
	sweeperAngle float32
	sweeperSpeed float32
	passedSweep1 bool
	passedSweep2 bool

	// Updates
	updateId   engine.UpdateId
	updateUIId engine.UpdateId
}

func (Game) PluginRegistry() []reflect.Type {
	return []reflect.Type{}
}

func (Game) ContentDatabase() (assets.Database, error) {
	if _, err := os.Stat(gameContentPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := gameCopyEditorContent(); err != nil {
			return nil, err
		}
	}
	return assets.NewFileDatabase(gameContentPath)
}

func (g *Game) Launch(host *engine.Host) {
	g.host = host

	// Set sky clear color: Deep midnight indigo
	host.SetSwapChainClearColor(matrix.NewColor(0.09, 0.11, 0.17, 1.0))

	// Setup perspective 3D camera looking down from front (+Z) at the arena and player
	if cam, ok := host.Cameras.Primary.Camera.(*cameras.StandardCamera); ok {
		cam.SetPosition(matrix.NewVec3(0, 4.6, 6.8))
		cam.SetLookAt(matrix.NewVec3(0, 1.0, 1.0))
	}

	// Add directional light from above-front for rich 3D shading
	lightEntity := engine.NewEntity(host.WorkGroup())
	light := rendering.NewLight(host.Window.GpuInstance.PrimaryDevice(),
		host.AssetDatabase(), host.MaterialCache(), rendering.LightTypeDirectional)
	light.SetDirection(matrix.NewVec3(-0.3, -1.0, -0.4).Normal())
	host.Lighting().Lights.Add(&lightEntity.Transform, light)

	// Cache materials and textures
	var err error
	g.matBasic, err = host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		panic("could not load basic material: " + err.Error())
	}
	g.texSquare, err = host.TextureCache().Texture(assets.TextureSquare, textures.TextureFilterLinear)
	if err != nil {
		panic("could not load texture: " + err.Error())
	}

	// Pre-load meshes
	cubeMesh := rendering.NewMeshCube(host.MeshCache())
	sphereMesh := rendering.NewMeshSphere(host.MeshCache(), 0.5, 32, 32)
	eyeMesh := rendering.NewMeshSphere(host.MeshCache(), 0.12, 16, 16)
	pupilMesh := rendering.NewMeshSphere(host.MeshCache(), 0.06, 16, 16)
	shadowMesh := rendering.NewMeshPlane(host.MeshCache())

	// 1. Build Arena
	// Main outer floor slab
	g.arenaFloor = engine.NewEntity(host.WorkGroup())
	g.arenaFloor.Transform.SetPosition(matrix.NewVec3(0, -0.22, 0))
	g.arenaFloor.Transform.SetScale(matrix.NewVec3(15.0, 0.4, 15.0))
	g.addDrawing(cubeMesh, g.arenaFloor, matrix.NewColor(0.12, 0.15, 0.22, 1.0))

	// Inner play ring/slab
	g.arenaRing = engine.NewEntity(host.WorkGroup())
	g.arenaRing.Transform.SetPosition(matrix.NewVec3(0, -0.20, 0))
	g.arenaRing.Transform.SetScale(matrix.NewVec3(12.5, 0.42, 12.5))
	g.addDrawing(cubeMesh, g.arenaRing, matrix.NewColor(0.16, 0.20, 0.30, 1.0))

	// Center spindle / hub
	g.centerHub = engine.NewEntity(host.WorkGroup())
	g.centerHub.Transform.SetPosition(matrix.NewVec3(0, 0.45, 0))
	g.centerHub.Transform.SetScale(matrix.NewVec3(1.1, 0.9, 1.1))
	g.addDrawing(cubeMesh, g.centerHub, matrix.NewColor(0.28, 0.33, 0.45, 1.0))

	// Corner accent pillars
	pillarCoords := [4][2]matrix.Float{
		{-5.8, -5.8}, {5.8, -5.8}, {-5.8, 5.8}, {5.8, 5.8},
	}
	pillarColors := [4]matrix.Color{
		matrix.NewColor(0.2, 0.8, 0.95, 1.0),
		matrix.NewColor(0.95, 0.6, 0.2, 1.0),
		matrix.NewColor(0.95, 0.6, 0.2, 1.0),
		matrix.NewColor(0.2, 0.8, 0.95, 1.0),
	}
	for i := 0; i < 4; i++ {
		g.pillars[i] = engine.NewEntity(host.WorkGroup())
		g.pillars[i].Transform.SetPosition(matrix.NewVec3(pillarCoords[i][0], 0.7, pillarCoords[i][1]))
		g.pillars[i].Transform.SetScale(matrix.NewVec3(0.5, 1.4, 0.5))
		g.addDrawing(cubeMesh, g.pillars[i], pillarColors[i])
	}

	// 2. Sweeper / Hurdle Bars
	g.hurdleBar = engine.NewEntity(host.WorkGroup())
	g.hurdleBar.Transform.SetScale(matrix.NewVec3(hurdleLength, hurdleThickness, hurdleThickness))
	g.hurdleSD = g.addDrawing(cubeMesh, g.hurdleBar, matrix.NewColor(1.0, 0.35, 0.1, 1.0))

	g.hurdleBar2 = engine.NewEntity(host.WorkGroup())
	g.hurdleBar2.Transform.SetScale(matrix.NewVec3(hurdleLength, hurdleThickness, hurdleThickness))
	g.hurdle2SD = g.addDrawing(cubeMesh, g.hurdleBar2, matrix.NewColor(0.15, 0.9, 1.0, 1.0))

	// 3. Player Character (foreground at Z = playerZ)
	g.playerBody = engine.NewEntity(host.WorkGroup())
	g.playerBody.Transform.SetPosition(matrix.NewVec3(0, groundY, playerZ))
	g.playerSD = g.addDrawing(sphereMesh, g.playerBody, matrix.NewColor(1.0, 0.82, 0.12, 1.0))

	// Player Eyes (facing camera towards +Z)
	g.playerEyes[0] = engine.NewEntity(host.WorkGroup())
	g.playerEyes[0].SetParent(g.playerBody)
	g.playerEyes[0].Transform.SetPosition(matrix.NewVec3(-0.16, 0.14, 0.42))
	g.addDrawing(eyeMesh, g.playerEyes[0], matrix.ColorWhite())

	g.playerEyes[1] = engine.NewEntity(host.WorkGroup())
	g.playerEyes[1].SetParent(g.playerBody)
	g.playerEyes[1].Transform.SetPosition(matrix.NewVec3(0.16, 0.14, 0.42))
	g.addDrawing(eyeMesh, g.playerEyes[1], matrix.ColorWhite())

	// Player Pupils (parented to eyes)
	g.playerPupils[0] = engine.NewEntity(host.WorkGroup())
	g.playerPupils[0].SetParent(g.playerEyes[0])
	g.playerPupils[0].Transform.SetPosition(matrix.NewVec3(0, 0, 0.08))
	g.addDrawing(pupilMesh, g.playerPupils[0], matrix.NewColor(0.06, 0.06, 0.12, 1.0))

	g.playerPupils[1] = engine.NewEntity(host.WorkGroup())
	g.playerPupils[1].SetParent(g.playerEyes[1])
	g.playerPupils[1].Transform.SetPosition(matrix.NewVec3(0, 0, 0.08))
	g.addDrawing(pupilMesh, g.playerPupils[1], matrix.NewColor(0.06, 0.06, 0.12, 1.0))

	// Player Blob Shadow (flat plane on the floor under player)
	g.playerShadow = engine.NewEntity(host.WorkGroup())
	g.playerShadow.Transform.SetPosition(matrix.NewVec3(0, 0.03, playerZ))
	g.playerShadow.Transform.SetScale(matrix.NewVec3(1.1, 0.01, 1.1))
	g.addDrawing(shadowMesh, g.playerShadow, matrix.NewColor(0.05, 0.07, 0.11, 0.8))

	// 4. UI Setup: Stylized HUD Panel
	g.ui = &ui.Manager{}
	g.ui.Init(g.host)

	hudPanel := g.ui.Add().ToPanel()
	hudPanel.Init(g.texSquare, ui.ElementTypePanel)
	hudPanel.SetColor(matrix.NewColor(0.08, 0.10, 0.16, 0.85))
	hudPanel.SetBorderSize(2, 2, 2, 2)
	hudPanel.SetBorderColor(matrix.ColorAqua(), matrix.ColorAqua(), matrix.ColorAqua(), matrix.ColorAqua())
	hudPanel.SetBorderRadius(8, 8, 8, 8)
	hudPanel.Base().Layout().SetPadding(14, 10, 14, 10)
	hudPanel.Base().Layout().SetOffset(24, 24)

	g.scoreLabel = g.ui.Add().ToLabel()
	g.scoreLabel.Init("SCORE: 0    BEST: 0    JUMPS: 0")
	g.scoreLabel.SetFontSize(22)
	g.scoreLabel.SetColor(matrix.ColorGold())
	g.scoreLabel.SetBGColor(matrix.ColorTransparent())
	hudPanel.AddChild(g.scoreLabel.Base())

	g.statusLabel = g.ui.Add().ToLabel()
	g.statusLabel.Init("PRESS [SPACE] OR [CLICK] TO JUMP!")
	g.statusLabel.SetFontSize(18)
	g.statusLabel.SetColor(matrix.ColorAqua())
	g.statusLabel.SetBGColor(matrix.ColorTransparent())
	hudPanel.AddChild(g.statusLabel.Base())

	// Initialize state
	g.resetGame()

	// Register loop updates
	g.updateId = host.Updater.AddUpdate(g.update)
	g.updateUIId = host.UIUpdater.AddUpdate(g.updateUI)

	g.playerBody.OnDestroy.Add(func() {
		host.Updater.RemoveUpdate(&g.updateId)
		host.UIUpdater.RemoveUpdate(&g.updateUIId)
	})

	if engine.LaunchParams.AutoTest {
		host.RunAfterFrames(8, func() {
			g.triggerJump()
		})
		host.RunAfterFrames(35, func() {
			captureScreenshot(host, "jump_game_screenshot.png")
			os.Exit(0)
		})
	}
}

func captureScreenshot(host *engine.Host, path string) {
	var pixels []byte
	var width, height int
	var err error
	host.RunOnRenderThread(func(device *rendering.GPUDevice) {
		pixels, err = device.ScreenshotRGBA()
		if err == nil && len(pixels) > 0 {
			size := device.LogicalDevice.SwapChain.Extent
			width = int(size.X())
			height = int(size.Y())
		}
	})
	if err != nil || len(pixels) == 0 {
		slog.Error("failed to capture screenshot", "error", err)
		return
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	copy(img.Pix, pixels)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err == nil {
		_ = os.WriteFile(path, buf.Bytes(), os.ModePerm)
		slog.Info("Captured game screenshot", "path", path)
	}
}

func (g *Game) addDrawing(mesh *rendering.Mesh, entity *engine.Entity, col matrix.Color) *shader_data_registry.ShaderDataStandard {
	sd := shader_data_registry.Create("basic")
	sds := sd.(*shader_data_registry.ShaderDataStandard)
	sds.Color = col
	draw := rendering.Drawing{
		Material:   g.matBasic.CreateInstance([]*rendering.Texture{g.texSquare}),
		Mesh:       mesh,
		ShaderData: sds,
		Transform:  &entity.Transform,
		ViewCuller: &g.host.Cameras.Primary,
	}
	g.host.Drawings.AddDrawing(draw)
	return sds
}

func (g *Game) resetGame() {
	g.state = StateReady
	g.score = 0
	g.posY = groundY
	g.velY = 0
	g.isGrounded = true
	g.squashTimer = 0
	g.deathRot = 0
	g.sweeperAngle = 0
	g.sweeperSpeed = 2.1
	g.passedSweep1 = false
	g.passedSweep2 = false
	g.statusText = "Press [SPACE] or [CLICK] to Jump!"
	g.playerSD.Color = matrix.NewColor(1.0, 0.82, 0.12, 1.0)
	g.updateHurdlePositions()
}

func (g *Game) triggerJump() {
	if g.state == StateReady {
		g.state = StatePlaying
		g.statusText = "JUMP OVER THE HURDLES!"
		g.statusTimer = 2.0
	} else if g.state == StateGameOver {
		g.resetGame()
		return
	}

	if g.isGrounded {
		g.velY = jumpForce
		g.isGrounded = false
		g.jumpCount++
		g.squashTimer = 0.18 // Stretch on jump
	}
}

func (g *Game) update(deltaTime float64) {
	dt := float32(deltaTime)
	if dt <= 0 || dt > 0.1 {
		dt = 0.016
	}
	g.gameTime += dt

	// Input handling: Space bar, Up arrow, W key, or Mouse Left Click
	kb := &g.host.Window.Keyboard
	mouse := &g.host.Window.Mouse
	if kb.KeyDown(hid.KeyboardKeySpace) || kb.KeyDown(hid.KeyboardKeyUp) ||
		kb.KeyDown(hid.KeyboardKeyW) || mouse.Pressed(hid.MouseButtonLeft) {
		g.triggerJump()
	}

	// Dynamic pillar idle bobbing & rotation
	for i := range g.pillars {
		rot := matrix.Float(g.gameTime*25.0 + float32(i)*45.0)
		g.pillars[i].Transform.SetRotation(matrix.NewVec3(0, rot, 0))
	}

	// Hurdle sweeper update
	if g.state == StatePlaying {
		g.sweeperAngle += g.sweeperSpeed * dt
		if g.sweeperAngle >= 2.0*math.Pi {
			g.sweeperAngle -= 2.0 * math.Pi
		}
	}
	g.updateHurdlePositions()

	// Player vertical physics
	if !g.isGrounded {
		g.velY += gravity * matrix.Float(dt)
		g.posY += g.velY * matrix.Float(dt)
		if g.posY <= groundY {
			g.posY = groundY
			g.velY = 0
			g.isGrounded = true
			g.squashTimer = -0.15 // Squash on landing
		}
	}

	// Squash and stretch scale calculation
	scaleX := matrix.Float(1.0)
	scaleY := matrix.Float(1.0)
	scaleZ := matrix.Float(1.0)

	if g.state == StateGameOver {
		g.deathRot += 220.0 * matrix.Float(dt)
		g.playerBody.Transform.SetRotation(matrix.NewVec3(g.deathRot, 0, g.deathRot*0.5))
	} else {
		if g.squashTimer > 0 { // Jumping stretch
			g.squashTimer -= dt
			scaleY = 1.22
			scaleX = 0.90
			scaleZ = 0.90
		} else if g.squashTimer < 0 { // Landing squash
			g.squashTimer += dt
			scaleY = 0.85
			scaleX = 1.12
			scaleZ = 1.12
		} else { // Idle gentle breathing
			breath := matrix.Float(math.Sin(float64(g.gameTime*4.0))) * 0.03
			scaleY = 1.0 + breath
			scaleX = 1.0 - breath*0.5
			scaleZ = 1.0 - breath*0.5
		}
		g.playerBody.Transform.SetRotation(matrix.Vec3Zero())
	}

	g.playerBody.Transform.SetPosition(matrix.NewVec3(0, g.posY, playerZ))
	g.playerBody.Transform.SetScale(matrix.NewVec3(scaleX, scaleY, scaleZ))

	// Blob Shadow: shrinks and fades when high in air
	heightAboveGround := g.posY - groundY
	shadowScale := max(0.35, 1.1-heightAboveGround*0.28)
	g.playerShadow.Transform.SetScale(matrix.NewVec3(shadowScale, 0.01, shadowScale))

	// Collision & Scoring check
	if g.state == StatePlaying {
		g.checkHurdleCollision()
	}

	// Status text timer
	if g.statusTimer > 0 {
		g.statusTimer -= dt
		if g.statusTimer <= 0 && g.state == StatePlaying {
			g.statusText = ""
		}
	}
}

func (g *Game) updateHurdlePositions() {
	// Hurdle 1: Arm extends from center out to length
	halfL := matrix.Float(hurdleLength * 0.5)
	c1 := matrix.Float(math.Cos(float64(g.sweeperAngle)))
	s1 := matrix.Float(math.Sin(float64(g.sweeperAngle)))
	g.hurdleBar.Transform.SetPosition(matrix.NewVec3(c1*halfL, hurdleHeightY, s1*halfL))
	deg1 := matrix.Float(-g.sweeperAngle * 180.0 / math.Pi)
	g.hurdleBar.Transform.SetRotation(matrix.NewVec3(0, deg1, 0))

	// Hurdle 2: Appears when score >= 8 on the opposite side
	if g.score >= 8 {
		g.hurdleBar2.Transform.SetPosition(matrix.NewVec3(-c1*halfL, hurdleHeightY, -s1*halfL))
		g.hurdleBar2.Transform.SetRotation(matrix.NewVec3(0, deg1+180.0, 0))
	} else {
		// Park under the floor when inactive
		g.hurdleBar2.Transform.SetPosition(matrix.NewVec3(0, -10.0, 0))
	}
}

func (g *Game) checkHurdleCollision() {
	targetAngle := float32(math.Pi * 0.5) // Angle pointing along +Z towards player
	diff1 := math.Abs(float64(normalizeAngle(g.sweeperAngle - targetAngle)))

	// Check Bar 1 collision window
	if diff1 < collisionWindow {
		if g.posY < hitThresholdY {
			g.handleHit()
			return
		}
	} else if g.sweeperAngle > targetAngle && !g.passedSweep1 {
		// Cleared Bar 1!
		g.passedSweep1 = true
		g.onHurdleCleared()
	}

	if g.sweeperAngle < targetAngle-collisionWindow {
		g.passedSweep1 = false
	}

	// Check Bar 2 collision window (opposite angle: 1.5*pi)
	if g.score >= 8 {
		targetAngle2 := float32(math.Pi * 1.5)
		diff2 := math.Abs(float64(normalizeAngle(g.sweeperAngle - targetAngle2)))
		if diff2 < collisionWindow {
			if g.posY < hitThresholdY {
				g.handleHit()
				return
			}
		} else if g.sweeperAngle > targetAngle2 && !g.passedSweep2 {
			g.passedSweep2 = true
			g.onHurdleCleared()
		}

		if g.sweeperAngle < targetAngle2-collisionWindow {
			g.passedSweep2 = false
		}
	}
}

func (g *Game) onHurdleCleared() {
	g.score++
	if g.score > g.highScore {
		g.highScore = g.score
	}

	// Speed up every 3 hurdles cleared
	if g.score%3 == 0 && g.sweeperSpeed < 4.6 {
		g.sweeperSpeed += 0.15
		g.statusText = "SPEED UP! 🔥"
		g.statusTimer = 1.2
	} else {
		cheers := []string{"NICE JUMP! +1", "PERFECT! ⭐", "CLEARED! ✨", "GREAT TIMING! 🎯"}
		g.statusText = cheers[g.score%len(cheers)]
		g.statusTimer = 1.0
	}
}

func (g *Game) handleHit() {
	g.state = StateGameOver
	g.playerSD.Color = matrix.NewColor(1.0, 0.2, 0.2, 1.0) // Flash red
	g.statusText = fmt.Sprintf("💥 BONK! Hit! Final Score: %d — Press [SPACE] to Retry!", g.score)
}

func (g *Game) updateUI(deltaTime float64) {
	g.scoreLabel.SetText(fmt.Sprintf("SCORE: %d     BEST: %d     JUMPS: %d", g.score, g.highScore, g.jumpCount))
	if g.statusText != "" {
		g.statusLabel.SetText(g.statusText)
		if g.state == StateGameOver {
			g.statusLabel.SetColor(matrix.ColorRed())
		} else if g.state == StateReady {
			g.statusLabel.SetColor(matrix.ColorAqua())
		} else {
			g.statusLabel.SetColor(matrix.ColorGreen())
		}
	} else {
		g.statusLabel.SetText("")
	}
}

func normalizeAngle(a float32) float32 {
	for a > math.Pi {
		a -= 2.0 * math.Pi
	}
	for a < -math.Pi {
		a += 2.0 * math.Pi
	}
	return a
}

func getGame() bootstrap.GameInterface { return &Game{} }

func gameCopyEditorContent() error {
	slog.Info("copying stock content to the project database")
	if err := os.MkdirAll(gameContentPath, os.ModePerm); err != nil {
		return err
	}
	rawPath := rawContentPath
	if _, err := os.Stat(rawPath); err != nil {
		rawPath = filepath.Join("src", rawContentPath)
	}
	top, err := os.ReadDir(rawPath)
	if err != nil {
		return err
	}
	all := []string{}
	var readSubDir func(path string) error
	readSubDir = func(path string) error {
		if strings.HasSuffix(path, "renderer/src") {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for i := range entries {
			subPath := filepath.ToSlash(filepath.Join(path, entries[i].Name()))
			if entries[i].IsDir() {
				if err := readSubDir(subPath); err != nil {
					return err
				}
				continue
			}
			all = append(all, subPath)
		}
		return nil
	}
	skip := []string{"editor"}
	for i := range top {
		if !top[i].IsDir() {
			continue
		}
		name := top[i].Name()
		if slices.Contains(skip, name) {
			continue
		}
		if err := readSubDir(filepath.ToSlash(filepath.Join(rawPath, name))); err != nil {
			return err
		}
	}
	for i := range all {
		outPath := filepath.Join(gameContentPath, filepath.Base(all[i]))
		data, err := os.ReadFile(all[i])
		if err != nil {
			return err
		}
		if err := os.WriteFile(outPath, data, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
