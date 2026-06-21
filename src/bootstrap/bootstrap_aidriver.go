//go:build ai_driver

/******************************************************************************/
/* bootstrap_aidriver.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package bootstrap

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/aidriver"
)

// startAIDriver brings up the localhost AI-driver control server. This variant
// is only compiled under `-tags ai_driver`; see bootstrap_aidriver_stub.go for
// the default no-op.
func startAIDriver(host *engine.Host) {
	aidriver.Start(host)
}
