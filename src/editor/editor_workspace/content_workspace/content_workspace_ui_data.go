package content_workspace

import (
	"kaiju/editor/project/project_database/content_database"
	"kaiju/klib"
)

type WorkspaceUIData struct {
	Filters []string
	Tags    []string
}

func (w *WorkspaceUIData) SetupUIData(cdb *content_database.Cache) []string {
	for _, cat := range content_database.ContentCategories {
		w.Filters = append(w.Filters, cat.TypeName())
	}
	list := cdb.List()
	ids := make([]string, 0, len(list))
	for i := range list {
		ids = append(ids, list[i].Id())
		for j := range list[i].Config.Tags {
			w.Tags = klib.AppendUnique(w.Tags, list[i].Config.Tags[j])
		}
	}
	return ids
}
