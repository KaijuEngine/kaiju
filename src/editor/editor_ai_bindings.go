package editor

import (
	_ "embed"
	"kaiju/klib"
	"kaiju/ollama"
)

//go:embed docs.md
var docs string

func init() {
	ollama.ReflectFuncToOllama(func() string { return docs },
		"docs", "Get the documentation text for the engine to know how to use it.")
	ollama.ReflectFuncToOllama(func(kind string) string {
		addr := "https://github.com/KaijuEngine/kaiju/issues"
		switch kind {
		case "bug":
			addr = "https://github.com/KaijuEngine/kaiju/issues/new?template=bug_report.md"
		case "feature":
			addr = "https://github.com/KaijuEngine/kaiju/issues/new?template=feature_request.md"
		case "view":
			addr = "https://github.com/KaijuEngine/kaiju/issues"
		}
		klib.OpenWebsite(addr)
		return "issue tracker opened for developer"
	}, "issue", "Open the web browser to show the GitHub issues",
		"kind", "The kind of issue ['bug', 'feature', 'view']")
}
