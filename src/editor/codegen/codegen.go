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

package codegen

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"kaiju/klib"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unsafe"
)

type structure struct {
	Doc  string
	Name string
	Spec *ast.StructType
}

func walkInternal(srcPath, pkgPrefix, ext string) ([]GeneratedType, error) {
	gens := []GeneratedType{}
	skips := []string{}
	sp := klib.ToUnixPath(srcPath) + "/"
	wErr := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ext {
			return nil
		}
		path = strings.TrimPrefix(klib.ToUnixPath(path), sp)
		if slices.Contains(skips, path) {
			return nil
		}
		if g, err := create(sp+path, ext, &skips); err == nil {
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

func Walk(srcPath, pkgPrefix string) ([]GeneratedType, error) {
	return walkInternal(srcPath, pkgPrefix, ".go")
}

func create(file, ext string, skips *[]string) ([]GeneratedType, error) {
	genTypes := []GeneratedType{}
	fs := token.NewFileSet()
	src, err := os.ReadFile(file)
	if err != nil {
		return genTypes, err
	}
	pkgPath := filepath.Dir(file)
	a, err := parser.ParseFile(fs, "", src, parser.ParseComments)
	if err != nil {
		return genTypes, err
	}
	for i := range a.Imports {
		path := strings.Trim(a.Imports[i].Path.Value, `"`)
		// TODO:  Only walk things in this path
		path = path[strings.Index(path, "/")+1:]
		dir, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		for i := range dir {
			if !dir[i].IsDir() && filepath.Ext(dir[i].Name()) == ext {
				p := filepath.Join(path, dir[i].Name())
				if slices.Contains(*skips, p) {
					continue
				}
				*skips = append(*skips, p)
				t, err := create(p, ext, skips)
				if err != nil {
					return genTypes, err
				}
				genTypes = append(genTypes, t...)
			}
		}
	}
	types := allTypes(a)
	for i := range types {
		g, err := generateStructType(a.Name.Name, pkgPath, types[i], genTypes)
		if err != nil {
			return genTypes, err
		}
		genTypes = append(genTypes, g)
	}
	return genTypes, nil
}

func toType(name string) (reflect.Type, error) {
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

func typeFromType(pkg, pkgPath string, typ ast.Expr, genTypes []GeneratedType, ptrDepth *int) (reflect.Type, error) {
	switch t := typ.(type) {
	case *ast.StarExpr:
		*ptrDepth += 1
		return typeFromType(pkg, pkgPath, t.X, genTypes, ptrDepth)
	case *ast.Ident:
		return toType(t.Name)
	case *ast.SelectorExpr:
		structPkg := t.X.(*ast.Ident).Name
		structName := t.Sel.Name
		if t, ok := registry[structPkg+"."+structName]; ok {
			return t, nil
		}
		for i := range genTypes {
			if genTypes[i].Pkg == structPkg && genTypes[i].Name == structName {
				return genTypes[i].Type, nil
			}
		}
		return nil, errors.New("failed to locate the struct type named " + structPkg + "." + structName)
	case *ast.StructType:
		g, err := generateStructType(pkg, pkgPath, structure{
			Doc:  "",
			Name: "",
			Spec: t,
		}, genTypes)
		if err != nil {
			return nil, err
		}
		return g.Type, nil
	case *ast.ArrayType:
		typ, err := typeFromType(pkg, pkgPath, t.Elt, genTypes, ptrDepth)
		if err != nil {
			return nil, err
		}
		if t.Len == nil {
			return reflect.SliceOf(typ), nil
		} else {
			count, err := strconv.Atoi(t.Len.(*ast.BasicLit).Value)
			if err != nil {
				return nil, err
			}
			return reflect.ArrayOf(count, typ), nil
		}
	case *ast.MapType:
		kType, err := typeFromType(pkg, pkgPath, t.Key, genTypes, ptrDepth)
		if err != nil {
			return nil, err
		}
		vType, err := typeFromType(pkg, pkgPath, t.Value, genTypes, ptrDepth)
		if err != nil {
			return nil, err
		}
		return reflect.MapOf(kType, vType), nil
	}
	return nil, errors.New("failed to correctly identify the type")
}

func generateStructType(pkg, pkgPath string, s structure, genTypes []GeneratedType) (GeneratedType, error) {
	g := GeneratedType{
		Pkg:     pkg,
		PkgPath: strings.ReplaceAll(pkgPath, "\\", "/"),
		Name:    s.Name,
		Fields:  make([]reflect.StructField, 0),
	}
	offset := uintptr(0)
	for _, f := range s.Spec.Fields.List {
		tag := ""
		if f.Tag != nil {
			tag = strings.Trim(f.Tag.Value, "`")
		}
		ptrDepth := 0
		typ, err := typeFromType(pkg, pkgPath, f.Type, genTypes, &ptrDepth)
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
	}
	g.Type = reflect.StructOf(g.Fields)
	return g, nil
}

func allTypes(a *ast.File) []structure {
	types := make([]structure, 0)
	for _, d := range a.Decls {
		if g, ok := d.(*ast.GenDecl); ok {
			if s, ok := g.Specs[0].(*ast.TypeSpec); ok {
				if satisfiesInterface(s, a.Decls) {
					doc := ""
					if g.Doc != nil {
						doc = strings.TrimSpace(g.Doc.Text())
					}
					types = append(types, structure{
						doc, s.Name.Name, s.Type.(*ast.StructType)})
				}
			}
		}
	}
	return types
}

func hasInterfaceReceiver(f *ast.FuncDecl, name string) bool {
	if len(f.Recv.List) != 1 {
		return false
	}
	if t, ok := f.Recv.List[0].Type.(*ast.StarExpr); !ok {
		return false
	} else if id, ok := t.X.(*ast.Ident); !ok {
		return false
	} else if id.Name != name {
		return false
	}
	return true
}

func hasEntityArg(t ast.Expr) bool {
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
