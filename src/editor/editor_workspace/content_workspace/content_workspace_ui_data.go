/******************************************************************************/
/* content_workspace_ui_data.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_workspace

import (
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/platform/profiler/tracing"
)

type WorkspaceUIData struct {
	Filters map[string]int
	Tags    map[string]int
}

func (w *WorkspaceUIData) SetupUIData(cdb *content_database.Cache) []string {
	defer tracing.NewRegion("ContentWorkspaceUIData.SetupUIData").End()
	w.Filters = make(map[string]int)
	w.Tags = make(map[string]int)

	for _, cat := range content_database.ContentCategories {
		w.Filters[cat.TypeName()]++
	}
	list := cdb.List()
	ids := make([]string, 0, len(list))
	for i := range list {
		ids = append(ids, list[i].Id())
		for tag := range list[i].Config.Tags {
			w.Tags[tag]++
		}
	}
	return ids
}
