//go:build windows && filedialog

package filesystem

import (
	"errors"
	"reflect"
	"testing"
)

func TestNormalizeDialogPattern(t *testing.T) {
	tests := []struct {
		name      string
		extension string
		want      string
	}{
		{name: "Empty", extension: "", want: "*.*"},
		{name: "Whitespace", extension: "   ", want: "*.*"},
		{name: "WildcardDot", extension: ".*", want: "*.*"},
		{name: "WildcardAll", extension: "*", want: "*.*"},
		{name: "WildcardAllDot", extension: "*.*", want: "*.*"},
		{name: "DotExt", extension: ".go", want: "*.go"},
		{name: "PlainExt", extension: "go", want: "*.go"},
		{name: "TrimmedDotExt", extension: " .json ", want: "*.json"},
		{name: "AlreadyPattern", extension: "*.png", want: "*.png"},
		{name: "WildcardWithPrefix", extension: "file.*", want: "file.*"},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeDialogPattern(tc.extension)
			if got != tc.want {
				t.Fatalf("normalizeDialogPattern(%q) = %q, want %q", tc.extension, got, tc.want)
			}
		})
	}
}

func TestExtensionsToDialogFilters(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		got := extensionsToDialogFilters(nil)
		want := []DialogFilter{
			{Name: "All Files (*.*)", Patterns: []string{"*.*"}},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("extensionsToDialogFilters(nil) = %#v, want %#v", got, want)
		}
	})

	t.Run("MapsExtensions", func(t *testing.T) {
		got := extensionsToDialogFilters([]DialogExtension{
			{Name: "Go Files", Extension: ".go"},
			{Name: "", Extension: "txt"},
		})
		want := []DialogFilter{
			{Name: "Go Files", Patterns: []string{"*.go"}},
			{Name: "Files", Patterns: []string{"*.txt"}},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("extensionsToDialogFilters(...) = %#v, want %#v", got, want)
		}
	})
}

func TestEncodeDialogFilters(t *testing.T) {
	t.Run("EmptyFilters", func(t *testing.T) {
		got := encodeDialogFilters(nil)
		if got != "" {
			t.Fatalf("encodeDialogFilters(nil) = %q, want empty string", got)
		}
	})

	got := encodeDialogFilters([]DialogFilter{
		{Name: "Go and Text", Patterns: []string{"*.go", "*.txt"}},
		{Name: "", Patterns: []string{"  ", "*.*"}},
		{Name: "Whitespace Only", Patterns: []string{" ", "\t"}},
	})
	want := "Go and Text|*.go;*.txt\nFiles|*.*\nWhitespace Only|*.*"
	if got != want {
		t.Fatalf("encodeDialogFilters(...) = %q, want %q", got, want)
	}
}

func TestEncodeAndParseDialogOptions(t *testing.T) {
	encoded := encodeDialogOptions([]DialogCustomOption{
		{Name: "Recursive import", Default: 1},
		{Name: "Import mode", Values: []string{"Copy", "Reference", "Link"}, Default: 2},
		{Name: " ", Values: []string{"ignored"}, Default: 0},
	})

	wantEncoded := "Recursive import|1|\nImport mode|2|Copy;Reference;Link"
	if encoded != wantEncoded {
		t.Fatalf("encodeDialogOptions(...) = %q, want %q", encoded, wantEncoded)
	}

	parsed := parseDialogSelectedOptions("Recursive import|b|1\nImport mode|i|2\nInvalid\nBad|x|10")
	wantParsed := map[string]any{
		"Recursive import": true,
		"Import mode":      2,
	}
	if !reflect.DeepEqual(parsed, wantParsed) {
		t.Fatalf("parseDialogSelectedOptions(...) = %#v, want %#v", parsed, wantParsed)
	}
}

func TestParseDialogSelectedOptionsInvalidData(t *testing.T) {
	got := parseDialogSelectedOptions("Flag|b|0\nMode|i|abc\nUnknown|x|1\n|b|1")
	want := map[string]any{
		"Flag": false,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseDialogSelectedOptions(...) = %#v, want %#v", got, want)
	}
}

func TestDefaultSelectedOptions(t *testing.T) {
	got := defaultSelectedOptions([]DialogCustomOption{
		{Name: "Recursive import", Default: 1},
		{Name: "Import mode", Values: []string{"Copy", "Reference", "Link"}, Default: -2},
		{Name: "Compression", Values: []string{"Off", "On"}, Default: 10},
		{Name: "  Trim Me  ", Values: []string{"A", "B"}, Default: 1},
		{Name: "   ", Default: 1},
	})

	want := map[string]any{
		"Recursive import": true,
		"Import mode":      0,
		"Compression":      1,
		"Trim Me":          1,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("defaultSelectedOptions(...) = %#v, want %#v", got, want)
	}
}

func TestOpenNativeDialogWindowValidation(t *testing.T) {
	err := openNativeDialogWindow(NativeDialogRequest{}, nil)
	if err == nil {
		t.Fatal("openNativeDialogWindow with nil callback should fail")
	}

	err = openNativeDialogWindow(NativeDialogRequest{
		Mode: NativeDialogModeOpenFolder + 1,
	}, func(NativeDialogResult) {})
	if err == nil {
		t.Fatal("openNativeDialogWindow with invalid mode should fail")
	}
}

func TestValidateNativeDialogRequestRejectsDelimiters(t *testing.T) {
	// These cases protect the raw delimiter wire format consumed by directory.win32.h.
	tests := []struct {
		name    string
		request NativeDialogRequest
	}{
		{
			name: "FilterNamePipe",
			request: NativeDialogRequest{
				Filters: []DialogFilter{{Name: "Go|Files", Patterns: []string{"*.go"}}},
			},
		},
		{
			name: "FilterPatternSemicolon",
			request: NativeDialogRequest{
				Filters: []DialogFilter{{Name: "Files", Patterns: []string{"*.go;*.txt"}}},
			},
		},
		{
			name: "OptionNameNewline",
			request: NativeDialogRequest{
				Options: []DialogCustomOption{{Name: "Import\nMode"}},
			},
		},
		{
			name: "OptionValueSemicolon",
			request: NativeDialogRequest{
				Options: []DialogCustomOption{{Name: "Mode", Values: []string{"Copy;Link"}}},
			},
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			if err := validateNativeDialogRequest(tc.request); err == nil {
				t.Fatal("validateNativeDialogRequest should reject unsupported delimiters")
			}
		})
	}
}

func TestOpenNativeDialogWindowWhenRuntimeClosing(t *testing.T) {
	windowsDialogRuntime.mu.Lock()
	prevClosing := windowsDialogRuntime.closing
	windowsDialogRuntime.closing = true
	windowsDialogRuntime.mu.Unlock()
	t.Cleanup(func() {
		windowsDialogRuntime.mu.Lock()
		windowsDialogRuntime.closing = prevClosing
		windowsDialogRuntime.mu.Unlock()
	})

	err := openNativeDialogWindow(NativeDialogRequest{
		Mode: NativeDialogModeOpenFile,
	}, func(NativeDialogResult) {})
	if err == nil {
		t.Fatal("openNativeDialogWindow should fail when runtime is shutting down")
	}
}

func TestMakeSimpleDialogCallbackRouting(t *testing.T) {
	var okPath string
	cancelCalled := 0
	callback := makeSimpleDialogCallback(func(path string) {
		okPath = path
	}, func() {
		cancelCalled++
	})

	callback(NativeDialogResult{
		Status: NativeDialogStatusAccepted,
		Paths:  []string{"C:/tmp/test.txt"},
	})
	if okPath != "C:/tmp/test.txt" {
		t.Fatalf("ok callback path = %q, want %q", okPath, "C:/tmp/test.txt")
	}
	if cancelCalled != 0 {
		t.Fatalf("cancel callback count = %d, want 0", cancelCalled)
	}

	callback(NativeDialogResult{
		Status: NativeDialogStatusAccepted,
		Paths:  nil,
	})
	if cancelCalled != 1 {
		t.Fatalf("cancel callback count after empty accepted result = %d, want 1", cancelCalled)
	}
}

func TestMakeSimpleDialogCallbackAllowsNilCancel(t *testing.T) {
	callback := makeSimpleDialogCallback(func(path string) {}, nil)
	callback(NativeDialogResult{
		Status: NativeDialogStatusFailed,
		Err:    errors.New("test failure"),
	})
}

func TestMakeSimpleDialogCallbackAllowsNilOk(t *testing.T) {
	cancelCalled := 0
	callback := makeSimpleDialogCallback(nil, func() {
		cancelCalled++
	})
	callback(NativeDialogResult{
		Status: NativeDialogStatusAccepted,
		Paths:  []string{"C:/tmp/test.txt"},
	})
	if cancelCalled != 1 {
		t.Fatalf("cancel callback count with nil ok callback = %d, want 1", cancelCalled)
	}
}

func TestOpenDialogWindowRejectsNilOkCallbacks(t *testing.T) {
	if err := openFileDialogWindow("", nil, nil, nil, nil); err == nil {
		t.Fatal("openFileDialogWindow should fail with nil ok callback")
	}
	if err := openSaveFileDialogWindow("", "file.txt", nil, nil, nil, nil); err == nil {
		t.Fatal("openSaveFileDialogWindow should fail with nil ok callback")
	}
	if err := openFolderDialogWindow("", nil, nil, nil); err == nil {
		t.Fatal("openFolderDialogWindow should fail with nil ok callback")
	}
}

func TestIsPathWithinDialogRoot(t *testing.T) {
	tests := []struct {
		name string
		root string
		path string
		want bool
	}{
		{
			name: "DriveRootContainsChild",
			root: `C:\`,
			path: `C:\a.txt`,
			want: true,
		},
		{
			name: "DriveRootRejectsOtherDrive",
			root: `C:\`,
			path: `D:\a.txt`,
			want: false,
		},
		{
			name: "BoundaryRejectsPrefixCollision",
			root: `C:\foo`,
			path: `C:\foobar\a.txt`,
			want: false,
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			got := isPathWithinDialogRoot(tc.root, tc.path)
			if got != tc.want {
				t.Fatalf("isPathWithinDialogRoot(%q, %q) = %v, want %v", tc.root, tc.path, got, tc.want)
			}
		})
	}
}

func TestApplyNativeDialogResultGuardsClearsPathsOnFailure(t *testing.T) {
	res := NativeDialogResult{
		Status: NativeDialogStatusFailed,
		Paths:  []string{`C:\tmp\a.txt`},
	}

	applyNativeDialogResultGuards(&res, "")

	if len(res.Paths) != 0 {
		t.Fatalf("expected failed dialog result to clear paths, got %#v", res.Paths)
	}
	if res.Err == nil {
		t.Fatal("expected failed dialog result without explicit error to get default error")
	}
}

func TestApplyNativeDialogResultGuardsRootViolation(t *testing.T) {
	res := NativeDialogResult{
		Status: NativeDialogStatusAccepted,
		Paths:  []string{`D:\picked.txt`},
	}

	applyNativeDialogResultGuards(&res, `C:\`)

	if res.Status != NativeDialogStatusFailed {
		t.Fatalf("expected status failed after root violation, got %v", res.Status)
	}
	if len(res.Paths) != 0 {
		t.Fatalf("expected paths to be cleared after root violation, got %#v", res.Paths)
	}
	if res.Err == nil {
		t.Fatal("expected error after root violation")
	}
}
