//go:build !editor

/******************************************************************************/
/* main.test.go                                                               */
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

package main

import (
	"kaiju/bootstrap"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
)

const rawContentPath = `editor/editor_embedded_content/editor_content`
const gameContentPath = `game_content`

type Game struct {
	host *engine.Host
	ball *engine.Entity
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
	updateId := host.Updater.AddUpdate(g.update)
	g.ball.OnDestroy.Add(func() {
		sd.Destroy()
		host.Updater.RemoveUpdate(&updateId)
	})
}

func (g *Game) update(deltaTime float64) {
	x := math.Sin(g.host.Runtime())
	g.ball.Transform.SetPosition(matrix.NewVec3(matrix.Float(x), 0, -3))
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
	skip := []string{"editor", "meshes"}
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
