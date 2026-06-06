/******************************************************************************/
/* query_linux_parse.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package power

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func queryLinuxPowerSupply(root string) (Status, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Status{Source: SourceUnknown, BatteryPercent: -1}, nil
		}
		return Status{Source: SourceUnknown, BatteryPercent: -1}, err
	}
	out := Status{Source: SourceUnknown, BatteryPercent: -1}
	mainsOnline := false
	mainsOffline := false
	batteryDischarging := false
	batteryChargingOrFull := false
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name())
		supplyType := strings.ToLower(strings.TrimSpace(readLinuxPowerSupplyValue(path, "type")))
		switch supplyType {
		case "battery":
			out.HasBattery = true
			if out.BatteryPercent < 0 {
				if percent, ok := readLinuxPowerSupplyPercent(path); ok {
					out.BatteryPercent = percent
				}
			}
			switch strings.ToLower(strings.TrimSpace(readLinuxPowerSupplyValue(path, "status"))) {
			case "discharging":
				batteryDischarging = true
			case "charging", "full":
				batteryChargingOrFull = true
			}
		case "mains", "usb", "usb-c", "usb_c", "usb-pd", "usb_pd":
			switch strings.TrimSpace(readLinuxPowerSupplyValue(path, "online")) {
			case "1":
				mainsOnline = true
			case "0":
				mainsOffline = true
			}
		}
	}
	switch {
	case mainsOnline:
		out.Source = SourceAC
	case batteryDischarging:
		out.Source = SourceBattery
	case out.HasBattery && mainsOffline:
		out.Source = SourceBattery
	case batteryChargingOrFull:
		out.Source = SourceAC
	}
	return out, nil
}

func readLinuxPowerSupplyValue(path, name string) string {
	data, err := os.ReadFile(filepath.Join(path, name))
	if err != nil {
		return ""
	}
	return string(data)
}

func readLinuxPowerSupplyPercent(path string) (int, bool) {
	if capacity := strings.TrimSpace(readLinuxPowerSupplyValue(path, "capacity")); capacity != "" {
		if percent, err := strconv.Atoi(capacity); err == nil {
			return percent, true
		}
	}
	chargeNow, chargeNowOk := readLinuxPowerSupplyInt(path, "charge_now")
	chargeFull, chargeFullOk := readLinuxPowerSupplyInt(path, "charge_full")
	if !chargeNowOk || !chargeFullOk || chargeFull <= 0 {
		return -1, false
	}
	return int((chargeNow * 100) / chargeFull), true
}

func readLinuxPowerSupplyInt(path, name string) (int64, bool) {
	value := strings.TrimSpace(readLinuxPowerSupplyValue(path, name))
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	return parsed, err == nil
}
