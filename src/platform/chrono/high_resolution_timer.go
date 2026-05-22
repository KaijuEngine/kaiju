/******************************************************************************/
/* high_resolution_timer.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package chrono

func (t *HighResolutionTimer) Start()                  { t.start() }
func (t *HighResolutionTimer) Stop() (seconds float64) { return t.stop() }
