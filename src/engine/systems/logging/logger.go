/******************************************************************************/
/* logger.go                                                                  */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package logging

import (
	"kaiju/build"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"regexp"
	"runtime"
	"strings"
)

type LogStream struct {
	OnInfo  Event
	OnWarn  TracedEvent
	OnError TracedEvent
}

func (l *LogStream) writeLine(line string) {
	if line == "" {
		return
	}
	if !strings.HasPrefix(line, "time=") {
		ExtPlatformLogInfo(line)
		println(line)
		return
	}
	levelOffset := 21
	if strings.HasPrefix(line, "time=") {
		levelOffset = 41
	}
	level := line[levelOffset:]
	if strings.HasPrefix(level, "WARN") {
		ExtPlatformLogWarn(line)
		if !l.OnWarn.IsEmpty() {
			l.OnWarn.Execute(line, klib.TraceStrings(line, 7))
		}
	} else if strings.HasPrefix(level, "ERROR") {
		ExtPlatformLogError(line)
		if !l.OnError.IsEmpty() {
			l.OnError.Execute(line, klib.TraceStrings(line, 7))
		}
	} else {
		ExtPlatformLogInfo(line)
		l.OnInfo.Execute(line)
	}
	if build.Debug && runtime.GOOS != "android" {
		os.Stdout.WriteString(line + "\n")
	}
}

func (l *LogStream) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for i := range lines {
		l.writeLine(lines[i])
	}
	return len(p), nil
}

func Initialize(opts *slog.HandlerOptions) *LogStream {
	defer tracing.NewRegion("logging.Initialize").End()
	stream := &LogStream{
		OnInfo:  newEvent(),
		OnWarn:  newTracedEvent(),
		OnError: newTracedEvent(),
	}
	logger := slog.New(newLogHandler(stream, opts))
	slog.SetDefault(logger)
	return stream
}

func ToMap(logMessage string) map[string]string {
	mapping := make(map[string]string)
	re := regexp.MustCompile(`(\w+)=("[^"]*"|.*?)(\s|$)`)
	matches := re.FindAllStringSubmatch(logMessage, -1)
	for _, match := range matches {
		key := match[1]
		value := strings.Trim(match[2], `"`)
		mapping[key] = value
	}
	return mapping
}
