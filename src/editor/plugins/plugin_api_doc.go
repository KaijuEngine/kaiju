//go:build editor

package plugins

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

var pkgSources = map[string][]string{}

const prefefinedAPIDocs = `---@Shape Pointer
--- Start an interactive debugger in the console window
function breakpoint() end
`

func reflectCreateDefaultForTypeName(name string) string {
	switch name {
	case "boolean":
		return "false"
	case "string":
		return `""`
	case "number":
		return "0"
	default:
		return name + ".New()"
	}
}

func reflectCommentDocCommonType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Pointer:
		//if t.Elem().Kind() == reflect.Pointer {
		return "Pointer"
		//}
		//return reflectCommentDocCommonType(t.Elem())
	default:
		return t.Name()
	}
}

func reflectCommentDocTypeHint(t reflect.Type) string {
	tName := reflectCommentDocCommonType(t)
	switch t.Kind() {
	case reflect.Map:
		// TODO:  Implement maps type
	case reflect.Array, reflect.Slice:
		if tName == "" {
			e := t.Elem()
			depth := 1
			for e.Kind() == reflect.Array || e.Kind() == reflect.Slice {
				e = e.Elem()
				depth++
			}
			tName = reflectCommentDocCommonType(t.Elem()) + strings.Repeat("[]", depth)
		}
	}
	return tName
}

func pullSourceForType(t reflect.Type) ([]string, error) {
	pkg := t.PkgPath()
	if sources, ok := pkgSources[pkg]; ok {
		return sources, nil
	}
	srcPath := strings.Replace(pkg, "kaiju/", "src/", 1)
	files, err := os.ReadDir(srcPath)
	if err != nil {
		return []string{}, err
	}
	sources := make([]string, 0, len(files))
	for i := range files {
		if files[i].IsDir() {
			continue
		}
		if filepath.Ext(files[i].Name()) != ".go" {
			continue
		}
		p := filepath.Join(srcPath, files[i].Name())
		src, err := os.ReadFile(p)
		if err != nil {
			return []string{}, err
		}
		sources = append(sources, string(src))
	}
	pkgSources[pkg] = sources
	return sources, nil
}

func readMethodDoc(methodName string, t reflect.Type, m reflect.Type, sources []string) (comment string, args []string) {
	src := ""
	search := regexp.MustCompile(fmt.Sprintf(`func \(\w+\s+\*{0,}%s\) %s\(`, t.Name(), methodName))
	for i := range sources {
		if search.MatchString(sources[i]) {
			src = sources[i]
			break
		}
	}
	args = make([]string, m.NumIn()-1)
	failExit := func() (string, []string) {
		for i := range len(args) {
			args[i] = fmt.Sprintf("arg%d", i)
		}
		return comment, args
	}
	if src == "" {
		return failExit()
	}
	lines := strings.Split(src, "\n")
	idx := -1
	for i := range lines {
		if search.MatchString(lines[i]) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return failExit()
	}
	sb := strings.Builder{}
	commentStart := idx
	for i := idx - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			if !strings.HasPrefix(line, "//") {
				break
			}
			commentStart--
		}
	}
	for i := commentStart; i < idx; i++ {
		sb.WriteString(strings.TrimSpace(strings.TrimLeft(lines[i], "/")))
		sb.WriteRune('\n')
	}
	comment = strings.TrimSpace(sb.String())
	argSearch := regexp.MustCompile(fmt.Sprintf(`(?s)%s\((.*?)\)`, methodName))
	funcLineEnd := idx
	for i := idx; i < len(lines); i++ {
		if strings.Contains(lines[i], "{") {
			break
		}
		funcLineEnd++
	}
	signature := strings.Join(lines[idx:funcLineEnd+1], " ")
	found := argSearch.FindAllStringSubmatch(signature, -1)
	if len(found) < 1 || len(found[0]) < 2 || found[0][1] == "" {
		return failExit()
	}
	parts := strings.Split(found[0][1], ",")
	for i := range min(len(args), len(parts)) {
		args[i] = strings.Split(parts[i], " ")[0]
	}
	return comment, args
}

func reflectStructAPI(t reflect.Type, apiOut io.StringWriter) {
	pt := reflect.PointerTo(t)
	sources, err := pullSourceForType(t)
	if err != nil {
		slog.Error("failed to pull the sources for package, api will be missing function comments and argument names", "package", t.PkgPath(), "error", err)
	}
	methods := make([]reflect.Method, 0, pt.NumMethod())
	for i := range pt.NumMethod() {
		methods = append(methods, pt.Method(i))
	}
	for _, m := range methods {
		mt := m.Type
		ins := make([]string, mt.NumIn()-1)
		comment, args := readMethodDoc(m.Name, t, mt, sources)
		if comment != "" {
			apiOut.WriteString(fmt.Sprintf("--- %s\n", comment))
		}
		for i := range mt.NumIn() - 1 {
			ins[i] = args[i]
			tName := reflectCommentDocTypeHint(mt.In(i + 1))
			apiOut.WriteString(fmt.Sprintf("---@param %s %s\n", ins[i], tName))
		}
		outs := make([]string, mt.NumOut())
		for i := range mt.NumOut() {
			o := mt.Out(i)
			tName := reflectCommentDocTypeHint(o)
			apiOut.WriteString(fmt.Sprintf("---@return %s\n", tName))
			switch o.Kind() {
			case reflect.Bool:
				outs[i] = "false"
			case reflect.String:
				outs[i] = `""`
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
				reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32,
				reflect.Uint64, reflect.Float32, reflect.Float64:
				outs[i] = "0"
			case reflect.Array, reflect.Slice:
				if o.Name() != "" {
					outs[i] = o.Name() + ".New()"
				} else {
					outs[i] = "{}"
				}
			case reflect.Pointer:
				outs[i] = reflectCreateDefaultForTypeName(
					reflectCommentDocCommonType(o))
			default:
				outs[i] = o.Name() + ".New();"
			}
		}
		out := "return " + strings.Join(outs, ", ")
		apiOut.WriteString(fmt.Sprintf("function %s.%s(%s) %s end\n",
			t.Name(), m.Name, strings.Join(ins, ", "), out))
	}
	apiOut.WriteString("\n")
}

func RegenerateAPI() error {
	const apiFile = "content/editor/plugins/api.lua"
	f, err := os.OpenFile(apiFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	clear(pkgSources)
	f.WriteString(prefefinedAPIDocs)
	for _, t := range reflectedTypes() {
		reflectStructAPI(t, f)
	}
	return nil
}
