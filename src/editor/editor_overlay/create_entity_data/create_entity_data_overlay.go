package create_entity_data

import (
	"fmt"
	"kaiju/editor/editor_overlay/file_browser"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
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
