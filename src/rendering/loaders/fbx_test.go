/******************************************************************************/
/* fbx_test.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package loaders

import (
	"testing"

	"kaijuengine.com/engine/assets"
	fbxdoc "kaijuengine.com/rendering/loaders/fbx"
)

func TestFBXReturnsExpectedScopeErrors(t *testing.T) {
	cases := []struct {
		name  string
		path  string
		files map[string][]byte
		want  string
	}{
		{
			name:  "file not found",
			path:  "missing.fbx",
			files: map[string][]byte{},
			want:  "file does not exist",
		},
		{
			name:  "invalid extension",
			path:  "model.obj",
			files: map[string][]byte{"model.obj": []byte("anything")},
			want:  "invalid file extension",
		},
		{
			name:  "invalid header",
			path:  "model.fbx",
			files: map[string][]byte{"model.fbx": []byte("not an fbx file")},
			want:  "invalid FBX file",
		},
		{
			name: "ascii not supported",
			path: "model.fbx",
			files: map[string][]byte{
				"model.fbx": []byte("; FBX 7.4.0 project file\n"),
			},
			want: "ASCII FBX is not supported yet",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := FBX(c.path, assets.NewMockDB(c.files))
			if err == nil {
				t.Fatalf("FBX returned nil error, want %q", c.want)
			}
			if err.Error() != c.want {
				t.Fatalf("FBX error = %q, want %q", err.Error(), c.want)
			}
		})
	}
}

func TestFBXAcceptsBinaryHeader(t *testing.T) {
	data := append([]byte(fbxdoc.BinaryHeader), 0xE8, 0x1C, 0x00, 0x00)
	_, err := FBX("model.fbx", assets.NewMockDB(map[string][]byte{"model.fbx": data}))
	if err != nil {
		t.Fatalf("FBX returned error for binary header: %v", err)
	}
}
