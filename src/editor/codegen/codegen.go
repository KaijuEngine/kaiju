/******************************************************************************/
/* codegen.go                                                                 */
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

package codegen

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unsafe"
)

var (
	registerRe = regexp.MustCompile(`engine\.RegisterEntityData\(("{0,1}.*?"{0,1}), &{0,}(.*?)\{\}\)`)
)

type structure struct {
	Doc                string
	Name               string
	Spec               *ast.StructType
	PrimSpec           ast.Expr
	EnumValues         map[string]any
	satisfiesInterface bool
}

func (s *structure) IsValid() bool {
	return s.Spec != nil || s.PrimSpec != nil
}

func (s *structure) IsPrimitiveType() bool {
	return s.Spec == nil && s.PrimSpec != nil
}

func Walk(srcRoot, readRoot *os.Root, pkgPrefix string) ([]GeneratedType, error) {
	defer tracing.NewRegion("codegen.Walk").End()
	return walkInternal(srcRoot, readRoot, pkgPrefix, ".go")
}

func walkInternal(srcRoot, readRoot *os.Root, pkgPrefix, ext string) ([]GeneratedType, error) {
	defer tracing.NewRegion("codegen.walkInternal").End()
	gens := []GeneratedType{}
	skips := []string{}
	registrations := map[string]string{}
	sp := filepath.ToSlash(readRoot.Name()) + "/"
	wErr := filepath.Walk(readRoot.Name(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ext {
			return nil
		}
		path = strings.TrimPrefix(filepath.ToSlash(path), sp)
		if slices.Contains(skips, path) {
			return nil
		}
		if srcRoot != readRoot {
			path = strings.TrimPrefix(filepath.Join(readRoot.Name(), path), srcRoot.Name())
			path = strings.TrimPrefix(filepath.ToSlash(path), "/")
		}
		if g, err := create(srcRoot, path, ext, &skips, &registrations); err == nil {
			for i := range g {
				g[i].PkgPath = pkgPrefix + "/" + strings.TrimPrefix(g[i].PkgPath, sp)
			}
			gens = append(gens, g...)
			return nil
		} else {
			slog.Warn("Failed to create types in file due to not being POD",
				slog.String("error", err.Error()), slog.String("file", path))
		}
		return nil
	})
	return gens, wErr
}

func readAst(srcRoot *os.Root, file string, registrations *map[string]string, localRegs *map[string]string) (*ast.File, error) {
	defer tracing.NewRegion("codegen.readAst").End()
	fs := token.NewFileSet()
	src, err := srcRoot.ReadFile(file)
	if err != nil {
		return nil, err
	}
	ast, err := parser.ParseFile(fs, "", src, parser.ParseComments)
	for _, r := range registerRe.FindAllStringSubmatch(string(src), -1) {
		if len(r) <= 2 {
			continue
		}
		key := r[1]
		if key[0] == '"' {
			key = strings.Trim(key, `"`)
		} else {
			key = findConstValueInAST(ast, key)
		}
		typeName := r[2]
		if key == "" {
			return nil, fmt.Errorf("the registration key for type name '%s' was empty", typeName)
		}
		if _, ok := (*registrations)[key]; ok {
			return nil, fmt.Errorf("the key '%s' has already been registered", key)
		}
		(*registrations)[key] = typeName
		(*localRegs)[typeName] = key
	}
	return ast, err
}

func create(srcRoot *os.Root, file, ext string, skips *[]string, registrations *map[string]string) ([]GeneratedType, error) {
	defer tracing.NewRegion("codegen.create").End()
	genTypes := []GeneratedType{}
	localRegs := map[string]string{}
	a, err := readAst(srcRoot, file, registrations, &localRegs)
	if err != nil {
		return genTypes, err
	}
	if len(localRegs) == 0 {
		return genTypes, nil
	}
	pkgPath := filepath.Dir(file)
	for i := range a.Imports {
		path := strings.Trim(a.Imports[i].Path.Value, `"`)
		// TODO:  Only walk things in this path
		path = path[strings.Index(path, "/")+1:]
		dir, err := os.ReadDir(filepath.Join(srcRoot.Name(), path))
		if err != nil {
			continue
		}
		for i := range dir {
			if !dir[i].IsDir() && filepath.Ext(dir[i].Name()) == ext {
				p := filepath.ToSlash(filepath.Join(path, dir[i].Name()))
				if slices.Contains(*skips, p) {
					continue
				}
				*skips = append(*skips, p)
				t, err := create(srcRoot, p, ext, skips, registrations)
				if err != nil {
					return genTypes, err
				}
				genTypes = append(genTypes, t...)
			}
		}
	}
	types := allTypes(a)
	// Find all sibling files and add them to the list of types
	dir, err := os.ReadDir(filepath.Dir(filepath.Join(srcRoot.Name(), file)))
	if err != nil {
		return genTypes, err
	}
	selfName := filepath.Base(file)
	selfDir := filepath.Dir(file)
	for i := range dir {
		if dir[i].IsDir() {
			continue
		}
		if filepath.Ext(dir[i].Name()) != ext {
			continue
		}
		if dir[i].Name() == selfName {
			continue
		}
		da, err := readAst(srcRoot, filepath.Join(selfDir, dir[i].Name()), registrations, &localRegs)
		if err != nil {
			continue
		}
		types = append(types, allTypes(da)...)
	}
	typesCount := 0
	var lastErr error
	for typesCount != len(types) && len(types) > 0 {
		lastErr = nil
		typesCount = len(types)
		for i := 0; i < len(types); i++ {
			g, err := generateStructType(a.Name.Name, pkgPath, types[i], genTypes)
			if err != nil {
				if _, ok := localRegs[types[i].Name]; ok {
					lastErr = err
				}
				continue
			}
			if k, ok := localRegs[g.Name]; ok {
				g.RegisterKey = k
			}
			g.satisfiesInterface = types[i].satisfiesInterface
			genTypes = append(genTypes, g)
			types = slices.Delete(types, i, i+1)
			i--
		}
	}
	for i := len(genTypes) - 1; i >= 0; i-- {
		if !genTypes[i].satisfiesInterface || genTypes[i].RegisterKey == "" {
			genTypes = klib.RemoveUnordered(genTypes, i)
		}
	}
	return genTypes, lastErr
}

func toType(name string) (reflect.Type, error) {
	defer tracing.NewRegion("codegen.toType").End()
	switch name {
	case "bool":
		return reflect.TypeOf(false), nil
	case "int":
		return reflect.TypeOf(int(0)), nil
	case "int8":
		return reflect.TypeOf(int8(0)), nil
	case "int16":
		return reflect.TypeOf(int16(0)), nil
	case "int32":
		return reflect.TypeOf(int32(0)), nil
	case "int64":
		return reflect.TypeOf(int64(0)), nil
	case "uint":
		return reflect.TypeOf(uint(0)), nil
	case "uint8":
		return reflect.TypeOf(uint8(0)), nil
	case "uint16":
		return reflect.TypeOf(uint16(0)), nil
	case "uint32":
		return reflect.TypeOf(uint32(0)), nil
	case "uint64":
		return reflect.TypeOf(uint64(0)), nil
	case "uintptr":
		return reflect.TypeOf(uintptr(0)), nil
	case "float32":
		return reflect.TypeOf(float32(0)), nil
	case "float64":
		return reflect.TypeOf(float64(0)), nil
	case "complex64":
		return reflect.TypeOf(complex64(0)), nil
	case "complex128":
		return reflect.TypeOf(complex128(0)), nil
	case "string":
		return reflect.TypeOf(""), nil
	case "unsafe.Pointer":
		return reflect.TypeOf(unsafe.Pointer(nil)), nil
	case "interface":
		fallthrough
	case "chan":
		fallthrough
	case "ptr":
		fallthrough
	case "func":
		return reflect.TypeOf(any(nil)), nil
	case "map":
		fallthrough
	case "slice":
		fallthrough
	case "array":
		fallthrough
	case "struct":
		fallthrough
	case "invalid":
		fallthrough
	default:
		return nil, errors.New("invalid type " + name)
	}
}

func genTypeByName(name string, genTypes []GeneratedType) (GeneratedType, bool) {
	for i := range genTypes {
		if genTypes[i].Name == name {
			return genTypes[i], true
		}
	}
	return GeneratedType{}, false
}

func typeFromType(pkg, pkgPath string, typ ast.Expr, genTypes []GeneratedType, ptrDepth *int) (reflect.Type, GeneratedType, error) {
	defer tracing.NewRegion("codegen.typeFromType").End()
	switch t := typ.(type) {
	case *ast.StarExpr:
		*ptrDepth += 1
		return typeFromType(pkg, pkgPath, t.X, genTypes, ptrDepth)
	case *ast.Ident:
		structName := t.Name
		for i := range genTypes {
			if genTypes[i].Name == structName {
				return genTypes[i].Type, genTypes[i], nil
			}
		}
		typ, err := toType(t.Name)
		return typ, GeneratedType{}, err
	case *ast.SelectorExpr:
		structPkg := t.X.(*ast.Ident).Name
		structName := t.Sel.Name
		if t, ok := registry[structPkg+"."+structName]; ok {
			return t, GeneratedType{}, nil
		}
		for i := range genTypes {
			if genTypes[i].Pkg == structPkg && genTypes[i].Name == structName {
				return genTypes[i].Type, GeneratedType{}, nil
			}
		}
		return nil, GeneratedType{}, errors.New("failed to locate the struct type named " + structPkg + "." + structName)
	case *ast.StructType:
		g, err := generateStructType(pkg, pkgPath, structure{
			Doc:  "",
			Name: "",
			Spec: t,
		}, genTypes)
		if err != nil {
			return nil, GeneratedType{}, err
		}
		return g.Type, GeneratedType{}, nil
	case *ast.ArrayType:
		typ, g, err := typeFromType(pkg, pkgPath, t.Elt, genTypes, ptrDepth)
		if err != nil {
			return nil, g, err
		}
		if t.Len == nil {
			return reflect.SliceOf(typ), g, nil
		} else {
			count, err := strconv.Atoi(t.Len.(*ast.BasicLit).Value)
			if err != nil {
				return nil, g, err
			}
			return reflect.ArrayOf(count, typ), g, nil
		}
	case *ast.MapType:
		kType, g, err := typeFromType(pkg, pkgPath, t.Key, genTypes, ptrDepth)
		if err != nil {
			return nil, g, err
		}
		vType, g, err := typeFromType(pkg, pkgPath, t.Value, genTypes, ptrDepth)
		if err != nil {
			return nil, g, err
		}
		return reflect.MapOf(kType, vType), g, nil
	}
	return nil, GeneratedType{}, errors.New("failed to correctly identify the type")
}

func generateStructType(pkg, pkgPath string, s structure, genTypes []GeneratedType) (GeneratedType, error) {
	defer tracing.NewRegion("codegen.generateStructType").End()
	g := GeneratedType{
		Pkg:        pkg,
		PkgPath:    strings.ReplaceAll(pkgPath, "\\", "/"),
		Name:       s.Name,
		Fields:     make([]reflect.StructField, 0),
		EnumValues: s.EnumValues,
	}
	offset := uintptr(0)
	if !s.IsPrimitiveType() {
		for _, f := range s.Spec.Fields.List {
			tag := ""
			if f.Tag != nil {
				tag = strings.Trim(f.Tag.Value, "`")
			}
			ptrDepth := 0
			typ, fGen, err := typeFromType(pkg, pkgPath, f.Type, genTypes, &ptrDepth)
			if err != nil {
				return g, err
			}
			for i := 0; i < ptrDepth; i++ {
				typ = reflect.PointerTo(typ)
			}
			n := f.Names[0].Name
			pkg := g.Pkg
			if !unicode.IsLower([]rune(n)[0]) {
				pkg = ""
			}
			gf := reflect.StructField{
				Name:      n,
				PkgPath:   pkg,
				Tag:       reflect.StructTag(tag),
				Offset:    offset,   // TODO:
				Index:     []int{0}, // TODO:
				Anonymous: false,
				Type:      typ,
			}
			offset += typ.Size()
			g.Fields = append(g.Fields, gf)
			g.FieldGens = append(g.FieldGens, fGen)
			// In Go, how can I make a generated
		}
		g.Type = reflect.StructOf(g.Fields)
	} else {
		ptrDepth := 0
		typ, _, err := typeFromType(pkg, pkgPath, s.PrimSpec, genTypes, &ptrDepth)
		if err != nil {
			return g, err
		}
		g.Type = typ
	}
	return g, nil
}

func allTypes(a *ast.File) []structure {
	defer tracing.NewRegion("codegen.allTypes").End()
	types := make([]structure, 0)
	for _, d := range a.Decls {
		if g, ok := d.(*ast.GenDecl); ok {
			if s, ok := g.Specs[0].(*ast.TypeSpec); ok {
				doc := ""
				if g.Doc != nil {
					doc = strings.TrimSpace(g.Doc.Text())
				}
				st, _ := s.Type.(*ast.StructType)
				ps, isPrim := s.Type.(*ast.Ident)
				target := structure{
					Doc:                doc,
					Name:               s.Name.Name,
					Spec:               st,
					PrimSpec:           ps,
					satisfiesInterface: satisfiesInterface(s, a.Decls),
				}
				if isPrim {
					target.EnumValues = locateAllEnumValues(a, s)
				}
				types = append(types, target)
			}
		}
	}
	return types
}

func hasInterfaceReceiver(f *ast.FuncDecl, name string) bool {
	defer tracing.NewRegion("codegen.hasInterfaceReceiver").End()
	if len(f.Recv.List) != 1 {
		return false
	}
	// if t, ok := f.Recv.List[0].Type.(*ast.StarExpr); !ok {
	// 	return false
	// } else if id, ok := t.X.(*ast.Ident); !ok {
	// 	return false
	// } else if id.Name != name {
	// 	return false
	// }
	if id, ok := f.Recv.List[0].Type.(*ast.Ident); !ok {
		return false
	} else if id.Name != name {
		return false
	}
	return true
}

func hasEntityArg(t ast.Expr) bool {
	defer tracing.NewRegion("codegen.hasEntityArg").End()
	if sx := t.(*ast.StarExpr); sx == nil {
		return false
	} else if sel := sx.X.(*ast.SelectorExpr); sel == nil {
		return false
	} else if sel.Sel.Name != "Entity" {
		return false
	} else if x := sel.X.(*ast.Ident); x == nil || x.Name != "engine" {
		return false
	}
	return true
}

func hasHostArg(t ast.Expr) bool {
	defer tracing.NewRegion("codegen.hasHostArg").End()
	if sx := t.(*ast.StarExpr); sx == nil {
		return false
	} else if sel := sx.X.(*ast.SelectorExpr); sel == nil {
		return false
	} else if sel.Sel.Name != "Host" {
		return false
	} else if x := sel.X.(*ast.Ident); x == nil || x.Name != "engine" {
		return false
	}
	return true
}

func satisfiesInterface(s *ast.TypeSpec, decl []ast.Decl) bool {
	defer tracing.NewRegion("codegen.satisfiesInterface").End()
	// TODO:  Use the reflect of the interface to validate
	for _, d := range decl {
		if f, ok := d.(*ast.FuncDecl); ok {
			if f.Name.Name == "Init" {
				if len(f.Recv.List) == 0 {
					continue
				}
				if len(f.Type.Params.List) != 2 {
					continue
				}
				if !hasInterfaceReceiver(f, s.Name.Name) {
					continue
				}
				if !hasEntityArg(f.Type.Params.List[0].Type) {
					continue
				}
				if !hasHostArg(f.Type.Params.List[1].Type) {
					continue
				}
				return true
			}
		}
	}
	return false
}

func findConstValueInAST(a *ast.File, name string) string {
	for _, d := range a.Decls {
		if g, ok := d.(*ast.GenDecl); ok {
			if s, ok := g.Specs[0].(*ast.ValueSpec); ok {
				if len(s.Names) == 0 || len(s.Values) == 0 {
					continue
				}
				n := s.Names[0].Name
				if n != name {
					continue
				}
				if bl, ok := s.Values[0].(*ast.BasicLit); ok {
					return strings.Trim(bl.Value, `"`)
				}
			}
		}
	}
	return ""
}
