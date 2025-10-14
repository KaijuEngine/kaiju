package font_to_msdf

import (
	"encoding/json"
	"kaiju/build"
	"kaiju/klib"
	"kaiju/rendering/loaders/kaiju_font"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const binDir = "../tools/content_tools/"

// ProcessTTF will take in a path to a TTF file and a set of characters that
// need to be rendered and generate the engine's proprietary font rendering
// data object. To do this, it uses the open source msdf-atlas-gen executable
// that is packaged with the engine, and generate the .json and .png files, then
// reads the .json file into the [kaiju_font.FontData] structure. A
// [kaiju_font.KaijuFont] is returned with the font details and the MSDF version
// of a PNG file.
func ProcessTTF(ttfFile string, charsetFile string) (kaiju_font.KaijuFont, error) {
	var f *os.File
	var err error
	if f, err = os.CreateTemp("", "*.json"); err != nil {
		return kaiju_font.KaijuFont{}, err
	}
	jsonFile := f.Name()
	f.Close()
	if f, err = os.CreateTemp("", "*.png"); err != nil {
		return kaiju_font.KaijuFont{}, err
	}
	pngFile := f.Name()
	f.Close()
	defer os.Remove(jsonFile)
	defer os.Remove(pngFile)
	if build.Debug && runtime.GOOS == "linux" {
		klib.NotYetImplemented(325)
	}
	cmd := exec.Command(binDir+"msdf-atlas-gen.exe",
		"-font", ttfFile,
		"-pxrange", "4",
		"-size", "64",
		"-charset", charsetFile,
		"-fontname", ttfFile,
		"-type", "msdf",
		"-format", "png",
		"-pots",
		"-json", jsonFile,
		"-imageout", pngFile)
	cmd.Run()
	jsonBin, err := os.ReadFile(jsonFile)
	if err != nil {
		return kaiju_font.KaijuFont{}, err
	}
	kf := kaiju_font.KaijuFont{}
	if err = json.NewDecoder(strings.NewReader(string(jsonBin))).Decode(&kf.Details); err != nil {
		return kaiju_font.KaijuFont{}, nil
	}
	kf.PNG, err = os.ReadFile(pngFile)
	return kf, err
}
