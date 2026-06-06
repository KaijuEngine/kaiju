/******************************************************************************/
/* project_custom_content_serializers.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"reflect"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/assets/content_archive"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

func (p *Project) initializeCustomSerializers() {
	p.contentSerializers = make(map[string]func(content_archive.FileReader, []byte) ([]byte, error))
	p.contentSerializers[(content_database.Stage{}).TypeName()] = p.stageArchiveSerializer
	toc := content_database.TableOfContents{}
	p.contentSerializers[toc.TypeName()] = toc.ArchiveSerializer
}

func (p *Project) stageArchiveSerializer(reader content_archive.FileReader, rawData []byte) ([]byte, error) {
	var ss stages.StageJson
	if err := json.Unmarshal(rawData, &ss); err != nil {
		return rawData, err
	}
	s := stages.Stage{}
	s.FromMinimized(ss)
	var removeUnpackedDataBindings func(desc *stages.EntityDescription) error
	removeUnpackedDataBindings = func(desc *stages.EntityDescription) error {
		for i := range desc.DataBinding {
			g, ok := p.EntityDataBinding(desc.DataBinding[i].RegistraionKey)
			if ok {
				de := entity_data_binding.EntityDataEntry{}
				de.ReadEntityDataBindingType(g)
				for k, v := range desc.DataBinding[i].Fields {
					de.SetFieldByName(k, v)
				}
				desc.RawDataBinding = append(desc.RawDataBinding, de.BoundData)
			} else {
				slog.Warn("failed to locate the data binding for registration key",
					"key", desc.DataBinding[i].RegistraionKey)
			}
		}
		desc.DataBinding = make([]stages.EntityDataBinding, 0)
		// Simpler than most ideas I had, essentially pull the shader data
		// the same way you would in a running game. Then cast all of the JSON
		// fields to the instance through the entity_data_binding.ToDataBinding
		// helpers. Then pull the actual value out for serialization.
		//
		// This is needed because the JSON serialization doesn't use the correct
		// types internally, int would be int64 and float32 would be float64. So
		// this will basically fix the values before serializing with POD.
		extractShaderData := func() {
			if desc.Material == "" {
				return
			}
			m, err := reader.Read(desc.Material)
			if err != nil {
				return
			}
			var mat rendering.MaterialData
			err = json.Unmarshal(m, &mat)
			if err != nil {
				return
			}
			s, err := reader.Read(mat.Shader)
			if err != nil {
				return
			}
			var sh rendering.ShaderData
			err = json.Unmarshal(s, &sh)
			if err != nil {
				return
			}
			sd := shader_data_registry.Create(sh.DrawInstanceDataName())
			v := reflect.ValueOf(sd)
			for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
				v = v.Elem()
			}
			db := entity_data_binding.ToDataBinding("", sd)
			for i := range desc.ShaderData {
				db.SetFieldByName(desc.ShaderData[i].Name, desc.ShaderData[i].Value)
				desc.ShaderData[i].Value = db.FieldValueByName(desc.ShaderData[i].Name)
			}
		}
		extractShaderData()
		for i := range desc.Children {
			if err := removeUnpackedDataBindings(&desc.Children[i]); err != nil {
				return err
			}
		}
		return nil
	}
	for i := range s.Entities {
		if err := removeUnpackedDataBindings(&s.Entities[i]); err != nil {
			return rawData, err
		}
	}
	stream := bytes.NewBuffer(rawData)
	stream.Reset()
	err := pod.NewEncoder(stream).Encode(s)
	return stream.Bytes(), err
}
