/******************************************************************************/
/* steam_api.go                                                               */
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

package steam

/*
#cgo CXXFLAGS: -std=c++11
#cgo windows LDFLAGS: -L../../libs -lsteam_api64 -lstdc++
#cgo steamdeck LDFLAGS: -L../../libs -lsteam_api -lstdc++ -Wl,-rpath=../../libs
#cgo linux LDFLAGS: -L../../libs -lsteam_api -lstdc++ -Wl,-rpath=../../libs
#include "steam_wrapper.h"

#cgo noescape   c_SteamAPI_Init
#cgo nocallback c_SteamAPI_Init
#cgo noescape   c_SteamAPI_Shutdown
#cgo nocallback c_SteamAPI_Shutdown
#cgo noescape   c_SteamAPI_RestartAppIfNecessary
#cgo nocallback c_SteamAPI_RestartAppIfNecessary
#cgo noescape   c_SteamAPI_RunCallbacks
#cgo noescape   c_SteamAPI_SteamFriends_GetPersonalName
#cgo nocallback c_SteamAPI_SteamFriends_GetPersonalName
#cgo noescape   c_SteamUser_BLoggedOn
#cgo nocallback c_SteamUser_BLoggedOn
#cgo noescape   c_SteamUserStats_RequestCurrentStats
#cgo nocallback c_SteamUserStats_RequestCurrentStats
#cgo noescape   c_SteamUtils_GetAppID
#cgo nocallback c_SteamUtils_GetAppID

*/
import "C"
import (
	"log/slog"
)

var (
	initialized    = false
	SteamFriends   steamFriends
	SteamUser      steamUser
	SteamUserStats steamUserStats
	SteamUtils     steamUtils
	Callbacks      steamCallbacks
)

type steamFriends struct{}
type steamUser struct{}
type steamUserStats struct{}
type steamUtils struct{}

type steamCallbacks struct {
	OnOverlayActivated  func(bool)
	OnUserStatsReceived func(gameId uint64, resultCode ResultCode)
	OnUserStatsStored   func()
}

func init() {
	Callbacks.OnOverlayActivated = func(b bool) {}
	Callbacks.OnUserStatsReceived = func(gameId uint64, resultCode ResultCode) {}
	Callbacks.OnUserStatsStored = func() {}
}

func IsInitialized() bool { return initialized }

func Initialize() {
	if IsInitialized() {
		return
	}
	if bool(C.c_SteamAPI_Init()) {
		initialized = true
	} else {
		slog.Error(`Failed to initialize the Steam API, possible reasons are:
- The Steam client isn't running
- The Steam client couldn't determine the App ID of game (check steam_appid.txt)
- Not running under same OS user context
- AppID not owned by the currently logged in Steam Account`)
	}
}

func Shutdown() {
	if !IsInitialized() {
		return
	}
	C.c_SteamAPI_Shutdown()
	initialized = false
}

func RestartAppIfNecessary(unOwnAppID uint32) bool {
	return bool(C.c_SteamAPI_RestartAppIfNecessary(C.uint32_t(unOwnAppID)))
}

func RunCallbacks() { C.c_SteamAPI_RunCallbacks() }

////////////////////////////////////////////////////////////////////////////////
// Steam Friends                                                              //
////////////////////////////////////////////////////////////////////////////////

func (s steamFriends) GetPersonalName() string {
	if !initialized {
		return ""
	}
	nameCStr := C.c_SteamAPI_SteamFriends_GetPersonalName()
	return C.GoString(nameCStr)
}

////////////////////////////////////////////////////////////////////////////////
// Steam User                                                                 //
////////////////////////////////////////////////////////////////////////////////

func (s steamUser) IsLoggedOn() bool {
	if !initialized {
		return false
	}
	return bool(C.c_SteamUser_BLoggedOn())
}

////////////////////////////////////////////////////////////////////////////////
// Steam User Stats                                                           //
////////////////////////////////////////////////////////////////////////////////

func (s steamUserStats) RequestCurrentStats() bool {
	if !initialized {
		return false
	}
	return bool(C.c_SteamUserStats_RequestCurrentStats())
}

////////////////////////////////////////////////////////////////////////////////
// Steam Utils                                                                //
////////////////////////////////////////////////////////////////////////////////

func (s steamUtils) GetAppID() int64 {
	if !initialized {
		return 0
	}
	return int64(C.c_SteamUtils_GetAppID())
}

////////////////////////////////////////////////////////////////////////////////
// Steam Callbacks                                                            //
////////////////////////////////////////////////////////////////////////////////

//export goOnGameOverlayActivated
func goOnGameOverlayActivated(active C.bool) {
	Callbacks.OnOverlayActivated(bool(active))
}

//export goOnUserStatsReceived
func goOnUserStatsReceived(gameId C.uint64_t, resultCode C.int) {
	Callbacks.OnUserStatsReceived(uint64(gameId), ResultCode(resultCode))
}

//export goOnUserStatsStored
func goOnUserStatsStored() {
	Callbacks.OnUserStatsStored()
}
