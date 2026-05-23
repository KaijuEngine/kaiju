//go:build linux || android

package localization

import "os"

func currentLocalization() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if loc := os.Getenv(key); loc != "" {
			return loc
		}
	}
	return ""
}
