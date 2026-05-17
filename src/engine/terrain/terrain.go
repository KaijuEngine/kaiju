/******************************************************************************/
/* terrain.go                                                                 */
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

package terrain

import (
	"errors"
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	defaultResolution        = 33
	defaultChunkSize         = 32
	defaultWorldSize         = matrix.Float(100)
	defaultMinHeight         = matrix.Float(-100)
	defaultMaxHeight         = matrix.Float(100)
	defaultRayStep           = matrix.Float(0.5)
	defaultBrushSpacingScale = matrix.Float(0.25)
)

type TerrainTexture struct {
	Key    string
	Filter rendering.TextureFilter
}

type TerrainConfig struct {
	Resolution    int
	WorldSize     matrix.Vec2
	MinHeight     matrix.Float
	MaxHeight     matrix.Float
	InitialHeight matrix.Float
	ChunkSize     int
	Material      string
	Textures      []TerrainTexture
	ShaderData    string
}

type DirtyRegion struct {
	MinX, MinZ int
	MaxX, MaxZ int
	Valid      bool
}

func (r DirtyRegion) Expand(padding, resolution int) DirtyRegion {
	if !r.Valid || resolution <= 0 {
		return DirtyRegion{}
	}
	minX := max(0, r.MinX-padding)
	minZ := max(0, r.MinZ-padding)
	maxX := min(resolution-1, r.MaxX+padding)
	maxZ := min(resolution-1, r.MaxZ+padding)
	if minX > maxX || minZ > maxZ {
		return DirtyRegion{}
	}
	return DirtyRegion{
		MinX:  minX,
		MinZ:  minZ,
		MaxX:  maxX,
		MaxZ:  maxZ,
		Valid: true,
	}
}

func (r DirtyRegion) Intersects(other DirtyRegion) bool {
	if !r.Valid || !other.Valid {
		return false
	}
	return r.MinX <= other.MaxX && r.MaxX >= other.MinX &&
		r.MinZ <= other.MaxZ && r.MaxZ >= other.MinZ
}

type HeightField struct {
	Resolution int
	Heights    []matrix.Float
	MinHeight  matrix.Float
	MaxHeight  matrix.Float

	dirty DirtyRegion
}

func NewHeightField(resolution int, minHeight, maxHeight, initialHeight matrix.Float) (*HeightField, error) {
	if resolution < 2 {
		return nil, errors.New("terrain heightfield resolution must be at least 2")
	}
	if minHeight > maxHeight {
		return nil, errors.New("terrain heightfield min height cannot be greater than max height")
	}
	h := &HeightField{
		Resolution: resolution,
		Heights:    make([]matrix.Float, resolution*resolution),
		MinHeight:  minHeight,
		MaxHeight:  maxHeight,
	}
	initialHeight = h.clampHeight(initialHeight)
	for i := range h.Heights {
		h.Heights[i] = initialHeight
	}
	h.markDirty(0, 0, resolution-1, resolution-1)
	return h, nil
}

func (h *HeightField) DirtyRegion() DirtyRegion { return h.dirty }
func (h *HeightField) ClearDirty()              { h.dirty = DirtyRegion{} }

func (h *HeightField) CopyRegion(region DirtyRegion) []matrix.Float {
	if !region.Valid {
		return nil
	}
	region = region.Expand(0, h.Resolution)
	width := region.MaxX - region.MinX + 1
	height := region.MaxZ - region.MinZ + 1
	out := make([]matrix.Float, width*height)
	for z := region.MinZ; z <= region.MaxZ; z++ {
		src := h.index(region.MinX, z)
		dst := (z - region.MinZ) * width
		copy(out[dst:dst+width], h.Heights[src:src+width])
	}
	return out
}

func (h *HeightField) SetRegion(region DirtyRegion, heights []matrix.Float) DirtyRegion {
	if !region.Valid {
		return DirtyRegion{}
	}
	region = region.Expand(0, h.Resolution)
	width := region.MaxX - region.MinX + 1
	height := region.MaxZ - region.MinZ + 1
	if len(heights) != width*height {
		return DirtyRegion{}
	}
	var dirty DirtyRegion
	for z := region.MinZ; z <= region.MaxZ; z++ {
		for x := region.MinX; x <= region.MaxX; x++ {
			idx := (x - region.MinX) + (z-region.MinZ)*width
			if h.SetHeight(x, z, heights[idx]) {
				dirty = mergeDirtyRegions(dirty, DirtyRegion{
					MinX: x, MinZ: z, MaxX: x, MaxZ: z, Valid: true,
				})
			}
		}
	}
	return dirty
}

func (h *HeightField) Height(x, z int) matrix.Float {
	if !h.inBounds(x, z) {
		return 0
	}
	return h.Heights[h.index(x, z)]
}

func (h *HeightField) SetHeight(x, z int, height matrix.Float) bool {
	if !h.inBounds(x, z) {
		return false
	}
	height = h.clampHeight(height)
	idx := h.index(x, z)
	if h.Heights[idx] == height {
		return false
	}
	h.Heights[idx] = height
	h.markDirty(x, z, x, z)
	return true
}

func (h *HeightField) AddHeight(x, z int, delta matrix.Float) bool {
	return h.SetHeight(x, z, h.Height(x, z)+delta)
}

func (h *HeightField) Sample(x, z matrix.Float) matrix.Float {
	if x < 0 {
		x = 0
	} else if x > matrix.Float(h.Resolution-1) {
		x = matrix.Float(h.Resolution - 1)
	}
	if z < 0 {
		z = 0
	} else if z > matrix.Float(h.Resolution-1) {
		z = matrix.Float(h.Resolution - 1)
	}
	x0 := int(matrix.Floor(x))
	z0 := int(matrix.Floor(z))
	x1 := min(x0+1, h.Resolution-1)
	z1 := min(z0+1, h.Resolution-1)
	tx := x - matrix.Float(x0)
	tz := z - matrix.Float(z0)
	h00 := h.Height(x0, z0)
	h10 := h.Height(x1, z0)
	h01 := h.Height(x0, z1)
	h11 := h.Height(x1, z1)
	return matrix.Lerp(matrix.Lerp(h00, h10, tx), matrix.Lerp(h01, h11, tx), tz)
}

func (h *HeightField) Paint(stroke PaintStroke) DirtyRegion {
	return paintHeightField(h, stroke)
}

func (h *HeightField) inBounds(x, z int) bool {
	return x >= 0 && z >= 0 && x < h.Resolution && z < h.Resolution
}

func (h *HeightField) index(x, z int) int { return x + z*h.Resolution }

func (h *HeightField) clampHeight(height matrix.Float) matrix.Float {
	return matrix.Clamp(height, h.MinHeight, h.MaxHeight)
}

func (h *HeightField) markDirty(minX, minZ, maxX, maxZ int) {
	minX = max(0, min(minX, h.Resolution-1))
	minZ = max(0, min(minZ, h.Resolution-1))
	maxX = max(0, min(maxX, h.Resolution-1))
	maxZ = max(0, min(maxZ, h.Resolution-1))
	if minX > maxX || minZ > maxZ {
		return
	}
	if !h.dirty.Valid {
		h.dirty = DirtyRegion{MinX: minX, MinZ: minZ, MaxX: maxX, MaxZ: maxZ, Valid: true}
		return
	}
	h.dirty.MinX = min(h.dirty.MinX, minX)
	h.dirty.MinZ = min(h.dirty.MinZ, minZ)
	h.dirty.MaxX = max(h.dirty.MaxX, maxX)
	h.dirty.MaxZ = max(h.dirty.MaxZ, maxZ)
}

type BrushMode int

const (
	BrushRaise BrushMode = iota
	BrushLower
	BrushSmooth
)

type BrushFalloff int

const (
	FalloffLinear BrushFalloff = iota
	FalloffSmooth
	FalloffConstant
)

type PaintStroke struct {
	Mode     BrushMode
	Center   matrix.Vec2
	Radius   matrix.Float
	Strength matrix.Float
	Falloff  BrushFalloff
	Spacing  matrix.Float
}

type TerrainChunk struct {
	Key        string
	StartX     int
	StartZ     int
	EndX       int
	EndZ       int
	Mesh       *rendering.Mesh
	Drawing    rendering.Drawing
	ShaderData rendering.DrawInstance
	Indexes    []uint32
}

type Terrain struct {
	Config      TerrainConfig
	Entity      *engine.Entity
	Transform   *matrix.Transform
	HeightField *HeightField
	MeshChunks  []TerrainChunk
	Material    *rendering.Material
	ShaderData  []rendering.DrawInstance

	host *engine.Host
}

type TerrainRayHit struct {
	Point      matrix.Vec3
	LocalPoint matrix.Vec3
	Normal     matrix.Vec3
	Distance   matrix.Float
}

func NewModel(config TerrainConfig) (*Terrain, error) {
	return newTerrainWithHeights(config, nil, nil, nil, nil)
}

func New(host *engine.Host, config TerrainConfig) (*Terrain, error) {
	defer tracing.NewRegion("terrain.New").End()
	if host == nil {
		return NewModel(config)
	}
	return newTerrainWithHeights(config, nil, host.WorkGroup(), host, nil)
}

func (t *Terrain) Destroy(host *engine.Host) {
	defer tracing.NewRegion("terrain.Destroy").End()
	for i := range t.ShaderData {
		t.ShaderData[i].Destroy()
	}
	if host != nil && t.Entity != nil {
		host.DestroyEntity(t.Entity)
	}
}

func (t *Terrain) Collision() (*graviton.TerrainCollision, error) {
	if t == nil || t.HeightField == nil {
		return nil, errors.New("terrain collision requires a terrain heightfield")
	}
	collision := t.NewCollision()
	if collision == nil {
		return nil, errors.New("terrain collision could not be created")
	}
	return collision, nil
}

// NewCollision creates a Graviton terrain collider backed by the terrain height
// storage. Height edits are visible to the collider immediately; do not mutate
// terrain heights concurrently with a physics step.
func (t *Terrain) NewCollision() *graviton.TerrainCollision {
	if t == nil || t.HeightField == nil {
		return nil
	}
	collision, err := graviton.NewTerrainCollision(
		t.HeightField.Resolution,
		t.Config.WorldSize,
		t.HeightField.Heights,
		t.HeightField.MinHeight,
		t.HeightField.MaxHeight,
	)
	if err != nil {
		return nil
	}
	return collision
}

// CollisionBounds returns the local-space bounds used by terrain collision.
func (t *Terrain) CollisionBounds() graviton.AABB {
	if t == nil || t.HeightField == nil {
		return graviton.NewAABB(matrix.Vec3Zero(), matrix.Vec3Zero())
	}
	minPoint := matrix.NewVec3(
		-t.Config.WorldSize.X()*0.5,
		t.HeightField.MinHeight,
		-t.Config.WorldSize.Y()*0.5,
	)
	maxPoint := matrix.NewVec3(
		t.Config.WorldSize.X()*0.5,
		t.HeightField.MaxHeight,
		t.Config.WorldSize.Y()*0.5,
	)
	return graviton.AABBFromMinMax(minPoint, maxPoint)
}

func (t *Terrain) HeightAtLocal(localXZ matrix.Vec2) matrix.Float {
	x, z := t.localToGrid(localXZ)
	return t.HeightField.Sample(x, z)
}

func (t *Terrain) HeightAtWorld(point matrix.Vec3) matrix.Float {
	local := t.Transform.InverseWorldMatrix().TransformPoint(point)
	local.SetY(t.HeightAtLocal(local.XZ()))
	return t.Transform.WorldMatrix().TransformPoint(local).Y()
}

func (t *Terrain) Paint(stroke PaintStroke) DirtyRegion {
	dirty := t.HeightField.Paint(t.localStrokeToGrid(stroke))
	t.ApplyDirty()
	return dirty
}

func (t *Terrain) PaintLine(from, to matrix.Vec2, stroke PaintStroke) DirtyRegion {
	var merged DirtyRegion
	t.VisitPaintLineStamps(from, to, stroke, func(stamp PaintStroke) bool {
		dirty := t.Paint(stamp)
		if dirty.Valid {
			merged = mergeDirtyRegions(merged, dirty)
		}
		return true
	})
	return merged
}

func (t *Terrain) VisitPaintLineStamps(from, to matrix.Vec2, stroke PaintStroke, visit func(PaintStroke) bool) {
	if visit == nil {
		return
	}
	delta := to.Subtract(from)
	distance := delta.Length()
	spacing := stroke.Spacing
	if spacing <= 0 {
		spacing = matrix.Max(stroke.Radius*defaultBrushSpacingScale, matrix.Tiny)
	}
	if distance <= matrix.Tiny {
		stroke.Center = from
		visit(stroke)
		return
	}
	steps := max(1, int(matrix.Ceil(distance/spacing)))
	for i := 0; i <= steps; i++ {
		stroke.Center = matrix.Vec2Lerp(from, to, matrix.Float(i)/matrix.Float(steps))
		if !visit(stroke) {
			return
		}
	}
}

func (t *Terrain) ApplyDirty() {
	dirty := t.HeightField.DirtyRegion()
	if !dirty.Valid || t.host == nil {
		return
	}
	vertexDirty := dirty.Expand(1, t.HeightField.Resolution)
	for i := range t.MeshChunks {
		chunkRegion := DirtyRegion{
			MinX:  t.MeshChunks[i].StartX,
			MinZ:  t.MeshChunks[i].StartZ,
			MaxX:  t.MeshChunks[i].EndX,
			MaxZ:  t.MeshChunks[i].EndZ,
			Valid: true,
		}
		if !vertexDirty.Intersects(chunkRegion) {
			continue
		}
		verts := t.buildChunkVertices(&t.MeshChunks[i])
		t.host.MeshCache().UpdateMeshVertices(t.MeshChunks[i].Key, verts)
	}
	t.HeightField.ClearDirty()
}

func (t *Terrain) ApplyHeightRegion(region DirtyRegion, heights []matrix.Float) DirtyRegion {
	dirty := t.HeightField.SetRegion(region, heights)
	t.ApplyDirty()
	return dirty
}

func (t *Terrain) StrokeRegion(stroke PaintStroke) DirtyRegion {
	return strokeDirtyRegion(t.HeightField, t.localStrokeToGrid(stroke))
}

func (t *Terrain) RayHit(ray graviton.Ray) (TerrainRayHit, bool) {
	inv := t.Transform.InverseWorldMatrix()
	localOrigin := inv.TransformPoint(ray.Origin)
	localTarget := inv.TransformPoint(ray.Origin.Add(ray.Direction))
	localDirection := localTarget.Subtract(localOrigin)
	if localDirection.IsZero() {
		return TerrainRayHit{}, false
	}
	localRay := graviton.Ray{
		Origin:    localOrigin,
		Direction: localDirection.Normal(),
	}
	hit, ok := t.RayHitLocal(localRay)
	if !ok {
		return TerrainRayHit{}, false
	}
	hit.Point = t.Transform.WorldMatrix().TransformPoint(hit.LocalPoint)
	hit.Distance = hit.Point.Distance(ray.Origin)
	return hit, true
}

func (t *Terrain) RayHitLocal(ray graviton.Ray) (TerrainRayHit, bool) {
	defer tracing.NewRegion("terrain.RayHitLocal").End()
	if ray.Direction.IsZero() {
		return TerrainRayHit{}, false
	}
	ray.Direction = ray.Direction.Normal()
	minBounds := matrix.NewVec3(
		-t.Config.WorldSize.X()*0.5,
		t.HeightField.MinHeight,
		-t.Config.WorldSize.Y()*0.5,
	)
	maxBounds := matrix.NewVec3(
		t.Config.WorldSize.X()*0.5,
		t.HeightField.MaxHeight,
		t.Config.WorldSize.Y()*0.5,
	)
	entry, exit, ok := rayBounds(ray, minBounds, maxBounds)
	if !ok {
		return TerrainRayHit{}, false
	}
	cell := matrix.Min(
		t.Config.WorldSize.X()/matrix.Float(t.HeightField.Resolution-1),
		t.Config.WorldSize.Y()/matrix.Float(t.HeightField.Resolution-1),
	)
	step := matrix.Max(cell*defaultRayStep, matrix.Tiny)
	if entry < 0 {
		entry = 0
	}
	lastDistance := entry
	lastPoint := ray.Point(float32(lastDistance))
	lastDelta := lastPoint.Y() - t.HeightAtLocal(lastPoint.XZ())
	if lastDelta <= 0 {
		return t.localHit(lastPoint, ray.Origin), true
	}
	for distance := entry + step; distance <= exit+matrix.Tiny; distance += step {
		point := ray.Point(float32(matrix.Min(distance, exit)))
		delta := point.Y() - t.HeightAtLocal(point.XZ())
		if delta <= 0 {
			hitPoint := t.refineRayHit(ray, lastDistance, matrix.Min(distance, exit))
			return t.localHit(hitPoint, ray.Origin), true
		}
		lastDistance = distance
		lastPoint = point
		lastDelta = delta
	}
	_ = lastDelta
	return TerrainRayHit{}, false
}

func (t *Terrain) refineRayHit(ray graviton.Ray, low, high matrix.Float) matrix.Vec3 {
	for i := 0; i < 12; i++ {
		mid := (low + high) * 0.5
		point := ray.Point(float32(mid))
		if point.Y() > t.HeightAtLocal(point.XZ()) {
			low = mid
		} else {
			high = mid
		}
	}
	return ray.Point(float32(high))
}

func (t *Terrain) localHit(localPoint, rayOrigin matrix.Vec3) TerrainRayHit {
	return TerrainRayHit{
		Point:      localPoint,
		LocalPoint: localPoint,
		Normal:     t.normalAtLocal(localPoint.XZ()),
		Distance:   localPoint.Distance(rayOrigin),
	}
}

func (t *Terrain) normalAtLocal(localXZ matrix.Vec2) matrix.Vec3 {
	x, z := t.localToGrid(localXZ)
	cellX := t.Config.WorldSize.X() / matrix.Float(t.HeightField.Resolution-1)
	cellZ := t.Config.WorldSize.Y() / matrix.Float(t.HeightField.Resolution-1)
	left := t.HeightField.Sample(x-1, z)
	right := t.HeightField.Sample(x+1, z)
	front := t.HeightField.Sample(x, z-1)
	back := t.HeightField.Sample(x, z+1)
	dx := (right - left) / (cellX * 2)
	dz := (back - front) / (cellZ * 2)
	return matrix.NewVec3(-dx, 1, -dz).Normal()
}

func newTerrainWithHeights(config TerrainConfig, heights []matrix.Float, workGroup *concurrent.WorkGroup, host *engine.Host, entity *engine.Entity) (*Terrain, error) {
	config = normalizeConfig(config)
	hf, err := NewHeightField(config.Resolution, config.MinHeight, config.MaxHeight, config.InitialHeight)
	if err != nil {
		return nil, err
	}
	if heights != nil {
		if len(heights) != len(hf.Heights) {
			return nil, fmt.Errorf("terrain expected %d heights, got %d", len(hf.Heights), len(heights))
		}
		for i := range heights {
			hf.Heights[i] = hf.clampHeight(heights[i])
		}
		hf.markDirty(0, 0, hf.Resolution-1, hf.Resolution-1)
	}
	if entity == nil {
		entity = engine.NewEntity(workGroup)
	}
	entity.SetName("Terrain")
	t := &Terrain{
		Config:      config,
		Entity:      entity,
		Transform:   &entity.Transform,
		HeightField: hf,
		MeshChunks:  make([]TerrainChunk, 0),
		ShaderData:  make([]rendering.DrawInstance, 0),
		host:        host,
	}
	if host != nil {
		if err := t.createRenderResources(host); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func normalizeConfig(config TerrainConfig) TerrainConfig {
	if config.Resolution < 2 {
		config.Resolution = defaultResolution
	}
	if config.WorldSize.X() <= 0 {
		config.WorldSize.SetX(defaultWorldSize)
	}
	if config.WorldSize.Y() <= 0 {
		config.WorldSize.SetY(defaultWorldSize)
	}
	if config.MinHeight == 0 && config.MaxHeight == 0 {
		config.MinHeight = defaultMinHeight
		config.MaxHeight = defaultMaxHeight
	}
	if config.MinHeight > config.MaxHeight {
		config.MinHeight, config.MaxHeight = config.MaxHeight, config.MinHeight
	}
	config.InitialHeight = matrix.Clamp(config.InitialHeight, config.MinHeight, config.MaxHeight)
	if config.ChunkSize <= 0 {
		config.ChunkSize = defaultChunkSize
	}
	if config.Material == "" {
		config.Material = assets.MaterialDefinitionTerrain
	}
	if config.ShaderData == "" || (config.Material == assets.MaterialDefinitionTerrain && config.ShaderData == "basic") {
		config.ShaderData = "terrain"
	}
	if len(config.Textures) == 0 {
		config.Textures = []TerrainTexture{{Key: assets.TextureSquare, Filter: rendering.TextureFilterLinear}}
	}
	for i := range config.Textures {
		if config.Textures[i].Filter < 0 || config.Textures[i].Filter >= rendering.TextureFilterMax {
			config.Textures[i].Filter = rendering.TextureFilterLinear
		}
	}
	return config
}

func (t *Terrain) SetBrushPreview(centerXZ matrix.Vec2, radius, ringWidth matrix.Float, color matrix.Color) {
	for i := range t.ShaderData {
		if sd, ok := t.ShaderData[i].(*shader_data_registry.ShaderDataTerrain); ok {
			sd.SetBrush(centerXZ, radius, ringWidth, color)
		}
	}
}

func (t *Terrain) ClearBrushPreview() {
	for i := range t.ShaderData {
		if sd, ok := t.ShaderData[i].(*shader_data_registry.ShaderDataTerrain); ok {
			sd.ClearBrush()
		}
	}
}

func (t *Terrain) createRenderResources(host *engine.Host) error {
	material, err := host.MaterialCache().Material(t.Config.Material)
	if err != nil {
		return err
	}
	textures := make([]*rendering.Texture, len(t.Config.Textures))
	for i := range t.Config.Textures {
		textures[i], err = host.TextureCache().Texture(t.Config.Textures[i].Key, t.Config.Textures[i].Filter)
		if err != nil {
			return err
		}
	}
	t.Material = material.CreateInstance(textures)
	t.createChunks(host)
	t.HeightField.ClearDirty()
	return nil
}

func (t *Terrain) createChunks(host *engine.Host) {
	cells := t.HeightField.Resolution - 1
	for z := 0; z < cells; z += t.Config.ChunkSize {
		for x := 0; x < cells; x += t.Config.ChunkSize {
			chunk := TerrainChunk{
				Key:    fmt.Sprintf("terrain_%p_%d_%d", t, x, z),
				StartX: x,
				StartZ: z,
				EndX:   min(x+t.Config.ChunkSize, cells),
				EndZ:   min(z+t.Config.ChunkSize, cells),
			}
			verts := t.buildChunkVertices(&chunk)
			indexes := t.buildChunkIndexes(&chunk)
			chunk.Indexes = indexes
			chunk.Mesh = host.MeshCache().DynamicMesh(chunk.Key, verts, indexes)
			chunk.ShaderData = shader_data_registry.Create(t.Config.ShaderData)
			chunk.Drawing = rendering.Drawing{
				Material:   t.Material,
				Mesh:       chunk.Mesh,
				ShaderData: chunk.ShaderData,
				Transform:  t.Transform,
				ViewCuller: &host.Cameras.Primary,
			}
			t.ShaderData = append(t.ShaderData, chunk.ShaderData)
			host.Drawings.AddDrawing(chunk.Drawing)
			t.MeshChunks = append(t.MeshChunks, chunk)
		}
	}
}

func (t *Terrain) buildChunkMeshData(chunk *TerrainChunk) ([]rendering.Vertex, []uint32) {
	return t.buildChunkVertices(chunk), t.buildChunkIndexes(chunk)
}

func (t *Terrain) buildChunkVertices(chunk *TerrainChunk) []rendering.Vertex {
	width := chunk.EndX - chunk.StartX + 1
	depth := chunk.EndZ - chunk.StartZ + 1
	verts := make([]rendering.Vertex, width*depth)
	for z := chunk.StartZ; z <= chunk.EndZ; z++ {
		for x := chunk.StartX; x <= chunk.EndX; x++ {
			local := t.gridToLocal(matrix.Float(x), matrix.Float(z))
			idx := (x - chunk.StartX) + (z-chunk.StartZ)*width
			verts[idx] = rendering.Vertex{
				Position: local,
				Normal:   t.normalAtLocal(local.XZ()),
				UV0: matrix.NewVec2(
					matrix.Float(x)/matrix.Float(t.HeightField.Resolution-1),
					matrix.Float(z)/matrix.Float(t.HeightField.Resolution-1),
				),
				Color: matrix.ColorWhite(),
			}
		}
	}
	return verts
}

func (t *Terrain) buildChunkIndexes(chunk *TerrainChunk) []uint32 {
	width := chunk.EndX - chunk.StartX + 1
	depth := chunk.EndZ - chunk.StartZ + 1
	indexes := make([]uint32, 0, (width-1)*(depth-1)*6)
	for z := 0; z < depth-1; z++ {
		for x := 0; x < width-1; x++ {
			i0 := uint32(x + z*width)
			i1 := uint32(x + (z+1)*width)
			i2 := uint32(x + 1 + (z+1)*width)
			i3 := uint32(x + 1 + z*width)
			indexes = append(indexes, i0, i1, i2, i0, i2, i3)
		}
	}
	return indexes
}

func (t *Terrain) localStrokeToGrid(stroke PaintStroke) PaintStroke {
	x, z := t.localToGrid(stroke.Center)
	cell := matrix.Min(
		t.Config.WorldSize.X()/matrix.Float(t.HeightField.Resolution-1),
		t.Config.WorldSize.Y()/matrix.Float(t.HeightField.Resolution-1),
	)
	stroke.Center = matrix.NewVec2(x, z)
	if cell > matrix.Tiny {
		stroke.Radius /= cell
		stroke.Spacing /= cell
	}
	return stroke
}

func (t *Terrain) localToGrid(localXZ matrix.Vec2) (matrix.Float, matrix.Float) {
	x := ((localXZ.X() / t.Config.WorldSize.X()) + 0.5) * matrix.Float(t.HeightField.Resolution-1)
	z := ((localXZ.Y() / t.Config.WorldSize.Y()) + 0.5) * matrix.Float(t.HeightField.Resolution-1)
	return x, z
}

func (t *Terrain) gridToLocal(x, z matrix.Float) matrix.Vec3 {
	gx := x / matrix.Float(t.HeightField.Resolution-1)
	gz := z / matrix.Float(t.HeightField.Resolution-1)
	return matrix.NewVec3(
		(gx-0.5)*t.Config.WorldSize.X(),
		t.HeightField.Sample(x, z),
		(gz-0.5)*t.Config.WorldSize.Y(),
	)
}

func paintHeightField(h *HeightField, stroke PaintStroke) DirtyRegion {
	if stroke.Radius <= 0 || stroke.Strength == 0 {
		return DirtyRegion{}
	}
	region := strokeDirtyRegion(h, stroke)
	if !region.Valid {
		return DirtyRegion{}
	}
	minX := region.MinX
	maxX := region.MaxX
	minZ := region.MinZ
	maxZ := region.MaxZ
	if minX > maxX || minZ > maxZ {
		return DirtyRegion{}
	}
	original := h.Heights
	if stroke.Mode == BrushSmooth {
		original = append([]matrix.Float(nil), h.Heights...)
	}
	var dirty DirtyRegion
	for z := minZ; z <= maxZ; z++ {
		for x := minX; x <= maxX; x++ {
			dx := matrix.Float(x) - stroke.Center.X()
			dz := matrix.Float(z) - stroke.Center.Y()
			distance := matrix.Sqrt(dx*dx + dz*dz)
			if distance > stroke.Radius {
				continue
			}
			weight := brushWeight(distance, stroke.Radius, stroke.Falloff)
			before := h.Height(x, z)
			after := before
			switch stroke.Mode {
			case BrushLower:
				after = before - stroke.Strength*weight
			case BrushSmooth:
				average := neighborAverage(original, h.Resolution, x, z)
				after = matrix.Lerp(before, average, matrix.Clamp(stroke.Strength*weight, 0, 1))
			case BrushRaise:
				fallthrough
			default:
				after = before + stroke.Strength*weight
			}
			if h.SetHeight(x, z, after) {
				dirty = mergeDirtyRegions(dirty, DirtyRegion{
					MinX: x, MinZ: z, MaxX: x, MaxZ: z, Valid: true,
				})
			}
		}
	}
	return dirty
}

func strokeDirtyRegion(h *HeightField, stroke PaintStroke) DirtyRegion {
	if stroke.Radius <= 0 {
		return DirtyRegion{}
	}
	minX := max(0, int(matrix.Floor(stroke.Center.X()-stroke.Radius)))
	maxX := min(h.Resolution-1, int(matrix.Ceil(stroke.Center.X()+stroke.Radius)))
	minZ := max(0, int(matrix.Floor(stroke.Center.Y()-stroke.Radius)))
	maxZ := min(h.Resolution-1, int(matrix.Ceil(stroke.Center.Y()+stroke.Radius)))
	if minX > maxX || minZ > maxZ {
		return DirtyRegion{}
	}
	return DirtyRegion{MinX: minX, MinZ: minZ, MaxX: maxX, MaxZ: maxZ, Valid: true}
}

func brushWeight(distance, radius matrix.Float, falloff BrushFalloff) matrix.Float {
	if radius <= 0 {
		return 0
	}
	t := matrix.Clamp(distance/radius, 0, 1)
	switch falloff {
	case FalloffConstant:
		return 1
	case FalloffSmooth:
		x := 1 - t
		return x * x * (3 - 2*x)
	case FalloffLinear:
		fallthrough
	default:
		return 1 - t
	}
}

func neighborAverage(heights []matrix.Float, resolution, x, z int) matrix.Float {
	var sum matrix.Float
	count := matrix.Float(0)
	for nz := max(0, z-1); nz <= min(resolution-1, z+1); nz++ {
		for nx := max(0, x-1); nx <= min(resolution-1, x+1); nx++ {
			sum += heights[nx+nz*resolution]
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / count
}

func mergeDirtyRegions(a, b DirtyRegion) DirtyRegion {
	if !a.Valid {
		return b
	}
	if !b.Valid {
		return a
	}
	return DirtyRegion{
		MinX:  min(a.MinX, b.MinX),
		MinZ:  min(a.MinZ, b.MinZ),
		MaxX:  max(a.MaxX, b.MaxX),
		MaxZ:  max(a.MaxZ, b.MaxZ),
		Valid: true,
	}
}

func rayBounds(ray graviton.Ray, minBounds, maxBounds matrix.Vec3) (matrix.Float, matrix.Float, bool) {
	tMin := matrix.Float(0)
	tMax := matrix.Inf(1)
	for axis := 0; axis < 3; axis++ {
		origin := ray.Origin[axis]
		direction := ray.Direction[axis]
		minValue := minBounds[axis]
		maxValue := maxBounds[axis]
		if matrix.Abs(direction) <= matrix.FloatSmallestNonzero {
			if origin < minValue || origin > maxValue {
				return 0, 0, false
			}
			continue
		}
		inv := 1 / direction
		t0 := (minValue - origin) * inv
		t1 := (maxValue - origin) * inv
		if t0 > t1 {
			t0, t1 = t1, t0
		}
		tMin = matrix.Max(tMin, t0)
		tMax = matrix.Min(tMax, t1)
		if tMin > tMax {
			return 0, 0, false
		}
	}
	return tMin, tMax, true
}
