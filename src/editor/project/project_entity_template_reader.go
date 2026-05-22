/******************************************************************************/
/* project_entity_template_reader.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import (
	"encoding/json"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/platform/profiler/tracing"
)

func (p *Project) ReadEntityTemplate(id string) (stages.EntityDescription, error) {
	defer tracing.NewRegion("StageManager.readTemplate").End()
	cc, err := p.cacheDatabase.Read(id)
	if err != nil {
		return stages.EntityDescription{}, err
	}
	f, err := p.fileSystem.Open(content_database.ToContentPath(cc.Path))
	if err != nil {
		return stages.EntityDescription{}, err
	}
	defer f.Close()
	var desc stages.EntityDescription
	if err = json.NewDecoder(f).Decode(&desc); err != nil {
		return stages.EntityDescription{}, err
	}
	desc.TemplateId = id
	return desc, nil
}
