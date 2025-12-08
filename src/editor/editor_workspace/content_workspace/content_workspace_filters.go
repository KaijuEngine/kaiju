/******************************************************************************/
/* content_workspace_filters.go                                               */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package content_workspace

import (
	"slices"
	"strings"

	"github.com/KaijuEngine/kaiju/editor/project/project_database/content_database"
	"github.com/KaijuEngine/kaiju/klib"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
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
