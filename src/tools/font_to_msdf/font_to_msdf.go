/******************************************************************************/
/* font_to_msdf.go                                                            */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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

package font_to_msdf

import (
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/KaijuEngine/kaiju/build"
	"github.com/KaijuEngine/kaiju/klib"
	"github.com/KaijuEngine/kaiju/rendering/loaders/kaiju_font"
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
