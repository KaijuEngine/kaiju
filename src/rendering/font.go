/******************************************************************************/
/* font.go                                                                    */
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

package rendering

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"kaiju/engine/assets"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

const (
	distanceFieldSize  = 64.0
	distanceFieldRange = 4.0
	invalidRuneProxy   = '_'
)

// We need overlapping letters to be drawn first when rendering blocks of text,
// otherwise, when they have a background, they will clip the text that they
// are overlapping
var overlappingLetters = []rune{'j', 'J'}

type FontJustify int

const (
	FontJustifyLeft = FontJustify(iota)
	FontJustifyCenter
	FontJustifyRight
)

type FontBaseline int

const (
	FontBaselineBottom = FontBaseline(iota)
	FontBaselineCenter
	FontBaselineTop
)

type FontFace string

func (f FontFace) IsBold() bool {
	return strings.Contains(string(f), "Bold")
}

func (f FontFace) IsExtraBold() bool {
	return strings.Contains(string(f), "ExtraBold")
}

func (f FontFace) IsItalic() bool {
	return strings.Contains(string(f), "Italic")
}

func (f FontFace) string() string { return string(f) }

const (
	FontCondensedBold                = FontFace("fonts/OpenSans_Condensed-Bold")
	FontCondensedBoldItalic          = FontFace("fonts/OpenSans_Condensed-BoldItalic")
	FontCondensedExtraBold           = FontFace("fonts/OpenSans_Condensed-ExtraBold")
	FontCondensedExtraBoldItalic     = FontFace("fonts/OpenSans_Condensed-ExtraBoldItalic")
	FontCondensedItalic              = FontFace("fonts/OpenSans_Condensed-Italic")
	FontCondensedLight               = FontFace("fonts/OpenSans_Condensed-Light")
	FontCondensedLightItalic         = FontFace("fonts/OpenSans_Condensed-LightItalic")
	FontCondensedMedium              = FontFace("fonts/OpenSans_Condensed-Medium")
	FontCondensedMediumItalic        = FontFace("fonts/OpenSans_Condensed-MediumItalic")
	FontCondensedRegular             = FontFace("fonts/OpenSans_Condensed-Regular")
	FontCondensedSemiBold            = FontFace("fonts/OpenSans_Condensed-SemiBold")
	FontCondensedSemiBoldItalic      = FontFace("fonts/OpenSans_Condensed-SemiBoldItalic")
	FontSemiCondensedBold            = FontFace("fonts/OpenSans_SemiCondensed-Bold")
	FontSemiCondensedBoldItalic      = FontFace("fonts/OpenSans_SemiCondensed-BoldItalic")
	FontSemiCondensedExtraBold       = FontFace("fonts/OpenSans_SemiCondensed-ExtraBold")
	FontSemiCondensedExtraBoldItalic = FontFace("fonts/OpenSans_SemiCondensed-ExtraBoldItalic")
	FontSemiCondensedItalic          = FontFace("fonts/OpenSans_SemiCondensed-Italic")
	FontSemiCondensedLight           = FontFace("fonts/OpenSans_SemiCondensed-Light")
	FontSemiCondensedLightItalic     = FontFace("fonts/OpenSans_SemiCondensed-LightItalic")
	FontSemiCondensedMedium          = FontFace("fonts/OpenSans_SemiCondensed-Medium")
	FontSemiCondensedMediumItalic    = FontFace("fonts/OpenSans_SemiCondensed-MediumItalic")
	FontSemiCondensedRegular         = FontFace("fonts/OpenSans_SemiCondensed-Regular")
	FontSemiCondensedSemiBold        = FontFace("fonts/OpenSans_SemiCondensed-SemiBold")
	FontSemiCondensedSemiBoldItalic  = FontFace("fonts/OpenSans_SemiCondensed-SemiBoldItalic")
	FontBold                         = FontFace("fonts/OpenSans-Bold")
	FontBoldItalic                   = FontFace("fonts/OpenSans-BoldItalic")
	FontExtraBold                    = FontFace("fonts/OpenSans-ExtraBold")
	FontExtraBoldItalic              = FontFace("fonts/OpenSans-ExtraBoldItalic")
	FontItalic                       = FontFace("fonts/OpenSans-Italic")
	FontLight                        = FontFace("fonts/OpenSans-Light")
	FontLightItalic                  = FontFace("fonts/OpenSans-LightItalic")
	FontMedium                       = FontFace("fonts/OpenSans-Medium")
	FontMediumItalic                 = FontFace("fonts/OpenSans-MediumItalic")
	FontRegular                      = FontFace("fonts/OpenSans-Regular")
	FontSemiBold                     = FontFace("fonts/OpenSans-SemiBold")
	FontSemiBoldItalic               = FontFace("fonts/OpenSans-SemiBoldItalic")

	DefaultFontEMSize = 14.0
)

type fontBinMetrics struct {
	EMSize, LineHeight, Ascender, Descender, UnderlineY, UnderlineThickness float32
}

type fontBinChar struct {
	letter                   rune
	advance                  float32
	planeBounds, atlasBounds [4]float32
}

type fontBin struct {
	texture                           *Texture
	width, height                     int32
	metrics                           fontBinMetrics
	letters                           map[rune]fontBinChar
	cachedLetters, cachedOrthoLetters map[rune]*cachedLetterMesh
}

type cachedLetterMesh struct {
	mesh           *Mesh
	pxRange        matrix.Vec2
	uvs            matrix.Vec4
	material       *Material
	texture        *Texture
	transformation matrix.Mat4
}

type FontCache struct {
	textMaterial                 *Material
	textOrthoMaterial            *Material
	textMaterialTransparent      *Material
	textOrthoMaterialTransparent *Material
	renderer                     Renderer
	renderCaches                 RenderCaches
	assetDb                      assets.Database
	fontFaces                    map[string]fontBin
	instanceKey                  int64
	FaceMutex                    sync.RWMutex
}

type TextShaderData struct {
	ShaderDataBase
	UVs     matrix.Vec4
	FgColor matrix.Color
	BgColor matrix.Color
	Scissor matrix.Vec4
	PxRange matrix.Vec2
}

func (s TextShaderData) Size() int {
	const size = int(unsafe.Sizeof(TextShaderData{}) - ShaderBaseDataStart)
	return size
}

func (cache *FontCache) TransparentMaterial(target *Material) *Material {
	defer tracing.NewRegion("FontCache.TransparentMaterial").End()
	if target.SelectRoot() == cache.textMaterial.SelectRoot() {
		return cache.textMaterialTransparent
	} else if target.SelectRoot() == cache.textOrthoMaterial.SelectRoot() {
		return cache.textOrthoMaterialTransparent
	} else if target.SelectRoot() == cache.textMaterialTransparent.SelectRoot() ||
		target == cache.textOrthoMaterialTransparent.SelectRoot() {
		return target
	}
	slog.Error("invalid material used for getting transparent text material", "material", target.Name)
	return nil
}

func (cache *FontCache) nextInstanceKey(key rune) string {
	cache.instanceKey++
	return fmt.Sprintf("font_%c_%d", key, cache.instanceKey)
}

func (cache *FontCache) requireFace(face FontFace) {
	defer tracing.NewRegion("FontCache.requireFace").End()
	cache.FaceMutex.RLock()
	if _, ok := cache.fontFaces[face.string()]; !ok {
		cache.FaceMutex.RUnlock()
		cache.FaceMutex.Lock()
		defer cache.FaceMutex.Unlock()
		cache.initFont(face, cache.renderer, cache.assetDb)
	} else {
		cache.FaceMutex.RUnlock()
	}
}

func (cache *FontCache) PreloadFace(face FontFace) {
	defer tracing.NewRegion("FontCache.PreloadFace").End()
	cache.requireFace(face)
}

func (cache *FontCache) EMSize(face FontFace) float32 {
	cache.requireFace(face)
	return cache.fontFaces[face.string()].metrics.EMSize * DefaultFontEMSize
}

func NewFontCache(renderer Renderer, assetDb assets.Database) FontCache {
	defer tracing.NewRegion("rendering.NewFontCache").End()
	return FontCache{
		renderer:  renderer,
		assetDb:   assetDb,
		fontFaces: make(map[string]fontBin),
	}
	// TODO:  Deal with the freeing of mesh/shaders/textures
}

func (c fontBinChar) Width() float32 {
	return c.planeBounds[2] - c.planeBounds[0]
}

func (c fontBinChar) Height() float32 {
	return c.planeBounds[3] - c.planeBounds[1]
}

func (c fontBinChar) AtlasWidth() float32 {
	return c.atlasBounds[2] - c.atlasBounds[0]
}

func (c fontBinChar) AtlasHeight() float32 {
	return c.atlasBounds[3] - c.atlasBounds[1]
}

func findBinChar(font fontBin, letter rune) fontBinChar {
	cached, ok := font.letters[letter]
	if !ok {
		cached = font.letters[invalidRuneProxy]
	}
	return cached
}

func (cache *FontCache) charCountInWidth(font fontBin, runes []rune, maxWidth, scale float32) int {
	defer tracing.NewRegion("FontCache.charCountInWidth").End()
	wrap := false
	spaceIndex := 0
	wx := float32(0.0)
	textLen := len(runes)
	for i, r := range runes {
		if r == '\n' {
			spaceIndex = i
			wrap = true
			break
		} else if unicode.IsSpace(r) {
			spaceIndex = i
		}
		// TODO:  Optimize this to use a linear array
		ch := findBinChar(font, r)
		wx += ch.advance * scale
		if wx >= maxWidth && spaceIndex != 0 {
			wrap = true
			break
		}
	}
	if !wrap {
		return textLen
	} else {
		return spaceIndex + 1
	}
}

func (cache *FontCache) cachedMeshLetter(font fontBin, letter rune, isOrtho bool) *cachedLetterMesh {
	defer tracing.NewRegion("FontCache.cachedMeshLetter").End()
	var outLetter *cachedLetterMesh
	var ok bool
	if isOrtho {
		outLetter, ok = font.cachedOrthoLetters[letter]
	} else {
		outLetter, ok = font.cachedLetters[letter]
	}
	if !ok {
		if isOrtho {
			outLetter = font.cachedOrthoLetters[invalidRuneProxy]
		} else {
			outLetter = font.cachedLetters[invalidRuneProxy]
		}
	}
	return outLetter
}

func (cache *FontCache) createLetterMesh(font fontBin, key rune, c fontBinChar, renderer Renderer, meshCache *MeshCache) {
	defer tracing.NewRegion("FontCache.createLetterMesh").End()
	mat := cache.textMaterial
	oMat := cache.textOrthoMaterial

	w := c.Width()
	h := -c.Height()

	mesh := NewMeshScreenQuad(meshCache)
	transformation := matrix.Mat4Identity()
	transformation.Scale(matrix.Vec3{w, h, 1})

	var clm cachedLetterMesh
	clm.mesh = mesh
	clm.material = mat
	clm.texture = font.texture
	clm.transformation = transformation
	uvx := c.atlasBounds[0]
	uvy := c.atlasBounds[3]
	uvw := c.atlasBounds[2] - c.atlasBounds[0]
	uvh := c.atlasBounds[1] - c.atlasBounds[3]
	clm.uvs = matrix.Vec4{
		uvx / float32(font.width), uvy / float32(font.height),
		uvw / float32(font.width), uvh / float32(font.height)}
	// TODO:  Figure out the distance field size
	clm.pxRange = matrix.Vec2{
		c.Width() / distanceFieldSize * distanceFieldRange,
		c.Height() / distanceFieldSize * distanceFieldRange}
	//clm.pxRange = matrix.Vec2{
	//	c.Width() / c.AtlasWidth() * 2.0,
	//	c.Height() / c.AtlasHeight() * 2.0}
	font.cachedLetters[key] = &clm

	clmCpy := clm
	clmCpy.material = oMat
	clmCpy.texture = font.texture
	// TODO:  [PORT] Do we need to clone the mesh anymore?
	//clmCpy.mesh = mesh.Clone()
	clmCpy.mesh = mesh
	font.cachedOrthoLetters[key] = &clmCpy
}

func (cache *FontCache) initFont(face FontFace, renderer Renderer, adb assets.Database) bool {
	defer tracing.NewRegion("FontCache.initFont").End()
	bin := fontBin{}
	bin.texture, _ = cache.renderCaches.TextureCache().Texture(face.string()+".png", TextureFilterLinear)
	bin.texture.MipLevels = 1
	bin.cachedLetters = make(map[rune]*cachedLetterMesh)
	bin.cachedOrthoLetters = make(map[rune]*cachedLetterMesh)
	out, _ := adb.Read(face.string() + ".bin")
	if bin.texture == nil || out == nil || len(out) == 0 {
		return false
	}
	read := bytes.NewReader(out)
	// Create an int32 variable named count that is read from read
	var count int32
	binary.Read(read, binary.LittleEndian, &count)
	binary.Read(read, binary.LittleEndian, &bin.width)
	binary.Read(read, binary.LittleEndian, &bin.height)
	// TODO:  Read the metrics into cache.[font]
	binary.Read(read, binary.LittleEndian, &bin.metrics)
	bin.letters = make(map[rune]fontBinChar, count)
	for i := int32(0); i < count; i++ {
		var fbc fontBinChar
		var letter uint32
		binary.Read(read, binary.LittleEndian, &letter)
		fbc.letter = rune(letter)
		//utf8_from_unicode(letter, &fbc.letter)
		binary.Read(read, binary.LittleEndian, &fbc.advance)
		binary.Read(read, binary.LittleEndian, &fbc.planeBounds)
		binary.Read(read, binary.LittleEndian, &fbc.atlasBounds)
		bin.letters[fbc.letter] = fbc
	}
	sample := findBinChar(bin, '-')
	cSpace := fontBinChar{
		letter:      ' ',
		advance:     sample.advance,
		planeBounds: sample.planeBounds,
		atlasBounds: [4]float32{0.999, 0.001, 1.0, 0.0},
	}
	const tabRune = '\t'
	const tabSize = 4
	cTab := fontBinChar{
		letter:  tabRune,
		advance: cSpace.advance * 4,
		planeBounds: [4]float32{
			cSpace.planeBounds[0] * tabSize / 2,
			cSpace.planeBounds[1],
			cSpace.planeBounds[2] * tabSize / 2,
			cSpace.planeBounds[3],
		},
		atlasBounds: cSpace.atlasBounds,
	}
	bin.letters[' '] = cSpace
	bin.letters[tabRune] = cTab
	cReturn := fontBinChar{letter: '\r', advance: 0.0,
		planeBounds: [4]float32{0, 0, 0, 0}, atlasBounds: [4]float32{0.999, 0.001, 1.0, 0.0}}
	bin.letters['\r'] = cReturn
	for i := int32(0); i < count; i++ {
		cache.createLetterMesh(bin, bin.letters[i].letter, bin.letters[i], renderer, cache.renderCaches.MeshCache())
	}
	cache.fontFaces[face.string()] = bin
	return true
}

func (cache *FontCache) Init(renderer Renderer, adb assets.Database, caches RenderCaches) error {
	defer tracing.NewRegion("FontCache.Init").End()
	var err error
	mc := caches.MaterialCache()
	if cache.textMaterial, err = mc.Material("text3d"); err != nil {
		slog.Error("failed to load the text3d material", "error", err)
		return err
	}
	if cache.textOrthoMaterial, err = mc.Material("text"); err != nil {
		slog.Error("failed to load the text material", "error", err)
		return err
	}
	if cache.textMaterialTransparent, err = mc.Material("text3d_transparent"); err != nil {
		slog.Error("failed to load the transparent text3d material", "error", err)
		return err
	}
	if cache.textOrthoMaterialTransparent, err = mc.Material("text_transparent"); err != nil {
		slog.Error("failed to load the transparent text material", "error", err)
		return err
	}
	cache.renderCaches = caches
	return nil
}

func (cache *FontCache) RenderMeshes(caches RenderCaches,
	text string, x, y, z, scale, maxWidth float32, fgColor, bgColor matrix.Color,
	justify FontJustify, baseline FontBaseline, rootScale matrix.Vec3, instanced,
	is3D bool, face FontFace, lineHeight float32) []Drawing {
	defer tracing.NewRegion("FontCache.RenderMeshes").End()
	cache.requireFace(face)
	cx := x
	cy := y

	es := rootScale
	if lineHeight != 0 {
		baseline = FontBaselineCenter
	}

	left := -es.X() * 0.5
	inverseWidth := 1.0 / es.X()
	inverseHeight := 1.0 / es.Y()

	fontFace := cache.fontFaces[face.string()]
	var material *Material
	if is3D {
		material = cache.textMaterial
	} else {
		material = cache.textOrthoMaterial
	}

	// Iterate through all characters
	runes := []rune(text)
	textLen := len(runes)
	charLen := textLen
	//size_t lenLeft = textLen;
	current := 0
	height := float32(0.0)

	fontMeshes := make([]Drawing, 0)
	maxHeight := fontFace.metrics.LineHeight * -scale
	if lineHeight != 0 {
		maxHeight = -lineHeight / fontFace.metrics.LineHeight
	}
	for current < textLen {
		if maxWidth > 0 {
			charLen = cache.charCountInWidth(fontFace, runes[current:], maxWidth, scale)
		}
		lineWidth := float32(0.0)
		if charLen > 0 || unicode.IsSpace(runes[current]) {
			for _, c := range runes[current : current+charLen] {
				if c != '\n' {
					ch := findBinChar(fontFace, c)
					lineWidth += ch.advance * scale
				}
			}
		}
		var xOffset, yOffset float32
		switch justify {
		case FontJustifyRight:
			xOffset = left + (maxWidth - lineWidth)
		case FontJustifyCenter:
			xOffset = (maxWidth * 0.5) - (lineWidth * 0.5)
		case FontJustifyLeft:
			xOffset = left
		default:
			xOffset = left
		}
		switch baseline {
		case FontBaselineTop:
			yOffset = (es.Y() * 0.5) + maxHeight
		case FontBaselineCenter:
			yOffset = maxHeight * 0.5
		case FontBaselineBottom:
		default:
			yOffset = es.Y() * -0.5
		}
		xOffset *= inverseWidth
		yOffset -= fontFace.metrics.Descender * scale
		yOffset *= inverseHeight

		if charLen > 0 || (unicode.IsSpace(runes[current]) && runes[current] != '\n') {
			for i := current; i < current+charLen; i++ {
				c := runes[i]
				if c == '\n' {
					continue
				}
				ch := findBinChar(fontFace, c)
				xPos := cx + (ch.planeBounds[0] * scale * inverseWidth)
				yPos := cy + (ch.planeBounds[1] * scale * inverseHeight)
				xPos += xOffset
				yPos += yOffset
				w := ch.Width() * scale * inverseWidth
				h := ch.Height() * scale * inverseHeight
				// TODO:  Figure out the distance field size
				pxRange := matrix.Vec2{
					(ch.Width() * scale) / distanceFieldSize * distanceFieldRange,
					(-ch.Height() * scale) / distanceFieldSize * distanceFieldRange}
				//pxRange := matrix.Vec2{
				//	(ch.Width() * scale) / ch.AtlasWidth() * distanceFieldRange,
				//	(ch.Height() * scale) / ch.AtlasHeight() * distanceFieldRange}
				var uvs matrix.Vec4
				var clm *cachedLetterMesh = nil
				if instanced {
					clm = cache.cachedMeshLetter(fontFace, c, !is3D)
				}
				var m *Mesh
				model := matrix.Mat4Identity()
				if clm == nil {
					zPos := z
					if slices.Contains(overlappingLetters, c) {
						zPos += 0.0001
					}
					var verts [4]Vertex
					verts[0].Position = matrix.Vec3{xPos, yPos, zPos}
					verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
					verts[0].UV0 = matrix.Vec2{0.0, 1.0}
					verts[0].Color = matrix.ColorWhite()
					verts[1].Position = matrix.Vec3{xPos, yPos + h, zPos}
					verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
					verts[1].UV0 = matrix.Vec2{0.0, 0.0}
					verts[1].Color = matrix.ColorWhite()
					verts[2].Position = matrix.Vec3{xPos + w, yPos + h, zPos}
					verts[2].Normal = matrix.Vec3{0.0, 0.0, 1.0}
					verts[2].UV0 = matrix.Vec2{1.0, 0.0}
					verts[2].Color = matrix.ColorWhite()
					verts[3].Position = matrix.Vec3{xPos + w, yPos, zPos}
					verts[3].Normal = matrix.Vec3{0.0, 0.0, 1.0}
					verts[3].UV0 = matrix.Vec2{1.0, 1.0}
					verts[3].Color = matrix.ColorWhite()
					indexes := [6]uint32{0, 1, 2, 0, 2, 3}
					caches.MeshCache().Mesh(cache.nextInstanceKey(c), verts[:], indexes[:])
					uvx := ch.atlasBounds[0]
					uvy := ch.atlasBounds[1]
					uvw := ch.atlasBounds[2] - ch.atlasBounds[0]
					uvh := ch.atlasBounds[3] - ch.atlasBounds[1]
					uvs = matrix.Vec4{
						uvx / float32(fontFace.width), uvy / float32(fontFace.height),
						uvw / float32(fontFace.width), uvh / float32(fontFace.height)}
				} else {
					// TODO:  Scale and place the mesh based on justify, baseline, etc.
					model.MultiplyAssign(clm.transformation)
					model.Scale(matrix.Vec3{scale * inverseWidth, scale * inverseHeight, 1.0})
					model.Translate(matrix.Vec3{xPos, (yPos + h), z})
					uvs = clm.uvs
					m = clm.mesh
				}
				shaderData := &TextShaderData{
					ShaderDataBase: NewShaderDataBase(),
					FgColor:        fgColor,
					BgColor:        bgColor,
					PxRange:        pxRange,
					UVs:            uvs,
					Scissor:        matrix.Vec4{-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax},
				}
				shaderData.SetModel(model)
				drawing := Drawing{
					Renderer:   cache.renderer,
					Material:   material.CreateInstance([]*Texture{fontFace.texture}),
					Mesh:       m,
					ShaderData: shaderData,
					Transform:  nil,
				}
				fontMeshes = append(fontMeshes, drawing)
				cx += ch.advance * scale * inverseWidth
				ay := fontFace.metrics.LineHeight * scale * inverseHeight
				height = matrix.Max(height, ay)
			}
		}
		cx = x
		cy -= height
		//lenLeft -= charLen;
		current += charLen
	}
	return fontMeshes
}

func (cache *FontCache) MeasureString(face FontFace, text string, scale float32) float32 {
	defer tracing.NewRegion("FontCache.MeasureString").End()
	cache.requireFace(face)
	x, maxX := float32(0.0), float32(0.0)
	for _, r := range text {
		if r == '\n' {
			x = 0.0
		} else {
			ch := findBinChar(cache.fontFaces[face.string()], r)
			x += ch.advance * scale
			maxX = matrix.Max(maxX, x)
		}
	}
	return maxX
}

func (cache *FontCache) MeasureStringWithin(face FontFace, text string, scale, maxWidth float32, lineHeight float32) matrix.Vec2 {
	defer tracing.NewRegion("FontCache.MeasureStringWithin").End()
	cache.requireFace(face)
	fontFace := cache.fontFaces[face.string()]
	maxHeight := fontFace.metrics.LineHeight * scale
	if lineHeight != 0 {
		maxHeight = max(maxHeight, lineHeight)
	}
	var x, y float32 = 0.0, 0.0
	clip := []rune(text)
	for len(clip) > 0 {
		count := klib.Clamp(cache.charCountInWidth(fontFace, clip, maxWidth, scale), 0, len(clip))
		x = max(x, cache.MeasureString(face, string(clip[:count]), scale))
		y += maxHeight
		clip = []rune(clip)[count:]
	}
	return matrix.Vec2{x, y}
}

func (cache *FontCache) StringRectsWithinNew(face FontFace, text string, scale, maxWidth float32) []matrix.Vec4 {
	defer tracing.NewRegion("FontCache.StringRectsWithinNew").End()
	cache.requireFace(face)
	fontFace := cache.fontFaces[face.string()]
	rects := make([]matrix.Vec4, 0)
	current := 0
	var x, y, height float32 = 0.0, 0.0, 0.0
	runes := []rune(text)
	for current < len(runes) {
		offset := current
		count := cache.charCountInWidth(fontFace, runes[current:], maxWidth, scale)
		x = 0.0
		for _, r := range runes[offset : offset+count] {
			ch := findBinChar(fontFace, r)
			w := ch.advance * scale
			h := fontFace.metrics.LineHeight * scale
			rects = append(rects, matrix.Vec4{x, y, w, h})
			current++
			x += w
			height = matrix.Max(height, h)
		}
		y += height
	}
	return rects
}

func (cache *FontCache) LineCountWithin(face FontFace, text string, scale, maxWidth float32) int {
	defer tracing.NewRegion("FontCache.LineCountWithin").End()
	cache.requireFace(face)
	lines := 0
	runes := []rune(text)
	textLen := len(runes)
	current := 0
	fontFace := cache.fontFaces[face.string()]
	for current < textLen {
		current += cache.charCountInWidth(fontFace, runes[current:], maxWidth, scale)
		lines++
	}
	return max(1, lines)
}

func (cache *FontCache) MeasureCharacter(face string, r rune, pixelSize float32) matrix.Vec2 {
	ch := findBinChar(cache.fontFaces[face], r)
	return matrix.Vec2{ch.Width() * pixelSize, ch.Height() * pixelSize}
}

func (cache *FontCache) PointOffsetWithin(face FontFace, text string, point matrix.Vec2, scale, maxWidth float32) int {
	defer tracing.NewRegion("FontCache.PointOffsetWithin").End()
	cache.requireFace(face)
	textLen := utf8.RuneCountInString(text)
	idx := textLen
	rects := cache.StringRectsWithinNew(face, text, scale, maxWidth)
	for i := 0; i < textLen; i++ {
		width := rects[i].Z()
		height := rects[i].W()
		if (rects[i].X()+width*0.5) > point.X() && ((rects[i].Y()+height) > point.Y() || point.Y() > rects[i].Y()) {
			idx = i
			break
		}
	}
	return idx
}

func (cache *FontCache) Destroy() {
	cache.textMaterial = nil
	cache.textOrthoMaterial = nil
	cache.textMaterialTransparent = nil
	cache.textOrthoMaterialTransparent = nil
	cache.fontFaces = make(map[string]fontBin)
}
