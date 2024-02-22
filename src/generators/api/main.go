package main

import (
	_ "embed"
	"fmt"
	"io"
	"kaiju/filesystem"
	"kaiju/klib"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
)

//go:embed api_index.md
var apiIndex string

func findRootFolder() (string, error) {
	wd, err := os.Getwd()
	if _, goMain, _, ok := runtime.Caller(0); ok {
		if newWd, pathErr := filepath.Abs(filepath.Dir(goMain)); pathErr == nil {
			wd = filepath.Dir(newWd + "/../../")
		}
	} else if err != nil {
		return "", err
	}
	return wd, nil
}

func generateRaw() {
	const apiDocs = "../docs/api/raw"
	os.Mkdir(apiDocs, os.ModePerm)
	filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		out := filepath.Join(apiDocs, path, "raw.txt")
		os.MkdirAll(filepath.Dir(out), os.ModePerm)
		f, err := os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd := exec.Command("go", "doc", "-C", path, "-all")
		cmd.Stdout = f
		if err := cmd.Start(); err != nil {
			return err
		}
		if err := cmd.Wait(); err != nil && err.Error() != "exit status 1" {
			return err
		}
		println("Copied the raw documentation for package:", path)
		return nil
	})
}

func readRaw() []string {
	var paths []string
	filepath.Walk("./raw", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() != "raw.txt" {
			return nil
		}
		path = path[4:]
		dir := strings.ReplaceAll(filepath.Dir(path), "\\", "/")
		if dir != "" && dir != "." {
			paths = append(paths, dir)
		}
		println("Reading the raw documentation for package:", path)
		err = rawToMarkdown(path)
		if err != nil {
			return err
		}
		println("Created the markdown documentation for package:", path)
		return nil
	})
	slices.Sort(paths)
	return paths
}

func rawToMarkdown(rawPath string) error {
	text, err := filesystem.ReadTextFile(filepath.Join("raw", rawPath))
	if err != nil {
		return err
	}
	mdDir := filepath.Dir(rawPath)
	markdownPath := filepath.Join(mdDir, "index.md")
	os.MkdirAll(filepath.Dir(markdownPath), os.ModePerm)
	f, err := os.Create(markdownPath)
	if err != nil {
		return err
	}
	defer f.Close()
	writeMarkdown(f, mdDir, text)
	return nil
}

type startEnd struct {
	start int
	end   int
}

func writeMarkdown(md io.StringWriter, dir, text string) {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return
	}
	e := len(text)
	if !strings.HasPrefix(text, "package") {
		return
	}
	end := strings.Index(text, "\n")
	if end < 0 {
		end = len(text)
	}
	writePackage(md, dir, text[:end])
	positions := make(map[string]startEnd)
	positions["CONSTANTS"] = startEnd{strings.Index(text, "\nCONSTANTS"), e}
	positions["VARIABLES"] = startEnd{strings.Index(text, "\nVARIABLES"), e}
	positions["FUNCTIONS"] = startEnd{strings.Index(text, "\nFUNCTIONS"), e}
	positions["TYPES"] = startEnd{strings.Index(text, "\nTYPES"), e}
	for k, v := range positions {
		if v.start > 0 {
			positions[k] = startEnd{v.start, v.end}
		} else {
			delete(positions, k)
		}
	}
	for k, v := range positions {
		for _, v2 := range positions {
			if v2.start <= v.start {
				continue
			}
			v.end = min(v.end, v2.start)
			positions[k] = v
		}
	}
	for k, v := range positions {
		v.start += len(k) + 2
		positions[k] = v
	}
	if p, ok := positions["CONSTANTS"]; ok {
		writeConstants(md, text[p.start:p.end])
	}
	if p, ok := positions["VARIABLES"]; ok {
		writeVariables(md, text[p.start:p.end])
	}
	if p, ok := positions["FUNCTIONS"]; ok {
		writeFunctions(md, text[p.start:p.end])
	}
	if p, ok := positions["TYPES"]; ok {
		writeTypes(md, text[p.start:p.end])
	}
}

func writePackage(md io.StringWriter, dir, line string) {
	parts := strings.Split(line, " // ")
	md.WriteString("# ")
	md.WriteString(parts[0])
	md.WriteString("\n```go\n")
	md.WriteString("import kaiju/" + dir)
	md.WriteString("\n```\n\n")
}

func writeConst(md io.StringWriter, name, value string) {
	md.WriteString("#### ")
	md.WriteString(name)
	if value != "" {
		md.WriteString("\n")
		if strings.Contains(value, ".") && value[0] != '"' {
			ps := strings.Split(value, ".")
			md.WriteString(fmt.Sprintf("[%s](../%s).[%s](../%s#%s)",
				ps[0], ps[0], ps[1], ps[0], ps[1]))
		} else {
			md.WriteString("`")
			md.WriteString(value)
			md.WriteString("`")
		}
	}
	md.WriteString("\n")
}

func writeVar(md io.StringWriter, name, value string) {
	md.WriteString("#### ")
	md.WriteString(name)
	if value != "" {
		md.WriteString("\n")
		if strings.Contains(value, "\n") {
			md.WriteString("```go\n")
			md.WriteString(value)
			md.WriteString("\n```")
		} else {
			md.WriteString("`")
			md.WriteString(value)
			md.WriteString("`")
		}
	}
	md.WriteString("\n")
}

func writeConstants(md io.StringWriter, text string) {
	// TODO:  Add the comment documentation
	md.WriteString("## Constants\n")
	lineReg := regexp.MustCompile(`const\s+(\w+)\s+=\s+(\w+)$`)
	blockReg := regexp.MustCompile(`(?s)const\s+\((.*?)\n\)`)
	blockLineReg := regexp.MustCompile(`\s+(\w+)(\s+=\s+(.*)|)`)
	blocks := blockReg.FindAllString(text, -1)
	lines := lineReg.FindAllStringSubmatch(text, -1)
	for _, match := range lines {
		writeConst(md, match[1], match[2])
	}
	for _, block := range blocks {
		lines := blockLineReg.FindAllStringSubmatch(block, -1)
		for _, line := range lines {
			writeConst(md, line[1], line[3])
		}
	}
}

func writeVariables(md io.StringWriter, text string) {
	// TODO:  Add the comment documentation
	md.WriteString("## Variables\n")
	lineReg := regexp.MustCompile(`(?s)var\s+(\w+)\s+=\s+(.*?)[^{,]$`)
	blockReg := regexp.MustCompile(`(?s)var\s+\((.*?)\n\)`)
	blockLineReg := regexp.MustCompile(`(?s)\s+(\w+)\s+=\s+(.*?)$`)
	blocks := blockReg.FindAllString(text, -1)
	lines := lineReg.FindAllStringSubmatch(text, -1)
	for _, match := range lines {
		writeVar(md, match[1], match[2])
	}
	for _, block := range blocks {
		lines := blockLineReg.FindAllStringSubmatch(block, -1)
		for _, line := range lines {
			writeVar(md, line[1], line[2])
		}
	}
}

func writeFunctions(md io.StringWriter, text string) {
	md.WriteString("## Functions\n")
	src := strings.ReplaceAll(text, "\r", "")
	src = strings.TrimSpace(src)
	lines := strings.Split(src, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "func ") {
			md.WriteString("### ")
			md.WriteString(strings.TrimSpace(line[5:strings.Index(line, "(")]))
			md.WriteString("\n")
			md.WriteString("```go\n")
			md.WriteString(line)
			md.WriteString("\n```\n\n")
		} else {
			md.WriteString(strings.TrimSpace(line))
			md.WriteString("\n")
		}
	}
}

func writeTypes(md io.StringWriter, text string) {
	md.WriteString("## Types\n")
	src := strings.ReplaceAll(text, "\r", "")
	src = strings.TrimSpace(src)
	lines := strings.Split(src, "\n")
	name := ""
	reg := regexp.MustCompile(`type\s+(\w+)[\s=]+([\[\]\w\.]+)`)
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "type ") {
			md.WriteString("### ")
			parts := reg.FindAllStringSubmatch(line, -1)
			name = parts[0][1]
			key := parts[0][2]
			md.WriteString(name)
			md.WriteString("\n")
			switch key {
			case "interface":
				md.WriteString("`interface`\n")
			case "struct":
				md.WriteString("`struct`\n")
			default:
				// TODO:  Check if this is a package type pkg.type
				if strings.Contains(key, ".") {
					p := strings.Split(key, ".")
					pkg := p[0]
					key = p[1]
					md.WriteString(fmt.Sprintf("[%s](../%s).[%s](../%s#%s)",
						pkg, pkg, key, pkg, key))
				} else {
					md.WriteString(fmt.Sprintf("[%s](#%s)", key, key))
				}
			}
			md.WriteString("\n")
			md.WriteString("```go\n")
			if strings.Contains(lines[i], "}") {
				md.WriteString(line)
			} else {
				for j := i; j < len(lines); j++ {
					if strings.Contains(lines[j], "}") {
						break
					}
					md.WriteString(lines[j])
					md.WriteString("\n")
					i++
				}
				md.WriteString("}")
			}
			md.WriteString("\n```\n\n")
		} else if strings.HasPrefix(line, "func ") {
			md.WriteString("#### ")
			n := line[5:]
			if strings.HasPrefix(line, "func (") {
				n = line[strings.Index(line, ")")+1:]
				md.WriteString(name)
				md.WriteString(".")
			}
			n = n[:strings.Index(n, "(")]
			md.WriteString(strings.TrimSpace(n))
			md.WriteString("\n")
			md.WriteString("```go\n")
			md.WriteString(line)
			md.WriteString("\n```\n\n")
		} else {
			md.WriteString(strings.TrimSpace(line))
			md.WriteString("\n")
		}
	}
}

func writeIndex(paths []string) {
	println("Writing the api/index.md file")
	idx := klib.MustReturn(os.Create("index.md"))
	defer idx.Close()
	idx.WriteString(apiIndex)
	idx.WriteString("\n")
	mapping := make(map[string]bool)
	for _, path := range paths {
		mapping[path] = true
	}
	for k := range mapping {
		parts := strings.Split(k, "/")
		for i := 0; i < len(parts)-1; i++ {
			k := strings.Join(parts[:i+1], "/")
			if _, ok := mapping[k]; !ok {
				mapping[k] = false
			}
		}
	}
	for _, path := range paths {
		tabs := strings.Count(path, "/")
		for i := 0; i < tabs; i++ {
			idx.WriteString("\t")
		}
		if mapping[path] {
			idx.WriteString(fmt.Sprintf("- [%s](%s/)\n", path, path))
		} else {
			idx.WriteString("- ")
			idx.WriteString(path)
			idx.WriteString("\n")
		}
	}
	println("Finished writing the api/index.md file")
}

func main() {
	root, err := findRootFolder()
	if err != nil {
		panic(err)
	}
	klib.Must(os.Chdir(root))
	generateRaw()
	klib.Must(os.Chdir("../docs/api"))
	paths := readRaw()
	writeIndex(paths)
	//rawToMarkdown("assets/raw.txt")
	println("Finished generating the api documentation")
}
