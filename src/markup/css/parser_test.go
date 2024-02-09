package css

import (
	"kaiju/markup/css/rules"
	"testing"
)

func TestParser(t *testing.T) {
	s := rules.NewStyleSheet()
	s.Parse(DefaultCSS)
	if len(s.Groups) == 0 {
		t.Error("No groups found")
	}
}
