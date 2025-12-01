package stage_workspace

import (
	"errors"
	"kaiju/matrix"
	"kaiju/ollama"
	"kaiju/platform/profiler/tracing"
	"weak"
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
