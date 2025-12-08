/******************************************************************************/
/* stage_spawner_llm.go                                                       */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package stage_workspace

import (
	"errors"
	"weak"

	"github.com/KaijuEngine/kaiju/matrix"
	"github.com/KaijuEngine/kaiju/ollama"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
)

type llmCreateCameraAction struct {
	ollama.LLMActionResultBase
	w        weak.Pointer[StageWorkspace]
	Position matrix.Vec3 `json:"position" desc:"The position to spawn the camera at"`
	LookAt   matrix.Vec3 `json:"lookAt" desc:"The point to make the camera look at"`
}

func (a llmCreateCameraAction) Execute() (any, error) {
	defer tracing.NewRegion("StageWorkspace.createCamera").End()
	w := a.w.Value()
	cam, ok := w.CreateNewCamera()
	if !ok {
		return nil, errors.New("failed to create the camera for some reason")
	}
	cam.Transform.SetPosition(a.Position)
	cam.Transform.LookAt(a.LookAt)
	return a, nil
}

func (w *StageWorkspace) initLLMActions() {
	a := llmCreateCameraAction{w: weak.Make(w)}
	ollama.ReflectFuncToOllama(a,
		"CreateCamera",
		"Creates a new camera at the given position with the given look at point.")
}
