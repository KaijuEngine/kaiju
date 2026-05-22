package integration_testing

import (
	"fmt"
	"reflect"

	"kaijuengine.com/editor/editor_embedded_content"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
)

var tests = map[string]func(host *engine.Host){}

type IntegrationGame struct {
	test func(host *engine.Host)
}

func IntegrationTestGame(testName string) (*IntegrationGame, error) {
	if test, ok := tests[testName]; ok {
		return &IntegrationGame{test: test}, nil
	}
	return nil, fmt.Errorf("could not find test named %s, perhaps you forgot to build the executable", testName)
}

func (IntegrationGame) PluginRegistry() []reflect.Type { return []reflect.Type{} }

func (IntegrationGame) ContentDatabase() (assets.Database, error) {
	// TODO:  Only do this if it is the editor, otherwise use standard content
	return &editor_embedded_content.EditorContent{}, nil
}

func (g *IntegrationGame) Launch(host *engine.Host) { g.test(host) }
