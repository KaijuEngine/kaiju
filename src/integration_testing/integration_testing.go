/******************************************************************************/
/* integration_testing.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"reflect"

	"kaijuengine.com/engine"
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

func (g *IntegrationGame) Launch(host *engine.Host) { g.test(host) }
