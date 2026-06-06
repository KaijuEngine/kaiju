/******************************************************************************/
/* content_workspace_filters.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_workspace

import (
	"strings"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
)

func ShouldShowContent(query, id string, typeFilters, tagFilters map[string]struct{}, cdb *content_database.Cache) bool {
	defer tracing.NewRegion("content_workspace.ShouldShowContent").End()
	cc, err := cdb.Read(id)
	if err != nil {
		return false
	}
	show := len(typeFilters) == 0 && len(tagFilters) == 0
	if !show && len(typeFilters) > 0 {
		_, hasType := typeFilters[cc.Config.Type]
		show = hasType
	}
	if !show || len(tagFilters) > 0 {
		show = filterThroughTags(&cc, tagFilters)
	}
	if show && query != "" {
		show = runQueryOnContent(&cc, query, tagFilters)
	}
	return show
}

func ShouldHideContent(id string, typeFilters, tagFilters map[string]struct{}, cdb *content_database.Cache) bool {
	defer tracing.NewRegion("content_workspace.ShouldHideContent").End()
	cc, err := cdb.Read(id)
	if err != nil {
		return false
	}
	_, hasType := typeFilters[cc.Config.Type]
	hide := hasType ||
		filterThroughTags(&cc, tagFilters)
	return hide
}

func filterThroughTags(cc *content_database.CachedContent, tagFilters klib.Set[string]) bool {
	defer tracing.NewRegion("content_workspace.filterThroughTags").End()
	for i := range cc.Config.Tags {
		_, hasTag := tagFilters[i]
		if hasTag {
			return true
		}
	}
	return false
}

func runQueryOnContent(cc *content_database.CachedContent, query string, tagFilters klib.Set[string]) bool {
	defer tracing.NewRegion("content_workspace.runQueryOnContent").End()
	// TODO:  Use filters like tag:..., type:..., etc.
	if strings.Contains(cc.Config.NameLower(), query) {
		return true
	}
	for tag := range cc.Config.Tags {
		if _, ok := tagFilters[tag]; ok {
			return true
		}
	}
	return false
}
