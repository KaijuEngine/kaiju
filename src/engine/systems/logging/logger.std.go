//go:build !android

package logging

func ExtPlatformLogVerbose(message string) {}
func ExtPlatformLogInfo(message string)    {}
func ExtPlatformLogWarn(message string)    {}
func ExtPlatformLogError(message string)   {}
