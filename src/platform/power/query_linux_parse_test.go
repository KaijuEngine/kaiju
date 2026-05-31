/******************************************************************************/
/* query_linux_parse_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package power

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxPowerSupplyACOnline(t *testing.T) {
	root := t.TempDir()
	writeLinuxPowerSupply(t, root, "AC", map[string]string{
		"type":   "Mains",
		"online": "1",
	})
	writeLinuxPowerSupply(t, root, "BAT0", map[string]string{
		"type":     "Battery",
		"status":   "Charging",
		"capacity": "79",
	})
	status, err := queryLinuxPowerSupply(root)
	if err != nil {
		t.Fatal(err)
	}
	if status.Source != SourceAC || !status.HasBattery || status.BatteryPercent != 79 {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestLinuxPowerSupplyBatteryDischarging(t *testing.T) {
	root := t.TempDir()
	writeLinuxPowerSupply(t, root, "AC", map[string]string{
		"type":   "Mains",
		"online": "0",
	})
	writeLinuxPowerSupply(t, root, "BAT0", map[string]string{
		"type":     "Battery",
		"status":   "Discharging",
		"capacity": "42",
	})
	status, err := queryLinuxPowerSupply(root)
	if err != nil {
		t.Fatal(err)
	}
	if status.Source != SourceBattery || !status.HasBattery || status.BatteryPercent != 42 {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestLinuxPowerSupplyBatteryChargingWithoutMains(t *testing.T) {
	root := t.TempDir()
	writeLinuxPowerSupply(t, root, "BAT0", map[string]string{
		"type":     "Battery",
		"status":   "Charging",
		"capacity": "54",
	})
	status, err := queryLinuxPowerSupply(root)
	if err != nil {
		t.Fatal(err)
	}
	if status.Source != SourceAC || !status.HasBattery || status.BatteryPercent != 54 {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestLinuxPowerSupplyMissingData(t *testing.T) {
	status, err := queryLinuxPowerSupply(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatal(err)
	}
	if status.Source != SourceUnknown || status.HasBattery || status.BatteryPercent != -1 {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func writeLinuxPowerSupply(t *testing.T, root, name string, values map[string]string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	for file, value := range values {
		if err := os.WriteFile(filepath.Join(path, file), []byte(value), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
