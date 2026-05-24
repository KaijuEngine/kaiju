/******************************************************************************/
/* lua.go                                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package lua

/*
//#cgo noescape lua_pushinteger
//#cgo noescape m_lua_tointeger

#cgo noescape m_lua_pop
#cgo noescape m_lua_tostring
#cgo noescape m_luaL_loadfile
#cgo noescape m_luaL_loadstring
#cgo noescape m_lua_pcall
#cgo noescape m_lua_isboolean
#cgo noescape m_lua_islightuserdata
#cgo noescape m_lua_istable
#cgo noescape m_lua_isfunction
#cgo noescape m_lua_tonumber
#cgo noescape lua_toboolean
#cgo noescape lua_isstring
#cgo noescape lua_isnumber
#cgo noescape lua_pushstring
#cgo noescape lua_pushnumber
#cgo noescape lua_pushboolean
#cgo noescape lua_pushvalue
#cgo noescape m_lua_remove
#cgo noescape m_lua_rawseti
#cgo noescape lua_setglobal
#cgo noescape lua_setfield
#cgo noescape lua_createtable
#cgo noescape lua_gettop
#cgo noescape lua_getglobal
#cgo noescape lua_getfield
#cgo noescape luaL_openlibs
#cgo noescape luaL_newstate
#cgo noescape lua_pushlightuserdata
#cgo noescape lua_close

#cgo linux CFLAGS: -DLUA_USE_LINUX
#cgo linux LDFLAGS: -lm -ldl
#cgo darwin CFLAGS: -DLUA_USE_MACOSX
#cgo windows CFLAGS: -DLUA_USE_WINDOWS

#include <stdlib.h>
#include <stdbool.h>

#define LUA_USE_LONGJMP 1

#include "lua.h"
#include "lualib.h"
#include "lauxlib.h"

extern int cCallGoFunc(int id, lua_State* L);

static void m_lua_pop(lua_State *L, int n) {
	lua_settop(L, -(n)-1);
}

static void m_lua_remove(lua_State *L, int idx) {
	lua_remove(L, idx);
}

static void m_lua_rawseti(lua_State *L, int idx, lua_Integer n) {
	lua_rawseti(L, idx, n);
}

static const char* m_lua_tostring(lua_State* L, int i) {
	return lua_tolstring(L, (i), NULL);
}

static int m_luaL_loadfile(lua_State* L, const char* s) {
	return luaL_loadfile(L, s);
}

static int m_luaL_loadstring(lua_State* L, const char* s) {
	return luaL_loadstring(L, s);
}

static int traceback_handler(lua_State* L) {
	const char* msg = lua_tostring(L, 1);
	if (msg) {
		luaL_traceback(L, L, msg, 1);
	} else {
		lua_pushliteral(L, "(error object is not a string)");
	}
	return 1;
}

static int m_lua_pcall(lua_State* L, int n, int r) {
	int base = lua_gettop(L) - n;
	lua_pushcfunction(L, traceback_handler);
	lua_insert(L, base);
	int status = lua_pcallk(L, n, r, base, 0, NULL);
	lua_remove(L, base);
	return status;
}

static int m_lua_isboolean(lua_State* L, int n) {
	return (lua_type(L, (n)) == LUA_TBOOLEAN);
}

static int m_lua_islightuserdata(lua_State* L, int n) {
	return (lua_type(L, (n)) == LUA_TLIGHTUSERDATA);
}

static int m_lua_istable(lua_State* L, int n) {
	return (lua_type(L, (n)) == LUA_TTABLE);
}

static int m_lua_isfunction(lua_State* L, int n) {
	return (lua_type(L, (n)) == LUA_TFUNCTION);
}

static double m_lua_tonumber(lua_State* L, int i) {
	return lua_tonumberx(L,(i),NULL);
}

static double m_lua_tointeger(lua_State* L, int i) {
	return lua_tointegerx(L,(i),NULL);
}

static int go_func_wrapper(lua_State* L) {
    if (!lua_isinteger(L, lua_upvalueindex(1))) {
        lua_pushstring(L, "invalid upvalue");
        lua_error(L);
        return 0;
    }
    int id = lua_tointeger(L, lua_upvalueindex(1));
    return cCallGoFunc(id, L);
}

static void push_go_function(lua_State* L, int id) {
    lua_pushinteger(L, id);
    lua_pushcclosure(L, go_func_wrapper, 1);
}
*/
import "C"
import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"unsafe"
)

const (
	errorOK     = 0
	multiReturn = -1
)

var (
	vms = map[*C.lua_State]*State{}
)

type pinnedPointer struct {
	pinner  *runtime.Pinner
	pointer any
}

type State struct {
	state      *C.lua_State
	pinned     map[unsafe.Pointer]pinnedPointer
	funcs      map[int]func(state *State) int
	nextFuncId int
	err        error
}

func pinPointer(ptr any) pinnedPointer {
	pp := pinnedPointer{
		pinner:  new(runtime.Pinner),
		pointer: ptr,
	}
	pp.pinner.Pin(ptr)
	// I don't believe I need to pin any innter-pointers as only the top
	// level pointer is accessed when coming into Go from lua
	return pp
}

func New() State {
	return State{
		state:      C.luaL_newstate(),
		nextFuncId: 1,
		pinned:     make(map[unsafe.Pointer]pinnedPointer),
		funcs:      make(map[int]func(state *State) int),
	}
}

func (l *State) OpenLibraries() error {
	C.luaL_openlibs(l.state)
	if l.state == nil {
		return errors.New("failed to open the libraries for lua")
	}
	vms[l.state] = l

	// Sandbox the Lua environment by removing dangerous libraries
	// that allow file system access and command execution
	l.disableDangerousLibraries()

	return nil
}

func (l *State) Close() {
	if l.state == nil {
		return
	}
	for ptr, pp := range l.pinned {
		pp.pinner.Unpin()
		delete(l.pinned, ptr)
	}
	state := l.state
	C.lua_close(l.state)
	delete(vms, state)
	l.state = nil
	l.pinned = nil
	l.funcs = nil
}

// disableDangerousLibraries removes dangerous libraries from the Lua state
// to prevent execution of arbitrary commands and file system access
func (l *State) disableDangerousLibraries() {
	// Use DoString to safely remove dangerous global functions
	// This is safer than direct C API calls

	// Remove os library - prevents OS command execution and file system access
	l.DoString("os = nil")

	// Remove io library - prevents arbitrary file I/O operations
	l.DoString("io = nil")

	// Remove debug library - prevents debugging and introspection attacks
	l.DoString("debug = nil")

	// Remove package library - prevents module loading attacks
	l.DoString("package = nil")
	l.DoString("require = function(_) return true end")

	// Remove load and loadstring functions - prevents dynamic code loading
	l.DoString("load = nil")
	l.DoString("loadstring = nil")

	// Remove dofile function - prevents loading and executing arbitrary files
	l.DoString("dofile = nil")
}

func (l *State) Pop(idx int) {
	C.m_lua_pop(l.state, C.int(idx))
}

func (l *State) DoFile(file string) error {
	cStr := C.CString(file)
	defer C.free(unsafe.Pointer(cStr))
	if C.m_luaL_loadfile(l.state, cStr) == errorOK {
		return l.ProtectedCall(0, multiReturn)
	} else {
		return l.stackErrorf("failed to load the lua file: %s", file)
	}
}

func (l *State) DoString(code string) error {
	cStr := C.CString(code)
	defer C.free(unsafe.Pointer(cStr))
	if C.m_luaL_loadstring(l.state, cStr) != errorOK {
		return l.stackErrorf("failed to load the code string")
	}
	return l.ProtectedCall(0, multiReturn)
}

func (l *State) stackErrorf(format string, args ...any) error {
	errText := ""
	if l.Top() > 0 {
		errText = C.GoString(C.m_lua_tostring(l.state, C.int(-1)))
		l.Pop(1)
	}
	if errText == "" {
		return fmt.Errorf(format, args...)
	}
	return fmt.Errorf("%s: %s", fmt.Sprintf(format, args...), errText)
}

func (l *State) Field(idx int, name string) {
	cStr := C.CString(name)
	defer C.free(unsafe.Pointer(cStr))
	C.lua_getfield(l.state, C.int(idx), cStr)
}

func (l *State) Global(name string) {
	cStr := C.CString(name)
	defer C.free(unsafe.Pointer(cStr))
	C.lua_getglobal(l.state, cStr)
}

func (l *State) Top() int {
	return int(C.lua_gettop(l.state))
}

func (l *State) NewTable() {
	C.lua_createtable(l.state, 0, 0)
}

func (l *State) CreateTable(arrLen, fieldLen int) {
	C.lua_createtable(l.state, C.int(arrLen), C.int(fieldLen))
}

func (l *State) SetField(idx int, name string) {
	cStr := C.CString(name)
	defer C.free(unsafe.Pointer(cStr))
	C.lua_setfield(l.state, C.int(idx), cStr)
}

func (l *State) RawSetInt(idx, n int) {
	C.m_lua_rawseti(l.state, C.int(idx), C.lua_Integer(n))
}

func (l *State) SetGlobal(name string) {
	cStr := C.CString(name)
	defer C.free(unsafe.Pointer(cStr))
	C.lua_setglobal(l.state, cStr)
}

func (l *State) PushValue(idx int) {
	C.lua_pushvalue(l.state, C.int(idx))
}

func (l *State) Remove(idx int) {
	C.m_lua_remove(l.state, C.int(idx))
}

func (l *State) PushBoolean(value bool) {
	if value {
		C.lua_pushboolean(l.state, 1)
	} else {
		C.lua_pushboolean(l.state, 0)
	}
}

func (l *State) PushNumber(value float64) {
	C.lua_pushnumber(l.state, C.lua_Number(value))
}

func (l *State) PushString(value string) {
	cStr := C.CString(value)
	defer C.free(unsafe.Pointer(cStr))
	C.lua_pushstring(l.state, cStr)
}

func (l *State) Error(message string) int {
	l.err = errors.New(message)
	l.PushString(message)
	return 1
}

func (l *State) PushUserData(value reflect.Value) {
	p := unsafe.Pointer(value.Pointer())
	if _, ok := l.pinned[p]; !ok {
		l.pinned[p] = pinPointer(value.Interface())
	}
	C.lua_pushlightuserdata(l.state, p)
}

func (l *State) ToUserData(idx int) any {
	ptr := unsafe.Pointer(C.lua_touserdata(l.state, C.int(idx)))
	pp := l.pinned[ptr]
	return pp.pointer
}

func (l *State) RemovePinnedPointer(idx int) {
	// Pointers are pinned for the VM lifetime. Multiple Lua wrapper tables can
	// refer to the same Go object, so table finalizers must not unpin eagerly.
}

func (l *State) PushGoFunction(fn func(state *State) int) {
	id := l.nextFuncId
	l.nextFuncId++
	l.funcs[id] = fn
	C.push_go_function(l.state, C.int(id))
}

//export cCallGoFunc
func cCallGoFunc(id C.int, L *C.lua_State) C.int {
	l := vms[L]
	fn, ok := l.funcs[int(id)]
	if !ok {
		slog.Error("go function with not found", "id", int(id))
		l.err = errors.New("Go function not found")
		return 0
	}
	return C.int(fn(l))
}

func (l *State) IsBoolean(idx int) bool {
	return C.m_lua_isboolean(l.state, C.int(idx)) == 1
}

func (l *State) IsNumber(idx int) bool {
	return C.lua_isnumber(l.state, C.int(idx)) == 1
}

func (l *State) IsString(idx int) bool {
	return C.lua_isstring(l.state, C.int(idx)) == 1
}

func (l *State) IsTable(idx int) bool {
	return C.m_lua_istable(l.state, C.int(idx)) == 1
}

func (l *State) IsFunction(idx int) bool {
	return C.m_lua_isfunction(l.state, C.int(idx)) == 1
}

func (l *State) IsUserData(idx int) bool {
	return C.m_lua_islightuserdata(l.state, C.int(idx)) == 1
}

func (l *State) ToBoolean(idx int) bool {
	return C.lua_toboolean(l.state, C.int(idx)) == 1
}

func (l *State) ToNumber(idx int) float64 {
	return float64(C.m_lua_tonumber(l.state, C.int(idx)))
}

func (l *State) ToString(idx int) string {
	return C.GoString(C.m_lua_tostring(l.state, C.int(idx)))
}

func (l *State) ProtectedCall(args, returns int) error {
	l.err = nil
	if C.m_lua_pcall(l.state, C.int(args), C.int(returns)) != errorOK {
		return l.stackErrorf("failed to execute lua")
	}
	if l.err != nil {
		return l.err
	}
	return nil
}

func (l *State) Call(args, returns int) {
	if err := l.ProtectedCall(args, returns); err != nil {
		panic(err)
	}
}
