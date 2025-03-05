package tracer

import (
	_ "embed"
	"fmt"
	"kaiju/engine"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/ui"
	"math"
)

const (
	maxStackDepth        = 16
	infoPanelId          = "info"
	areaSelectId         = "areaSelect"
	templateBoxId        = "templateBox"
	templateThreadNameId = "templateThreadName"
	templateBoxInnerId   = "templateBoxInner"
	hoverInfoId          = "hoverInfo"
)

var (
	boxColors = []matrix.Color{
		matrix.NewColor(0.4, 0.0, 0.0, 1.0),    // Reddish Brown
		matrix.NewColor(0.0, 0.35, 0.0, 1.0),   // Forest Green
		matrix.NewColor(0.0, 0.0, 0.45, 1.0),   // Navy Blue
		matrix.NewColor(0.35, 0.2, 0.35, 1.0),  // Plum
		matrix.NewColor(0.45, 0.15, 0.25, 1.0), // Dark Coral
		matrix.NewColor(0.0, 0.4, 0.4, 1.0),    // Teal
		matrix.NewColor(0.45, 0.35, 0.15, 1.0), // Dark Khaki
		matrix.NewColor(0.35, 0.0, 0.35, 1.0),  // Purple
		matrix.NewColor(0.25, 0.0, 0.4, 1.0),   // Dark Violet
		matrix.NewColor(0.4, 0.0, 0.4, 1.0),    // Dark Magenta
	}
)

//go:embed tracer.html
var tracerHTML string

type TracerSelectRange struct {
	from float64
	to   float64
}

type Tracer struct {
	host           *engine.Host
	doc            *document.Document
	tracksDoc      *document.Document
	showingInfo    *ui.UI
	uiManTimeline  ui.Manager
	uiMan          ui.Manager
	file           TraceFile
	threadLabels   []*document.Element
	selectRange    TracerSelectRange
	dragStart      matrix.Vec2
	offset         matrix.Vec2
	selectStart    float32
	zoom           float32
	lastSpec       int
	renderIndex    int
	selectingRange bool
}

func (t *Tracer) findSpec(target string) *document.Element {
	s, _ := t.doc.GetElementById(target)
	return s
}

func (t *Tracer) findTrackSpec(target string) *document.Element {
	s, _ := t.tracksDoc.GetElementById(target)
	return s
}

func toDisplayTime(time float64, out *int) string {
	if time >= 1 {
		*out = (int)(time)
		return "s"
	} else if time >= 0.001+math.SmallestNonzeroFloat64 {
		*out = (int)(time * 1000)
		return "ms"
	} else if time >= 0.000001+math.SmallestNonzeroFloat64 {
		*out = (int)(time * 1000000)
		return "us"
	} else {
		*out = (int)(time * 1000000000)
		return "ns"
	}
}

func New(host *engine.Host, file string) {

}

func (t *Tracer) zoomScale() float32 {
	return matrix.Pow(t.zoom, 2.5)
}

func (t *Tracer) timelineViewWidth() float32 {
	viewWidth := float32(t.host.Window.Width())
	info := t.findSpec(infoPanelId)
	viewWidth -= info.UI.Layout().Stretch().Left()
	return viewWidth
}

func (t *Tracer) updateSelectRangeText() {
	tpl := t.findTrackSpec(templateBoxId)
	r := math.Abs(t.selectRange.from-t.selectRange.to) /
		float64(tpl.UI.Layout().PixelSize().Width())
	aCounter := 0
	bCounter := 0
	aSize := toDisplayTime(r, &aCounter)
	var bSize string
	if r >= 1 {
		bSize = toDisplayTime(r-float64(aCounter), &bCounter)
	} else if r >= 0.001+math.SmallestNonzeroFloat64 {
		bSize = toDisplayTime(((r*1000.0)-float64(aCounter))/1000.0, &bCounter)
	} else {
		bSize = toDisplayTime(((r*1000000.0)-float64(aCounter))/1000000.0, &bCounter)
	}
	time := fmt.Sprintf("%d%s %d%s", aCounter, aSize, bCounter, bSize)
	areaSelect := t.findSpec(areaSelectId).Children[0].UI.ToLabel()
	areaSelect.SetText(time)
}

func (t *Tracer) updateSelectRange() {
	from := t.selectRange.from
	to := t.selectRange.to
	if to < from {
		to, from = from, to
	}
	areaSelect := t.findSpec(areaSelectId)
	if math.Abs(from-to) <= math.SmallestNonzeroFloat64 {
		areaSelect.UI.Hide()
		return
	}
	scale := t.zoomScale()
	scaleX := float32(to-from) * scale
	xOffset := t.offset.X() + float32(from*float64(scale))
	l := areaSelect.UI.Layout()
	l.SetOffset(xOffset, l.Offset().Y())
	l.Scale(scaleX, l.PixelSize().Height())
	areaSelect.UI.Show()
}

func (t *Tracer) updateChanges() {
	tpl := t.findTrackSpec(templateBoxId)
	width := tpl.UI.Layout().PixelSize().Width()
	height := tpl.UI.Layout().PixelSize().Height()
	scale := t.zoomScale()
	widthScale := width * scale
	viewHeight := float32(t.host.Window.Height())
	viewWidth := t.timelineViewWidth()
	t.updateSelectRange()
	for i := range t.file.values {
		v := &t.file.values[i]
		if !v.isStart || v.veryTiny || !v.rendered || v.timelineSpec == nil {
			continue
		}
		ui := v.timelineSpec.UI
		if !ui.Entity().IsRoot() {
			continue
		}
		scaleX := widthScale * float32(t.file.values[v.other].time-v.time)
		l := ui.Layout()
		xOffset := float32(v.time*float64(widthScale)) + t.offset.X()
		l.SetOffset(xOffset, v.y-t.offset.Y())
		l.Scale(scaleX, height)
		xOutOfBounds := (scaleX < 1) || (xOffset+scaleX < 0) || (xOffset > viewWidth)
		yOutOfBounds := ((l.Offset().Y() + height) < 0) || l.Offset().Y() > viewHeight
		if xOutOfBounds || yOutOfBounds {
			ui.Hide()
			ui.SetDontClean(true)
		} else {
			ui.Show()
			ui.SetDontClean(false)
		}
		xOffset += scaleX
		ui.Clean()
	}
}

func (t *Tracer) averageForFunc(key uint16) float64 {
	count := 0
	time := 0.0
	for i := range len(t.file.values) {
		e := &t.file.values[i]
		if e.isStart && e.key == key {
			time += (t.file.values[e.other].time - e.time)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return time / float64(count)
}

func (t *Tracer) onEntryEnter(e *document.Element) {
	label := t.findSpec(templateBoxInnerId).Children[0].UI.ToLabel()
	info := t.findSpec(hoverInfoId)
	infoLabel := t.findSpec("info").Children[0].UI.ToLabel()
	info.UI.Show()
	t.showingInfo = e.UI
	infoLabel.SetText(label.Text())
	info.UI.Layout().ScaleWidth(t.host.FontCache().MeasureString(
		label.FontFace(), label.Text(), label.FontSize()) + 10)
}

func (t *Tracer) onEntryExit(e *document.Element) {
	if e.UI == t.showingInfo {
		info := t.findSpec(hoverInfoId)
		info.UI.Hide()
		t.showingInfo = nil
	}
}

func (t *Tracer) renderTrace(v *EntryValue, idx int, tpl *document.Element,
	zoomScale, viewWidth, y, yMax float32) bool {
	if v.rendered {
		return false
	}
	v.rendered = true
	if !v.isStart || v.veryTiny {
		return false
	}
	if y < yMax {
		var copy *document.Element
		ui_spec_clone(tpl, t.host, &copy)
		c := boxColors[t.file.keys[v.key].funcHash%uint64(len(boxColors))]
		spec_as_panel(copy.kids.data[0]).color = c
		spec_as_label(copy.kids.data[0].kids.data[0]).bgColor = c
		xOffset := float32(v.time*tpl.width) + t.offset.x
		copy.x = xOffset
		copy.y = copy.y + (y * copy.height)
		v.y = copy.y
		entryTime := v.other.time - v.time
		copy.width *= zoomScale * (float)(entryTime)
		copy.events[UI_EVENT_TYPE_ENTER] = EventEntry{
			senderCall: local_entry_enter,
			state:      t,
		}
		copy.events[UI_EVENT_TYPE_EXIT] = EventEntry{
			senderCall: local_entry_exit,
			state:      t,
		}
		// Disable anything too small to render
		copy.isDisabled = (copy.width < 1) ||
			(copy.x+copy.width < 0) || (copy.x > viewWidth)
		l := spec_as_label(copy.kids.data[0].kids.data[0])
		var buff string
		count := 0
		size := local_to_display_time(entryTime, &count)
		snprintf(buff, sizeof(buff), "%s (%d %s)", v.key.funcName, count, size)
		strclone(buff, &l.text)
		t.trackSpecs.add(copy)
		v.timelineSpec = copy
	}
	if t.uiManTimeline.pools.lastPoolId == t.uiManTimeline.pools.pools.len {
		t.uiManTimeline.reserve(0xFF * 100)
	}
	children := t.file.selectChildCalls(idx)
	for i := range len(children) {
		id := children.data[i]
		t.renderTrace(&t.file.values[id], id, tpl, zoomScale, viewWidth, y+1, yMax)
	}
	return true
}
