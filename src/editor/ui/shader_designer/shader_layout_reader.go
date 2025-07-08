/******************************************************************************/
/* shader_layout_reader.go                                                    */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

package shader_designer

import (
	"bufio"
	"kaiju/klib"
	"kaiju/klib/string_equations"
	"kaiju/rendering"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	layoutReg = regexp.MustCompile(`(?s)\s*layout\s*\(([\w\s=\d,\+\-\*\/]+)\)\s*(?:readonly\s+)?(in|out|uniform)\s+([a-zA-Z0-9]+)\s+([a-zA-Z0-9_]+){0,1}(?:\s*\{(.*?)\})?\s*(\w+){0,1}`)
)

type shaderSource struct {
	src     string
	file    string
	defines map[string]any
	layouts []rendering.ShaderLayout
}

func (s *shaderSource) defineAsString(name string) string {
	if d, ok := s.defines[name]; ok {
		if v, ok := d.(string); ok {
			return v
		} else {
			return strconv.FormatFloat(d.(float64), 'G', 10, 64)
		}
	}
	return name
}

func (s *shaderSource) processArrayField(value string) (string, error) {
	re := regexp.MustCompile(`\[([[\w\d\s\+\-\*\/]+)\]`)
	start := strings.Index(value, "[")
	if start < 0 {
		return value, nil
	}
	matches := re.FindAllStringSubmatch(value, -1)
	sb := strings.Builder{}
	sb.WriteString(value[:start])
	for i := range matches {
		v, err := s.processDefineEquation(matches[i][1])
		if err != nil {
			slog.Error("failed to process the equation for the array size", "size", matches[i][0], "error", err)
			return "", err
		}
		sb.WriteRune('[')
		sb.WriteString(strconv.FormatFloat(v, 'G', 10, 64))
		sb.WriteRune(']')
	}
	return sb.String(), nil
}

func (s *shaderSource) processDefineEquation(value string) (float64, error) {
	const replace = "+-*/"
	for i := range replace {
		value = strings.ReplaceAll(value, string(replace[i]), " "+string(replace[i])+" ")
	}
	fields := strings.Fields(value)
	for i := range fields {
		fields[i] = s.defineAsString(fields[i])
	}
	return string_equations.CalculateSimpleStringExpression(strings.Join(fields, " "))
}

func (s *shaderSource) readDefines() {
	re := regexp.MustCompile(`#define\s+(\w+)\s+([\w\d\s\+\-\*\/]+)`)
	ops := []string{"+", "-", "*", "/"}
	scan := bufio.NewScanner(strings.NewReader(s.src))
	for scan.Scan() {
		line := scan.Text()
		match := re.FindStringSubmatch(line)
		if len(match) == 3 {
			name := match[1]
			value := match[2]
			isEquation := false
			for j := range ops {
				isEquation = isEquation || strings.Contains(value, ops[j])
			}
			if isEquation {
				if v, err := s.processDefineEquation(value); err == nil {
					s.defines[name] = v
				} else {
					log.Fatalf("error processing equation (%s): %s", match[0], err)
				}
			} else {
				if f, err := strconv.ParseFloat(value, 64); err == nil {
					s.defines[name] = f
				} else {
					s.defines[name] = value
				}
			}
		}
	}
}

func (s *shaderSource) readLayouts() error {
	matches := layoutReg.FindAllStringSubmatch(s.src, -1)
	s.layouts = make([]rendering.ShaderLayout, len(matches))
	for i := range matches {
		name := matches[i][4]
		if name == "" {
			name = matches[i][6]
		}
		s.layouts[i] = rendering.ShaderLayout{
			Location:        -1,
			Binding:         -1,
			Set:             -1,
			InputAttachment: -1,
			Type:            matches[i][3],
			Name:            name,
			Source:          matches[i][2],
		}
		attrs := strings.Split(matches[i][1], ",")
		for j := range attrs {
			parts := strings.Fields(attrs[j])
			val, err := s.processDefineEquation(strings.Join(parts[2:], " "))
			if err != nil {
				log.Fatalf("invalid value for layout (%s): %s", matches[i][0], err)
			}
			switch parts[0] {
			case "location":
				s.layouts[i].Location = int(val)
			case "binding":
				s.layouts[i].Binding = int(val)
			case "set":
				s.layouts[i].Set = int(val)
			case "input_attachment_index":
				s.layouts[i].InputAttachment = int(val)
			}
		}
		if matches[i][5] != "" {
			fields := strings.Split(strings.TrimSpace(matches[i][5]), ";")
			if len(fields) > 0 && fields[len(fields)-1] == "" {
				fields = fields[:len(fields)-1]
			}
			s.layouts[i].Fields = make([]rendering.ShaderLayoutStructField, len(fields))
			for j := range fields {
				parts := strings.Fields(fields[j])
				name, err := s.processArrayField(parts[1])
				if err != nil {
					return err
				}
				s.layouts[i].Fields[j] = rendering.ShaderLayoutStructField{
					Type: parts[0],
					Name: name,
				}
			}
		}
	}
	return nil
}

func readShaderImports(fs *os.Root, inSrc, path string) string {
	src := strings.Builder{}
	scan := bufio.NewScanner(strings.NewReader(inSrc))
	re := regexp.MustCompile(`\s{0,}#include\s+\"([\w\.]+)\"`)
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		match := re.FindStringSubmatch(line)
		if len(match) == 2 && match[1] != "" {
			importSrc, err := klib.ReadRootFile(fs, filepath.Join(path, match[1]))
			if err != nil {
				log.Fatalf("failed to load import file (%s): %s", match[1], err)
			}
			src.WriteString(readShaderImports(fs, string(importSrc), path))
		} else {
			src.WriteString(line + "\n")
		}
	}
	return src.String()
}

func readShaderCode(fs *os.Root, file string) (shaderSource, error) {
	source := shaderSource{
		file:    file,
		defines: make(map[string]any),
	}
	data, err := klib.ReadRootFile(fs, source.file)
	if err != nil {
		slog.Error("failed to read the file", "file", file, "error", err)
		return source, err
	}
	source.src = readShaderImports(fs, string(data), filepath.Dir(source.file))
	source.readDefines()
	if err := source.readLayouts(); err != nil {
		return source, err
	}
	return source, nil
}

func importShaderLayout(shader rendering.ShaderData) (rendering.ShaderData, error) {
	shader.LayoutGroups = make([]rendering.ShaderLayoutGroup, 0)
	fs, err := os.OpenRoot(shaderSrcFolder)
	if err != nil {
		return shader, err
	}
	if shader.Vertex != "" {
		s := filepath.ToSlash(shader.Vertex)
		c, err := readShaderCode(fs, strings.TrimPrefix(s, shaderSrcFolder+"/"))
		if err != nil {
			return shader, err
		}
		shader.LayoutGroups = append(shader.LayoutGroups, rendering.ShaderLayoutGroup{
			Type:    "Vertex",
			Layouts: c.layouts,
		})
	}
	if shader.Fragment != "" {
		s := filepath.ToSlash(shader.Fragment)
		c, err := readShaderCode(fs, strings.TrimPrefix(s, shaderSrcFolder+"/"))
		if err != nil {
			return shader, err
		}
		shader.LayoutGroups = append(shader.LayoutGroups, rendering.ShaderLayoutGroup{
			Type:    "Fragment",
			Layouts: c.layouts,
		})
	}
	if shader.Geometry != "" {
		s := filepath.ToSlash(shader.Geometry)
		c, err := readShaderCode(fs, strings.TrimPrefix(s, shaderSrcFolder+"/"))
		if err != nil {
			return shader, err
		}
		shader.LayoutGroups = append(shader.LayoutGroups, rendering.ShaderLayoutGroup{
			Type:    "Geometry",
			Layouts: c.layouts,
		})
	}
	if shader.TessellationControl != "" {
		s := filepath.ToSlash(shader.TessellationControl)
		c, err := readShaderCode(fs, strings.TrimPrefix(s, shaderSrcFolder+"/"))
		if err != nil {
			return shader, err
		}
		shader.LayoutGroups = append(shader.LayoutGroups, rendering.ShaderLayoutGroup{
			Type:    "TessellationControl",
			Layouts: c.layouts,
		})
	}
	if shader.TessellationEvaluation != "" {
		s := filepath.ToSlash(shader.TessellationEvaluation)
		c, err := readShaderCode(fs, strings.TrimPrefix(s, shaderSrcFolder+"/"))
		if err != nil {
			return shader, err
		}
		shader.LayoutGroups = append(shader.LayoutGroups, rendering.ShaderLayoutGroup{
			Type:    "TessellationEvaluation",
			Layouts: c.layouts,
		})
	}
	return shader, nil
}
