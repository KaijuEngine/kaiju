/******************************************************************************/
/* font_faces.go                                                              */
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

package rendering

import "strings"

type FontFace string

const (
	FontBold            = FontFace("OpenSans-Bold")
	FontBoldItalic      = FontFace("OpenSans-BoldItalic")
	FontExtraBold       = FontFace("OpenSans-ExtraBold")
	FontExtraBoldItalic = FontFace("OpenSans-ExtraBoldItalic")
	FontItalic          = FontFace("OpenSans-Italic")
	FontLight           = FontFace("OpenSans-Light")
	FontLightItalic     = FontFace("OpenSans-LightItalic")
	FontRegular         = FontFace("OpenSans-Regular")
	FontSemiBold        = FontFace("OpenSans-SemiBold")
	FontSemiBoldItalic  = FontFace("OpenSans-SemiBoldItalic")
)

func (f FontFace) IsBold() bool {
	return strings.Contains(string(f), "Bold")
}

func (f FontFace) IsExtraBold() bool {
	return strings.Contains(string(f), "ExtraBold")
}

func (f FontFace) IsItalic() bool {
	return strings.Contains(string(f), "Italic")
}

func (f FontFace) AsBold() FontFace {
	if f.IsItalic() {
		return FontFace(string(f.Base()) + "-BoldItalic")
	}
	return FontFace(string(f.Base()) + "-Bold")
}

func (f FontFace) AsExtraBold() FontFace {
	if f.IsItalic() {
		return FontFace(string(f.Base()) + "-ExtraBoldItalic")
	}
	return FontFace(string(f.Base()) + "-ExtraBold")
}

func (f FontFace) AsLight() FontFace {
	if f.IsItalic() {
		return FontFace(string(f.Base()) + "-LightItalic")
	}
	return FontFace(string(f.Base()) + "-Light")
}

func (f FontFace) AsMedium() FontFace {
	if f.IsItalic() {
		return FontFace(string(f.Base()) + "-MediumItalic")
	}
	return FontFace(string(f.Base()) + "-Medium")
}

func (f FontFace) AsSemiBold() FontFace {
	if f.IsItalic() {
		return FontFace(string(f.Base()) + "-SemiBoldItalic")
	}
	return FontFace(string(f.Base()) + "-SemiBold")
}

func (f FontFace) AsItalic() FontFace {
	if f.IsBold() {
		return FontFace(string(f.Base()) + "-BoldItalic")
	} else if f.IsExtraBold() {
		return FontFace(string(f.Base()) + "-ExtraBoldItalic")
	}
	return FontFace(string(f.Base()) + "-Italic")
}

func (f FontFace) RemoveBold() FontFace {
	if f.IsItalic() {
		return FontFace(string(f.Base()) + "-Italic")
	}
	return f.AsRegular()
}

func (f FontFace) RemoveItalic() FontFace {
	if f.IsBold() {
		return FontFace(string(f.Base()) + "-Bold")
	} else if f.IsExtraBold() {
		return FontFace(string(f.Base()) + "-ExtraBold")
	}
	return f.AsRegular()
}

func (f FontFace) AsRegular() FontFace {
	return FontFace(string(f.Base()) + "-Regular")
}

func (f FontFace) Base() FontFace { return FontFace(strings.Split(string(f), "-")[0]) }

func (f FontFace) string() string { return string(f) }
