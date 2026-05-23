package localization

import (
	"strings"

	"golang.org/x/text/language"
	"kaijuengine.com/platform/hid"
)

const defaultLocalization = "en-US"

type Localization interface {
	KeyToRune(keyboard *hid.Keyboard, key hid.KeyboardKey) rune
}

func Select() Localization {
	switch String() {
	case "en-US":
		return AmericanEnglish{}
	default:
		return AmericanEnglish{}
	}
}

// String returns the user's current localization as a canonical language tag.
func String() string {
	return normalizeLocalization(currentLocalization())
}

func normalizeLocalization(loc string) string {
	loc = strings.TrimSpace(loc)
	if loc == "" || loc == "C" || loc == "POSIX" {
		return defaultLocalization
	}
	if i := strings.Index(loc, "."); i >= 0 {
		loc = loc[:i]
	}
	if i := strings.Index(loc, "@"); i >= 0 {
		loc = loc[:i]
	}
	loc = strings.ReplaceAll(loc, "_", "-")
	tag, err := language.Parse(loc)
	if err != nil {
		return defaultLocalization
	}
	return tag.String()
}
