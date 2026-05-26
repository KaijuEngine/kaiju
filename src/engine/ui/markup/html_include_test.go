/******************************************************************************/
/* html_include_test.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package markup

import (
	"strings"
	"testing"

	"kaijuengine.com/engine/assets"
)

func TestExpandHTMLIncludes(t *testing.T) {
	db := assets.NewMockDB(map[string][]byte{
		"editor/ui/workspace/root.go.html":   []byte(`<html><body><kaiju-include src="part.go.html"></kaiju-include></body></html>`),
		"editor/ui/workspace/part.go.html":   []byte(`<div>{{ .Name }}</div><kaiju-include src="nested.go.html"></kaiju-include>`),
		"editor/ui/workspace/nested.go.html": []byte(`<span>Nested</span>`),
	})
	root, err := db.ReadText("editor/ui/workspace/root.go.html")
	if err != nil {
		t.Fatal(err)
	}
	got, err := expandHTMLIncludes(db, "editor/ui/workspace/root.go.html", root, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`<div>{{ .Name }}</div>`, `<span>Nested</span>`} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected composed html to contain %q, got %q", want, got)
		}
	}
}

func TestExpandHTMLIncludesMissingSource(t *testing.T) {
	db := assets.NewMockDB(map[string][]byte{})
	_, err := expandHTMLIncludes(db, "root.go.html", `<kaiju-include></kaiju-include>`, map[string]bool{})
	if err == nil {
		t.Fatal("expected missing src to return an error")
	}
}

func TestExpandHTMLIncludesDetectsCycle(t *testing.T) {
	db := assets.NewMockDB(map[string][]byte{
		"a.go.html": []byte(`<kaiju-include src="b.go.html"></kaiju-include>`),
		"b.go.html": []byte(`<kaiju-include src="a.go.html"></kaiju-include>`),
	})
	root, err := db.ReadText("a.go.html")
	if err != nil {
		t.Fatal(err)
	}
	_, err = expandHTMLIncludes(db, "a.go.html", root, map[string]bool{})
	if err == nil {
		t.Fatal("expected include cycle to return an error")
	}
}
