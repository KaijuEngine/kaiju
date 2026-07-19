/******************************************************************************/
/* probe_baker.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gi

import (
	"context"
	"errors"
	"math"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

type BakeTriangle struct {
	Points   [3]matrix.Vec3
	Albedo   matrix.Vec3
	Emissive matrix.Vec3
}

type BakeDirectionalLight struct {
	TowardLight matrix.Vec3
	Radiance    matrix.Vec3
}

type BakePointLight struct {
	Position  matrix.Vec3
	Intensity matrix.Vec3
	Range     matrix.Float
}

type BakeSpotLight struct {
	Position    matrix.Vec3
	Direction   matrix.Vec3
	Intensity   matrix.Vec3
	Range       matrix.Float
	InnerCutoff matrix.Float
	OuterCutoff matrix.Float
}

type BakeInput struct {
	Bounds            graviton.AABB
	ProbeSpacing      matrix.Float
	RaysPerProbe      uint32
	MaxRayDistance    matrix.Float
	Scenario          string
	Environment       matrix.Vec3
	Triangles         []BakeTriangle
	DirectionalLights []BakeDirectionalLight
	PointLights       []BakePointLight
	SpotLights        []BakeSpotLight
	GeometryHash      [32]byte
	LightingHash      [32]byte
	Progress          func(completed, total int)
}

type bakeTriangleData struct {
	triangle graviton.DetailedTriangle
	material BakeTriangle
}

type probeBakerScene struct {
	bvh       *graviton.BVH
	triangles []bakeTriangleData
}

// BakeProbes is the deterministic, platform-neutral reference baker. Its BVH
// path makes it suitable for CI and modest scenes; a future GPU baker can emit
// the exact same ProbeAsset format without changing runtime providers.
func BakeProbes(ctx context.Context, input BakeInput) (ProbeAsset, error) {
	if input.ProbeSpacing <= 0 {
		return ProbeAsset{}, errors.New("GI bake probe spacing must be greater than zero")
	}
	if input.RaysPerProbe < 32 {
		return ProbeAsset{}, errors.New("GI bake requires at least 32 rays per probe")
	}
	if input.MaxRayDistance <= 0 {
		return ProbeAsset{}, errors.New("GI bake maximum ray distance must be greater than zero")
	}
	if input.Bounds.Extent.X() <= 0 || input.Bounds.Extent.Y() <= 0 || input.Bounds.Extent.Z() <= 0 {
		return ProbeAsset{}, errors.New("GI bake bounds must have positive extent")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	dimensions := probeGridDimensions(input.Bounds.Size(), input.ProbeSpacing)
	total := int(uint64(dimensions[0]) * uint64(dimensions[1]) * uint64(dimensions[2]))
	minPoint := input.Bounds.Min()
	gridMax := minPoint.Add(matrix.Vec3{
		matrix.Float(dimensions[0]-1) * input.ProbeSpacing,
		matrix.Float(dimensions[1]-1) * input.ProbeSpacing,
		matrix.Float(dimensions[2]-1) * input.ProbeSpacing,
	})
	asset := ProbeAsset{
		Bounds:       graviton.AABBFromMinMax(minPoint, gridMax),
		Spacing:      input.ProbeSpacing,
		Dimensions:   dimensions,
		Scenario:     input.Scenario,
		GeometryHash: input.GeometryHash,
		LightingHash: input.LightingHash,
		Probes:       make([]Probe, total),
	}
	scene := buildProbeBakerScene(input.Triangles)
	for z := uint32(0); z < dimensions[2]; z++ {
		for y := uint32(0); y < dimensions[1]; y++ {
			for x := uint32(0); x < dimensions[0]; x++ {
				if err := ctx.Err(); err != nil {
					return ProbeAsset{}, err
				}
				index := asset.ProbeIndex(x, y, z)
				position := matrix.Vec3{
					minPoint.X() + matrix.Float(x)*input.ProbeSpacing,
					minPoint.Y() + matrix.Float(y)*input.ProbeSpacing,
					minPoint.Z() + matrix.Float(z)*input.ProbeSpacing,
				}
				asset.Probes[index] = bakeProbe(scene, input, position, index)
				if input.Progress != nil {
					input.Progress(index+1, total)
				}
			}
		}
	}
	return asset, asset.Validate()
}

func probeGridDimensions(size matrix.Vec3, spacing matrix.Float) [3]uint32 {
	return [3]uint32{
		max(2, uint32(math.Ceil(float64(size.X()/spacing)))+1),
		max(2, uint32(math.Ceil(float64(size.Y()/spacing)))+1),
		max(2, uint32(math.Ceil(float64(size.Z()/spacing)))+1),
	}
}

func buildProbeBakerScene(triangles []BakeTriangle) probeBakerScene {
	scene := probeBakerScene{triangles: make([]bakeTriangleData, len(triangles))}
	for i := range triangles {
		detailed := graviton.DetailedTriangleFromPoints(triangles[i].Points)
		scene.triangles[i] = bakeTriangleData{triangle: detailed, material: triangles[i]}
		graviton.InsertBVH(&scene.bvh, detailed, nil, i)
	}
	return scene
}

func bakeProbe(scene probeBakerScene, input BakeInput, position matrix.Vec3, probeIndex int) Probe {
	probe := Probe{Position: position, Validity: 1}
	distanceSum := matrix.Float(0)
	distanceSquaredSum := matrix.Float(0)
	weight := matrix.Float(4 * math.Pi / float64(input.RaysPerProbe))
	for rayIndex := uint32(0); rayIndex < input.RaysPerProbe; rayIndex++ {
		direction := fibonacciSphereDirection(rayIndex, input.RaysPerProbe, uint32(probeIndex))
		radiance, distance := scene.traceRadiance(position, direction, input)
		basis := evaluateSHBasis(direction)
		for coefficient := range probe.RadianceSH {
			probe.RadianceSH[coefficient].AddAssign(radiance.Scale(basis[coefficient] * weight))
		}
		distanceSum += distance
		distanceSquaredSum += distance * distance
	}
	probe.MeanDistance = distanceSum / matrix.Float(input.RaysPerProbe)
	meanSquared := distanceSquaredSum / matrix.Float(input.RaysPerProbe)
	probe.DistanceVariance = max(0, meanSquared-probe.MeanDistance*probe.MeanDistance)
	return probe
}

func fibonacciSphereDirection(index, count, scramble uint32) matrix.Vec3 {
	const goldenRatioConjugate = 0.6180339887498949
	u := (float64(index) + 0.5) / float64(count)
	v := math.Mod(float64(index)*goldenRatioConjugate+float64(hashProbe(scramble))/4294967296.0, 1)
	z := 1 - 2*u
	radius := math.Sqrt(max(0, 1-z*z))
	angle := 2 * math.Pi * v
	return matrix.NewVec3(radius*math.Cos(angle), radius*math.Sin(angle), z)
}

func hashProbe(value uint32) uint32 {
	value ^= value >> 16
	value *= 0x7feb352d
	value ^= value >> 15
	value *= 0x846ca68b
	value ^= value >> 16
	return value
}

func (s probeBakerScene) trace(origin, direction matrix.Vec3, maxDistance matrix.Float) (int, matrix.Vec3, matrix.Float, bool) {
	if s.bvh == nil {
		return 0, matrix.Vec3{}, maxDistance, false
	}
	data, point, hit := s.bvh.RayIntersect(graviton.Ray{Origin: origin, Direction: direction}, float32(maxDistance))
	if !hit {
		return 0, matrix.Vec3{}, maxDistance, false
	}
	index, ok := data.(int)
	if !ok || index < 0 || index >= len(s.triangles) {
		return 0, matrix.Vec3{}, maxDistance, false
	}
	return index, point, point.Subtract(origin).Length(), true
}

func (s probeBakerScene) traceRadiance(origin, direction matrix.Vec3, input BakeInput) (matrix.Vec3, matrix.Float) {
	triangleIndex, point, distance, hit := s.trace(origin, direction, input.MaxRayDistance)
	if !hit {
		return input.Environment, input.MaxRayDistance
	}
	triangle := s.triangles[triangleIndex]
	normal := triangle.triangle.Normal
	if normal.Dot(direction) > 0 {
		normal = normal.Negative()
	}
	radiance := triangle.material.Emissive.Add(triangle.material.Albedo.Multiply(input.Environment))
	originBias := point.Add(normal.Scale(0.001))
	pi := matrix.Float(math.Pi)
	for i := range input.DirectionalLights {
		lightDirection := input.DirectionalLights[i].TowardLight
		if lightDirection.LengthSquared() == 0 {
			continue
		}
		lightDirection = lightDirection.Normal()
		nDotL := max(matrix.Float(0), normal.Dot(lightDirection))
		if nDotL == 0 || s.occluded(originBias, lightDirection, input.MaxRayDistance) {
			continue
		}
		radiance.AddAssign(triangle.material.Albedo.Multiply(input.DirectionalLights[i].Radiance).Scale(nDotL / pi))
	}
	for i := range input.PointLights {
		toLight := input.PointLights[i].Position.Subtract(point)
		distanceSquared := toLight.LengthSquared()
		if distanceSquared <= matrix.FloatSmallestNonzero {
			continue
		}
		lightDistance := matrix.Sqrt(distanceSquared)
		if input.PointLights[i].Range > 0 && lightDistance > input.PointLights[i].Range {
			continue
		}
		lightDirection := toLight.Scale(1 / lightDistance)
		nDotL := max(matrix.Float(0), normal.Dot(lightDirection))
		if nDotL == 0 || s.occluded(originBias, lightDirection, lightDistance-0.002) {
			continue
		}
		incident := input.PointLights[i].Intensity.Scale(1 / max(distanceSquared, matrix.Float(0.01)))
		radiance.AddAssign(triangle.material.Albedo.Multiply(incident).Scale(nDotL / pi))
	}
	for i := range input.SpotLights {
		light := input.SpotLights[i]
		toLight := light.Position.Subtract(point)
		distanceSquared := toLight.LengthSquared()
		if distanceSquared <= matrix.FloatSmallestNonzero {
			continue
		}
		lightDistance := matrix.Sqrt(distanceSquared)
		if light.Range > 0 && lightDistance > light.Range {
			continue
		}
		lightDirection := toLight.Scale(1 / lightDistance)
		spotDirection := light.Direction
		if spotDirection.LengthSquared() == 0 {
			continue
		}
		coneCosine := spotDirection.Normal().Dot(lightDirection.Negative())
		coneRange := max(matrix.Float(0.0001), light.InnerCutoff-light.OuterCutoff)
		cone := max(matrix.Float(0), min(matrix.Float(1), (coneCosine-light.OuterCutoff)/coneRange))
		nDotL := max(matrix.Float(0), normal.Dot(lightDirection))
		if cone == 0 || nDotL == 0 || s.occluded(originBias, lightDirection, lightDistance-0.002) {
			continue
		}
		incident := light.Intensity.Scale(cone / max(distanceSquared, matrix.Float(0.01)))
		radiance.AddAssign(triangle.material.Albedo.Multiply(incident).Scale(nDotL / pi))
	}
	return radiance, distance
}

func (s probeBakerScene) occluded(origin, direction matrix.Vec3, distance matrix.Float) bool {
	_, _, _, hit := s.trace(origin, direction, distance)
	return hit
}
