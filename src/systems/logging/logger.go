/*****************************************************************************/
/* logger.go                                                                 */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package logging

import (
	"kaiju/klib"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

type LogStream struct {
	OnInfo  Event
	OnWarn  TracedEvent
	OnError TracedEvent
}

func (l *LogStream) Write(p []byte) (n int, err error) {
	str := string(p)
	levelOffset := 21
	if strings.HasPrefix(str, "time=") {
		levelOffset = 41
	}
	level := str[levelOffset:]
	if strings.HasPrefix(level, "INFO") {
		l.OnInfo.Execute(str)
	} else if strings.HasPrefix(level, "WARN") {
		l.OnWarn.Execute(str, klib.TraceStrings(str, 7))
	} else if strings.HasPrefix(level, "ERROR") {
		l.OnError.Execute(str, klib.TraceStrings(str, 7))
	}
	os.Stdout.WriteString(str)
	return len(p), nil
}

func Initialize(opts *slog.HandlerOptions) *LogStream {
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
