package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"kaiju/klib"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Rect struct {
	Left   float32 `json:"left"`
	Top    float32 `json:"top"`
	Right  float32 `json:"right"`
	Bottom float32 `json:"bottom"`
}

type Glyph struct {
	Unicode     int  `json:"unicode"`
	Advance     int  `json:"advance"`
	PlaneBounds Rect `json:"planeBounds"`
	AtlasBounds Rect `json:"atlasBounds"`
}

type Atlas struct {
	Width  int32 `json:"width"`
	Height int32 `json:"height"`
}

type Metrics struct {
	EmSize             float32 `json:"emSize"`
	LineHeight         float32 `json:"lineHeight"`
	Ascender           float32 `json:"ascender"`
	Descender          float32 `json:"descender"`
	UnderlineY         float32 `json:"underlineY"`
	UnderlineThickness float32 `json:"underlineThickness"`
}

type Kerning struct {
	Unicode1 int32   `json:"unicode1"`
	Unicode2 int32   `json:"unicode2"`
	Advance  float32 `json:"advance"`
}

type FontData struct {
	Glyphs  []Glyph   `json:"glyphs"`
	Atlas   Atlas     `json:"atlas"`
	Metrics Metrics   `json:"metrics"`
	Kerning []Kerning `json:"kerning"`
}

func main() {
	ttfName := os.Args[1]
	ttfFile := ttfName + ".ttf"
	jsonFile := ttfName + ".json"
	binFile := ttfName + ".bin"
	pngFile := ttfName + ".png"
	cmd := exec.Command("./msdf-atlas-gen.exe",
		"-font", ttfFile,
		"-charset", "charset.txt",
		"-fontname", ttfName,
		"-type", "msdf",
		"-format", "png",
		"-pots",
		"-json", jsonFile,
		"-imageout", pngFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out := klib.MustReturn(cmd.StdoutPipe())
	scanner := bufio.NewScanner(out)
	klib.Must(cmd.Start())
	for scanner.Scan() {
		println(scanner.Text())
	}
	klib.Must(cmd.Wait())
	jsonBin := klib.MustReturn(os.ReadFile(jsonFile))
	fout := klib.MustReturn(os.Create(binFile))
	defer fout.Close()
	var fontData FontData
	klib.Must(json.NewDecoder(strings.NewReader(string(jsonBin))).Decode(&fontData))
	binary.Write(fout, binary.LittleEndian, int32(len(fontData.Glyphs)))
	binary.Write(fout, binary.LittleEndian, fontData.Atlas.Width)
	binary.Write(fout, binary.LittleEndian, fontData.Atlas.Height)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.EmSize)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.LineHeight)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.Ascender)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.Descender)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.UnderlineY)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.UnderlineThickness)
	for _, glyph := range fontData.Glyphs {
		binary.Write(fout, binary.LittleEndian, int32(glyph.Unicode))
		binary.Write(fout, binary.LittleEndian, glyph.Advance)
		binary.Write(fout, binary.LittleEndian, glyph.PlaneBounds.Left)
		binary.Write(fout, binary.LittleEndian, glyph.PlaneBounds.Top)
		binary.Write(fout, binary.LittleEndian, glyph.PlaneBounds.Right)
		binary.Write(fout, binary.LittleEndian, glyph.PlaneBounds.Bottom)
		binary.Write(fout, binary.LittleEndian, glyph.AtlasBounds.Left)
		binary.Write(fout, binary.LittleEndian, glyph.AtlasBounds.Top)
		binary.Write(fout, binary.LittleEndian, glyph.AtlasBounds.Right)
		binary.Write(fout, binary.LittleEndian, glyph.AtlasBounds.Bottom)
	}
	os.Remove(jsonFile)
}
