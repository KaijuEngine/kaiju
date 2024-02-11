/*****************************************************************************/
/* color.go                                                                  */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md)    */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining     */
/* a copy of this software and associated documentation files (the           */
/* "Software"), to deal in the Software without restriction, including       */
/* without limitation the rights to use, copy, modify, merge, publish,       */
/* distribute, sublicense, and/or sell copies of the Software, and to        */
/* permit persons to whom the Software is furnished to do so, subject to     */
/* the following conditions:                                                 */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,           */
/* EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF        */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY      */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,      */
/* TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE         */
/* SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                    */
/*****************************************************************************/

package matrix

import (
	"fmt"
	"strings"
)

type Color Vec4
type Color8 struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func (c Color) R() Float                           { return c[R] }
func (c Color) G() Float                           { return c[G] }
func (c Color) B() Float                           { return c[B] }
func (c Color) A() Float                           { return c[A] }
func (c *Color) PR() *Float                        { return &c[R] }
func (c *Color) PG() *Float                        { return &c[G] }
func (c *Color) PB() *Float                        { return &c[B] }
func (c *Color) PA() *Float                        { return &c[A] }
func (c Color) RGBA() (Float, Float, Float, Float) { return c[R], c[G], c[B], c[A] }
func (c *Color) SetR(r Float)                      { c[R] = r }
func (c *Color) SetG(g Float)                      { c[G] = g }
func (c *Color) SetB(b Float)                      { c[B] = b }
func (c *Color) SetA(a Float)                      { c[A] = a }

func NewColor(r, g, b, a Float) Color {
	return Color{r, g, b, a}
}

func NewColor8(r, g, b, a uint8) Color8 {
	return Color8{r, g, b, a}
}

func ColorFromColor8(c Color8) Color {
	return Color{
		Float(c.R) / 255.0,
		Float(c.G) / 255.0,
		Float(c.B) / 255.0,
		Float(c.A) / 255.0,
	}
}

func Color8FromColor(c Color) Color8 {
	return Color8{
		uint8(c.R() * 255),
		uint8(c.G() * 255),
		uint8(c.B() * 255),
		uint8(c.A() * 255),
	}
}

func ColorFromVec3(v Vec3) Color {
	return Color{v.X(), v.Y(), v.Z(), 1}
}

func ColorFromVec4(v Vec4) Color {
	return Color{v.X(), v.Y(), v.Z(), v.W()}
}

func ColorRGBAInt(r, g, b, a int) Color {
	return Color{Float(r) / 255.0, Float(g) / 255.0, Float(b) / 255.0, Float(a) / 255.0}
}

func ColorRGBInt(r, g, b int) Color {
	return Color{Float(r) / 255.0, Float(g) / 255.0, Float(b) / 255.0, 1.0}
}

func (c Color) AsColor8() Color8 { return Color8FromColor(c) }
func (c Color8) AsColor() Color  { return ColorFromColor8(c) }

func (c Color8) Equal(rhs Color8) bool {
	return c.R == rhs.R && c.G == rhs.G && c.B == rhs.B && c.A == rhs.A
}

func ColorMix(lhs, rhs Color, amount Float) Color {
	return Color{
		lhs.R() + (rhs.R()-lhs.R())*amount,
		lhs.G() + (rhs.G()-lhs.G())*amount,
		lhs.B() + (rhs.B()-lhs.B())*amount,
		lhs.A() + (rhs.A()-lhs.A())*amount,
	}
}

func ColorFromHexString(str string) (Color, error) {
	c8, err := Color8FromHexString(str)
	return ColorFromColor8(c8), err
}

func (c Color) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x%02x", uint8(c.R()*255), uint8(c.G()*255), uint8(c.B()*255), uint8(c.A()*255))
}

func Color8FromHexString(str string) (Color8, error) {
	rgba := Color8{255, 255, 255, 255}
	var err error
	str = strings.TrimPrefix(str, "#")
	if len(str) == 8 {
		fmt.Sscanf(str, "%02x%02x%02x%02x", &rgba.R, &rgba.G, &rgba.B, &rgba.A)
	} else if len(str) == 6 {
		fmt.Sscanf(str, "%02x%02x%02x", &rgba.R, &rgba.G, &rgba.B)
	} else if len(str) == 3 {
		str = fmt.Sprintf("%c%c%c%c%c%c", str[0], str[0], str[1], str[1], str[2], str[2])
		fmt.Sscanf(str, "%02x%02x%02x", &rgba.R, &rgba.G, &rgba.B)
	} else {
		rgba = Color8FromColor(ColorMagenta())
		err = fmt.Errorf("invalid hex color string: %s", str)
	}
	return rgba, err
}

func (c Color8) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}

func ColorRed() Color                  { return Color{1, 0, 0, 1} }
func ColorWhite() Color                { return Color{1, 1, 1, 1} }
func ColorBlue() Color                 { return Color{0, 0, 1, 1} }
func ColorBlack() Color                { return Color{0, 0, 0, 1} }
func ColorGreen() Color                { return Color{0, 1, 0, 1} }
func ColorYellow() Color               { return Color{1, 1, 0, 1} }
func ColorOrange() Color               { return Color{1, 0.647, 0, 1} }
func ColorClear() Color                { return Color{0, 0, 0, 0} }
func ColorGray() Color                 { return Color{0.5, 0.5, 0.5, 1} }
func ColorPurple() Color               { return Color{0.5, 0, 0.5, 1} }
func ColorBrown() Color                { return Color{0.647, 0.165, 0.165, 1} }
func ColorPink() Color                 { return Color{1, 0.753, 0.796, 1} }
func ColorCyan() Color                 { return Color{0, 1, 1, 1} }
func ColorMagenta() Color              { return Color{1, 0, 1, 1} }
func ColorTeal() Color                 { return Color{0, 0.5, 0.5, 1} }
func ColorLime() Color                 { return Color{0, 1, 0, 1} }
func ColorMaroon() Color               { return Color{0.5, 0, 0, 1} }
func ColorOlive() Color                { return Color{0.5, 0.5, 0, 1} }
func ColorNavy() Color                 { return Color{0, 0, 0.5, 1} }
func ColorSilver() Color               { return Color{0.753, 0.753, 0.753, 1} }
func ColorGold() Color                 { return Color{1, 0.843, 0, 1} }
func ColorSky() Color                  { return Color{0.529, 0.808, 0.922, 1} }
func ColorViolet() Color               { return Color{0.933, 0.51, 0.933, 1} }
func ColorIndigo() Color               { return Color{0.294, 0, 0.51, 1} }
func ColorTurquoise() Color            { return Color{0.251, 0.878, 0.816, 1} }
func ColorAzure() Color                { return Color{0.941, 1, 1, 1} }
func ColorChartreuse() Color           { return Color{0.498, 1, 0, 1} }
func ColorCoral() Color                { return Color{1, 0.498, 0.314, 1} }
func ColorCrimson() Color              { return Color{0.863, 0.078, 0.235, 1} }
func ColorFuchsia() Color              { return Color{1, 0, 1, 1} }
func ColorKhaki() Color                { return Color{0.941, 0.902, 0.549, 1} }
func ColorLavender() Color             { return Color{0.902, 0.902, 0.98, 1} }
func ColorMoccasin() Color             { return Color{1, 0.894, 0.71, 1} }
func ColorSalmon() Color               { return Color{0.98, 0.502, 0.447, 1} }
func ColorSienna() Color               { return Color{0.627, 0.322, 0.176, 1} }
func ColorTan() Color                  { return Color{0.824, 0.706, 0.549, 1} }
func ColorTomato() Color               { return Color{1, 0.388, 0.278, 1} }
func ColorWheat() Color                { return Color{0.961, 0.871, 0.702, 1} }
func ColorAqua() Color                 { return Color{0, 1, 1, 1} }
func ColorAquamarine() Color           { return Color{0.498, 1, 0.831, 1} }
func ColorBeige() Color                { return Color{0.961, 0.961, 0.863, 1} }
func ColorBisque() Color               { return Color{1, 0.894, 0.769, 1} }
func ColorBlanchedAlmond() Color       { return Color{1, 0.922, 0.804, 1} }
func ColorBlueViolet() Color           { return Color{0.541, 0.169, 0.886, 1} }
func ColorBurlyWood() Color            { return Color{0.871, 0.722, 0.529, 1} }
func ColorCadetBlue() Color            { return Color{0.373, 0.62, 0.627, 1} }
func ColorChocolate() Color            { return Color{0.824, 0.412, 0.118, 1} }
func ColorCornflowerBlue() Color       { return Color{0.392, 0.584, 0.929, 1} }
func ColorCornSilk() Color             { return Color{1, 0.973, 0.863, 1} }
func ColorDarkBlue() Color             { return Color{0, 0, 0.545, 1} }
func ColorDarkCyan() Color             { return Color{0, 0.545, 0.545, 1} }
func ColorDarkGoldenrod() Color        { return Color{0.722, 0.525, 0.043, 1} }
func ColorDarkGray() Color             { return Color{0.663, 0.663, 0.663, 1} }
func ColorDarkModeGrayBG() Color       { return Color{0.2, 0.2, 0.2, 1} }
func ColorDarkModeGrayFG() Color       { return Color{0.3, 0.3, 0.3, 1} }
func ColorDarkGreen() Color            { return Color{0, 0.392, 0, 1} }
func ColorDarkKhaki() Color            { return Color{0.741, 0.718, 0.42, 1} }
func ColorDarkMagenta() Color          { return Color{0.545, 0, 0.545, 1} }
func ColorDarkOliveGreen() Color       { return Color{0.333, 0.42, 0.184, 1} }
func ColorDarkOrange() Color           { return Color{1, 0.549, 0, 1} }
func ColorDarkOrchid() Color           { return Color{0.6, 0.196, 0.8, 1} }
func ColorDarkRed() Color              { return Color{0.545, 0, 0, 1} }
func ColorDarkSalmon() Color           { return Color{0.914, 0.588, 0.478, 1} }
func ColorDarkSeaGreen() Color         { return Color{0.561, 0.737, 0.561, 1} }
func ColorDarkSlateBlue() Color        { return Color{0.282, 0.239, 0.545, 1} }
func ColorDarkSlateGray() Color        { return Color{0.184, 0.31, 0.31, 1} }
func ColorDarkTurquoise() Color        { return Color{0, 0.808, 0.82, 1} }
func ColorDarkViolet() Color           { return Color{0.58, 0, 0.827, 1} }
func ColorDeepPink() Color             { return Color{1, 0.078, 0.576, 1} }
func ColorDeepSkyBlue() Color          { return Color{0, 0.749, 1, 1} }
func ColorDimGray() Color              { return Color{0.412, 0.412, 0.412, 1} }
func ColorDodgerBlue() Color           { return Color{0.118, 0.565, 1, 1} }
func ColorFirebrick() Color            { return Color{0.698, 0.133, 0.133, 1} }
func ColorFloralWhite() Color          { return Color{1, 0.98, 0.941, 1} }
func ColorForestGreen() Color          { return Color{0.133, 0.545, 0.133, 1} }
func ColorGainsboro() Color            { return Color{0.863, 0.863, 0.863, 1} }
func ColorGhostWhite() Color           { return Color{0.973, 0.973, 1, 1} }
func ColorGoldenrod() Color            { return Color{0.855, 0.647, 0.125, 1} }
func ColorGreenYellow() Color          { return Color{0.678, 1, 0.184, 1} }
func ColorHoneydew() Color             { return Color{0.941, 1, 0.941, 1} }
func ColorHotPink() Color              { return Color{1, 0.412, 0.706, 1} }
func ColorIndianRed() Color            { return Color{0.804, 0.361, 0.361, 1} }
func ColorIvory() Color                { return Color{1, 1, 0.941, 1} }
func ColorLavenderBlush() Color        { return Color{1, 0.941, 0.961, 1} }
func ColorLawnGreen() Color            { return Color{0.486, 0.988, 0, 1} }
func ColorLemonChiffon() Color         { return Color{1, 0.98, 0.804, 1} }
func ColorLightBlue() Color            { return Color{0.678, 0.847, 0.902, 1} }
func ColorLightCoral() Color           { return Color{0.941, 0.502, 0.502, 1} }
func ColorLightCyan() Color            { return Color{0.878, 1, 1, 1} }
func ColorLightGoldenrodYellow() Color { return Color{0.98, 0.98, 0.824, 1} }
func ColorLightGreen() Color           { return Color{0.565, 0.933, 0.565, 1} }
func ColorLightGrey() Color            { return Color{0.827, 0.827, 0.827, 1} }
func ColorLightPink() Color            { return Color{1, 0.714, 0.757, 1} }
func ColorLightSalmon() Color          { return Color{1, 0.627, 0.478, 1} }
func ColorLightSeaGreen() Color        { return Color{0.125, 0.698, 0.667, 1} }
func ColorLightSkyBlue() Color         { return Color{0.529, 0.808, 0.98, 1} }
func ColorLightSlateGray() Color       { return Color{0.467, 0.533, 0.6, 1} }
func ColorLightSteelBlue() Color       { return Color{0.69, 0.769, 0.871, 1} }
func ColorLightYellow() Color          { return Color{1, 1, 0.878, 1} }
func ColorLimeGreen() Color            { return Color{0.196, 0.804, 0.196, 1} }
func ColorLinen() Color                { return Color{0.98, 0.941, 0.902, 1} }
func ColorMediumAquamarine() Color     { return Color{0.4, 0.804, 0.667, 1} }
func ColorMediumBlue() Color           { return Color{0, 0, 0.804, 1} }
func ColorMediumOrchid() Color         { return Color{0.729, 0.333, 0.827, 1} }
func ColorMediumPurple() Color         { return Color{0.576, 0.439, 0.859, 1} }
func ColorMediumSeaGreen() Color       { return Color{0.235, 0.702, 0.443, 1} }
func ColorMediumSlateBlue() Color      { return Color{0.482, 0.408, 0.933, 1} }
func ColorMediumSpringGreen() Color    { return Color{0, 0.98, 0.604, 1} }
func ColorMediumTurquoise() Color      { return Color{0.282, 0.82, 0.8, 1} }
func ColorMediumVioletRed() Color      { return Color{0.78, 0.082, 0.522, 1} }
func ColorMidnightBlue() Color         { return Color{0.098, 0.098, 0.439, 1} }
func ColorMintCream() Color            { return Color{0.961, 1, 0.98, 1} }
func ColorMistyRose() Color            { return Color{1, 0.894, 0.882, 1} }
func ColorNavajoWhite() Color          { return Color{1, 0.871, 0.678, 1} }
func ColorOldLace() Color              { return Color{0.992, 0.961, 0.902, 1} }
func ColorOliveDrab() Color            { return Color{0.42, 0.557, 0.137, 1} }
func ColorOrangeRed() Color            { return Color{1, 0.271, 0, 1} }
func ColorOrchid() Color               { return Color{0.855, 0.439, 0.839, 1} }
func ColorPaleGoldenrod() Color        { return Color{0.933, 0.91, 0.667, 1} }
func ColorPaleGreen() Color            { return Color{0.596, 0.984, 0.596, 1} }
func ColorPaleTurquoise() Color        { return Color{0.686, 0.933, 0.933, 1} }
func ColorPaleVioletred() Color        { return Color{0.859, 0.439, 0.576, 1} }
func ColorPapayaWhip() Color           { return Color{1, 0.937, 0.835, 1} }
func ColorPeachPuff() Color            { return Color{1, 0.855, 0.725, 1} }
func ColorPeru() Color                 { return Color{0.804, 0.522, 0.247, 1} }
func ColorPlum() Color                 { return Color{0.867, 0.627, 0.867, 1} }
func ColorPowderBlue() Color           { return Color{0.69, 0.878, 0.902, 1} }
func ColorRosyBrown() Color            { return Color{0.737, 0.561, 0.561, 1} }
func ColorRoyalBlue() Color            { return Color{0.255, 0.412, 0.882, 1} }
func ColorSaddleBrown() Color          { return Color{0.545, 0.271, 0.075, 1} }
func ColorSandyBrown() Color           { return Color{0.957, 0.643, 0.376, 1} }
func ColorSeaGreen() Color             { return Color{0.18, 0.545, 0.341, 1} }
func ColorSeashell() Color             { return Color{1, 0.961, 0.933, 1} }
func ColorSkyBlue() Color              { return Color{0.529, 0.808, 0.922, 1} }
func ColorSlateBlue() Color            { return Color{0.416, 0.353, 0.804, 1} }
func ColorSlateGray() Color            { return Color{0.439, 0.502, 0.565, 1} }
func ColorSlateGrey() Color            { return Color{0.439, 0.502, 0.565, 1} }
func ColorSnow() Color                 { return Color{1, 0.98, 0.98, 1} }
func ColorSpringGreen() Color          { return Color{0, 1, 0.498, 1} }
func ColorSteelBlue() Color            { return Color{0.275, 0.51, 0.706, 1} }
func ColorThistle() Color              { return Color{0.847, 0.749, 0.847, 1} }
func ColorWhiteSmoke() Color           { return Color{0.961, 0.961, 0.961, 1} }
func ColorYellowGreen() Color          { return Color{0.604, 0.804, 0.196, 1} }
func ColorTransparent() Color          { return Color{1, 1, 1, 0} }
func ColorZero() Color                 { return Color{0, 0, 0, 0} }

func (lhs Color8) Similar(rhs Color8, tolerance uint8) bool {
	return uint8(AbsInt(int(lhs.R)-int(rhs.R))) <= tolerance &&
		uint8(AbsInt(int(lhs.G)-int(rhs.G))) <= tolerance &&
		uint8(AbsInt(int(lhs.B)-int(rhs.B))) <= tolerance &&
		uint8(AbsInt(int(lhs.A)-int(rhs.A))) <= tolerance
}

func (lhs Color) Equals(rhs Color) bool { return Vec4(lhs).Equals(Vec4(rhs)) }
func (c Color) IsZero() bool            { return c.Equals(ColorZero()) }

func (c Color) ScaleWithoutAlpha(scale Float) Color {
	return Color{c.R() * scale, c.G() * scale, c.B() * scale, c.A()}
}

func (c *Color) MultiplyAssign(other Color) {
	c[R] *= other[R]
	c[G] *= other[G]
	c[B] *= other[B]
	c[A] *= other[A]
}
