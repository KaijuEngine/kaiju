/******************************************************************************/
/* probe_asset.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

const (
	probeAssetVersion        = uint16(1)
	probeAssetHeaderBytes    = 4 + 2 + 2 + 12 + 4 + 24 + 2 + 64 + 4
	probeAssetFloatsPerProbe = 33
)

var probeAssetMagic = [4]byte{'K', 'J', 'G', 'I'}

// Probe stores second-order RGB spherical harmonics plus distance moments.
// Coefficients contain incident radiance; EvaluateIrradiance applies the
// Lambertian convolution when resolving against a surface normal.
type Probe struct {
	Position         matrix.Vec3
	RadianceSH       [9]matrix.Vec3
	MeanDistance     matrix.Float
	DistanceVariance matrix.Float
	Validity         matrix.Float
}

type ProbeAsset struct {
	Bounds       graviton.AABB
	Spacing      matrix.Float
	Dimensions   [3]uint32
	Scenario     string
	GeometryHash [32]byte
	LightingHash [32]byte
	Probes       []Probe
}

func (a ProbeAsset) Validate() error {
	if a.Spacing <= 0 {
		return errors.New("GI probe spacing must be greater than zero")
	}
	count := uint64(1)
	for i := range a.Dimensions {
		if a.Dimensions[i] == 0 {
			return fmt.Errorf("GI probe dimension %d is zero", i)
		}
		count *= uint64(a.Dimensions[i])
	}
	if count > uint64(^uint(0)>>1) || int(count) != len(a.Probes) {
		return fmt.Errorf("GI probe count is %d, expected %d", len(a.Probes), count)
	}
	if a.Bounds.Extent.X() <= 0 || a.Bounds.Extent.Y() <= 0 || a.Bounds.Extent.Z() <= 0 {
		return errors.New("GI probe bounds must have positive extent")
	}
	for i := range a.Probes {
		if a.Probes[i].Validity < 0 || a.Probes[i].Validity > 1 {
			return fmt.Errorf("GI probe %d validity must be in [0, 1]", i)
		}
		if a.Probes[i].DistanceVariance < 0 {
			return fmt.Errorf("GI probe %d has negative distance variance", i)
		}
	}
	return nil
}

func (a ProbeAsset) MarshalBinary() ([]byte, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}
	if len(a.Scenario) > math.MaxUint16 {
		return nil, errors.New("GI scenario name is too long")
	}
	probeBytes := probeAssetFloatsPerProbe * 4
	out := make([]byte, 0, probeAssetHeaderBytes+len(a.Scenario)+len(a.Probes)*probeBytes)
	out = append(out, probeAssetMagic[:]...)
	out = appendUint16(out, probeAssetVersion)
	out = appendUint16(out, 0)
	for i := range a.Dimensions {
		out = appendUint32(out, a.Dimensions[i])
	}
	out = appendFloat(out, a.Spacing)
	minPoint, maxPoint := a.Bounds.Min(), a.Bounds.Max()
	out = appendVec3(out, minPoint)
	out = appendVec3(out, maxPoint)
	out = appendUint16(out, uint16(len(a.Scenario)))
	out = append(out, a.Scenario...)
	out = append(out, a.GeometryHash[:]...)
	out = append(out, a.LightingHash[:]...)
	out = appendUint32(out, uint32(len(a.Probes)))
	for i := range a.Probes {
		out = appendVec3(out, a.Probes[i].Position)
		for coefficient := range a.Probes[i].RadianceSH {
			out = appendVec3(out, a.Probes[i].RadianceSH[coefficient])
		}
		out = appendFloat(out, a.Probes[i].MeanDistance)
		out = appendFloat(out, a.Probes[i].DistanceVariance)
		out = appendFloat(out, a.Probes[i].Validity)
	}
	return out, nil
}

func UnmarshalProbeAsset(data []byte) (ProbeAsset, error) {
	decoder := probeAssetDecoder{data: data}
	magic, err := decoder.bytes(4)
	if err != nil || string(magic) != string(probeAssetMagic[:]) {
		return ProbeAsset{}, errors.New("invalid GI probe asset magic")
	}
	version, err := decoder.uint16()
	if err != nil {
		return ProbeAsset{}, err
	}
	if version != probeAssetVersion {
		return ProbeAsset{}, fmt.Errorf("unsupported GI probe asset version %d", version)
	}
	if _, err = decoder.uint16(); err != nil { // Reserved flags.
		return ProbeAsset{}, err
	}
	asset := ProbeAsset{}
	for i := range asset.Dimensions {
		if asset.Dimensions[i], err = decoder.uint32(); err != nil {
			return ProbeAsset{}, err
		}
	}
	if asset.Spacing, err = decoder.float(); err != nil {
		return ProbeAsset{}, err
	}
	minPoint, err := decoder.vec3()
	if err != nil {
		return ProbeAsset{}, err
	}
	maxPoint, err := decoder.vec3()
	if err != nil {
		return ProbeAsset{}, err
	}
	asset.Bounds = graviton.AABBFromMinMax(minPoint, maxPoint)
	scenarioLength, err := decoder.uint16()
	if err != nil {
		return ProbeAsset{}, err
	}
	scenario, err := decoder.bytes(int(scenarioLength))
	if err != nil {
		return ProbeAsset{}, err
	}
	asset.Scenario = string(scenario)
	geometryHash, err := decoder.bytes(len(asset.GeometryHash))
	if err != nil {
		return ProbeAsset{}, err
	}
	copy(asset.GeometryHash[:], geometryHash)
	lightingHash, err := decoder.bytes(len(asset.LightingHash))
	if err != nil {
		return ProbeAsset{}, err
	}
	copy(asset.LightingHash[:], lightingHash)
	count, err := decoder.uint32()
	if err != nil {
		return ProbeAsset{}, err
	}
	expectedBytes := uint64(count) * probeAssetFloatsPerProbe * 4
	if expectedBytes > uint64(decoder.remaining()) {
		return ProbeAsset{}, errors.New("truncated GI probe data")
	}
	asset.Probes = make([]Probe, count)
	for i := range asset.Probes {
		if asset.Probes[i].Position, err = decoder.vec3(); err != nil {
			return ProbeAsset{}, err
		}
		for coefficient := range asset.Probes[i].RadianceSH {
			if asset.Probes[i].RadianceSH[coefficient], err = decoder.vec3(); err != nil {
				return ProbeAsset{}, err
			}
		}
		if asset.Probes[i].MeanDistance, err = decoder.float(); err != nil {
			return ProbeAsset{}, err
		}
		if asset.Probes[i].DistanceVariance, err = decoder.float(); err != nil {
			return ProbeAsset{}, err
		}
		if asset.Probes[i].Validity, err = decoder.float(); err != nil {
			return ProbeAsset{}, err
		}
	}
	if decoder.remaining() != 0 {
		return ProbeAsset{}, fmt.Errorf("GI probe asset has %d trailing bytes", decoder.remaining())
	}
	if err := asset.Validate(); err != nil {
		return ProbeAsset{}, err
	}
	return asset, nil
}

func appendUint16(out []byte, value uint16) []byte {
	return binary.LittleEndian.AppendUint16(out, value)
}

func appendUint32(out []byte, value uint32) []byte {
	return binary.LittleEndian.AppendUint32(out, value)
}

func appendFloat(out []byte, value matrix.Float) []byte {
	return appendUint32(out, math.Float32bits(float32(value)))
}

func appendVec3(out []byte, value matrix.Vec3) []byte {
	for component := range value {
		out = appendFloat(out, value[component])
	}
	return out
}

type probeAssetDecoder struct {
	data   []byte
	offset int
}

func (d *probeAssetDecoder) remaining() int { return len(d.data) - d.offset }

func (d *probeAssetDecoder) bytes(count int) ([]byte, error) {
	if count < 0 || count > d.remaining() {
		return nil, errors.New("truncated GI probe asset")
	}
	value := d.data[d.offset : d.offset+count]
	d.offset += count
	return value, nil
}

func (d *probeAssetDecoder) uint16() (uint16, error) {
	data, err := d.bytes(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(data), nil
}

func (d *probeAssetDecoder) uint32() (uint32, error) {
	data, err := d.bytes(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), nil
}

func (d *probeAssetDecoder) float() (matrix.Float, error) {
	value, err := d.uint32()
	return matrix.Float(math.Float32frombits(value)), err
}

func (d *probeAssetDecoder) vec3() (matrix.Vec3, error) {
	value := matrix.Vec3{}
	for component := range value {
		decoded, err := d.float()
		if err != nil {
			return matrix.Vec3{}, err
		}
		value[component] = decoded
	}
	return value, nil
}

func evaluateSHBasis(direction matrix.Vec3) [9]matrix.Float {
	x, y, z := direction.X(), direction.Y(), direction.Z()
	return [9]matrix.Float{
		0.2820947918,
		0.4886025119 * y,
		0.4886025119 * z,
		0.4886025119 * x,
		1.0925484306 * x * y,
		1.0925484306 * y * z,
		0.3153915653 * (3*z*z - 1),
		1.0925484306 * x * z,
		0.5462742153 * (x*x - y*y),
	}
}

func (p Probe) EvaluateIrradiance(normal matrix.Vec3) matrix.Vec3 {
	if normal.LengthSquared() == 0 {
		return matrix.Vec3Zero()
	}
	basis := evaluateSHBasis(normal.Normal())
	pi := matrix.Float(math.Pi)
	bandScale := [9]matrix.Float{
		pi,
		2 * pi / 3, 2 * pi / 3, 2 * pi / 3,
		pi / 4, pi / 4, pi / 4, pi / 4, pi / 4,
	}
	result := matrix.Vec3Zero()
	for i := range p.RadianceSH {
		result.AddAssign(p.RadianceSH[i].Scale(basis[i] * bandScale[i]))
	}
	return matrix.Vec3Max(result, matrix.Vec3Zero()).Scale(p.Validity)
}

func (a ProbeAsset) ProbeIndex(x, y, z uint32) int {
	return int(x + a.Dimensions[0]*(y+a.Dimensions[1]*z))
}

// SampleIrradiance trilinearly samples the probe lattice. A conservative
// distance-moment weight suppresses probes likely to be behind nearby walls.
func (a ProbeAsset) SampleIrradiance(position, normal matrix.Vec3) matrix.Vec3 {
	if len(a.Probes) == 0 {
		return matrix.Vec3Zero()
	}
	minPoint := a.Bounds.Min()
	grid := position.Subtract(minPoint).Scale(1 / a.Spacing)
	base := [3]int{}
	fraction := [3]matrix.Float{}
	for axis := range 3 {
		maxCell := max(0, int(a.Dimensions[axis])-2)
		cell := int(math.Floor(float64(grid[axis])))
		cell = max(0, min(cell, maxCell))
		base[axis] = cell
		fraction[axis] = matrix.Clamp(grid[axis]-matrix.Float(cell), 0, 1)
	}
	result := matrix.Vec3Zero()
	totalWeight := matrix.Float(0)
	for z := range 2 {
		for y := range 2 {
			for x := range 2 {
				coordinates := [3]uint32{uint32(base[0] + x), uint32(base[1] + y), uint32(base[2] + z)}
				for axis := range 3 {
					coordinates[axis] = min(coordinates[axis], a.Dimensions[axis]-1)
				}
				probe := a.Probes[a.ProbeIndex(coordinates[0], coordinates[1], coordinates[2])]
				weight := []matrix.Float{1 - fraction[0], fraction[0]}[x] *
					[]matrix.Float{1 - fraction[1], fraction[1]}[y] *
					[]matrix.Float{1 - fraction[2], fraction[2]}[z]
				distance := position.Subtract(probe.Position).Length()
				if probe.MeanDistance > 0 && distance > probe.MeanDistance {
					variance := max(probe.DistanceVariance, matrix.Float(0.0001))
					delta := distance - probe.MeanDistance
					weight *= variance / (variance + delta*delta)
				}
				weight *= probe.Validity
				result.AddAssign(probe.EvaluateIrradiance(normal).Scale(weight))
				totalWeight += weight
			}
		}
	}
	if totalWeight <= matrix.FloatSmallestNonzero {
		return matrix.Vec3Zero()
	}
	return result.Scale(1 / totalWeight)
}
