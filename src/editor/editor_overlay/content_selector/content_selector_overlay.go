/******************************************************************************/
/* content_selector_overlay.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_selector

import (
	"log/slog"
	"strings"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

type ContentSelector struct {
	doc      *document.Document
	uiMan    ui.Manager
	keyKb    hid.KeyCallbackId
	onSelect func(id string)
	onClose  func()
	list     *document.Element
}

type contentSelectorData struct {
	Options []contentSelectorEntry
}

type contentSelectorEntry struct {
	Id      string
	Name    string
	Texture string
}

func Show(host *engine.Host, typeName string, cache *content_database.Cache, onSelect func(id string), onClose func()) (*ContentSelector, error) {
	defer tracing.NewRegion("content_selector.Show").End()
	o := &ContentSelector{onSelect: onSelect, onClose: onClose}
	o.uiMan.Init(host)
	var err error
	all := cache.ListByType(typeName)
	data := contentSelectorData{
		Options: make([]contentSelectorEntry, 0, len(all)+2),
	}
	for i := range all {
		if typeName == (content_database.Mesh{}).TypeName() && len(contentSelectorMeshSubmeshes(all[i].Config.Mesh)) > 1 {
			for _, submesh := range contentSelectorMeshSubmeshes(all[i].Config.Mesh) {
				data.Options = append(data.Options, contentSelectorEntry{
					Id:      kaiju_mesh.MeshRefString(all[i].Id(), submesh.Key),
					Name:    all[i].Config.Name + " / " + contentSelectorMeshName(submesh.Name, submesh.Key),
					Texture: "editor/textures/icons/file.png",
				})
			}
		} else {
			entry := contentSelectorEntry{
				Id:   all[i].Id(),
				Name: all[i].Config.Name,
			}
			if all[i].Config.Type == (content_database.Texture{}).TypeName() {
				entry.Texture = entry.Id
			} else {
				entry.Texture = "editor/textures/icons/file.png"
			}
			data.Options = append(data.Options, entry)
		}
	}
	if typeName == (content_database.Texture{}).TypeName() {
		data.Options = append(data.Options, contentSelectorEntry{
			Id:      assets.TextureSquare,
			Name:    assets.TextureSquare,
			Texture: assets.TextureSquare,
		})
	}
	if typeName == (content_database.Mesh{}).TypeName() {
		primitives := []struct {
			mesh rendering.PrimitiveMesh
			name string
		}{
			{rendering.PrimitiveMeshTexturableCube, "Cube"},
			{rendering.PrimitiveMeshSphere, "Sphere"},
			{rendering.PrimitiveMeshPlane, "Plane"},
			{rendering.PrimitiveMeshCapsule, "Capsule"},
			{rendering.PrimitiveMeshCylinder, "Cylinder"},
			{rendering.PrimitiveMeshCone, "Cone"},
			{rendering.PrimitiveMeshArrow, "Arrow"},
		}
		for _, p := range primitives {
			data.Options = append(data.Options, contentSelectorEntry{
				Id:      string(p.mesh),
				Name:    p.name,
				Texture: "editor/textures/icons/Mesh.png",
			})
		}
	}
	data.Options = append(data.Options, contentSelectorEntry{
		Name:    "None",
		Texture: "editor/textures/icons/none.png",
	})
	o.doc, err = markup.DocumentFromHTMLAsset(&o.uiMan, "editor/ui/overlay/content_selector_overlay.go.html",
		data, map[string]func(*document.Element){
			"search":        o.search,
			"selectContent": o.selectContent,
		})
	if err != nil {
		return o, err
	}
	o.keyKb = host.Window.Keyboard.AddKeyCallback(func(keyId int, keyState hid.KeyState) {
		if keyId == hid.KeyboardKeyEscape {
			o.Close()
		}
	})
	o.list, _ = o.doc.GetElementById("list")
	return o, err
}

func contentSelectorMeshSubmeshes(cfg *content_database.MeshConfig) []content_database.MeshSubmeshConfig {
	if cfg == nil || len(cfg.Submeshes) <= 1 {
		return nil
	}
	out := make([]content_database.MeshSubmeshConfig, 0, len(cfg.Submeshes))
	for i := range cfg.Submeshes {
		if !cfg.Submeshes[i].Missing {
			out = append(out, cfg.Submeshes[i])
		}
	}
	if len(out) <= 1 {
		return nil
	}
	return out
}

func contentSelectorMeshName(name, key string) string {
	if strings.TrimSpace(name) != "" {
		parts := strings.Split(name, "/")
		return parts[len(parts)-1]
	}
	if strings.TrimSpace(key) != "" {
		return key
	}
	return "Mesh"
}

func (o *ContentSelector) Close() {
	defer tracing.NewRegion("ContentSelector.Close").End()
	o.closeInternal()
	if o.onClose == nil {
		slog.Warn("onClose was not set on the ContentSelector")
		return
	}
	o.onClose()
}

func (o *ContentSelector) closeInternal() {
	o.uiMan.Host.Window.CursorStandard()
	o.doc.Destroy()
	o.uiMan.Host.Window.Keyboard.RemoveKeyCallback(o.keyKb)
}

func (o *ContentSelector) search(e *document.Element) {
	defer tracing.NewRegion("ContentSelector.search").End()
	q := strings.ToLower(e.UI.ToInput().Text())
	for _, c := range o.list.Children {
		lbl := strings.ToLower(c.Children[1].InnerLabel().Text())
		if strings.Contains(lbl, q) {
			c.UI.Show()
		} else {
			c.UI.Hide()
		}
	}
}

func (o *ContentSelector) selectContent(e *document.Element) {
	defer tracing.NewRegion("ContentSelector.selectContent").End()
	id := e.Attribute("id")
	o.closeInternal()
	o.onSelect(id)
}
