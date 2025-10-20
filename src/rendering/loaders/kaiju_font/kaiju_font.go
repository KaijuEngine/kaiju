/******************************************************************************/
/* kaiju_font.go                                                              */
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

package kaiju_font

import (
	"bytes"
	"encoding/gob"
)

// KaijuFont is a base primitive representing a single font including each
// character's details along with the MSDF PNG image. Typically this structure
// is generated when importing a font into the engine using
// a function like [font_to_msdf.ProcessTTF]. From this point, it is typically
// serialized and stored into the content database. When reading a font from the
// content database, it will return a KaijuFont.
type KaijuFont struct {
	Details FontData
	PNG     []byte
}

type Rect struct {
	Left   float32 `json:"left"`
	Top    float32 `json:"top"`
	Right  float32 `json:"right"`
	Bottom float32 `json:"bottom"`
}

type Glyph struct {
	Unicode     int     `json:"unicode"`
	Advance     float32 `json:"advance"`
	PlaneBounds Rect    `json:"planeBounds"`
	AtlasBounds Rect    `json:"atlasBounds"`
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

// Serialize will convert a [KaijuFont] into a byte array for saving to the
// database or later use. This serialization uses the built-in [gob.Encoder]
func (f *KaijuFont) Serialize() ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(w)
	err := enc.Encode(f)
	return w.Bytes(), err
}

// Deserialize will construct a [KaijuFont] from the given array of bytes. This
// deserialization uses the built-in [gob.Decoder]
func Deserialize(data []byte) (KaijuFont, error) {
	r := bytes.NewReader(data)
	dec := gob.NewDecoder(r)
	var kf KaijuFont
	err := dec.Decode(&kf)
	return kf, err
}
