/******************************************************************************/
/* editor_actions_lua.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/plugins"
	"kaijuengine.com/plugins/lua"
)

type luaRunScriptArgs struct {
	Path   string `json:"path"`
	Script string `json:"script"`
}

const (
	ActionLuaRunScript editor_action.ActionID = "lua.runScript"
)

func init() {
	registerEditorActionProvider(registerLuaActions)
}

func registerLuaActions(ed *Editor, mustRegister editorActionRegistrar) {
	mustRegister(editor_action.Definition{
		ID:          ActionLuaRunScript,
		Label:       "Run Lua Script",
		Description: "Runs a Lua script with access to the editor actions bridge.",
		Category:    "Lua",
		Tags:        []string{"lua", "script", "automation"},
		Parameters: []editor_action.Parameter{
			{Name: "path", Label: "Path", Type: "string"},
			{Name: "script", Label: "Script", Type: "string"},
		},
		NewParams:  func() any { return &luaRunScriptArgs{} },
		UndoPolicy: editor_action.UndoPolicyNone,
		Visible:    false,
	}, ed.actionRunLuaScript, nil)
}

func (ed *Editor) actionRunLuaScript(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	args, ok := editor_action.Param[luaRunScriptArgs](req)
	if !ok {
		return editor_action.Failure("path or script is required")
	}
	code := strings.TrimSpace(args.Script)
	name := "editor lua action"
	root := ed.ProjectFileSystem().FullPath("")
	if code == "" {
		path := strings.TrimSpace(args.Path)
		if path == "" {
			return editor_action.Failure("path or script is required")
		}
		if !filepath.IsAbs(path) {
			path = ed.ProjectFileSystem().FullPath(path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return editor_action.Failure(err.Error())
		}
		code = string(data)
		name = "@" + path
		root = filepath.Dir(path)
	}
	vm, err := plugins.NewScriptVM(ed.host.AssetDatabase(), root)
	if err != nil {
		return editor_action.Failure(err.Error())
	}
	defer vm.Close()
	ed.installLuaActionBridge(vm)
	if err := vm.DoStringNamed(code, name); err != nil {
		return editor_action.Failure(err.Error())
	}
	return editor_action.Success("lua script completed")
}

func (ed *Editor) installLuaActionBridge(vm *plugins.LuaVM) {
	vm.SetGlobalGoFunction("__kaiju_editor_action_run", func(state *lua.State) int {
		req := editor_action.Request{
			ID:     editor_action.ActionID(luaStringArg(state, 1)),
			Params: luaStringArg(state, 2),
			Source: editor_action.SourceLua,
		}
		result := ed.Actions().Run(req)
		data, _ := json.Marshal(result)
		state.PushString(string(data))
		return 1
	})
	vm.SetGlobalGoFunction("__kaiju_editor_action_can_run", func(state *lua.State) int {
		req := editor_action.Request{
			ID:     editor_action.ActionID(luaStringArg(state, 1)),
			Params: luaStringArg(state, 2),
			Source: editor_action.SourceLua,
		}
		result := ed.Actions().CanRun(req)
		state.PushBoolean(result.OK)
		if result.OK {
			state.PushString(result.Message)
		} else {
			state.PushString(result.Error)
		}
		return 2
	})
	vm.SetGlobalGoFunction("__kaiju_editor_action_list", func(state *lua.State) int {
		entries := ed.Actions().Search(luaStringArg(state, 1))
		data, _ := json.Marshal(entries)
		state.PushString(string(data))
		return 1
	})
	_ = vm.DoStringNamed(`
actions = {
	run = function(id, argsJson)
		return __kaiju_editor_action_run(id, argsJson or "")
	end,
	can_run = function(id, argsJson)
		return __kaiju_editor_action_can_run(id, argsJson or "")
	end,
	list = function(query)
		return __kaiju_editor_action_list(query or "")
	end
}`, "editor actions lua bridge")
}

func luaStringArg(state *lua.State, idx int) string {
	if state.Top() < idx || state.IsNil(idx) {
		return ""
	}
	return state.ToString(idx)
}
