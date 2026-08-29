package svg

import (
	"os"
	"path/filepath"
	"testing"
)

const testSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="40" height="30">
  <rect width="40" height="30" fill="red"/>
</svg>`

func TestRenderString(t *testing.T) {
	img, err := RenderString(testSVG, 40, 30)
	if err != nil {
		t.Fatalf("RenderString error: %v", err)
	}
	if img.Width != 40 || img.Height != 30 {
		t.Fatalf("unexpected size: %dx%d", img.Width, img.Height)
	}
	if len(img.RGBA) != img.Width*img.Height*4 {
		t.Fatalf("unexpected buffer length %d", len(img.RGBA))
	}
	for i := 3; i < len(img.RGBA); i += 4 {
		if img.RGBA[i] == 0 {
			t.Fatalf("non-opaque pixel at %d", i)
		}
	}
}

func TestRenderFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.svg")
	if err := os.WriteFile(path, []byte(testSVG), 0o644); err != nil {
		t.Fatal(err)
	}
	img, err := RenderFile(path, 40, 30)
	if err != nil {
		t.Fatalf("RenderFile error: %v", err)
	}
	if img.Width != 40 || img.Height != 30 {
		t.Fatalf("unexpected size: %dx%d", img.Width, img.Height)
	}
	if len(img.RGBA) != img.Width*img.Height*4 {
		t.Fatalf("unexpected buffer length %d", len(img.RGBA))
	}
}

func TestRenderMissingFile(t *testing.T) {
	if _, err := RenderFile(filepath.Join(t.TempDir(), "missing.svg"), 40, 30); err == nil {
		t.Fatal("expected error for missing file")
	}
}
