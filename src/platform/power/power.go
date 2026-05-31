/******************************************************************************/
/* power.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package power

// Source is the currently detected system power source.
type Source int

const (
	SourceUnknown Source = iota
	SourceAC
	SourceBattery
)

// Status is a snapshot of the system power state.
type Status struct {
	Source         Source
	HasBattery     bool
	BatteryPercent int
}

// OnBattery returns true when the machine is currently running on battery.
func (s Status) OnBattery() bool { return s.Source == SourceBattery }
