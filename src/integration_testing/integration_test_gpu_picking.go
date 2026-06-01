//go:build editor

/******************************************************************************/
/* integration_test_gpu_picking.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"

	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const gpuPickingRenderName = "integration-gpu-picking"

func init() {
	tests["gpu-picking-visible-pixels"] = IntegrationTestGPUPickingVisiblePixels
}

func IntegrationTestGPUPickingVisiblePixels(host *engine.Host) {
	history := &memento.History{}
	history.Initialize(128)
	manager := &editor_stage_manager.StageManager{}
	manager.Initialize(host, history, nil)

	material, err := host.MaterialCache().Material(assets.MaterialDefinitionEditorPicking)
	if err != nil {
		gpuPickingIntegrationFail("load picking material", err)
	}
	back := createPickingCube(host, manager, material, "back-occluded",
		matrix.NewVec3(0, 0, -0.45), matrix.NewVec3(1, 1, 1))
	front := createPickingCube(host, manager, material, "front-visible",
		matrix.NewVec3(0, 0, 0.2), matrix.NewVec3(1.3, 1.3, 1.3))
	side := createPickingCube(host, manager, material, "side-visible",
		matrix.NewVec3(1.8, 0, 0), matrix.NewVec3(0.85, 0.85, 0.85))

	host.PrimaryCamera().SetPositionAndLookAt(
		matrix.NewVec3(0, 0, 5),
		matrix.NewVec3(0, 0, 0),
	)
	target, err := host.RenderTargets.Create(rendering.RenderTargetOptions{
		Name:   gpuPickingRenderName,
		Width:  max(1, host.Window.Width()),
		Height: max(1, host.Window.Height()),
		Depth:  true,
	})
	if err != nil {
		gpuPickingIntegrationFail("create picking render target", err)
	}
	_, err = host.RenderViews.Create(rendering.RenderViewOptions{
		Name:      gpuPickingRenderName,
		Target:    target,
		Camera:    host.PrimaryCamera(),
		LayerMask: rendering.RenderLayerEditorPicking,
		Clear:     true,
		Sort:      -90,
	})
	if err != nil {
		gpuPickingIntegrationFail("create picking render view", err)
	}

	host.RunAfterFrames(12, func() {
		visibleIDs, centerID, err := readVisiblePickingIDs(host, material)
		if err != nil {
			gpuPickingIntegrationFail("read picking attachment", err)
		}
		if centerID != front.PickID {
			gpuPickingIntegrationFail("center click picked wrong entity",
				fmt.Errorf("center pick ID = %d, want front ID %d", centerID, front.PickID))
		}
		centerEntities := manager.EntitiesByPickIDs([]uint32{centerID})
		manager.SelectEntities(centerEntities, editor_stage_manager.SelectionModeReplace)
		assertSelectionOrder(manager, []*editor_stage_manager.StageEntity{front}, "center click")

		visibleEntities := manager.EntitiesByPickIDs(visibleIDs)
		manager.SelectEntities(visibleEntities, editor_stage_manager.SelectionModeReplace)
		assertSelectionOrder(manager, []*editor_stage_manager.StageEntity{front, side}, "box select")
		if manager.IsSelected(back) {
			gpuPickingIntegrationFail("occluded entity selected", nil)
		}
		os.Exit(0)
	})
}

func createPickingCube(host *engine.Host, manager *editor_stage_manager.StageManager,
	material *rendering.Material, name string, position, scale matrix.Vec3) *editor_stage_manager.StageEntity {
	e := manager.AddEntity(name, position)
	e.Transform.SetScale(scale)
	mesh := rendering.NewMeshCube(host.MeshCache())
	e.StageData.Mesh = mesh
	pickID := manager.AssignPickID(e)
	sd := shader_data_registry.Create("editor_pick")
	sd.(*shader_data_registry.ShaderDataEditorPicking).PickID = pickID
	e.StageData.PickingShaderData = sd
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  &e.Transform,
		ViewCuller: &host.Cameras.Primary,
		Layer:      rendering.RenderLayerEditorPicking,
	})
	return e
}

func readVisiblePickingIDs(host *engine.Host, material *rendering.Material) ([]uint32, uint32, error) {
	tex := material.RenderPass().Texture(0)
	if tex == nil || !tex.RenderId.IsValid() {
		return nil, 0, fmt.Errorf("picking texture is not ready")
	}
	w := tex.RenderId.Width
	h := tex.RenderId.Height
	var allData []byte
	var centerData []byte
	var err error
	host.RunOnRenderThread(func(device *rendering.GPUDevice) {
		allData, err = device.TextureReadRegion(tex, matrix.Vec4i{0, 0, int32(w), int32(h)})
		if err != nil {
			return
		}
		centerData, err = device.TextureReadRegion(tex, matrix.Vec4i{int32(w / 2), int32(h / 2), 1, 1})
	})
	if err != nil {
		return nil, 0, err
	}
	centerID := uint32(0)
	if len(centerData) >= 4 {
		centerID = binary.LittleEndian.Uint32(centerData[:4])
	}
	return decodeIntegrationPickIDs(allData), centerID, nil
}

func decodeIntegrationPickIDs(data []byte) []uint32 {
	ids := make([]uint32, 0)
	for i := 0; i+4 <= len(data); i += 4 {
		id := binary.LittleEndian.Uint32(data[i : i+4])
		if id == 0 || containsIntegrationPickID(ids, id) {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func containsIntegrationPickID(ids []uint32, id uint32) bool {
	for i := range ids {
		if ids[i] == id {
			return true
		}
	}
	return false
}

func assertSelectionOrder(manager *editor_stage_manager.StageManager,
	want []*editor_stage_manager.StageEntity, label string) {
	got := manager.Selection()
	if len(got) != len(want) {
		gpuPickingIntegrationFail(label+" selection count",
			fmt.Errorf("got %d, want %d", len(got), len(want)))
	}
	for i := range want {
		if got[i] != want[i] {
			gpuPickingIntegrationFail(label+" selection order",
				fmt.Errorf("entity %d = %s, want %s", i, got[i].Name(), want[i].Name()))
		}
	}
}

func gpuPickingIntegrationFail(message string, err error) {
	if err != nil {
		slog.Error("GPU picking integration test failed", "message", message, "error", err)
	} else {
		slog.Error("GPU picking integration test failed", "message", message)
	}
	os.Exit(1)
}
