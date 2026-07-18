/******************************************************************************/
/* main.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/lighting/gi"
	"kaijuengine.com/matrix"
)

type vector3 [3]float32

func (v vector3) matrix() matrix.Vec3 { return matrix.NewVec3(v[0], v[1], v[2]) }

type bakeTriangleJSON struct {
	Points   [3]vector3 `json:"points"`
	Albedo   vector3    `json:"albedo"`
	Emissive vector3    `json:"emissive"`
}

type bakeDirectionalLightJSON struct {
	TowardLight vector3 `json:"towardLight"`
	Radiance    vector3 `json:"radiance"`
}

type bakePointLightJSON struct {
	Position  vector3 `json:"position"`
	Intensity vector3 `json:"intensity"`
	Range     float32 `json:"range"`
}

type bakeFile struct {
	Bounds struct {
		Min vector3 `json:"min"`
		Max vector3 `json:"max"`
	} `json:"bounds"`
	ProbeSpacing      float32                    `json:"probeSpacing"`
	RaysPerProbe      uint32                     `json:"raysPerProbe"`
	MaxRayDistance    float32                    `json:"maxRayDistance"`
	Scenario          string                     `json:"scenario"`
	Environment       vector3                    `json:"environment"`
	Triangles         []bakeTriangleJSON         `json:"triangles"`
	DirectionalLights []bakeDirectionalLightJSON `json:"directionalLights"`
	PointLights       []bakePointLightJSON       `json:"pointLights"`
	GeometryHash      string                     `json:"geometryHash"`
	LightingHash      string                     `json:"lightingHash"`
}

func main() {
	input := flag.String("input", "", "path to the GI bake scene JSON")
	output := flag.String("output", "", "path for the generated .kjgi asset")
	flag.Parse()
	if err := run(context.Background(), *input, *output); err != nil {
		fmt.Fprintln(os.Stderr, "kaiju-gi-bake:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, inputPath, outputPath string) error {
	if inputPath == "" || outputPath == "" {
		return errors.New("both -input and -output are required")
	}
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	input := bakeFile{}
	if err := decoder.Decode(&input); err != nil {
		return fmt.Errorf("decode input: %w", err)
	}
	bakeInput, err := input.toBakeInput()
	if err != nil {
		return err
	}
	asset, err := gi.BakeProbes(ctx, bakeInput)
	if err != nil {
		return fmt.Errorf("bake probes: %w", err)
	}
	encoded, err := asset.MarshalBinary()
	if err != nil {
		return fmt.Errorf("encode probes: %w", err)
	}
	if err := os.WriteFile(outputPath, encoded, 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func (input bakeFile) toBakeInput() (gi.BakeInput, error) {
	triangles := make([]gi.BakeTriangle, len(input.Triangles))
	for i := range input.Triangles {
		triangles[i] = gi.BakeTriangle{
			Points: [3]matrix.Vec3{
				input.Triangles[i].Points[0].matrix(),
				input.Triangles[i].Points[1].matrix(),
				input.Triangles[i].Points[2].matrix(),
			},
			Albedo:   input.Triangles[i].Albedo.matrix(),
			Emissive: input.Triangles[i].Emissive.matrix(),
		}
	}
	directional := make([]gi.BakeDirectionalLight, len(input.DirectionalLights))
	for i := range input.DirectionalLights {
		directional[i] = gi.BakeDirectionalLight{
			TowardLight: input.DirectionalLights[i].TowardLight.matrix(),
			Radiance:    input.DirectionalLights[i].Radiance.matrix(),
		}
	}
	points := make([]gi.BakePointLight, len(input.PointLights))
	for i := range input.PointLights {
		points[i] = gi.BakePointLight{
			Position:  input.PointLights[i].Position.matrix(),
			Intensity: input.PointLights[i].Intensity.matrix(),
			Range:     matrix.Float(input.PointLights[i].Range),
		}
	}
	geometryPayload, _ := json.Marshal(input.Triangles)
	lightingPayload, _ := json.Marshal(struct {
		Environment       vector3
		DirectionalLights []bakeDirectionalLightJSON
		PointLights       []bakePointLightJSON
	}{input.Environment, input.DirectionalLights, input.PointLights})
	geometryHash, err := parseOrHash(input.GeometryHash, geometryPayload)
	if err != nil {
		return gi.BakeInput{}, fmt.Errorf("geometryHash: %w", err)
	}
	lightingHash, err := parseOrHash(input.LightingHash, lightingPayload)
	if err != nil {
		return gi.BakeInput{}, fmt.Errorf("lightingHash: %w", err)
	}
	return gi.BakeInput{
		Bounds:            graviton.AABBFromMinMax(input.Bounds.Min.matrix(), input.Bounds.Max.matrix()),
		ProbeSpacing:      matrix.Float(input.ProbeSpacing),
		RaysPerProbe:      input.RaysPerProbe,
		MaxRayDistance:    matrix.Float(input.MaxRayDistance),
		Scenario:          input.Scenario,
		Environment:       input.Environment.matrix(),
		Triangles:         triangles,
		DirectionalLights: directional,
		PointLights:       points,
		GeometryHash:      geometryHash,
		LightingHash:      lightingHash,
	}, nil
}

func parseOrHash(value string, payload []byte) ([32]byte, error) {
	if value == "" {
		return sha256.Sum256(payload), nil
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return [32]byte{}, errors.New("must be exactly 64 hexadecimal characters")
	}
	result := [32]byte{}
	copy(result[:], decoded)
	return result, nil
}
