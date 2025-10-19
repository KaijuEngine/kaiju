//go:build !steam

package bootstrap

import "kaiju/engine"

func initExternalGameService()                         {}
func initExternalGameServiceRuntime(host *engine.Host) {}
func terminateExternalGameService()                    {}
