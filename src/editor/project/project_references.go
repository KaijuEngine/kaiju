/******************************************************************************/
/* project_references.go                                                      */
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

package project

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine/stages"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type ContentReference struct {
	Id           string
	Name         string
	Source       string
	SubReference []ContentReference
}

func (p *Project) FindReferences(id string) ([]ContentReference, error) {
	defer tracing.NewRegion("Project.FindReferences").End()
	refs := []ContentReference{}
	err := p.FindReferencesWithCallback(id, func(ref ContentReference) {
		refs = append(refs, ref)
	})
	return refs, err
}

func (p *Project) FindReferencesWithCallback(id string, onFound func(ref ContentReference)) error {
	defer tracing.NewRegion("Project.FindReferencesWithCallback").End()
	var err error
	funcs := []func(id string, onFound func(ref ContentReference)) error{
		p.findReferencesStages,
		p.findReferencesTemplates,
		p.findReferencesMaterial,
		p.findReferencesShader,
		p.findReferencesTableOfContents,
		p.findReferencesHtml,
		p.findReferencesCss,
		p.findReferencesCode,
	}
	for i := range funcs {
		if err = funcs[i](id, onFound); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) findReferencesStages(id string, onFound func(ref ContentReference)) error {
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
		}, onFound)
}

func (p *Project) findReferencesTemplates(id string, onFound func(ref ContentReference)) error {
	defer tracing.NewRegion("Project.findReferencesTemplates").End()
	return p.findRefsOnFolderAndDo(id, project_file_system.ContentTemplateFolder,
		func(name string, src []byte) (ContentReference, error) {
			ref := ContentReference{
				Id:     name,
				Source: content_database.Template{}.TypeName(),
			}
			var desc stages.EntityDescription
			if err := json.Unmarshal(src, &desc); err != nil {
				return ref, err
			}
			ref.SubReference = p.findEntityRefs(&desc, id)
			return ref, nil
		}, onFound)
}

func (p *Project) findReferencesMaterial(id string, onFound func(ref ContentReference)) error {
	defer tracing.NewRegion("Project.findReferencesMaterial").End()
	return p.findRefsOnFolderAndDo(id, project_file_system.ContentMaterialFolder,
		func(name string, src []byte) (ContentReference, error) {
			return ContentReference{
				Id:     name,
				Source: content_database.Material{}.TypeName(),
			}, nil
		}, onFound)
}

func (p *Project) findReferencesShader(id string, onFound func(ref ContentReference)) error {
	defer tracing.NewRegion("Project.findReferencesShader").End()
	return p.findRefsOnFolderAndDo(id, project_file_system.ContentShaderFolder,
		func(name string, src []byte) (ContentReference, error) {
			return ContentReference{
				Id:     name,
				Source: content_database.Shader{}.TypeName(),
			}, nil
		}, onFound)
}

func (p *Project) findReferencesTableOfContents(id string, onFound func(ref ContentReference)) error {
	defer tracing.NewRegion("Project.findReferencesTableOfContents").End()
	return p.findRefsOnFolderAndDo(id, project_file_system.ContentTableOfContentsFolder,
		func(name string, src []byte) (ContentReference, error) {
			return ContentReference{
				Id:     name,
				Source: content_database.TableOfContents{}.TypeName(),
			}, nil
		}, onFound)
}

func (p *Project) findReferencesHtml(id string, onFound func(ref ContentReference)) error {
	defer tracing.NewRegion("Project.findReferencesHtml").End()
	return p.findRefsOnFolderAndDo(id, project_file_system.ContentHtmlFolder,
		func(name string, src []byte) (ContentReference, error) {
			return ContentReference{
				Id:     name,
				Source: content_database.Html{}.TypeName(),
			}, nil
		}, onFound)
}

func (p *Project) findReferencesCss(id string, onFound func(ref ContentReference)) error {
	defer tracing.NewRegion("Project.findReferencesCss").End()
	return p.findRefsOnFolderAndDo(id, project_file_system.ContentCssFolder,
		func(name string, src []byte) (ContentReference, error) {
			return ContentReference{
				Id:     name,
				Source: content_database.Css{}.TypeName(),
			}, nil
		}, onFound)
}

func (p *Project) findReferencesCode(id string, onFound func(ref ContentReference)) error {
	defer tracing.NewRegion("Project.findReferencesCode").End()
	paths := []string{
		p.fileSystem.FullPath(project_file_system.KaijuSrcFolder),
		p.fileSystem.FullPath(project_file_system.ProjectCodeFolder),
	}
	var wErr error
	buffer := bytes.NewBuffer([]byte{})
	for i := range paths {
		wErr = filepath.Walk(paths[i], func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return err
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			buffer.Reset()
			if _, err = io.Copy(buffer, f); err != nil {
				return err
			}
			if strings.Contains(buffer.String(), id) {
				nml := p.fileSystem.NormalizePath(path)
				onFound(ContentReference{
					Id:     nml,
					Name:   filepath.Base(nml),
					Source: nml,
				})
			}
			return err
		})
		if wErr != nil {
			return wErr
		}
	}
	return nil
}

func (p *Project) findRefsOnFolderAndDo(id, folder string, do func(name string, src []byte) (ContentReference, error), onFound func(ref ContentReference)) error {
	defer tracing.NewRegion("Project.findRefsOnFolderAndDo").End()
	dir := filepath.Join(project_file_system.ContentFolder, folder)
	entries, err := p.fileSystem.ReadDir(dir)
	if err != nil {
		slog.Error("failed to read the target content", "error", err)
		return err
	}
	buffer := bytes.NewBuffer([]byte{})
	for i := range entries {
		if entries[i].Name()[0] == '.' {
			continue
		}
		entryName := entries[i].Name()
		if entryName == id {
			continue
		}
		name := filepath.Join(dir, entryName)
		f, err := p.fileSystem.Open(name)
		if err != nil {
			slog.Error("failed to open the target content file", "file", name, "error", err)
			return err
		}
		defer f.Close()
		buffer.Reset()
		if _, err := io.Copy(buffer, f); err != nil {
			slog.Error("failed to read the target content file", "file", name, "error", err)
			return err
		}
		data := buffer.Bytes()
		if strings.Contains(string(data), id) {
			r, err := do(entryName, data)
			if err != nil {
				return err
			}
			if cc, err := p.cacheDatabase.Read(entryName); err == nil {
				r.Name = cc.Config.Name
			}
			onFound(r)
		}
	}
	return nil
}

func (p *Project) findEntityRefs(e *stages.EntityDescription, id string) []ContentReference {
	refs := []ContentReference{}
	sub := ContentReference{
		Id:     e.Id,
		Name:   e.Name,
		Source: "entity",
	}
	if e.Material == id {
		sub.SubReference = append(sub.SubReference, ContentReference{
			Id:     id,
			Name:   "<self>",
			Source: content_database.Material{}.TypeName(),
		})
	}
	for i := range e.Textures {
		if e.Textures[i] == id {
			sub.SubReference = append(sub.SubReference, ContentReference{
				Id:     id,
				Name:   "<self>",
				Source: content_database.Texture{}.TypeName(),
			})
		}
	}
	if e.Mesh == id {
		sub.SubReference = append(sub.SubReference, ContentReference{
			Id:     id,
			Name:   "<self>",
			Source: content_database.Mesh{}.TypeName(),
		})
	}
	if e.TemplateId == id {
		sub.SubReference = append(sub.SubReference, ContentReference{
			Id:     id,
			Name:   "<self>",
			Source: content_database.Template{}.TypeName(),
		})
	}
	for i := range e.DataBinding {
		for k, v := range e.DataBinding[i].Fields {
			if s, ok := v.(string); ok && s == id {
				sub.SubReference = append(sub.SubReference, ContentReference{
					Id:     id,
					Name:   k,
					Source: "databind",
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

func (p *Project) updateReferences(from, to string) error {
	defer tracing.NewRegion("Project.updateReferences").End()
	refs, err := p.FindReferences(from)
	if err != nil {
		return err
	}
	var fixRef func(ref *ContentReference) error
	fixRef = func(ref *ContentReference) error {
		cc, err := p.cacheDatabase.Read(ref.Id)
		if err != nil {
			return err
		}
		data, err := p.fileSystem.ReadFile(cc.ContentPath())
		if err != nil {
			return err
		}
		str := strings.ReplaceAll(string(data), from, to)
		if err = p.fileSystem.WriteFile(cc.ContentPath(), []byte(str), os.ModePerm); err != nil {
			return err
		}
		for i := range ref.SubReference {
			if err := fixRef(&ref.SubReference[i]); err != nil {
				return err
			}
		}
		return nil
	}
	for i := range refs {
		if err := fixRef(&refs[i]); err != nil {
			return err
		}
	}
	return nil
}
