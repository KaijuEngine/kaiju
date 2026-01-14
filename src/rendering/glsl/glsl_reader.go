package glsl

import (
	"bufio"
	"fmt"
	"kaiju/klib/string_equations"
	"kaiju/rendering"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	layoutReg        = regexp.MustCompile(`(?s)\s*layout\s*\(([\w\s=\d,\+\-\*\/]+)\)\s*(?:readonly\s+)?(in|out|uniform|buffer)\s+([a-zA-Z0-9]+)\s+([a-zA-Z0-9_]+){0,1}(?:\s*\{(.*?)\})?\s*(\w+){0,1}(\[(.*?)\]){0,1}`)
	computeLayoutReg = regexp.MustCompile(`(?s)\s*layout\s*\(\s*local_size_x\s*=\s*(\d+)\s*,\s*local_size_y\s*=\s*(\d+)\s*,\s*local_size_z\s*=\s*(\d+)\)\s*in\s*;`)
)

type ShaderSource struct {
	file            string
	src             string
	defines         map[string]any
	WorkGroups      [3]uint32
	Layouts         []rendering.ShaderLayout
	preprocDepth    int
	preprocRead     []bool
	preprocReadAny  []bool
	multilineDefine bool
	lastDefineKey   string
}

func Parse(path string, args string) (ShaderSource, error) {
	source := ShaderSource{
		file:    path,
		defines: make(map[string]any),
	}
	data, err := os.ReadFile(source.file)
	if err != nil {
		slog.Error("failed to read the file", "file", path, "error", err)
		return source, err
	}
	source.readArgs(args)
	source.src = readImports(string(data), filepath.Dir(source.file))
	source.readPreprocessor()
	if source.IsCompute() {
		if err := source.readComputeLayouts(); err != nil {
			return source, err
		}
	} else if err := source.readLayouts(); err != nil {
		return source, err
	}
	return source, nil
}

func (s *ShaderSource) Type() string {
	switch filepath.Ext(s.file) {
	case ".vert":
		return "Vertex"
	case ".frag":
		return "Fragment"
	case ".geom":
		return "Geometry"
	case ".tesc":
		return "TessellationControl"
	case ".tese":
		return "TessellationEvaluation"
	case ".comp":
		return "Compute"
	default:
		slog.Error("invalid shader file extension", "extension", filepath.Ext(s.file))
		return ""
	}
}

func (s *ShaderSource) IsCompute() bool {
	return s.Type() == "Compute"
}

func (s *ShaderSource) readArgs(args string) {
	a := strings.Split(args, "-")
	for i := range a {
		if len(a[i]) == 0 {
			continue
		}
		if rune(a[i][0]) == 'D' {
			parts := strings.Split(a[i], " ")
			if len(parts) > 1 {
				s.defines[parts[0]] = parts[1]
			} else {
				s.defines[parts[0]] = nil
			}
		}
	}
}

type srcLine string

func srcLineFromString(str string) srcLine {
	return srcLine(strings.TrimSpace(str))
}

func (s srcLine) prefixed(pre string) bool {
	return strings.HasPrefix(string(s), pre)
}

func (s srcLine) suffixed(pre string) bool {
	return strings.HasSuffix(string(s), pre)
}

func (s srcLine) isComment() bool       { return s.prefixed("//") }
func (s srcLine) isPreprocDefine() bool { return s.prefixed("#define") }
func (s srcLine) isPreprocIf() bool     { return s.prefixed("#if") }
func (s srcLine) isPreprocElseIf() bool { return s.prefixed("#elif") }
func (s srcLine) isPreprocElse() bool   { return s.prefixed("#else") }
func (s srcLine) isPreprocEndIf() bool  { return s.prefixed("#endif") }
func (s srcLine) isPreprocIfDef() bool  { return s.prefixed("#ifdef") }
func (s srcLine) isPreprocIfNDef() bool { return s.prefixed("#ifndef") }
func (s srcLine) hasDefineSlash() bool  { return s.suffixed("\\") }

func (s srcLine) string() string {
	str := strings.Split(string(s), "//")[0]
	return strings.TrimSuffix(str, "\\")
}

func (s *ShaderSource) readPreprocessor() {
	re := regexp.MustCompile(`\s*#define\s+(\w+)(?:\s+([\w\d\s\+\-\*\/]+))?`)
	cRe := regexp.MustCompile(`#(if[n]*def)\s+(\w+)`)
	c2Re := regexp.MustCompile(`#(e{0,1}l{0,1}if)\s+(!{0,1})defined\((\w+)\)`)
	ops := []string{"+", "-", "*", "/"}
	scan := bufio.NewScanner(strings.NewReader(s.src))
	sb := strings.Builder{}
	sb.Grow(len(s.src))
	s.preprocRead = []bool{true}
	s.preprocReadAny = []bool{true}
	for scan.Scan() {
		rawLine := scan.Text()
		line := srcLineFromString(rawLine)
		if line.isComment() {
			continue
		}
		if line.isPreprocIf() {
			s.preprocRead = append(s.preprocRead, s.preprocRead[s.preprocDepth])
			s.preprocReadAny = append(s.preprocReadAny, false)
			s.preprocDepth++
		} else if line.isPreprocEndIf() {
			s.preprocRead = s.preprocRead[:s.preprocDepth]
			s.preprocReadAny = s.preprocReadAny[:s.preprocDepth]
			s.preprocDepth--
			continue
		} else if line.isPreprocElse() {
			if s.preprocRead[s.preprocDepth-1] && !s.preprocReadAny[s.preprocDepth] {
				s.preprocRead[s.preprocDepth] = true
			}
			continue
		}
		if s.preprocRead[s.preprocDepth] && (line.isPreprocIfDef() || line.isPreprocIfNDef()) {
			cMatch := cRe.FindStringSubmatch(line.string())
			if len(cMatch) == 3 {
				_, ok := s.defines[cMatch[2]]
				switch cMatch[1] {
				case "ifndef":
					s.preprocRead[s.preprocDepth] = !ok
				case "ifdef":
					s.preprocRead[s.preprocDepth] = ok
				}
				continue
			}
		}
		if line.isPreprocIf() || line.isPreprocElseIf() {
			c2Match := c2Re.FindStringSubmatch(line.string())
			if len(c2Match) == 4 {
				_, ok := s.defines[c2Match[3]]
				if c2Match[2] == "!" {
					ok = !ok
				}
				switch c2Match[1] {
				case "if":
					s.preprocRead[s.preprocDepth] = ok
				case "elif":
					s.preprocRead[s.preprocDepth] = !s.preprocRead[s.preprocDepth] && ok
				}
				continue
			}
		}
		if !s.preprocRead[s.preprocDepth] {
			continue
		} else {
			s.preprocReadAny[s.preprocDepth] = true
		}
		if s.lastDefineKey != "" {
			s.multilineDefine = line.hasDefineSlash()
			if v, ok := s.defines[s.lastDefineKey]; ok {
				if str, ok := v.(string); ok {
					str += "\n" + line.string()
					s.defines[s.lastDefineKey] = str
				}
			}
			if !s.multilineDefine {
				s.lastDefineKey = ""
			}
		} else {
			s.multilineDefine = false
		}
		if line.isPreprocDefine() {
			s.multilineDefine = line.hasDefineSlash()
			match := re.FindStringSubmatch(line.string())
			if len(match) == 3 {
				name := match[1]
				value := match[2]
				isEquation := false
				if value != "" {
					for j := range ops {
						isEquation = isEquation || strings.Contains(value, ops[j])
					}
				}
				if isEquation {
					if v, err := s.processDefineEquation(value); err == nil {
						s.defines[name] = v
					} else {
						slog.Error("error processing equation", "equation", match[0], "error", err)
						return
					}
				} else {
					if value == "" {
						s.defines[name] = nil
					} else if f, err := strconv.ParseFloat(value, 64); err == nil {
						s.defines[name] = f
					} else {
						s.defines[name] = srcLineFromString(value).string()
					}
				}
				if s.multilineDefine {
					s.lastDefineKey = name
				}
			}
		} else if rawLine != "" {
			sb.WriteString(rawLine)
			sb.WriteRune('\n')
		}
	}
	s.src = sb.String()
	for k, v := range s.defines {
		if v != nil {
			repl := regexp.MustCompile(fmt.Sprintf(`\b%s\b`, k))
			s.src = repl.ReplaceAllString(s.src, fmt.Sprintf("%v", v))
		}
	}
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

func (s *ShaderSource) readComputeLayouts() error {
	matches := computeLayoutReg.FindAllStringSubmatch(s.src, -1)
	s.Layouts = make([]rendering.ShaderLayout, 0, len(matches))
	for i := range matches {
		x, _ := strconv.Atoi(matches[i][1])
		y, _ := strconv.Atoi(matches[i][2])
		z, _ := strconv.Atoi(matches[i][3])
		s.WorkGroups = [3]uint32{uint32(x), uint32(y), uint32(z)}
		break
	}
	if err := s.readLayouts(); err != nil {
		return err
	}
	for i := range s.Layouts {
		// TODO:  Accurate?
		s.Layouts[i].Type = "StorageBuffer"
	}
	return nil
}

func (s *ShaderSource) readLayouts() error {
	matches := layoutReg.FindAllStringSubmatch(s.src, -1)
	s.Layouts = make([]rendering.ShaderLayout, 0, len(matches))
	for i := range matches {
		name := matches[i][4]
		if name == "" || matches[i][3] == "flat" {
			name = matches[i][6]
		}
		s.Layouts = append(s.Layouts, rendering.ShaderLayout{
			Location:        -1,
			Binding:         -1,
			Count:           1,
			Set:             -1,
			InputAttachment: -1,
			Type:            matches[i][3],
			Name:            name,
			Source:          matches[i][2],
		})
		layout := &s.Layouts[len(s.Layouts)-1]
		if matches[i][8] != "" {
			v, err := s.processDefineEquation(matches[i][8])
			if err != nil {
				slog.Error("invalid array value for layout", "value", matches[i][8], "error", err)
				return err
			}
			layout.Count = int(v)
		}
		attrs := strings.Split(matches[i][1], ",")
		for j := range attrs {
			parts := strings.Fields(attrs[j])
			val, err := s.processDefineEquation(strings.Join(parts[2:], " "))
			if err != nil {
				slog.Error("invalid value for layout", "value", matches[i][0], "error", err)
			}
			switch parts[0] {
			case "location":
				layout.Location = int(val)
				if layout.Count > 1 {
					for j := 1; j < layout.Count; j++ {
						l := *layout
						l.Count = 1
						l.Location = layout.Location + j
						s.Layouts = append(s.Layouts, l)
					}
					layout.Count = 1
				}
			case "binding":
				layout.Binding = int(val)
			case "set":
				layout.Set = int(val)
			case "input_attachment_index":
				layout.InputAttachment = int(val)
			}
		}
		if matches[i][5] != "" {
			fields := strings.Split(strings.TrimSpace(matches[i][5]), ";")
			if len(fields) > 0 && fields[len(fields)-1] == "" {
				fields = fields[:len(fields)-1]
			}
			layout.Fields = make([]rendering.ShaderLayoutStructField, len(fields))
			for j := range fields {
				parts := strings.Fields(fields[j])
				name, err := s.processArrayField(parts[1])
				if err != nil {
					return err
				}
				layout.Fields[j] = rendering.ShaderLayoutStructField{
					Type: parts[0],
					Name: name,
				}
			}
		}
	}
	return nil
}

func (s *ShaderSource) processArrayField(value string) (string, error) {
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
				slog.Error("failed to load import file", "file", match[1], "error", err)
				return ""
			}
			src.WriteString(readImports(string(importSrc), path))
		} else {
			src.WriteString(line + "\n")
		}
	}
	return src.String()
}
