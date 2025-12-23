/******************************************************************************/
/* spawn_template.go                                                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package framework

import (
	"encoding/json"
	"kaiju/build"
	"kaiju/engine"
	"kaiju/engine/stages"
	"kaiju/klib"
	"kaiju/matrix"
	"log/slog"
)

// SpawnTemplate loads an entity template asset identified by `id` from the
// host's asset database, deserializes it into a `stages.EntityDescription`
// (using JSON on desktop debug builds or the archive deserializer in other
// builds), and creates a new entity configured from that description. If
// `parent` is non-nil the new entity will be parented to it. The created
// entity is initialized via `stages.SetupEntityFromDescription` and returned,
// or an error is returned if the asset cannot be read or deserialized.
func SpawnTemplate(id string, host *engine.Host, parent *engine.Entity) (*engine.Entity, error) {
	data, err := host.AssetDatabase().Read(id)
	if err != nil {
		return nil, err
	}
	var desc stages.EntityDescription
	if build.Debug && !klib.IsMobile() {
		err = json.Unmarshal(data, &desc)
	} else {
		desc, err = stages.EntityDescriptionArchiveDeserializer(data)
	}
	if err != nil {
		slog.Error("failed to deserialize the template data", "template", id, "error", err)
		return nil, err
	}
	e := host.NewEntity(host.WorkGroup())
	if parent != nil {
		e.SetParent(parent)
	}
	return stages.SetupEntityFromDescription(e, host, &desc)
}

// SpawnTemplateWithTransform loads an entity template asset identified by
// `id`, creates a new entity from that template (optionally parented to
// `parent`), then sets the entity's transform to the provided position,
// rotation and scale. It is a thin wrapper around `SpawnTemplate` that
// applies the transform before returning the created entity or any
// encountered error.
//
// Parameters:
//   - id: template asset id to load from the host's asset database
//   - host: the engine host used to create the entity and read assets
//   - parent: optional parent entity to attach the new entity to
//   - pos, rot, scale: transform components applied to the new entity
//
// Returns the newly created entity or a non-nil error if the template
// could not be read or deserialized.
func SpawnTemplateWithTransform(id string, host *engine.Host, parent *engine.Entity, pos, rot, scale matrix.Vec3) (*engine.Entity, error) {
	e, err := SpawnTemplate(id, host, parent)
	if err != nil {
		return e, err
	}
	e.Transform.SetPosition(pos)
	e.Transform.SetRotation(rot)
	e.Transform.SetScale(scale)
	return e, nil
}
