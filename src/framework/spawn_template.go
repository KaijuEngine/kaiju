/******************************************************************************/
/* spawn_template.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package framework

import (
	"encoding/json"
	"log/slog"

	"kaijuengine.com/build"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
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
	stages.RegenerateEntityIdsAndRewriteReferences(&desc)
	return spawnTemplateEntities(host, parent, &desc)
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

func spawnTemplateEntities(host *engine.Host, parent *engine.Entity, desc *stages.EntityDescription) (*engine.Entity, error) {
	stage := stages.Stage{Entities: []stages.EntityDescription{*desc}}
	res := stage.Load(host)
	if len(res.Roots) == 0 {
		return nil, nil
	}
	root := res.Roots[0]
	if parent != nil {
		root.SetParent(parent)
	}
	return root, nil
}
