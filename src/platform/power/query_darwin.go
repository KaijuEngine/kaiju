//go:build darwin && cgo

/******************************************************************************/
/* query_darwin.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package power

/*
#cgo darwin LDFLAGS: -framework IOKit -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/ps/IOPowerSources.h>
#include <IOKit/ps/IOPSKeys.h>

static int kaiju_power_source_status(int* hasBattery, int* onBattery, int* percent) {
	*hasBattery = 0;
	*onBattery = 0;
	*percent = -1;
	CFTypeRef info = IOPSCopyPowerSourcesInfo();
	if (!info) {
		return 0;
	}
	CFArrayRef sources = IOPSCopyPowerSourcesList(info);
	if (!sources) {
		CFRelease(info);
		return 0;
	}
	CFIndex count = CFArrayGetCount(sources);
	for (CFIndex i = 0; i < count; i++) {
		CFTypeRef source = CFArrayGetValueAtIndex(sources, i);
		CFDictionaryRef desc = IOPSGetPowerSourceDescription(info, source);
		if (!desc) {
			continue;
		}
		CFStringRef state = CFDictionaryGetValue(desc, CFSTR(kIOPSPowerSourceStateKey));
		if (state && CFGetTypeID(state) == CFStringGetTypeID()) {
			*hasBattery = 1;
			if (CFStringCompare(state, CFSTR(kIOPSBatteryPowerValue), 0) == kCFCompareEqualTo) {
				*onBattery = 1;
			}
		}
		CFNumberRef current = CFDictionaryGetValue(desc, CFSTR(kIOPSCurrentCapacityKey));
		CFNumberRef max = CFDictionaryGetValue(desc, CFSTR(kIOPSMaxCapacityKey));
		if (*percent < 0 && current && max &&
				CFGetTypeID(current) == CFNumberGetTypeID() &&
				CFGetTypeID(max) == CFNumberGetTypeID()) {
			int c = 0;
			int m = 0;
			CFNumberGetValue(current, kCFNumberIntType, &c);
			CFNumberGetValue(max, kCFNumberIntType, &m);
			if (m > 0) {
				*percent = (c * 100) / m;
			}
		}
	}
	CFRelease(sources);
	CFRelease(info);
	return 1;
}
*/
import "C"

func Query() (Status, error) {
	hasBattery := C.int(0)
	onBattery := C.int(0)
	percent := C.int(-1)
	if C.kaiju_power_source_status(&hasBattery, &onBattery, &percent) == 0 {
		return Status{Source: SourceUnknown, BatteryPercent: -1}, nil
	}
	source := SourceUnknown
	if hasBattery == 0 {
		source = SourceUnknown
	} else if onBattery != 0 {
		source = SourceBattery
	} else {
		source = SourceAC
	}
	return Status{
		Source:         source,
		HasBattery:     hasBattery != 0,
		BatteryPercent: int(percent),
	}, nil
}
