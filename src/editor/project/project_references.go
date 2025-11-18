package project

import (
	"encoding/json"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"kaiju/stages"
	"log/slog"
	"path/filepath"
	"strings"
)

type ContentReference struct {
	Id           string
	Source       string
	FieldName    string
	SubReference []ContentReference
}

func (p *Project) FindReferences(id string) ([]ContentReference, error) {
	defer tracing.NewRegion("Project.FindReferences").End()
	refs := []ContentReference{}
	var add []ContentReference
	var err error
	if add, err = p.findReferencesStages(id); err != nil {
		return refs, err
	}
	refs = append(refs, add...)
	if add, err = p.findReferencesTemplates(id); err != nil {
		return refs, err
	}
	refs = append(refs, add...)
	if add, err = p.findReferencesMaterial(id); err != nil {
		return refs, err
	}
	refs = append(refs, add...)
	if add, err = p.findReferencesShader(id); err != nil {
		return refs, err
	}
	refs = append(refs, add...)
	return refs, nil
}

func (p *Project) findReferencesStages(id string) ([]ContentReference, error) {
	defer tracing.NewRegion("Project.findReferencesStages").End()
	return p.findRefsOnFolderAndDo(id, project_file_system.ContentStageFolder,
		func(name string, src []byte) (ContentReference, error) {
			ref := ContentReference{
				Id:     name,
				Source: content_database.Stage{}.TypeName(),
			}
			var ss stages.StageJson
			if err := json.Unmarshal(src, &ss); err != nil {
				return ref, err
			}
			s := stages.Stage{}
			s.FromMinimized(ss)
			for i := range s.Entities {
				subs := p.findEntityRefs(&s.Entities[i], id)
				if len(subs) > 0 {
					ref.SubReference = append(ref.SubReference, subs...)
				}
			}
			return ref, nil
		})
}

func (p *Project) findReferencesTemplates(id string) ([]ContentReference, error) {
	defer tracing.NewRegion("Project.findReferencesTemplates").End()
	return p.findRefsOnFolderAndDo(id, project_file_system.ContentTemplateFolder,
		func(name string, src []byte) (ContentReference, error) {
			ref := ContentReference{
				Id:     name,
				Source: content_database.Stage{}.TypeName(),
			}
			var desc stages.EntityDescription
			if err := json.Unmarshal(src, &desc); err != nil {
				return ref, err
			}
			ref.SubReference = p.findEntityRefs(&desc, id)
			return ref, nil
		})
}

func (p *Project) findReferencesMaterial(id string) ([]ContentReference, error) {
	defer tracing.NewRegion("Project.findReferencesMaterial").End()
	return p.findRefsOnFolderAndDo(id, project_file_system.ContentMaterialFolder,
		func(name string, src []byte) (ContentReference, error) {
			return ContentReference{
				Id:     name,
				Source: content_database.Material{}.TypeName(),
			}, nil
		})
}

func (p *Project) findReferencesShader(id string) ([]ContentReference, error) {
	defer tracing.NewRegion("Project.findReferencesShader").End()
	return p.findRefsOnFolderAndDo(id, project_file_system.ContentShaderFolder,
		func(name string, src []byte) (ContentReference, error) {
			return ContentReference{
				Id:     name,
				Source: content_database.Shader{}.TypeName(),
			}, nil
		})
}

func (p *Project) findRefsOnFolderAndDo(id, folder string, do func(name string, src []byte) (ContentReference, error)) ([]ContentReference, error) {
	defer tracing.NewRegion("Project.findRefsOnFolderAndDo").End()
	refs := []ContentReference{}
	dir := project_file_system.ContentFolderPath(folder)
	entries, err := p.fileSystem.ReadDir(dir)
	if err != nil {
		slog.Error("failed to read the target content", "error", err)
		return refs, err
	}
	for i := range entries {
		if entries[i].Name()[0] == '.' {
			continue
		}
		entryName := entries[i].Name()
		name := filepath.Join(dir, entryName)
		data, err := p.fileSystem.ReadFile(name)
		if err != nil {
			slog.Error("failed to read the target content file", "file", name, "error", err)
			return refs, err
		}
		if strings.Contains(string(data), id) {
			r, err := do(entryName, data)
			if err != nil {
				return refs, err
			}
			refs = append(refs, r)
		}
	}
	return refs, nil
}

func (p *Project) findEntityRefs(e *stages.EntityDescription, id string) []ContentReference {
	refs := []ContentReference{}
	sub := ContentReference{
		Id:     e.Id,
		Source: "entity",
	}
	if e.Material == id {
		sub.SubReference = append(sub.SubReference, ContentReference{
			Id:     id,
			Source: content_database.Material{}.TypeName(),
		})
	}
	for i := range e.Textures {
		if e.Textures[i] == id {
			sub.SubReference = append(sub.SubReference, ContentReference{
				Id:     id,
				Source: content_database.Texture{}.TypeName(),
			})
		}
	}
	if e.Mesh == id {
		sub.SubReference = append(sub.SubReference, ContentReference{
			Id:     id,
			Source: content_database.Mesh{}.TypeName(),
		})
	}
	if e.TemplateId == id {
		sub.SubReference = append(sub.SubReference, ContentReference{
			Id:     id,
			Source: content_database.Template{}.TypeName(),
		})
	}
	for i := range e.DataBinding {
		for k, v := range e.DataBinding[i].Fields {
			if s, ok := v.(string); ok && s == id {
				sub.SubReference = append(sub.SubReference, ContentReference{
					Id:        id,
					FieldName: k,
					Source:    "databind",
				})
			}
		}
	}
	if e.Id == id || len(sub.SubReference) > 0 {
		refs = append(refs, sub)
	}
	for i := range e.Children {
		cr := p.findEntityRefs(&e.Children[i], id)
		if len(cr) > 0 {
			refs = append(refs, cr...)
		}
	}
	return refs
}
