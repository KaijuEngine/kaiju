//go:build !editor

/******************************************************************************/
/* main.test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"fmt"
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
	"kaijuengine.com/engine/systems/console"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders"
)

const rawContentPath = `editor/editor_embedded_content/editor_content`
const gameContentPath = `game_content`

type Game struct {
	host  *engine.Host
	ball  *engine.Entity
	cube  *engine.Entity
	ui    *ui.Manager
	label *ui.Label
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
	// TODO:  The world is your oyster
	g.host = host
	g.ui = &ui.Manager{}
	g.ui.Init(g.host)
	sphere := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)
	sd := shader_data_registry.Create("basic")
	g.ball = engine.NewEntity(host.WorkGroup())
	sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorRed()
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		panic("you've probably got the wrong asset database path")
	}
	tex, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err != nil {
		panic("you've probably got the wrong asset database path")
	}
	draw := rendering.Drawing{
		Material:   mat.CreateInstance([]*rendering.Texture{tex}),
		Mesh:       sphere,
		ShaderData: sd,
		Transform:  &g.ball.Transform,
		ViewCuller: &host.Cameras.Primary,
	}
	host.Drawings.AddDrawing(draw)

	res, err := loaders.OBJ("cube.obj", host.AssetDatabase())
	if err != nil {
		panic("could not read cube.obj")
	}
	meshCache := host.MeshCache()
	slog.Info("read cube.obj", "verts", len(res.Meshes[0].Verts), "indexes", len(res.Meshes[0].Indexes))
	cube := meshCache.Mesh("cube", res.Meshes[0].Verts, res.Meshes[0].Indexes)
	meshCache.AddMesh(cube)
	g.cube = engine.NewEntity(host.WorkGroup())
	g.cube.Transform.SetPosition(matrix.NewVec3(2, 0, -5))
	g.cube.Transform.SetRotation(matrix.NewVec3(-20, 20, 20))
	sd = shader_data_registry.Create("cube")
	sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorBlue()
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   mat.CreateInstance([]*rendering.Texture{tex}),
		Mesh:       cube,
		ShaderData: sd,
		Transform:  &g.cube.Transform,
		ViewCuller: &host.Cameras.Primary,
	})

	updateId := host.Updater.AddUpdate(g.update)
	g.ball.OnDestroy.Add(func() {
		sd.Destroy()
		host.Updater.RemoveUpdate(&updateId)
	})
	g.label = g.ui.Add().ToLabel()
	g.label.Init("FPS: -")
	g.label.SetColor(matrix.ColorAqua())
	g.label.SetBGColor(matrix.ColorTransparent())
	p := g.ui.Add().ToPanel()
	p.Init(tex, ui.ElementTypePanel)
	p.SetColor(matrix.ColorTransparent())
	p.AddChild(g.label.Base())

	updateUIID := g.host.UIUpdater.AddUpdate(g.updateUI)
	g.label.Base().Entity().OnDestroy.Add(func() {
		host.UIUpdater.RemoveUpdate(&updateUIID)
	})
}

func (g *Game) updateUI(deltaTime float64) {
	g.label.SetText(fmt.Sprintf("FPS: %.2f", 1/deltaTime))
}

func (g *Game) update(deltaTime float64) {
	x := math.Sin(g.host.Runtime())
	console.For(g.host).Write(g.ball.Transform.Position().String()) // Open console in game (debug flag) with F1
	g.ball.Transform.SetPosition(matrix.NewVec3(matrix.Float(x), 0, -3))
	g.cube.Transform.SetRotation(g.cube.Transform.Rotation().Add(matrix.NewVec3(0, 0, 10).Scale(float32(deltaTime))))
}

func getGame() bootstrap.GameInterface { return &Game{} }

func gameCopyEditorContent() error {
	slog.Info("copying stock content to the project database")
	if err := os.MkdirAll(gameContentPath, os.ModePerm); err != nil {
		return err
	}
	top, err := os.ReadDir(rawContentPath)
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
		if err := readSubDir(filepath.ToSlash(filepath.Join(rawContentPath, name))); err != nil {
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
