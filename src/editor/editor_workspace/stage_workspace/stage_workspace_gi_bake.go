package stage_workspace

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"kaijuengine.com/editor/editor_overlay/confirm_prompt"
	"kaijuengine.com/editor/editor_overlay/input_prompt"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/lighting/gi"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/engine_entity_data/engine_entity_data_light"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	maxEditorGIProbeCount = uint64(1_000_000)
	expensiveGIBakeRays   = uint64(10_000_000)
	encodedGIProbeBytes   = uint64(33 * 4)
)

type editorGIBakeSnapshot struct {
	input      gi.BakeInput
	probeCount uint64
	totalRays  uint64
	assetBytes uint64
	warnings   []string
}

func (g *WorkspaceGIUI) bake(*document.Element) {
	if g.baking {
		g.setStatus("A GI bake is already running.")
		return
	}
	stageGI := g.current()
	if stageGI.ProbeAsset == "" {
		g.bakeAs(nil)
		return
	}
	name := "Stage GI"
	createNew := false
	if cached, err := g.workspace.ed.Cache().Read(stageGI.ProbeAsset); err == nil {
		name = cached.Config.Name
	} else {
		createNew = true
	}
	g.prepareBake(name, createNew)
}

func (g *WorkspaceGIUI) bakeAs(*document.Element) {
	if g.baking {
		g.setStatus("A GI bake is already running.")
		return
	}
	g.workspace.ed.BlurInterface()
	_, err := input_prompt.Show(g.workspace.Host, input_prompt.Config{
		Title:       "Bake global illumination",
		Description: "Name the baked probe scenario for this stage.",
		Placeholder: "Stage GI",
		Value:       "Stage GI",
		ConfirmText: "Bake",
		CancelText:  "Cancel",
		OnConfirm: func(name string) {
			g.workspace.ed.FocusInterface()
			name = strings.TrimSpace(name)
			if name == "" {
				g.setStatus("A GI probe asset name is required.")
				return
			}
			g.prepareBake(name, true)
		},
		OnCancel: g.workspace.ed.FocusInterface,
	})
	if err != nil {
		g.workspace.ed.FocusInterface()
		g.setStatus(err.Error())
	}
}

func (g *WorkspaceGIUI) prepareBake(name string, createNew bool) {
	snapshot, err := g.captureBakeSnapshot(name)
	if err != nil {
		g.setStatus("Bake preflight failed: " + err.Error())
		return
	}
	start := func() { g.runBake(name, createNew, snapshot) }
	if snapshot.totalRays > expensiveGIBakeRays {
		g.workspace.ed.BlurInterface()
		confirm_prompt.Show(g.workspace.Host, confirm_prompt.Config{
			Title:       "Expensive GI bake",
			Description: fmt.Sprintf("This bake will trace approximately %d million rays. Continue?", snapshot.totalRays/1_000_000),
			ConfirmText: "Bake",
			CancelText:  "Cancel",
			OnConfirm:   func() { g.workspace.ed.FocusInterface(); start() },
			OnCancel:    g.workspace.ed.FocusInterface,
		})
		return
	}
	start()
}

func (g *WorkspaceGIUI) runBake(name string, createNew bool, snapshot editorGIBakeSnapshot) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancel = cancel
	g.baking = true
	g.setBakeProgress(0, 1)
	g.setStatus(fmt.Sprintf("Baking %d probes (%.1f MB)...", snapshot.probeCount, float64(snapshot.assetBytes)/(1024*1024)))
	lastProgress := time.Time{}
	snapshot.input.Progress = func(completed, total int) {
		now := time.Now()
		if completed != total && now.Sub(lastProgress) < 100*time.Millisecond {
			return
		}
		lastProgress = now
		g.workspace.Host.RunOnMainThread(func() {
			g.setBakeProgress(completed, total)
			g.setStatus(fmt.Sprintf("Baking GI: %d/%d probes (%.0f%%)", completed, total, float64(completed)*100/float64(total)))
		})
	}
	go func() {
		asset, err := gi.BakeProbes(ctx, snapshot.input)
		var data []byte
		if err == nil {
			data, err = asset.MarshalBinary()
		}
		g.workspace.Host.RunOnMainThread(func() {
			g.finishBake(name, createNew, data, err, snapshot.warnings)
		})
	}()
}

func (g *WorkspaceGIUI) cancelBake(*document.Element) {
	if g.cancel != nil {
		g.cancel()
		g.setStatus("Cancelling GI bake...")
	}
}

func (g *WorkspaceGIUI) finishBake(name string, createNew bool, data []byte, bakeErr error, warnings []string) {
	g.baking = false
	g.hideBakeProgress()
	g.cancel = nil
	if bakeErr != nil {
		if errors.Is(bakeErr, context.Canceled) {
			g.setStatus("GI bake cancelled; the previous probe asset was preserved.")
		} else {
			g.setStatus("GI bake failed: " + bakeErr.Error())
		}
		return
	}
	if _, err := gi.UnmarshalProbeAsset(data); err != nil {
		g.setStatus("GI bake validation failed: " + err.Error())
		return
	}
	stageGI := g.current()
	assetID := stageGI.ProbeAsset
	if createNew || assetID == "" {
		ids := content_database.ImportRaw(name, data, content_database.GIProbe{}, g.workspace.ed.ProjectFileSystem(), g.workspace.ed.Cache())
		if len(ids) != 1 {
			g.setStatus("Failed to import the baked GI probe asset.")
			return
		}
		assetID = ids[0]
		g.workspace.ed.Events().OnContentAdded.Execute(ids)
	} else if err := g.replaceProbeAsset(assetID, data); err != nil {
		g.setStatus("Failed to replace the GI probe asset: " + err.Error())
		return
	} else {
		g.workspace.ed.Events().OnContentChangesSaved.Execute(assetID)
	}
	stageGI.ProbeAsset = assetID
	g.apply(stageGI, createNew || g.current().ProbeAsset != assetID)
	var override *gi.Settings
	if stageGI.OverrideProjectSettings {
		override = &stageGI.Settings
	}
	if err := g.workspace.Host.GlobalIllumination().ApplyStageSettings(override, assetID); err != nil {
		g.setStatus("Baked, but preview failed: " + err.Error())
		return
	}
	message := "GI bake complete and previewing."
	if len(warnings) > 0 {
		message += " Warnings: " + strings.Join(warnings, "; ")
	}
	g.setStatus(message)
	g.syncInputs()
}

func (g *WorkspaceGIUI) replaceProbeAsset(id string, data []byte) error {
	cached, err := g.workspace.ed.Cache().Read(id)
	if err != nil {
		return err
	}
	path := content_database.ToContentPath(cached.Path)
	temporary := path + ".bake.tmp"
	fs := g.workspace.ed.ProjectFileSystem()
	if err := fs.WriteFile(temporary, data, os.ModePerm); err != nil {
		return err
	}
	if err := fs.Rename(temporary, path); err != nil {
		fs.Remove(temporary)
		return err
	}
	cached.Config.SrcPath = ""
	if err := content_database.WriteConfig(cached.Path, cached.Config, fs); err != nil {
		return err
	}
	g.workspace.ed.Cache().IndexCachedContent(cached)
	return nil
}

func (g *WorkspaceGIUI) captureBakeSnapshot(name string) (editorGIBakeSnapshot, error) {
	stageGI := g.current()
	stageGI.Normalize(g.workspace.ed.Project().Settings.GlobalIllumination)
	if err := validateStageGIBakeSettings(stageGI.BakeSettings); err != nil {
		return editorGIBakeSnapshot{}, err
	}
	triangles, bounds, geometryHash, warnings := g.captureBakeGeometry()
	if len(triangles) == 0 {
		return editorGIBakeSnapshot{}, errors.New("the stage has no contributed static mesh or terrain triangles")
	}
	if stageGI.BakeSettings.BoundsMode == stages.GIBakeBoundsManual {
		bounds = graviton.NewAABB(stageGI.BakeSettings.ManualCenter, stageGI.BakeSettings.ManualSize.Scale(0.5))
	} else {
		padding := matrix.NewVec3(stageGI.BakeSettings.BoundsPadding, stageGI.BakeSettings.BoundsPadding, stageGI.BakeSettings.BoundsPadding)
		bounds = graviton.NewAABB(bounds.Center, bounds.Extent.Add(padding))
	}
	geometryHash = hashBakeConfiguration(geometryHash, bounds, stageGI.BakeSettings)
	directional, points, spots, lightingHash := g.captureBakeLights(stageGI.BakeSettings)
	dimensions := bakeGridDimensions(bounds.Size(), stageGI.BakeSettings.ProbeSpacing)
	probeCount := uint64(dimensions[0]) * uint64(dimensions[1]) * uint64(dimensions[2])
	assetBytes := probeCount*encodedGIProbeBytes + 256
	settings := g.effectiveSettings(stageGI)
	memoryLimitMB := uint64(settings.MemoryBudgetMB)
	if memoryLimitMB == 0 {
		memoryLimitMB = 256
	}
	if probeCount > maxEditorGIProbeCount {
		return editorGIBakeSnapshot{}, fmt.Errorf("estimated probe count %d exceeds the editor limit %d; increase spacing or reduce bounds", probeCount, maxEditorGIProbeCount)
	}
	if assetBytes > memoryLimitMB*1024*1024 {
		return editorGIBakeSnapshot{}, fmt.Errorf("estimated %.1f MB probe asset exceeds the %d MB GI memory budget", float64(assetBytes)/(1024*1024), memoryLimitMB)
	}
	input := gi.BakeInput{
		Bounds:            bounds,
		ProbeSpacing:      stageGI.BakeSettings.ProbeSpacing,
		RaysPerProbe:      stageGI.BakeSettings.RaysPerProbe,
		MaxRayDistance:    stageGI.BakeSettings.MaxRayDistance,
		Scenario:          name,
		Environment:       stageGI.BakeSettings.EnvironmentColor,
		Triangles:         triangles,
		DirectionalLights: directional,
		PointLights:       points,
		SpotLights:        spots,
		GeometryHash:      geometryHash,
		LightingHash:      lightingHash,
	}
	return editorGIBakeSnapshot{input: input, probeCount: probeCount, totalRays: probeCount * uint64(input.RaysPerProbe), assetBytes: assetBytes, warnings: warnings}, nil
}

func bakeGridDimensions(size matrix.Vec3, spacing matrix.Float) [3]uint32 {
	return [3]uint32{
		max(2, uint32(math.Ceil(float64(size.X()/spacing)))+1),
		max(2, uint32(math.Ceil(float64(size.Y()/spacing)))+1),
		max(2, uint32(math.Ceil(float64(size.Z()/spacing)))+1),
	}
}

func (g *WorkspaceGIUI) captureBakeGeometry() ([]gi.BakeTriangle, graviton.AABB, [32]byte, []string) {
	triangles := make([]gi.BakeTriangle, 0)
	points := make([]matrix.Vec3, 0)
	warnings := make([]string, 0)
	for _, entity := range g.workspace.stageView.Manager().List() {
		contribution := entity.StageData.Description.GIContribution
		if contribution == stages.GIContributionExcluded || contribution == stages.GIContributionRigid || contribution == stages.GIContributionReceivesOnly {
			continue
		}
		albedo, emissive, supported := bakeMaterialValues(entity.StageData.ShaderData)
		if !supported && entity.StageData.Description.Mesh != "" {
			warnings = append(warnings, entity.Name()+" uses an unsupported bake material; neutral albedo was used")
		}
		world := entity.Transform.WorldMatrix()
		if meshID := entity.StageData.Description.Mesh; meshID != "" && entity.StageData.ShaderData != nil && entity.StageData.ShaderData.SkinningHeader() == nil {
			mesh, _, ok := g.workspace.meshById(meshID)
			if ok {
				appendBakeMesh(&triangles, &points, mesh.Verts, mesh.Indexes, world, albedo, emissive)
			}
		}
		for _, data := range entity.NamedData("Terrain") {
			if model, ok := data.(*terrain.Terrain); ok {
				verts, indices := model.BakeMeshSnapshot()
				appendBakeMesh(&triangles, &points, verts, indices, world, albedo, emissive)
			}
		}
	}
	if len(points) == 0 {
		return triangles, graviton.AABB{}, sha256.Sum256(nil), warnings
	}
	hash := hashBakeTriangles(triangles)
	return triangles, graviton.AABBFromPoints(points), hash, warnings
}

func appendBakeMesh(out *[]gi.BakeTriangle, points *[]matrix.Vec3, verts []rendering.Vertex, indices []uint32, world matrix.Mat4, albedo, emissive matrix.Vec3) {
	for i := 0; i+2 < len(indices); i += 3 {
		if int(indices[i]) >= len(verts) || int(indices[i+1]) >= len(verts) || int(indices[i+2]) >= len(verts) {
			continue
		}
		triangle := gi.BakeTriangle{Albedo: albedo, Emissive: emissive}
		triangle.Points[0] = world.TransformPoint(verts[indices[i]].Position)
		triangle.Points[1] = world.TransformPoint(verts[indices[i+1]].Position)
		triangle.Points[2] = world.TransformPoint(verts[indices[i+2]].Position)
		*out = append(*out, triangle)
		*points = append(*points, triangle.Points[:]...)
	}
}

func bakeMaterialValues(data rendering.DrawInstance) (matrix.Vec3, matrix.Vec3, bool) {
	color := func(c matrix.Color) matrix.Vec3 { return matrix.NewVec3(c[0], c[1], c[2]) }
	neutral := matrix.NewVec3(0.8, 0.8, 0.8)
	switch value := data.(type) {
	case *shader_data_registry.ShaderDataStandard:
		return color(value.Color), matrix.Vec3Zero(), true
	case *shader_data_registry.ShaderDataStandardSkinned:
		return color(value.Color), matrix.Vec3Zero(), true
	case *shader_data_registry.ShaderDataPBR:
		albedo := color(value.VertColors)
		return albedo, albedo.Scale(value.MeRoEmAo[2]), true
	case *shader_data_registry.ShaderDataPbrSkinned:
		albedo := color(value.VertColors)
		return albedo, albedo.Scale(value.MeRoEmAo[2]), true
	case *shader_data_registry.ShaderDataTerrain:
		return color(value.Color), matrix.Vec3Zero(), true
	default:
		return neutral, matrix.Vec3Zero(), false
	}
}

func (g *WorkspaceGIUI) captureBakeLights(settings stages.StageGIBakeSettings) ([]gi.BakeDirectionalLight, []gi.BakePointLight, []gi.BakeSpotLight, [32]byte) {
	directional := make([]gi.BakeDirectionalLight, 0)
	points := make([]gi.BakePointLight, 0)
	spots := make([]gi.BakeSpotLight, 0)
	for _, entity := range g.workspace.stageView.Manager().List() {
		for _, binding := range entity.DataBindingsByKey(engine_entity_data_light.BindingKey()) {
			data, ok := binding.BoundData.(*engine_entity_data_light.LightEntityData)
			if !ok || data == nil {
				continue
			}
			intensity := data.Diffuse.Scale(matrix.Float(data.Intensity))
			switch data.Type {
			case engine_entity_data_light.LightTypeDirectional:
				directional = append(directional, gi.BakeDirectionalLight{TowardLight: entity.Transform.Up(), Radiance: intensity})
			case engine_entity_data_light.LightTypePoint:
				points = append(points, gi.BakePointLight{Position: entity.Transform.WorldPosition(), Intensity: intensity, Range: settings.MaxRayDistance})
			case engine_entity_data_light.LightTypeSpot:
				spots = append(spots, gi.BakeSpotLight{Position: entity.Transform.WorldPosition(), Direction: entity.Transform.Up().Negative(), Intensity: intensity, Range: settings.MaxRayDistance, InnerCutoff: matrix.Float(data.Cutoff), OuterCutoff: matrix.Float(data.OuterCutoff)})
			}
		}
	}
	return directional, points, spots, hashBakeLights(settings.EnvironmentColor, directional, points, spots)
}

func hashBakeTriangles(triangles []gi.BakeTriangle) [32]byte {
	h := sha256.New()
	for i := range triangles {
		for _, point := range triangles[i].Points {
			writeHashVec3(h, point)
		}
		writeHashVec3(h, triangles[i].Albedo)
		writeHashVec3(h, triangles[i].Emissive)
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

func hashBakeConfiguration(base [32]byte, bounds graviton.AABB, settings stages.StageGIBakeSettings) [32]byte {
	h := sha256.New()
	h.Write(base[:])
	writeHashVec3(h, bounds.Center)
	writeHashVec3(h, bounds.Extent)
	writeHashFloat(h, settings.ProbeSpacing)
	writeHashFloat(h, matrix.Float(settings.RaysPerProbe))
	writeHashFloat(h, settings.MaxRayDistance)
	var mode [1]byte
	mode[0] = byte(settings.BoundsMode)
	h.Write(mode[:])
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

func hashBakeLights(environment matrix.Vec3, directional []gi.BakeDirectionalLight, points []gi.BakePointLight, spots []gi.BakeSpotLight) [32]byte {
	h := sha256.New()
	writeHashVec3(h, environment)
	for _, light := range directional {
		writeHashVec3(h, light.TowardLight)
		writeHashVec3(h, light.Radiance)
	}
	for _, light := range points {
		writeHashVec3(h, light.Position)
		writeHashVec3(h, light.Intensity)
		writeHashFloat(h, light.Range)
	}
	for _, light := range spots {
		writeHashVec3(h, light.Position)
		writeHashVec3(h, light.Direction)
		writeHashVec3(h, light.Intensity)
		writeHashFloat(h, light.Range)
		writeHashFloat(h, light.InnerCutoff)
		writeHashFloat(h, light.OuterCutoff)
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

type hashWriter interface{ Write([]byte) (int, error) }

func writeHashVec3(h hashWriter, value matrix.Vec3) {
	for _, component := range value {
		writeHashFloat(h, component)
	}
}
func writeHashFloat(h hashWriter, value matrix.Float) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], math.Float32bits(float32(value)))
	h.Write(data[:])
}
