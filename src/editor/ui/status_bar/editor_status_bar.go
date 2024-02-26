package status_bar

import (
	"kaiju/editor/ui/log_window"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/systems/logging"
	"kaiju/ui"
	"regexp"
	"strconv"
	"strings"
)

type StatusBar struct {
	doc *document.Document
	msg *ui.Label
	log *ui.Label
}

func New(host *engine.Host, logWindow *log_window.LogWindow) *StatusBar {
	const html = "editor/ui/status.html"
	s := &StatusBar{}
	s.doc = klib.MustReturn(markup.DocumentFromHTMLAsset(host, html, nil,
		map[string]func(*document.DocElement){
			"openLogWindow": func(*document.DocElement) { logWindow.Show() },
		}))
	m, _ := s.doc.GetElementById("msg")
	l, _ := s.doc.GetElementById("log")
	s.msg = m.HTML.Children[0].DocumentElement.UI.(*ui.Label)
	s.log = l.HTML.Children[0].DocumentElement.UI.(*ui.Label)
	host.LogStream.OnInfo.Add(func(msg string) {
		host.RunAfterFrames(1, func() { s.setLog(msg, matrix.ColorWhite()) })
	})
	host.LogStream.OnWarn.Add(func(msg string, _ []string) {
		host.RunAfterFrames(1, func() { s.setLog(msg, matrix.ColorYellow()) })
	})
	host.LogStream.OnError.Add(func(msg string, _ []string) {
		host.RunAfterFrames(1, func() { s.setLog(msg, matrix.ColorLightCoral()) })
	})
	return s
}

func (s *StatusBar) setLog(msg string, color matrix.Color) {
	s.log.SetColor(color)
	res := logging.ToMap(msg)
	if m, ok := res["msg"]; ok {
		s.log.SetText(m)
	} else {
		s.log.SetText(msg)
	}
}

func (s *StatusBar) SetMessage(status string) {
	t := s.msg.Text()
	if strings.HasSuffix(t, status) {
		count := 1
		if strings.HasPrefix(t, "(") {
			re := regexp.MustCompile(`\((\d+)\)\s`)
			res := re.FindAllStringSubmatch(t, -1)
			if len(res) > 0 && len(res[0]) > 1 {
				count, _ = strconv.Atoi(res[0][1])
				count++
			}
		}
		status = "(" + strconv.Itoa(count) + ") " + status
	}
	s.msg.SetText(status)
}
