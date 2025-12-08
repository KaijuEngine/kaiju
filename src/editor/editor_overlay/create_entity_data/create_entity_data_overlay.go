/******************************************************************************/
/* create_entity_data_overlay.go                                              */
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

package create_entity_data

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/KaijuEngine/kaiju/editor/editor_overlay/file_browser"
	"github.com/KaijuEngine/kaiju/editor/project/project_file_system"
	"github.com/KaijuEngine/kaiju/engine"
	"github.com/KaijuEngine/kaiju/engine/ui"
	"github.com/KaijuEngine/kaiju/engine/ui/markup"
	"github.com/KaijuEngine/kaiju/engine/ui/markup/document"
	"github.com/KaijuEngine/kaiju/klib"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
)

type CreateEntityDataOverlay struct {
	fs        *project_file_system.FileSystem
	doc       *document.Document
	uiMan     ui.Manager
	config    Config
	nameInput *document.Element
	folder    *document.Element
	errorElm  *document.Element
}

type Config struct {
	OnCreate func()
	OnCancel func()
}

type FileTemplateData struct {
	FolderName string
	StructName string
}

func Show(host *engine.Host, fs *project_file_system.FileSystem, config Config) (*CreateEntityDataOverlay, error) {
	defer tracing.NewRegion("create_entity_data.Show").End()
	c := &CreateEntityDataOverlay{
		fs:     fs,
		config: config,
	}
	c.uiMan.Init(host)
	var err error
	c.doc, err = markup.DocumentFromHTMLAsset(&c.uiMan, "editor/ui/overlay/create_entity_data.go.html",
		c.config, map[string]func(*document.Element){
			"browse":  c.browse,
			"confirm": c.confirm,
			"cancel":  c.cancel,
		})
	if err != nil {
		return c, err
	}
	c.nameInput, _ = c.doc.GetElementById("nameInput")
	c.folder, _ = c.doc.GetElementById("folder")
	c.errorElm, _ = c.doc.GetElementById("error")
	c.errorElm.UI.Hide()
	return c, err
}

func (c *CreateEntityDataOverlay) Close() {
	defer tracing.NewRegion("CreateEntityDataOverlay.Close").End()
	c.doc.Destroy()
}

func (c *CreateEntityDataOverlay) browse(e *document.Element) {
	c.uiMan.DisableUpdate()
	file_browser.Show(c.uiMan.Host, file_browser.Config{
		Title:        "Target folder for entity data",
		StartingPath: c.fs.FullPath("src"),
		OnConfirm: func(paths []string) {
			c.folder.UI.ToInput().SetText(paths[0])
			c.uiMan.EnableUpdate()
		},
		OnCancel:    c.uiMan.EnableUpdate,
		OnlyFolders: true,
	})
}

func (c *CreateEntityDataOverlay) confirm(e *document.Element) {
	defer tracing.NewRegion("CreateEntityDataOverlay.confirm").End()
	name := strings.TrimSpace(c.nameInput.UI.ToInput().Text())
	folder := c.folder.UI.ToInput().Text()
	nameRunes := []rune(name)
	errMsg := ""
	if name == "" {
		errMsg = "Struct name is required"
	} else if !unicode.IsLetter(nameRunes[0]) {
		errMsg = "Struct name must start with an upper case letter"
	} else if !unicode.IsUpper(nameRunes[0]) {
		errMsg = "Struct name must start with an upper case letter (exported)"
	}
	if errMsg != "" {
		c.writeError(errMsg, nil)
		return
	}
	tplData := FileTemplateData{
		FolderName: filepath.Base(folder),
		StructName: name,
	}
	data, err := c.fs.ReadFile(filepath.Join(project_file_system.ProjectFileTemplates,
		"entity_data_file_template.go.txt"))
	if err != nil {
		c.writeError("failed to read "+project_file_system.ProjectFileTemplates+"/entity_data_file_template.go.txt", err)
		return
	}
	t, err := template.New("EntityData").Parse(string(data))
	if err != nil {
		c.writeError("failed to parse "+project_file_system.ProjectFileTemplates+"/entity_data_file_template.go.txt", err)
		return
	}
	sb := strings.Builder{}
	if err = t.Execute(&sb, tplData); err != nil {
		c.writeError("failed to execute "+project_file_system.ProjectFileTemplates+"/entity_data_file_template.go.txt", err)
		return
	}
	outPath := filepath.Join(folder, klib.ToSnakeCase(name)+".go")
	if _, err := os.Stat(outPath); err == nil {
		c.writeError(fmt.Sprintf("the output file path '%s' already exists", outPath), nil)
		return
	}
	if err = os.WriteFile(outPath, []byte(sb.String()), os.ModePerm); err != nil {
		c.writeError(fmt.Sprintf("failed to write the entity data file: %s", outPath), err)
		return
	}
	c.Close()
	if c.config.OnCreate == nil {
		slog.Error("the input prompt didn't have a OnConfirm set, nothing to do")
		return
	}
	c.config.OnCreate()
	exec.Command("code", c.fs.FullPath(""), outPath).Run()
}

func (c *CreateEntityDataOverlay) writeError(msg string, err error) {
	if err != nil {
		slog.Error(msg, "error", err)
	} else {
		slog.Error(msg)
	}
	c.errorElm.InnerLabel().SetText(msg)
	c.errorElm.UI.Show()
}

func (c *CreateEntityDataOverlay) cancel(e *document.Element) {
	defer tracing.NewRegion("CreateEntityDataOverlay.cancel").End()
	c.Close()
	if c.config.OnCancel != nil {
		c.config.OnCancel()
	}
}
