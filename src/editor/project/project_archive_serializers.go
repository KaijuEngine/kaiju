package project

import (
	"bytes"
	"encoding/json"
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/engine/runtime/encoding/gob"
	"kaiju/stages"
	"log/slog"
)

func (p *Project) stageArchiveSerializer(rawData []byte) ([]byte, error) {
	var ss stages.StageJson
	if err := json.Unmarshal(rawData, &ss); err != nil {
		return rawData, err
	}
	s := stages.Stage{}
	s.FromMinimized(ss)
	var removeUnpackedDataBindings func(desc *stages.EntityDescription)
	removeUnpackedDataBindings = func(desc *stages.EntityDescription) {
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
		desc.DataBinding = desc.DataBinding[0:]
		for i := range desc.Children {
			removeUnpackedDataBindings(&desc.Children[i])
		}
	}
	for i := range s.Entities {
		removeUnpackedDataBindings(&s.Entities[i])
	}
	stream := bytes.NewBuffer(rawData)
	stream.Reset()
	err := gob.NewEncoder(stream).Encode(s)
	return stream.Bytes(), err
}
