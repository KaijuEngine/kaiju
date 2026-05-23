package localization

import "testing"

func TestNormalizeLocalization(t *testing.T) {
	tests := map[string]string{
		"":                     defaultLocalization,
		"C":                    defaultLocalization,
		"POSIX":                defaultLocalization,
		"en_US.UTF-8":          "en-US",
		"de_DE.UTF-8@euro":     "de-DE",
		"zh_Hant_TW.UTF-8":     "zh-Hant-TW",
		"fr-CA":                "fr-CA",
		"invalid localization": defaultLocalization,
	}
	for input, expected := range tests {
		if actual := normalizeLocalization(input); actual != expected {
			t.Fatalf("normalizeLocalization(%q) = %q, expected %q", input, actual, expected)
		}
	}
}
