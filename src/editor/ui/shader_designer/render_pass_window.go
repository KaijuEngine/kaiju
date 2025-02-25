package shader_designer

import (
	"encoding/json"
	"kaiju/editor/alert"
	"kaiju/editor/editor_config"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const (
	renderPassHTML = "editor/ui/shader_designer/render_pass_window.html"
)

type renderPassHTMLData struct {
	rendering.RenderPassData
}

func (d renderPassHTMLData) SrcStageMaskFlagState(index int, a rendering.RenderPassSubpassDependency) flagState {
	return flagState{
		List:    klib.MapKeysSorted(rendering.StringVkPipelineStageFlagBits),
		Current: a.SrcStageMask,
		Array:   "SubpassDependencies",
		Field:   "SrcStageMask",
		Index:   index,
	}
}

func (d renderPassHTMLData) DstStageMaskFlagState(index int, a rendering.RenderPassSubpassDependency) flagState {
	return flagState{
		List:    klib.MapKeysSorted(rendering.StringVkPipelineStageFlagBits),
		Current: a.DstStageMask,
		Array:   "SubpassDependencies",
		Field:   "DstStageMask",
		Index:   index,
	}
}

func (d renderPassHTMLData) SrcAccessMaskFlagState(index int, a rendering.RenderPassSubpassDependency) flagState {
	return flagState{
		List:    klib.MapKeysSorted(rendering.StringVkAccessFlagBits),
		Current: a.SrcAccessMask,
		Array:   "SubpassDependencies",
		Field:   "SrcAccessMask",
		Index:   index,
	}
}

func (d renderPassHTMLData) DstAccessMaskFlagState(index int, a rendering.RenderPassSubpassDependency) flagState {
	return flagState{
		List:    klib.MapKeysSorted(rendering.StringVkAccessFlagBits),
		Current: a.DstAccessMask,
		Array:   "SubpassDependencies",
		Field:   "DstAccessMask",
		Index:   index,
	}
}

func (d renderPassHTMLData) DependencyFlagsState(index int, a rendering.RenderPassSubpassDependency) flagState {
	return flagState{
		List:    klib.MapKeysSorted(rendering.StringVkDependencyFlagBits),
		Current: a.DependencyFlags,
		Array:   "SubpassDependencies",
		Field:   "DependencyFlags",
		Index:   index,
	}
}

func OpenRenderPass(path string) {
	setup(func(win *ShaderDesigner) {
		data, err := os.ReadFile(path)
		if err != nil {
			slog.Error("failed to load the render pass file", "file", path, "error", err)
			return
		}
		if err := json.Unmarshal(data, &win.renderPass); err != nil {
			slog.Error("failed to unmarshal the render pass data", "error", err)
			return
		}
		win.reloadRenderPassDoc()
	})
}

func setupRenderPassDefaults() rendering.RenderPassData {
	return rendering.RenderPassData{
		Name:                   "",
		AttachmentDescriptions: make([]rendering.RenderPassAttachmentDescription, 0),
		SubpassDescriptions:    make([]rendering.RenderPassSubpassDescription, 0),
		SubpassDependencies:    make([]rendering.RenderPassSubpassDependency, 0),
	}
}

func setupRenderPassDoc(win *ShaderDesigner) {
	win.renderPass = setupRenderPassDefaults()
	win.reloadRenderPassDoc()
	//win.renderPassDoc.Deactivate()
}

func (win *ShaderDesigner) reloadRenderPassDoc() {
	sy := float32(0)
	if win.renderPassDoc != nil {
		content := win.renderPassDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.renderPassDoc.Destroy()
	}
	data := renderPassHTMLData{win.renderPass}
	win.renderPassDoc, _ = markup.DocumentFromHTMLAsset(&win.man, renderPassHTML,
		data, map[string]func(*document.Element){
			"showTooltip":                             showPipelineTooltip,
			"valueChanged":                            win.renderPassValueChanged,
			"nameChanged":                             win.renderPassNameChanged,
			"addAttachmentDescription":                win.renderPassAddAttachmentDescription,
			"deleteAttachmentDescription":             win.renderPassDeleteAttachmentDescription,
			"addSubpassDescription":                   win.renderPassAddSubpassDescription,
			"deleteSubpassDescription":                win.renderPassDeleteSubpassDescription,
			"addSubpassDescriptionColorRef":           win.renderPassAddSubpassDescriptionColorRef,
			"deleteSubpassDescriptionColorRef":        win.renderPassDeleteSubpassDescriptionColorRef,
			"addSubpassDescriptionInputRef":           win.renderPassAddSubpassDescriptionInputRef,
			"deleteSubpassDescriptionInputRef":        win.renderPassDeleteSubpassDescriptionInputRef,
			"addSubpassDescriptionResolveRef":         win.renderPassAddSubpassDescriptionResolveRef,
			"deleteSubpassDescriptionResolveRef":      win.renderPassDeleteSubpassDescriptionResolveRef,
			"addSubpassDescriptionDepthStencilRefs":   win.renderPassAddSubpassDescriptionDepthStencilRefs,
			"deleteSubpassDescriptionDepthStencilRef": win.renderPassDeleteSubpassDescriptionDepthStencilRef,
			"addSubpassDependency":                    win.renderPassAddSubpassDependency,
			"deleteSubpassDependency":                 win.renderPassDeleteSubpassDependency,
			"saveRenderPass":                          win.renderPassSaveRenderPass,
		})
	if sy != 0 {
		content := win.renderPassDoc.GetElementsByClass("topFields")[0]
		win.man.Host.RunAfterFrames(2, func() {
			content.UIPanel.SetScrollY(sy)
		})
	}
}

func (win *ShaderDesigner) renderPassValueChanged(e *document.Element) {
	path := e.Attribute("data-path")
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(&win.renderPass).Elem()
	for i := range parts {
		if idx, err := strconv.Atoi(parts[i]); err == nil {
			v = v.Index(idx)
		} else {
			v = v.FieldByName(parts[i])
		}
	}
	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.String {
		// TODO:  Ensure switch e.UI.Type() == ui.ElementTypeCheckbox
		add := e.UI.ToCheckbox().IsChecked()
		str := e.Attribute("name")
		var slice []string
		if !v.IsNil() {
			slice = v.Interface().([]string)
		} else {
			slice = []string{}
		}
		if add {
			for _, s := range slice {
				if s == str {
					return // Already exists, no change
				}
			}
			slice = append(slice, str)
		} else {
			for i, s := range slice {
				if s == str {
					slice = slices.Delete(slice, i, i+1)
					break
				}
			}
		}
		v.Set(reflect.ValueOf(slice))
	} else {
		var val reflect.Value
		switch e.UI.Type() {
		case ui.ElementTypeInput:
			res := klib.StringToTypeValue(v.Type().String(), e.UI.ToInput().Text())
			val = reflect.ValueOf(res)
		case ui.ElementTypeSelect:
			val = reflect.ValueOf(e.UI.ToSelect().Value())
		case ui.ElementTypeCheckbox:
			val = reflect.ValueOf(e.UI.ToCheckbox().IsChecked())
		}
		v.Set(val)
	}
}

func (win *ShaderDesigner) renderPassNameChanged(e *document.Element) {
	win.renderPass.Name = e.UI.ToInput().Text()
}

func (win *ShaderDesigner) renderPassAddAttachmentDescription(*document.Element) {
	win.renderPass.AttachmentDescriptions = append(win.renderPass.AttachmentDescriptions,
		rendering.RenderPassAttachmentDescription{
			// TODO:  Set the defaults
			Format:         "",
			Samples:        "",
			LoadOp:         "",
			StoreOp:        "",
			StencilLoadOp:  "",
			StencilStoreOp: "",
			InitialLayout:  "",
			FinalLayout:    "",
		})
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassDeleteAttachmentDescription(e *document.Element) {
	ok := <-alert.New("Delete entry?", "Are you sure you want to delete this attachment description entry? The action currently can't be undone.", "Yes", "No", win.man.Host)
	if !ok {
		return
	}
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	win.renderPass.AttachmentDescriptions = slices.Delete(
		win.renderPass.AttachmentDescriptions, idx, idx+1)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassAddSubpassDescription(*document.Element) {
	win.renderPass.SubpassDescriptions = append(win.renderPass.SubpassDescriptions,
		rendering.RenderPassSubpassDescription{
			PipelineBindPoint:         "Graphics",
			ColorAttachmentReferences: make([]rendering.RenderPassAttachmentReference, 0),
			InputAttachmentReferences: make([]rendering.RenderPassAttachmentReference, 0),
			ResolveAttachments:        make([]rendering.RenderPassAttachmentReference, 0),
			DepthStencilAttachment:    make([]rendering.RenderPassAttachmentReference, 0),
			PreserveAttachments:       make([]uint32, 0),
		})
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassDeleteSubpassDescription(e *document.Element) {
	ok := <-alert.New("Delete entry?", "Are you sure you want to delete this subpass description entry? The action currently can't be undone.", "Yes", "No", win.man.Host)
	if !ok {
		return
	}
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	win.renderPass.SubpassDescriptions = slices.Delete(
		win.renderPass.SubpassDescriptions, idx, idx+1)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassAddSubpassDescriptionColorRef(e *document.Element) {
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	win.renderPass.SubpassDescriptions[idx].ColorAttachmentReferences = append(
		win.renderPass.SubpassDescriptions[idx].ColorAttachmentReferences,
		rendering.RenderPassAttachmentReference{})
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassDeleteSubpassDescriptionColorRef(e *document.Element) {
	ok := <-alert.New("Delete entry?", "Are you sure you want to delete this subpass description color reference entry? The action currently can't be undone.", "Yes", "No", win.man.Host)
	if !ok {
		return
	}
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	subIdx, _ := strconv.Atoi(e.Attribute("data-subindex"))
	win.renderPass.SubpassDescriptions[idx].ColorAttachmentReferences = slices.Delete(
		win.renderPass.SubpassDescriptions[idx].ColorAttachmentReferences, subIdx, subIdx+1)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassAddSubpassDescriptionInputRef(e *document.Element) {
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	win.renderPass.SubpassDescriptions[idx].InputAttachmentReferences = append(
		win.renderPass.SubpassDescriptions[idx].InputAttachmentReferences,
		rendering.RenderPassAttachmentReference{})
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassDeleteSubpassDescriptionInputRef(e *document.Element) {
	ok := <-alert.New("Delete entry?", "Are you sure you want to delete this subpass description input reference entry? The action currently can't be undone.", "Yes", "No", win.man.Host)
	if !ok {
		return
	}
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	subIdx, _ := strconv.Atoi(e.Attribute("data-subindex"))
	win.renderPass.SubpassDescriptions[idx].InputAttachmentReferences = slices.Delete(
		win.renderPass.SubpassDescriptions[idx].InputAttachmentReferences, subIdx, subIdx+1)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassAddSubpassDescriptionResolveRef(e *document.Element) {
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	win.renderPass.SubpassDescriptions[idx].ResolveAttachments = append(
		win.renderPass.SubpassDescriptions[idx].ResolveAttachments,
		rendering.RenderPassAttachmentReference{})
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassDeleteSubpassDescriptionResolveRef(e *document.Element) {
	ok := <-alert.New("Delete entry?", "Are you sure you want to delete this subpass description resolve reference entry? The action currently can't be undone.", "Yes", "No", win.man.Host)
	if !ok {
		return
	}
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	subIdx, _ := strconv.Atoi(e.Attribute("data-subindex"))
	win.renderPass.SubpassDescriptions[idx].ResolveAttachments = slices.Delete(
		win.renderPass.SubpassDescriptions[idx].ResolveAttachments, subIdx, subIdx+1)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassAddSubpassDescriptionDepthStencilRefs(e *document.Element) {
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	win.renderPass.SubpassDescriptions[idx].DepthStencilAttachment = append(
		win.renderPass.SubpassDescriptions[idx].DepthStencilAttachment,
		rendering.RenderPassAttachmentReference{})
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassDeleteSubpassDescriptionDepthStencilRef(e *document.Element) {
	ok := <-alert.New("Delete entry?", "Are you sure you want to delete this subpass description depth stencil reference entry? The action currently can't be undone.", "Yes", "No", win.man.Host)
	if !ok {
		return
	}
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	subIdx, _ := strconv.Atoi(e.Attribute("data-subindex"))
	win.renderPass.SubpassDescriptions[idx].DepthStencilAttachment = slices.Delete(
		win.renderPass.SubpassDescriptions[idx].DepthStencilAttachment, subIdx, subIdx+1)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassAddSubpassDependency(*document.Element) {
	win.renderPass.SubpassDependencies = append(win.renderPass.SubpassDependencies,
		rendering.RenderPassSubpassDependency{
			SrcSubpass:      0,
			DstSubpass:      0,
			SrcStageMask:    make([]string, 0),
			DstStageMask:    make([]string, 0),
			SrcAccessMask:   make([]string, 0),
			DstAccessMask:   make([]string, 0),
			DependencyFlags: make([]string, 0),
		})
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassDeleteSubpassDependency(e *document.Element) {
	ok := <-alert.New("Delete entry?", "Are you sure you want to delete this subpass dependency entry? The action currently can't be undone.", "Yes", "No", win.man.Host)
	if !ok {
		return
	}
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	win.renderPass.SubpassDependencies = slices.Delete(
		win.renderPass.SubpassDependencies, idx, idx+1)
	win.reloadRenderPassDoc()
}

func (win *ShaderDesigner) renderPassSaveRenderPass(e *document.Element) {
	const saveRoot = "content/shaders/passes"
	if err := os.MkdirAll(saveRoot, os.ModePerm); err != nil {
		slog.Error("failed to create the render pass folder",
			"folder", saveRoot, "error", err)
	}
	path := filepath.Join(saveRoot, win.renderPass.Name+editor_config.FileExtensionRenderPass)
	if _, err := os.Stat(path); err == nil {
		ok := <-alert.New("Overwrite?", "You are about to overwrite a render pass with the same name, would you like to continue?", "Yes", "No", win.man.Host)
		if !ok {
			return
		}
	}
	res, err := json.Marshal(win.renderPass)
	if err != nil {
		slog.Error("failed to marshal the render pass data", "error", err)
		return
	}
	if err := os.WriteFile(path, res, os.ModePerm); err != nil {
		slog.Error("failed to write the render pass data to file", "error", err)
		return
	}
	slog.Info("render pass successfully saved", "file", path)
	// TODO:  Show an in-window popup for prompting that things saved
	if len(e.Children) > 0 {
		u := e.Children[0].UI
		if u.IsType(ui.ElementTypeLabel) {
			u.ToLabel().SetText("File saved!")
		}
	}
}
