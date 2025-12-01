/******************************************************************************/
/* editor_ai_bindings.go                                                      */
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

package editor

import (
	_ "embed"
	"kaiju/klib"
	"kaiju/ollama"
)

//go:embed docs.md
var docs string

type llmDocs struct{}

type llmDocsResult struct {
	ollama.LLMActionResultBase
	Docs string `json:"docs"`
}

type llmIssues struct {
	Kind string `json:"kind" enum:"bug,feature,view" desc:"The kind of issue"`
}

func (llmDocs) Execute() (any, error) {
	return llmDocsResult{Docs: docs}, nil
}

func (a llmIssues) Execute() (any, error) {
	addr := "https://github.com/KaijuEngine/kaiju/issues"
	switch a.Kind {
	case "bug":
		addr = "https://github.com/KaijuEngine/kaiju/issues/new?template=bug_report.md"
	case "feature":
		addr = "https://github.com/KaijuEngine/kaiju/issues/new?template=feature_request.md"
	case "view":
		addr = "https://github.com/KaijuEngine/kaiju/issues"
	}
	klib.OpenWebsite(addr)
	return nil, nil
}

func init() {
	ollama.ReflectFuncToOllama(llmDocs{},
		"docs", "Get the documentation text for the engine to know how to use it.")
	ollama.ReflectFuncToOllama(llmIssues{},
		"issue", "Open the web browser to show the GitHub issues")
}
