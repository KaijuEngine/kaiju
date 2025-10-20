/******************************************************************************/
/* content_database_texture.go                                                */
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

package content_database

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"os"
	"path/filepath"

	"golang.org/x/image/bmp"
)

func init() { addCategory(Texture{}) }

// Texture is a [ContentCategory] represented by a file with a ".png", ".jpg",
// or ".jpeg" extension. Textures are as they seem.
type Texture struct{}
type TextureConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Texture) Path() string       { return project_file_system.ContentTextureFolder }
func (Texture) TypeName() string   { return "texture" }
func (Texture) ExtNames() []string { return []string{".png", ".jpg", ".jpeg", ".bmp"} }

func (Texture) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Texture.Import").End()
	var decoder func(r io.Reader) (image.Image, error) = nil
	switch filepath.Ext(src) {
	case ".png":
		decoder = png.Decode
	case ".jpg":
		fallthrough
	case ".jpeg":
		decoder = jpeg.Decode
	case ".bmp":
		decoder = bmp.Decode
	}
	if decoder != nil {
		imgData, err := os.Open(src)
		if err != nil {
			return ProcessedImport{}, ImageImportError{err, "open"}
		}
		defer imgData.Close()
		img, err := decoder(imgData)
		if err != nil {
			return ProcessedImport{}, ImageImportError{err, "decode"}
		}
		buff := bytes.NewBuffer([]byte{})
		if err := png.Encode(buff, img); err != nil {
			return ProcessedImport{}, ImageImportError{err, "encode"}
		}
		return ProcessedImport{Variants: []ImportVariant{
			{Name: fileNameNoExt(src), Data: buff.Bytes()},
		}}, nil
	}
	return ProcessedImport{}, ImageImportError{
		errors.New("this error shouldn't happen, if it does, an image format is missing for import"),
		"formatting",
	}
}
