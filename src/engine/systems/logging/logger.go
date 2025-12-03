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
	"bufio"
	"fmt"
	"kaiju/build"
	"kaiju/klib"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	maxLogFileHistory = 10
	logFileName       = "kaiju.log"
)

type LogStream struct {
	OnInfo  Event
	OnWarn  TracedEvent
	OnError TracedEvent
	File    *os.File
}

func Initialize(opts *slog.HandlerOptions) *LogStream {
	defer tracing.NewRegion("logging.Initialize").End()
	stream := &LogStream{
		OnInfo:  newEvent(),
		OnWarn:  newTracedEvent(),
		OnError: newTracedEvent(),
	}
	setupErr := setupLogHistory(stream)
	slog.SetDefault(slog.New(newLogHandler(stream, opts)))
	if setupErr != nil {
		slog.Error("failed to setup the log file", "error", setupErr)
	}
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

func (l *LogStream) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for i := range lines {
		l.writeLine(lines[i])
	}
	return len(p), nil
}

func (l *LogStream) Close() {
	if l.File != nil {
		l.File.Close()
	}
}

func setupLogHistory(l *LogStream) error {
	logsDir, err := selectLogsFolder()
	if err != nil {
		return err
	}
	logFilePath := filepath.Join(logsDir, logFileName)
	if s, ok := logAlreadyExists(logFilePath); ok {
		renameOldLogFile(logFilePath, s, logsDir)
	}
	entries, err := os.ReadDir(logsDir)
	cleanupOldLogs(err, entries, logsDir)
	if f, err := os.Create(logFilePath); err == nil {
		l.File = f
	}
	return nil
}

func LogFolderPath() (string, error) {
	appData, err := filesystem.GameDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(appData, "logs"), nil
}

func selectLogsFolder() (string, error) {
	dir, err := LogFolderPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func logAlreadyExists(logFilePath string) (os.FileInfo, bool) {
	if s, err := os.Stat(logFilePath); err == nil && s.Size() > 0 {
		return s, true
	} else {
		return s, false
	}
}

func cleanupOldLogs(err error, entries []os.DirEntry, logsDir string) {
	if err == nil {
		var rotated []os.DirEntry
		for _, e := range entries {
			name := e.Name()
			if strings.HasPrefix(name, "kaiju-") && strings.HasSuffix(name, ".log") {
				rotated = append(rotated, e)
			}
		}
		if len(rotated) >= 10 {
			sort.Slice(rotated, func(i, j int) bool {
				ei, _ := rotated[i].Info()
				ej, _ := rotated[j].Info()
				return ei.ModTime().Before(ej.ModTime())
			})
			for len(rotated) >= maxLogFileHistory {
				oldest := rotated[0]
				_ = os.Remove(filepath.Join(logsDir, oldest.Name()))
				rotated = rotated[1:]
			}
		}
	}
}

func renameOldLogFile(logFilePath string, s os.FileInfo, logsDir string) {
	f, err := os.Open(logFilePath)
	if err == nil {
		scanner := bufio.NewScanner(f)
		firstLine := ""
		for scanner.Scan() {
			t := strings.TrimSpace(scanner.Text())
			if t != "" {
				firstLine = t
				break
			}
		}
		f.Close()
		if firstLine != "" {
			mapping := ToMap(firstLine)
			timeStr := mapping["time"]
			if timeStr == "" {
				timeStr = s.ModTime().UTC().Format(time.RFC3339)
			}
			sanitized := sanitizeForFilename(timeStr)
			newName := fmt.Sprintf("kaiju-%s.log", sanitized)
			newPath := filepath.Join(logsDir, newName)
			if _, err := os.Stat(newPath); err == nil {
				newPath = filepath.Join(logsDir, fmt.Sprintf("kaiju-%s-%d.log", sanitized, time.Now().Unix()))
			}
			_ = os.Rename(logFilePath, newPath)
		}
	}
}

func sanitizeForFilename(s string) string {
	// replace common separators and remove problematic chars
	s = strings.Trim(s, " \t\n\r\"'")
	s = strings.ReplaceAll(s, ":", "-")
	s = strings.ReplaceAll(s, " ", "T")
	// collapse any remaining non-alnum._- into _
	re := regexp.MustCompile(`[^0-9A-Za-z._-]`)
	s = re.ReplaceAllString(s, "_")
	return s
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
	level := line[strings.Index(line, "level=")+len("level="):]
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
	if l.File != nil {
		l.File.WriteString(line + "\n")
	}
}
