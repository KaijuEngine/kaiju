//go:build !editor

package main

import (
	"kaiju/bootstrap"
	"kaiju/games/editor"
	"reflect"
)

type DummyGame struct{}

func (DummyGame) PluginRegistry() []reflect.Type {
	return []reflect.Type{}
}

func (DummyGame) ContentDatabase() (assets.Database, error) {
	return assets.NewFileDatabase("content")
}

func (DummyGame) Launch(*engine.Host) {}

func getGame() bootstrap.GameInterface { return editor.DummyGame{} }
