//go:build !ai_driver

/******************************************************************************/
/* bootstrap_aidriver_stub.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package bootstrap

import "kaijuengine.com/engine"

// startAIDriver is a no-op in default builds. Build with `-tags ai_driver` to
// compile in the control server (bootstrap_aidriver.go).
func startAIDriver(host *engine.Host) {}
