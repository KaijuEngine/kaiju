package content_workspace

import (
	"kaiju/editor/project/project_database/content_database"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"slices"
	"strings"
)

func ShouldShowContent(query, id string, typeFilters, tagFilters []string, cdb *content_database.Cache) bool {
	defer tracing.NewRegion("content_workspace.ShouldShowContent").End()
	cc, err := cdb.Read(id)
	if err != nil {
		return false
	}
	show := len(typeFilters) == 0 && len(tagFilters) == 0
	if !show && len(typeFilters) > 0 {
		show = slices.Contains(typeFilters, cc.Config.Type)
	}
	if !show || len(tagFilters) > 0 {
		show = filterThroughTags(&cc, tagFilters)
	}
	if show && query != "" {
		show = runQueryOnContent(&cc, query, tagFilters)
	}
	return show
}

func filterThroughTags(cc *content_database.CachedContent, tagFilters []string) bool {
	defer tracing.NewRegion("content_workspace.filterThroughTags").End()
	for i := range cc.Config.Tags {
		if klib.StringsContainsCaseInsensitive(tagFilters, cc.Config.Tags[i]) {
			return true
		}
	}
	return false
}

func runQueryOnContent(cc *content_database.CachedContent, query string, tagFilters []string) bool {
	defer tracing.NewRegion("content_workspace.runQueryOnContent").End()
	// TODO:  Use filters like tag:..., type:..., etc.
	if strings.Contains(cc.Config.NameLower(), query) {
		return true
	}
	for i := range cc.Config.Tags {
		if slices.Contains(tagFilters, cc.Config.Tags[i]) {
			return true
		}
	}
	return false
}
