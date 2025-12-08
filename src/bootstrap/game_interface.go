/******************************************************************************/
/* game_interface.go                                                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package bootstrap

import (
	"kaiju/engine"
	"kaiju/engine/assets"
	"reflect"
)

// GameInterface is the primary interface to implement in order to bootstrap
// a game/application.
type GameInterface interface {
	// Launch is used to bootstrap a game, the game should fill out this
	// function's details to initialize itself. No updates are provided by the
	// engine, so it is on the the implementing code to take care of registering
	// any udpates with the supplied host.
	Launch(*engine.Host)

	// PluginRegistry is used to expose types to be exported for use in Lua.
	// Any type returned here will have it's members and functions mapped to
	// be called by Lua. You can run the engine with the command line argument
	// "generate=pluginapi" to dump a Lua API file and ensure your exposed
	// types have been correctly inserted.
	PluginRegistry() []reflect.Type

	// ContentDatabase must return the database interface for the engine to use
	// when it is trying to access content. You can use exsiting types that
	// implement [assets.Database], or you can create your own.
	ContentDatabase() (assets.Database, error)
}
