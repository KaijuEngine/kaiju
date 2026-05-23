package localization

import (
	"strings"

	"golang.org/x/text/language"
)

const defaultLocalization = "en-US"

// Current returns the user's current localization as a canonical language tag.
func Current() string {
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
