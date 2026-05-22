//go:build !steam

/******************************************************************************/
/* bootstrap_without_external_game_service.go                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package bootstrap

import "kaijuengine.com/engine"

func initExternalGameService()                         {}
func initExternalGameServiceRuntime(host *engine.Host) {}
func terminateExternalGameService()                    {}
