/******************************************************************************/
/* font_faces.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
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
