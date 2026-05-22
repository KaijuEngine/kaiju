/******************************************************************************/
/* main.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/glsl"
)

type Prebuilt struct {
	Src     string
	Spv     string
	Shaders []string
}

//go:embed prebuilt.json
var prebuiltJson []byte

func main() {
	var pb Prebuilt
	if err := json.Unmarshal(prebuiltJson, &pb); err != nil {
		panic(err)
	}
	srcRoot := openRoot(pb.Src)
	spvRoot := openRoot(pb.Spv)
	for i := range pb.Shaders {
		sd := parseShader(pb.Shaders[i])
		flagSuffix := ""
		if sd.EnableDebug {
			flagSuffix += " -g"
		}
		if sd.Vertex != "" {
			i := pathJoin(srcRoot.Name(), sd.Vertex)
			o := pathJoin(spvRoot.Name(), sd.VertexSpv)
			parseFile(&sd, i, sd.VertexFlags)
			compileFile(i, sd.VertexFlags+flagSuffix, o)
		}
		if sd.Fragment != "" {
			i := pathJoin(srcRoot.Name(), sd.Fragment)
			o := pathJoin(spvRoot.Name(), sd.FragmentSpv)
			parseFile(&sd, i, sd.FragmentFlags)
			compileFile(i, sd.FragmentFlags+flagSuffix, o)
		}
		if sd.Geometry != "" {
			i := pathJoin(srcRoot.Name(), sd.Geometry)
			o := pathJoin(spvRoot.Name(), sd.GeometrySpv)
			parseFile(&sd, i, sd.GeometryFlags)
			compileFile(i, sd.GeometryFlags+flagSuffix, o)
		}
		if sd.TessellationControl != "" {
			i := pathJoin(srcRoot.Name(), sd.TessellationControl)
			o := pathJoin(spvRoot.Name(), sd.TessellationControlSpv)
			parseFile(&sd, i, sd.TessellationControlFlags)
			compileFile(i, sd.TessellationControlFlags+flagSuffix, o)
		}
		if sd.TessellationEvaluation != "" {
			i := pathJoin(srcRoot.Name(), sd.TessellationEvaluation)
			o := pathJoin(spvRoot.Name(), sd.TessellationEvaluationSpv)
			parseFile(&sd, i, sd.TessellationEvaluationFlags)
			compileFile(i, sd.TessellationEvaluationFlags+flagSuffix, o)
		}
		if sd.Compute != "" {
			i := pathJoin(srcRoot.Name(), sd.Compute)
			o := pathJoin(spvRoot.Name(), sd.ComputeSpv)
			parseFile(&sd, i, sd.ComputeFlags)
			compileFile(i, sd.ComputeFlags+flagSuffix, o)
		}
		data, err := json.Marshal(sd)
		if err != nil {
			panic(err)
		}
		if err = os.WriteFile(pb.Shaders[i], data, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func pathJoin(a, b string) string {
	return filepath.ToSlash(filepath.Join(a, b))
}

func openRoot(path string) *os.Root {
	root, err := os.OpenRoot(filepath.ToSlash(path))
	if err != nil {
		panic(err)
	}
	return root
}

func parseFile(sd *rendering.ShaderData, fileType, flags string) {
	src, err := glsl.Parse(fileType, flags)
	if err != nil {
		panic(err)
	}
	sd.LayoutGroups = append(sd.LayoutGroups, rendering.ShaderLayoutGroup{
		Type:       src.Type(),
		WorkGroups: src.WorkGroups,
		Layouts:    src.Layouts,
	})
}

func compileFile(file, flags, out string) {
	args := []string{file, "-o", out}
	flags = strings.TrimSpace(flags)
	if flags != "" {
		args = append(args, strings.Split(flags, " ")...)
	}
	cmd := exec.Command("glslc", args...)
	if errStr, err := cmd.CombinedOutput(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			panic(string(errStr))
		}
		panic(err)
	}
}

func parseShader(file string) rendering.ShaderData {
	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	if len(data) == 0 {
		panic("the shader file was empty")
	}
	var shader rendering.ShaderData
	if err = json.Unmarshal(data, &shader); err != nil {
		panic(err)
	}
	shader.LayoutGroups = make([]rendering.ShaderLayoutGroup, 0)
	return shader
}
