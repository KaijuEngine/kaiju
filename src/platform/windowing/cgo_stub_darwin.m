//go:build darwin && !ios
// This file exists only to ensure the windowing package is cgo-enabled on macOS.
#import <Cocoa/Cocoa.h>
// No symbols needed.
