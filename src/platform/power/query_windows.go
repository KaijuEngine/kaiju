//go:build windows

/******************************************************************************/
/* query_windows.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package power

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	windowsACOffline       = 0
	windowsACOnline        = 1
	windowsACUnknown       = 255
	windowsNoSystemBattery = 128
	windowsBatteryUnknown  = 255
)

type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

var getSystemPowerStatusProc = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetSystemPowerStatus")

func Query() (Status, error) {
	status := systemPowerStatus{}
	ret, _, err := getSystemPowerStatusProc.Call(uintptr(unsafe.Pointer(&status)))
	if ret == 0 {
		return Status{Source: SourceUnknown, BatteryPercent: -1}, err
	}
	out := Status{
		Source:         SourceUnknown,
		HasBattery:     status.BatteryFlag != windowsNoSystemBattery && status.BatteryFlag != windowsBatteryUnknown,
		BatteryPercent: int(status.BatteryLifePercent),
	}
	if status.BatteryLifePercent == windowsBatteryUnknown {
		out.BatteryPercent = -1
	}
	switch status.ACLineStatus {
	case windowsACOffline:
		out.Source = SourceBattery
	case windowsACOnline:
		out.Source = SourceAC
	case windowsACUnknown:
		out.Source = SourceUnknown
	}
	return out, nil
}
