package plugins

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"kaiju/platform/profiler/tracing"
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
	defer tracing.NewRegion("plugins.reflectCreateDefaultForTypeName").End()
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
	defer tracing.NewRegion("plugins.reflectCommentDocCommonType").End()
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
	defer tracing.NewRegion("plugins.reflectCommentDocTypeHint").End()
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
	defer tracing.NewRegion("plugins.pullSourceForType").End()
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
	defer tracing.NewRegion("plugins.readMethodDoc").End()
	src := ""
	tName := t.Name()
	search := regexp.MustCompile(fmt.Sprintf(`func \(\w+\s+\*{0,}%s\) %s\(`, tName, methodName))
	for i := range sources {
		if search.MatchString(sources[i]) {
			src = sources[i]
			break
		}
	}
	argLen := m.NumIn() - 1
	args = make([]string, 0, argLen)
	failExit := func() (string, []string) {
		for i := range argLen - len(args) {
			args = append(args, fmt.Sprintf("arg%d", i))
		}
		return comment, args
	}
	if src == "" {
		return failExit()
	}
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return failExit()
	}
	found := false
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Recv == nil || fn.Recv.List == nil || len(fn.Recv.List) == 0 {
				continue
			}
			var ident *ast.Ident
			if ptr, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
				ident = ptr.X.(*ast.Ident)
			} else if ident, ok = fn.Recv.List[0].Type.(*ast.Ident); !ok {
				continue
			}
			if ident.Name != tName {
				continue
			}
			if fn.Name.Name == methodName {
				comment = fn.Doc.Text()
				for _, param := range fn.Type.Params.List {
					for _, name := range param.Names {
						args = append(args, name.Name)
					}
				}
				found = true
				break
			}
		}
	}
	if !found {
		return failExit()
	}
	return comment, args
}

func reflectStructAPI(t reflect.Type, apiOut io.StringWriter) {
	defer tracing.NewRegion("plugins.reflectStructAPI").End()
	pt := reflect.PointerTo(t)
	sources, err := pullSourceForType(t)
	if err != nil {
		slog.Error("failed to pull the sources for package, api will be missing function comments and argument names", "package", t.PkgPath(), "error", err)
	}
	methods := make([]reflect.Method, 0, pt.NumMethod())
	for i := range pt.NumMethod() {
		methods = append(methods, pt.Method(i))
	}
	apiOut.WriteString(fmt.Sprintf("---@class %s\n", t.Name()))
	apiOut.WriteString(fmt.Sprintf("%s = {}\n\n", t.Name()))
	apiOut.WriteString(fmt.Sprintf("---@return %s\n", t.Name()))
	apiOut.WriteString(fmt.Sprintf("function %s.New() return nil end\n", t.Name()))
	for _, m := range methods {
		mt := m.Type
		comment, args := readMethodDoc(m.Name, t, mt, sources)
		if comment != "" {
			apiOut.WriteString(fmt.Sprintf("--- %s\n", comment))
		}
		for i := range mt.NumIn() - 1 {
			tName := reflectCommentDocTypeHint(mt.In(i + 1))
			apiOut.WriteString(fmt.Sprintf("---@param %s %s\n", args[i], tName))
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
		apiOut.WriteString(fmt.Sprintf("function %s:%s(%s) %s end\n",
			t.Name(), m.Name, strings.Join(args, ", "), out))
	}
	apiOut.WriteString("\n")
}

func RegenerateAPI() error {
	defer tracing.NewRegion("plugins.RegenerateAPI").End()
	const apiFile = "content/plugins/api.lua"
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
