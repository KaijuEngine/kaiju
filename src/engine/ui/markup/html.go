/******************************************************************************/
/* html.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package markup

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"weak"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

var (
	htmlIncludeTagRE = regexp.MustCompile(`(?is)<kaiju-include\b([^>]*)>\s*</kaiju-include>|<kaiju-include\b([^>]*)/>`)
	htmlIncludeSrcRE = regexp.MustCompile(`(?is)\bsrc\s*=\s*"([^"]+)"|\bsrc\s*=\s*'([^']+)'`)
)

func DocumentFromHTMLAsset(uiMan *ui.Manager, htmlPath string, withData any, funcMap map[string]func(*document.Element)) (*document.Document, error) {
	host := uiMan.Host
	m, err := host.AssetDatabase().ReadText(htmlPath)
	if err != nil {
		return nil, err
	}
	m, err = expandHTMLIncludes(host.AssetDatabase(), htmlPath, m, map[string]bool{})
	if err != nil {
		return nil, err
	}
	doc := DocumentFromHTMLString(uiMan, m, "", withData, funcMap, nil)
	return doc, nil
}

func DocumentFromHTMLAssetRooted(uiMan *ui.Manager, htmlPath string, withData any, funcMap map[string]func(*document.Element), root *document.Element) (*document.Document, error) {
	host := uiMan.Host
	m, err := host.AssetDatabase().ReadText(htmlPath)
	if err != nil {
		return nil, err
	}
	m, err = expandHTMLIncludes(host.AssetDatabase(), htmlPath, m, map[string]bool{})
	if err != nil {
		return nil, err
	}
	doc := DocumentFromHTMLString(uiMan, m, "", withData, funcMap, root)
	return doc, nil
}

func expandHTMLIncludes(db assets.Database, ownerPath, html string, stack map[string]bool) (string, error) {
	var firstErr error
	expanded := htmlIncludeTagRE.ReplaceAllStringFunc(html, func(includeTag string) string {
		if firstErr != nil {
			return includeTag
		}
		matches := htmlIncludeTagRE.FindStringSubmatch(includeTag)
		attrs := ""
		if len(matches) > 1 {
			attrs = strings.TrimSpace(matches[1])
		}
		if attrs == "" && len(matches) > 2 {
			attrs = strings.TrimSpace(matches[2])
		}
		srcMatches := htmlIncludeSrcRE.FindStringSubmatch(attrs)
		if len(srcMatches) == 0 {
			firstErr = fmt.Errorf("kaiju-include in %s is missing a src attribute", ownerPath)
			return includeTag
		}
		includePath := srcMatches[1]
		if includePath == "" {
			includePath = srcMatches[2]
		}
		includePath = resolveHTMLIncludePath(ownerPath, includePath)
		if stack[includePath] {
			firstErr = fmt.Errorf("kaiju-include cycle detected while reading %s", includePath)
			return includeTag
		}
		stack[includePath] = true
		includeHTML, err := db.ReadText(includePath)
		if err != nil {
			firstErr = fmt.Errorf("failed to read kaiju-include %s from %s: %w", includePath, ownerPath, err)
			delete(stack, includePath)
			return includeTag
		}
		includeHTML, err = expandHTMLIncludes(db, includePath, includeHTML, stack)
		delete(stack, includePath)
		if err != nil {
			firstErr = err
			return includeTag
		}
		return includeHTML
	})
	if firstErr != nil {
		return "", firstErr
	}
	return expanded, nil
}

func resolveHTMLIncludePath(ownerPath, includePath string) string {
	includePath = path.Clean(filepathToSlash(includePath))
	if strings.HasPrefix(includePath, "/") ||
		strings.HasPrefix(includePath, "editor/") ||
		strings.Contains(includePath, "://") {
		return strings.TrimPrefix(includePath, "/")
	}
	return path.Clean(path.Join(path.Dir(filepathToSlash(ownerPath)), includePath))
}

func filepathToSlash(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}

func DocumentFromHTMLString(uiMan *ui.Manager, html, cssStr string, withData any, funcMap map[string]func(*document.Element), root *document.Element) *document.Document {
	host := uiMan.Host
	window := host.Window
	doc := document.DocumentFromHTMLString(uiMan, html, withData, funcMap)
	if root != nil {
		// Root the HTML body under the provided root element so layout behaves
		// like a normal document tree. Reparenting only top children can break
		// width containment and lead to runaway size feedback in previews.
		if bodyElms := doc.GetElementsByTagName("body"); len(bodyElms) > 0 {
			body := bodyElms[0]
			root.UIPanel.AddChild(body.UI)
			root.Children = append(root.Children, body)
			body.Parent = weak.Make(root)
			doc.TopElements = []*document.Element{body}
		}
	}
	s := rules.NewStyleSheet()
	s.Parse(css.DefaultCSS, window)
	if css.OverrideCSS != "" {
		s.Parse(css.OverrideCSS, window)
	}
	s.Parse(cssStr, window)
	for i := range doc.HeadElements {
		if doc.HeadElements[i].Data == "style" {
			if len(doc.HeadElements[i].Children) > 0 {
				s.Parse(doc.HeadElements[i].Children[0].Data, window)
			}
		} else if doc.HeadElements[i].Data == "link" {
			if doc.HeadElements[i].Attribute("rel") == "stylesheet" {
				cssPath := doc.HeadElements[i].Attribute("href")
				css, err := host.AssetDatabase().ReadText(cssPath)
				if err != nil {
					continue
				}
				s.Parse(css, window)
			}
		}
	}
	if root == nil {
		ui.RunDirtyBatch(uiMan, func() {
			doc.SetupStyle(s, host, css.Stylizer{Window: uiMan.Host.Window})
		})
	} else {
		doc.SetupStyle(s, host, css.Stylizer{Window: uiMan.Host.Window})
	}
	return doc
}
