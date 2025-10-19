package status_bar

import (
	"kaiju/editor/common_interfaces"
	"kaiju/editor/editor_logging"
	"kaiju/engine"
	"kaiju/engine/systems/logging"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
	"regexp"
	"strconv"
	"strings"
	"weak"
)

type StatusBar struct {
	doc              *document.Document
	msg              *document.Element
	log              *document.Element
	logPopup         *document.Element
	logEntryTemplate *document.Element
	uiMan            ui.Manager
	logging          *editor_logging.Logging
	outerInterface   common_interfaces.Focusable
	inPopup          bool
}

func (b *StatusBar) Initialize(host *engine.Host, logging *editor_logging.Logging, outerInterface common_interfaces.Focusable) error {
	defer tracing.NewRegion("StatusBar.Initialize").End()
	b.logging = logging
	b.outerInterface = outerInterface
	b.uiMan.Init(host)
	var err error
	b.doc, err = markup.DocumentFromHTMLAsset(&b.uiMan, "editor/ui/global/status_bar.go.html",
		nil, map[string]func(*document.Element){
			"openLogWindow": b.openLogWindow,
			"closePopup":    b.closePopup,
		})
	b.setupUIReferences()
	b.bindToSlog(host)
	return err
}

func (b *StatusBar) Focus() { b.uiMan.EnableUpdate() }
func (b *StatusBar) Blur() {
	if b.inPopup {
		return
	}
	b.uiMan.DisableUpdate()
}

func (b *StatusBar) setupUIReferences() {
	defer tracing.NewRegion("StatusBar.setupUIReferences").End()
	b.msg, _ = b.doc.GetElementById("msg")
	b.log, _ = b.doc.GetElementById("log")
	b.logPopup, _ = b.doc.GetElementById("logPopup")
	b.logEntryTemplate, _ = b.doc.GetElementById("logEntryTemplate")
	b.logPopup.UI.Hide()
}

func (b *StatusBar) bindToSlog(host *engine.Host) {
	defer tracing.NewRegion("StatusBar.bindToSlog").End()
	wb := weak.Make(b)
	infoEvtId := host.LogStream.OnInfo.Add(func(msg string) {
		bar := wb.Value()
		if bar == nil {
			return
		}
		bar.doc.SetElementClasses(bar.log, "")
		bar.setLog(msg)
	})
	warnEvtId := host.LogStream.OnWarn.Add(func(msg string, _ []string) {
		bar := wb.Value()
		if bar == nil {
			return
		}
		bar.doc.SetElementClasses(bar.log, "statusLogWarn")
		bar.setLog(msg)
	})
	errEvtId := host.LogStream.OnError.Add(func(msg string, _ []string) {
		bar := wb.Value()
		if bar == nil {
			return
		}
		bar.doc.SetElementClasses(bar.log, "statusLogError")
		bar.setLog(msg)
	})
	logStream := host.LogStream
	host.OnClose.Add(func() {
		logStream.OnInfo.Remove(infoEvtId)
		logStream.OnWarn.Remove(warnEvtId)
		logStream.OnError.Remove(errEvtId)
	})
}

func (b *StatusBar) setLog(msg string) {
	defer tracing.NewRegion("StatusBar.setLog").End()
	res := logging.ToMap(msg)
	lbl := b.log.Children[0].UI.ToLabel()
	if m, ok := res["msg"]; ok {
		lbl.SetText(m)
	} else {
		lbl.SetText(msg)
	}
}

func (b *StatusBar) SetMessage(status string) {
	defer tracing.NewRegion("StatusBar.SetMessage").End()
	lbl := b.msg.Children[0].UI.ToLabel()
	t := lbl.Text()
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
	lbl.SetText(status)
}

func (b *StatusBar) openLogWindow(*document.Element) {
	defer tracing.NewRegion("StatusBar.openLogWindow").End()
	if b.inPopup {
		return
	}
	b.inPopup = true
	b.logPopup.UI.Show()
	b.logEntryTemplate.UI.Hide()
	for i := len(b.logPopup.Children) - 1; i >= 1; i-- {
		b.doc.RemoveElement(b.logPopup.Children[i])
	}
	all := b.logging.All()
	elms := b.doc.DuplicateElementRepeat(b.logEntryTemplate, len(all))
	for i := range elms {
		elms[i].Children[0].UI.ToLabel().SetText(all[i].Time + ": " + all[i].Message)
		b.doc.SetElementClassesWithoutApply(elms[i], "logLine", all[i].Category)
	}
	b.doc.ApplyStyles()
	pnl := b.logPopup.UI.ToPanel()
	b.uiMan.Host.RunAfterFrames(2, func() {
		pnl.SetScrollY(pnl.MaxScroll().Y())
	})
	b.outerInterface.BlurInterface()
}

func (b *StatusBar) closePopup(*document.Element) {
	b.logPopup.UI.Hide()
	b.outerInterface.FocusInterface()
	b.inPopup = false
}
