package main

import (
	"bufio"
	"encoding/json"
	"kaiju/klib/string_equations"
	"kaiju/rendering"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	layoutReg = regexp.MustCompile(`(?s)\s*layout\s*\(([\w\s=\d,\+\-\*\/]+)\)\s*(?:readonly\s+)?(in|out|uniform)\s+([a-zA-Z0-9]+)\s+([a-zA-Z0-9_]+){0,1}(?:\s*\{(.*?)\})?\s*(\w+){0,1}`)
)

type ShaderSource struct {
	src     string
	file    string
	defines map[string]any
	layouts []rendering.ShaderLayout
}

func (s *ShaderSource) defineAsString(name string) string {
	if d, ok := s.defines[name]; ok {
		if v, ok := d.(string); ok {
			return v
		} else {
			return strconv.FormatFloat(d.(float64), 'G', 10, 64)
		}
	}
	return name
}

func (s *ShaderSource) processArrayField(value string) string {
	re := regexp.MustCompile(`\[([[\w\d\s\+\-\*\/]+)\]`)
	start := strings.Index(value, "[")
	if start < 0 {
		return value
	}
	matches := re.FindAllStringSubmatch(value, -1)
	sb := strings.Builder{}
	sb.WriteString(value[:start])
	for i := range matches {
		v, err := s.processDefineEquation(matches[i][1])
		if err != nil {
			log.Fatalf("failed to process the equation for the array size (%s): %s", matches[i][0], err)
		}
		sb.WriteRune('[')
		sb.WriteString(strconv.FormatFloat(v, 'G', 10, 64))
		sb.WriteRune(']')
	}
	return sb.String()
}

func (s *ShaderSource) processDefineEquation(value string) (float64, error) {
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

func (s *ShaderSource) readDefines() {
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

func (s *ShaderSource) readLayouts() {
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
				s.layouts[i].Fields[j] = rendering.ShaderLayoutStructField{
					Type: parts[0],
					Name: s.processArrayField(parts[1]),
				}
			}
		}
	}
}

func readImports(inSrc, path string) string {
	src := strings.Builder{}
	scan := bufio.NewScanner(strings.NewReader(inSrc))
	re := regexp.MustCompile(`\s{0,}#include\s+\"([\w\.]+)\"`)
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		match := re.FindStringSubmatch(line)
		if len(match) == 2 && match[1] != "" {
			importSrc, err := os.ReadFile(filepath.Join(path, match[1]))
			if err != nil {
				log.Fatalf("failed to load import file (%s): %s", match[1], err)
			}
			src.WriteString(readImports(string(importSrc), path))
		} else {
			src.WriteString(line + "\n")
		}
	}
	return src.String()
}

func readShaderCode(file string) ShaderSource {
	file = strings.Replace(strings.TrimSuffix(file, ".spv"), "/spv/", "/", 1)
	source := ShaderSource{
		file:    "content/" + file,
		defines: make(map[string]any),
	}
	data, err := os.ReadFile(source.file)
	if err != nil {
		log.Fatalf("failed to read the file: %s", err)
	}
	source.src = readImports(string(data), filepath.Dir(source.file))
	source.readDefines()
	source.readLayouts()
	return source
}

func processFile(jsonFile string) {
	d, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("failed to read the shader definition file: %s", err)
	}
	def, err := rendering.ShaderDefFromJson(string(d))
	if err != nil {
		log.Fatalf("failed to parse the shader definition file: %s", err)
	}
	def.LayoutGroups = make([]rendering.ShaderLayoutGroup, 0)
	if def.Vulkan.Vert != "" {
		c := readShaderCode(def.Vulkan.Vert)
		def.LayoutGroups = append(def.LayoutGroups, rendering.ShaderLayoutGroup{
			Type:    "Vertex",
			Layouts: c.layouts,
		})
	}
	if def.Vulkan.Frag != "" {
		c := readShaderCode(def.Vulkan.Frag)
		def.LayoutGroups = append(def.LayoutGroups, rendering.ShaderLayoutGroup{
			Type:    "Fragment",
			Layouts: c.layouts,
		})
	}
	if def.Vulkan.Geom != "" {
		c := readShaderCode(def.Vulkan.Geom)
		def.LayoutGroups = append(def.LayoutGroups, rendering.ShaderLayoutGroup{
			Type:    "Geometry",
			Layouts: c.layouts,
		})
	}
	if def.Vulkan.Tesc != "" {
		c := readShaderCode(def.Vulkan.Tesc)
		def.LayoutGroups = append(def.LayoutGroups, rendering.ShaderLayoutGroup{
			Type:    "TessellationControl",
			Layouts: c.layouts,
		})
	}
	if def.Vulkan.Tese != "" {
		c := readShaderCode(def.Vulkan.Tese)
		def.LayoutGroups = append(def.LayoutGroups, rendering.ShaderLayoutGroup{
			Type:    "TessellationEvaluation",
			Layouts: c.layouts,
		})
	}
	if out, err := json.Marshal(def); err == nil {
		if err := os.WriteFile(jsonFile, out, os.ModePerm); err != nil {
			log.Fatalf("failed to write the shader definition file %s", jsonFile)
		} else {
			log.Printf("Updated shader description: %s", jsonFile)
		}
	} else {
		log.Fatalf("failed to serialize the layout for %s", jsonFile)
	}
}

func main() {
	const defFolder = "content/shaders/definitions"
	entries, err := os.ReadDir(defFolder)
	if err != nil {
		log.Fatalf("failed to read the shader definition folder %s", defFolder)
	}
	for i := range entries {
		if entries[i].IsDir() {
			continue
		}
		processFile(filepath.Join(defFolder, entries[i].Name()))
	}
}
