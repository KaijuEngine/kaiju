/******************************************************************************/
/* editor_embedded_content_test.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_embedded_content

import (
	"os"
	"testing"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
)

func TestEditorContentReadsProjectContentByIndexedId(t *testing.T) {
	pfs, err := project_file_system.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer pfs.Close()
	const id = "texture-id"
	const cfgPath = "database/config/texture/" + id + ".json"
	const contentPath = "database/content/texture/" + id
	if err := pfs.MkdirAll("database/content/texture", os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := pfs.WriteFile(contentPath, []byte("indexed texture"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	db := &EditorContent{Pfs: &pfs}
	db.SetProjectContentIndex([]content_database.CachedContent{{
		Path: cfgPath,
		Config: content_database.ContentConfig{
			Type: content_database.Texture{}.TypeName(),
		},
	}})

	data, err := db.Read(id)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "indexed texture" {
		t.Fatalf("Read(%q) = %q, want indexed texture", id, data)
	}
	if !db.Exists(id) {
		t.Fatalf("Exists(%q) = false, want true", id)
	}
}
