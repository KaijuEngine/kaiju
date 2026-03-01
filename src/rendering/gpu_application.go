/******************************************************************************/
/* gpu_application.go                                                         */
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

package rendering

import (
	"kaijuengine.com/engine/assets"
	"log/slog"
)

const (
	engineVersionMajor = int(1)
	engineVersionMinor = int(0)
	engineVersionPatch = int(0)
)

type GPUApplication struct {
	Name      string
	Version   GPUApplicationVersion
	Instances []*GPUApplicationInstance
}

type GPUApplicationVersion struct {
	Major int
	Minor int
	Patch int
}

func (g *GPUApplication) Setup(name string, version GPUApplicationVersion) {
	g.Name = name
	g.Version = version
}

func (g *GPUApplication) IsValid() bool { return len(g.Instances) > 0 }

func (g *GPUApplication) FirstInstance() *GPUApplicationInstance {
	return g.Instances[0]
}

func (g *GPUApplication) Instance(index int) (*GPUApplicationInstance, bool) {
	if index < 0 || index > len(g.Instances) {
		slog.Error("index out of range for the instances", "has", len(g.Instances), "wants", index)
		return nil, false
	}
	return g.Instances[index], true
}

func (g *GPUApplication) ApplicationVersion() (major int, minor int, patch int) {
	return g.Version.Major, g.Version.Minor, g.Version.Patch
}

func (g *GPUApplication) EngineVersion() (major int, minor int, patch int) {
	return engineVersionMajor, engineVersionMinor, engineVersionPatch
}

func (g *GPUApplication) CreateInstance(window RenderingContainer, assets assets.Database) (*GPUApplicationInstance, error) {
	slog.Info("creating kaiju gpu instance")
	g.Instances = append(g.Instances, &GPUApplicationInstance{})
	if err := g.Instances[len(g.Instances)-1].Initialize(window, g, assets); err != nil {
		return nil, err
	}
	return g.Instances[len(g.Instances)-1], nil
}

func (g *GPUApplication) Destroy() {
	for i := range g.Instances {
		g.Instances[i].Destroy()
	}
}
