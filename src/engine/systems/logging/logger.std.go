//go:build !android

/******************************************************************************/
/* logger.std.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package logging

func ExtPlatformLogVerbose(message string) {}
func ExtPlatformLogInfo(message string)    {}
func ExtPlatformLogWarn(message string)    {}
func ExtPlatformLogError(message string)   {}
