/******************************************************************************/
/* stage.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stages

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"reflect"
	"sort"

	"kaijuengine.com/build"
	"kaijuengine.com/debug"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine/lighting/gi"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

const EntryPointAssetKey = "entryPointStage"

type Stage struct {
	Id                 string
	GlobalIllumination StageGlobalIllumination
	Entities           []EntityDescription
}

type GIContribution uint8

const (
	GIContributionAutomatic GIContribution = iota
	GIContributionExcluded
	GIContributionStatic
	GIContributionRigid
	GIContributionReceivesOnly
)

type GIBakeBoundsMode uint8

const (
	GIBakeBoundsAutomatic GIBakeBoundsMode = iota
	GIBakeBoundsManual
)

type StageGIBakeSettings struct {
	BoundsMode       GIBakeBoundsMode
	ManualCenter     matrix.Vec3
	ManualSize       matrix.Vec3
	BoundsPadding    matrix.Float
	ProbeSpacing     matrix.Float
	RaysPerProbe     uint32
	MaxRayDistance   matrix.Float
	EnvironmentColor matrix.Vec3
}

type StageGlobalIllumination struct {
	OverrideProjectSettings bool
	Settings                gi.Settings
	ProbeAsset              string
	BakeSettings            StageGIBakeSettings
}

func (s *StageGlobalIllumination) Normalize(projectDefaults gi.Settings) {
	effective := projectDefaults
	if s.OverrideProjectSettings {
		effective = s.Settings
	}
	defaults := DefaultStageGIBakeSettings(effective)
	uninitialized := s.BakeSettings.ProbeSpacing == 0 && s.BakeSettings.RaysPerProbe == 0 &&
		s.BakeSettings.MaxRayDistance == 0 && s.BakeSettings.ManualSize.LengthSquared() == 0 &&
		s.BakeSettings.BoundsPadding == 0 && s.BakeSettings.EnvironmentColor.LengthSquared() == 0
	if uninitialized {
		s.BakeSettings = defaults
		return
	}
	if s.BakeSettings.ManualSize.X() <= 0 || s.BakeSettings.ManualSize.Y() <= 0 || s.BakeSettings.ManualSize.Z() <= 0 {
		s.BakeSettings.ManualSize = defaults.ManualSize
	}
	if s.BakeSettings.BoundsPadding < 0 {
		s.BakeSettings.BoundsPadding = defaults.BoundsPadding
	}
	if s.BakeSettings.ProbeSpacing <= 0 {
		s.BakeSettings.ProbeSpacing = defaults.ProbeSpacing
	}
	if s.BakeSettings.RaysPerProbe < 32 {
		s.BakeSettings.RaysPerProbe = defaults.RaysPerProbe
	}
	if s.BakeSettings.MaxRayDistance <= 0 {
		s.BakeSettings.MaxRayDistance = defaults.MaxRayDistance
	}
}

func DefaultStageGIBakeSettings(settings gi.Settings) StageGIBakeSettings {
	spacing := matrix.Float(settings.ProbeSpacing)
	if spacing <= 0 {
		spacing = 2
	}
	maxDistance := matrix.Float(settings.CoverageDistance)
	if maxDistance < 64 {
		maxDistance = 64
	}
	return StageGIBakeSettings{
		BoundsMode:       GIBakeBoundsAutomatic,
		ManualSize:       matrix.NewVec3(32, 16, 32),
		BoundsPadding:    1,
		ProbeSpacing:     spacing,
		RaysPerProbe:     64,
		MaxRayDistance:   maxDistance,
		EnvironmentColor: matrix.NewVec3(0.03, 0.03, 0.03),
	}
}

func (s *Stage) ApplyGlobalIllumination(host *engine.Host) error {
	var override *gi.Settings
	if s.GlobalIllumination.OverrideProjectSettings {
		override = &s.GlobalIllumination.Settings
	}
	return host.GlobalIllumination().ApplyStageSettings(override, s.GlobalIllumination.ProbeAsset)
}

type LoadResult struct {
	Roots        []*engine.Entity
	Entities     []*engine.Entity
	EntitiesById map[engine.EntityId]*engine.Entity
}

type StageJson struct {
	Id                 string
	GlobalIllumination StageGlobalIllumination `json:",omitempty"`
	Meshes             []string                `json:",omitempty"`
	Materials          []string                `json:",omitempty"`
	Textures           []string                `json:",omitempty"`
	Entities           []EntityDescriptionJson `json:",omitempty"`
}

type EntityDescriptionShaderDataField struct {
	Name  string
	Index int32
	Value any
}

type EntityDescription struct {
	Id             string
	TemplateId     string
	Name           string
	Locked         bool
	Mesh           string
	Material       string
	Textures       []string
	Position       matrix.Vec3
	Rotation       matrix.Vec3
	Scale          matrix.Vec3
	DataBinding    []EntityDataBinding
	Children       []EntityDescription
	ShaderData     []EntityDescriptionShaderDataField
	GIContribution GIContribution
	RawDataBinding []any
}

type EntityDescriptionJson struct {
	Id             string
	TemplateId     string
	Name           string
	Locked         bool `json:"omitempty"`
	Mesh           int
	Material       int                     `json:"Mat"`
	Textures       []int                   `json:"Tex,omitempty"`
	Position       matrix.Vec3             `json:"P"`
	Rotation       matrix.Vec3             `json:"R"`
	Scale          matrix.Vec3             `json:"S"`
	DataBinding    []EntityDataBinding     `json:"Data,omitempty"`
	Children       []EntityDescriptionJson `json:"Kids,omitempty"`
	ShaderData     map[string]EntityDescriptionShaderDataField
	GIContribution GIContribution `json:",omitempty"`
}

type EntityDataBinding struct {
	RegistraionKey string
	Fields         map[string]any `json:",omitempty"`
}

func init() {
	pod.Register(Stage{})
	pod.Register(EntityDataBinding{})
	pod.Register(EntityDescription{})
	pod.Register(EntityDescriptionShaderDataField{})
	pod.Register(StageGlobalIllumination{})
	pod.Register(StageGIBakeSettings{})
	pod.Register(GIContribution(0))
	pod.Register(GIBakeBoundsMode(0))
	pod.Register(gi.Settings{})
	pod.Register(gi.Mode(0))
	pod.Register(gi.QualityPreset(0))
	pod.Register(gi.FallbackPolicy(0))
	pod.Register(gi.ContactDetailMode(0))
	pod.Register(gi.DynamicGeometryMode(0))
	pod.Register(gi.EmissiveParticipationMode(0))
}

func debugEnsureStructsMatch() {
	if build.Debug {
		ra := reflect.TypeFor[Stage]()
		rb := reflect.TypeFor[StageJson]()
		debug.Assert(ra.NumField() == rb.NumField()-3,
			"the Stage field has been modified but the matching StageSerialized was not updated")
		ea := reflect.TypeFor[EntityDescription]()
		eb := reflect.TypeFor[EntityDescriptionJson]()
		debug.Assert((ea.NumField()-1) == eb.NumField(), // -1 due to raw data field
			"the EntityDescription field has been modified but the matching EntityDescriptionSerialized was not updated")
	}
}

func (s *Stage) ToMinimized() StageJson {
	debugEnsureStructsMatch()
	ss := StageJson{
		Id:                 s.Id,
		GlobalIllumination: s.GlobalIllumination,
		Entities:           make([]EntityDescriptionJson, len(s.Entities)),
	}
	meshMap := map[string]int{}
	matMap := map[string]int{}
	texMap := map[string]int{}
	// Add a blank string into each for the case that they are not assigned
	meshMap[""] = 0
	matMap[""] = 0
	texMap[""] = 0
	var gatherMappings func(desc *EntityDescription)
	gatherMappings = func(desc *EntityDescription) {
		meshMap[desc.Mesh] = 0
		matMap[desc.Material] = 0
		for j := range desc.Textures {
			texMap[desc.Textures[j]] = 0
		}
		for i := range desc.Children {
			gatherMappings(&desc.Children[i])
		}
	}
	for i := range s.Entities {
		gatherMappings(&s.Entities[i])
	}
	for k := range meshMap {
		meshMap[k] = len(ss.Meshes)
		ss.Meshes = append(ss.Meshes, k)
	}
	for k := range matMap {
		matMap[k] = len(ss.Materials)
		ss.Materials = append(ss.Materials, k)
	}
	for k := range texMap {
		texMap[k] = len(ss.Textures)
		ss.Textures = append(ss.Textures, k)
	}
	var proc func(from *EntityDescription, to *EntityDescriptionJson)
	proc = func(from *EntityDescription, to *EntityDescriptionJson) {
		to.Id = from.Id
		to.TemplateId = from.TemplateId
		to.Name = from.Name
		to.Locked = from.Locked
		to.Position = from.Position
		to.Rotation = from.Rotation
		to.Scale = from.Scale
		to.DataBinding = from.DataBinding
		to.GIContribution = from.GIContribution
		to.Mesh = meshMap[from.Mesh]
		to.Material = matMap[from.Material]
		to.Textures = make([]int, len(from.Textures))
		for i := range from.Textures {
			to.Textures[i] = texMap[from.Textures[i]]
		}
		to.ShaderData = make(map[string]EntityDescriptionShaderDataField)
		for i := range from.ShaderData {
			to.ShaderData[from.ShaderData[i].Name] = from.ShaderData[i]
		}
		to.Children = make([]EntityDescriptionJson, len(from.Children))
		for i := range from.Children {
			proc(&from.Children[i], &to.Children[i])
		}
	}
	for i := range s.Entities {
		proc(&s.Entities[i], &ss.Entities[i])
	}
	return ss
}

func (s *Stage) FromMinimized(ss StageJson) {
	debugEnsureStructsMatch()
	s.Id = ss.Id
	s.GlobalIllumination = ss.GlobalIllumination
	s.Entities = make([]EntityDescription, len(ss.Entities))
	var proc func(from *EntityDescriptionJson, to *EntityDescription)
	proc = func(from *EntityDescriptionJson, to *EntityDescription) {
		to.Id = from.Id
		to.TemplateId = from.TemplateId
		to.Name = from.Name
		to.Locked = from.Locked
		to.Position = from.Position
		to.Rotation = from.Rotation
		to.Scale = from.Scale
		to.DataBinding = from.DataBinding
		to.GIContribution = from.GIContribution
		to.Mesh = ss.Meshes[from.Mesh]
		to.Material = ss.Materials[from.Material]
		to.Textures = make([]string, len(from.Textures))
		for i := range from.Textures {
			to.Textures[i] = ss.Textures[from.Textures[i]]
		}
		for _, v := range from.ShaderData {
			to.ShaderData = append(to.ShaderData, v)
		}
		to.Children = make([]EntityDescription, len(from.Children))
		for i := range from.Children {
			proc(&from.Children[i], &to.Children[i])
		}
	}
	for i := range ss.Entities {
		proc(&ss.Entities[i], &s.Entities[i])
	}
}

func Deserialize(rawData []byte) (Stage, error) {
	if build.Debug && !klib.IsMobile() {
		j := StageJson{}
		if err := json.Unmarshal(rawData, &j); err != nil {
			return Stage{}, err
		}
		s := Stage{}
		s.FromMinimized(j)
		return s, nil
	} else {
		return ArchiveDeserializer(rawData)
	}
}

func ArchiveDeserializer(rawData []byte) (Stage, error) {
	var s Stage
	err := pod.NewDecoder(bytes.NewReader(rawData)).Decode(&s)
	return s, err
}

func EntityDescriptionArchiveDeserializer(rawData []byte) (EntityDescription, error) {
	var desc EntityDescription
	err := pod.NewDecoder(bytes.NewReader(rawData)).Decode(&desc)
	return desc, err
}

func (s *Stage) Load(host *engine.Host) LoadResult {
	res := LoadResult{
		EntitiesById: make(map[engine.EntityId]*engine.Entity),
	}
	type entityBindingInit struct {
		phase engine.EntityDataPhase
		init  func()
	}
	entityBindings := []entityBindingInit{}
	addEntityBinding := func(data engine.EntityData, entity *engine.Entity) {
		entityBindings = append(entityBindings, entityBindingInit{
			phase: engine.EntityDataInitPhase(data),
			init: func() {
				data.Init(entity, host)
			},
		})
	}
	var proc func(se *EntityDescription, parent *engine.Entity)
	proc = func(se *EntityDescription, parent *engine.Entity) {
		e := engine.NewEntity(host.WorkGroup())
		res.Entities = append(res.Entities, e)
		if se.Id != "" {
			id := engine.EntityId(se.Id)
			if host.SetEntityId(e, id) {
				res.EntitiesById[id] = e
			}
		}
		if parent != nil {
			e.SetParent(parent)
		} else {
			res.Roots = append(res.Roots, e)
		}
		e.SetName(se.Name)
		e.Transform.SetPosition(se.Position)
		e.Transform.SetRotation(se.Rotation)
		e.Transform.SetScale(se.Scale)
		// TODO:  Entity data should have been serialized
		if build.Debug {
			for i := range se.DataBinding {
				b, ok := engine.DebugEntityDataRegistry[se.DataBinding[i].RegistraionKey]
				if ok {
					bi := reflect.ValueOf(b).Interface()
					nb := reflect.New(reflect.TypeOf(bi)).Elem()
					for k, v := range se.DataBinding[i].Fields {
						f := nb.FieldByName(k)
						engine.ReflectValueFromJson(v, f)
					}
					reflect.ValueOf(&b).Elem().Set(nb)
					addEntityBinding(b, e)
				} else {
					slog.Error("failed to locate the registered key", "key", se.DataBinding[i].RegistraionKey)
				}
			}
		} else {
			for i := range se.RawDataBinding {
				if data, ok := se.RawDataBinding[i].(engine.EntityData); ok {
					addEntityBinding(data, e)
				} else {
					slog.Error("raw data binding does not implement engine.EntityData",
						"type", reflect.TypeOf(se.RawDataBinding[i]))
				}
			}
		}
		if se.Mesh != "" {
			// TODO:  Handle error?
			SetupEntityFromDescription(e, host, se)
		}
		for i := range se.Children {
			proc(&se.Children[i], e)
		}
		// TODO:  Call the init for bound data after all have been created
	}
	for i := range s.Entities {
		proc(&s.Entities[i], nil)
	}
	sort.SliceStable(entityBindings, func(i, j int) bool {
		return entityBindings[i].phase < entityBindings[j].phase
	})
	for i := range entityBindings {
		entityBindings[i].init()
	}
	return res
}

func SetupEntityFromDescription(e *engine.Entity, host *engine.Host, se *EntityDescription) (*engine.Entity, error) {
	ad := host.AssetDatabase()
	meshId := se.Mesh
	materialId := se.Material
	textureIds := se.Textures
	var km kaiju_mesh.KaijuMesh
	var err error
	var builtIn bool
	meshRef := kaiju_mesh.ParseMeshRef(meshId)
	if meshRef.Key == "" {
		km.Verts, km.Indexes, builtIn = rendering.BuiltInMeshData(meshId)
	}
	if !builtIn {
		km, err = kaiju_mesh.ReadMesh(meshId, host)
	}
	if err != nil {
		slog.Error("failed to deserialize the mesh data", "id", meshId, "error", err)
		return nil, err
	}
	mesh := host.MeshCache().Mesh(meshId, km.Verts, km.Indexes)
	var mat *rendering.Material
	if materialId == "" {
		slog.Warn("no material provided for SpawnMesh, will use fallback material")
		materialId = assets.MaterialDefinitionBasic
	}
	mat, err = host.MaterialCache().Material(materialId)
	if err != nil {
		slog.Error("failed to create the standard material", "error", err)
		return nil, err
	}
	texs := make([]*rendering.Texture, 0, len(textureIds))
	for i := range textureIds {
		texData, err := ad.Read(textureIds[i])
		if err != nil {
			slog.Error("failed to read the texture file", "id", textureIds[i], "error", err)
			return nil, err
		}
		// TODO:  Should be reading the filter from the configuration file
		tex, err := host.TextureCache().InsertRawTexture(textureIds[i],
			texData, 0, 0, rendering.TextureFilterLinear)
		if err != nil {
			slog.Error("failed to create the texture from it's data", "id", textureIds[i], "error", err)
			return nil, err
		}
		texs = append(texs, tex)
	}
	if len(textureIds) == 0 {
		texs = append(texs, mat.Textures...)
	}
	if len(texs) == 0 {
		tex, err := host.TextureCache().Texture(assets.TextureSquare,
			rendering.TextureFilterLinear)
		if err != nil {
			slog.Error("failed to create the default texture", "error", err)
		}
		texs = append(texs, tex)
	}
	mat = mat.CreateInstance(texs)
	sd := shader_data_registry.Create(mat.Shader.DrawInstanceDataName())
	host.RunOnRenderThread(func(device *rendering.GPUDevice) {
		for i := range texs {
			texs[i].DelayedCreate(device)
		}
	})
	draw := rendering.Drawing{
		Material:   mat,
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  &e.Transform,
		ViewCuller: &host.Cameras.Primary,
	}
	e.StoreShaderData(sd)
	// TODO:  Keeping this simple reflection for now so that this is flexible
	// for the future. I want to think this through and not be locked into any
	// one way of doing things.
	if len(se.ShaderData) > 0 {
		v := reflect.ValueOf(sd)
		for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		if build.Debug {
			for i := range se.ShaderData {
				f := v.Field(int(se.ShaderData[i].Index))
				engine.ReflectValueFromJson(se.ShaderData[i].Value, f)
			}
		} else {
			for i := range se.ShaderData {
				f := v.Field(int(se.ShaderData[i].Index))
				f.Set(reflect.ValueOf(se.ShaderData[i].Value))
			}
		}
	}
	host.Drawings.AddDrawing(draw)
	e.OnDestroy.Add(func() { sd.Destroy() })
	return e, nil
}
