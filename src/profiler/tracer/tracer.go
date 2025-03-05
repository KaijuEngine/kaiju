package tracer

import (
	_ "embed"
	"fmt"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/ui"
	"math"
	"os"
	"strings"
	"unsafe"
	"weak"
)

const (
	maxStackDepth = 16
	infoPanelId   = "info"
	templateBoxId = "templateBox"
	areaSelectId  = "areaSelect"
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

type EntryKey struct {
	file        string
	funcName    string
	infoSpec    *document.Element
	funcHash    uint64
	averageTime float64
	tempTime    float64
	line        int32
	isStart     bool
	outOfRange  bool
}

type EntryValue struct {
	key          uint16
	other        weak.Pointer[EntryValue]
	timelineSpec *document.Element
	otherIdx     int
	time         float64
	y            float32
	isStart      bool
	rendered     bool
	veryTiny     bool
}

type TraceFile struct {
	keys         map[uint16]EntryKey
	values       []EntryValue
	frames       []int
	maxTime      float64
	maxFrameTime float64
}

type traceFileRead struct {
	time float64
	key  uint16
}

type TracerSelectRange struct {
	from float64
	to   float64
}

type Tracer struct {
	host           *engine.Host
	showingInfo    *ui.UI
	uiManTimeline  ui.Manager
	uiMan          ui.Manager
	file           TraceFile
	trackSpecs     []*document.Element
	specs          []*document.Element
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

func findSpec(specs []*document.Element, target string) *document.Element {
	for i := range specs {
		if specs[i].Attribute("id") == target {
			return specs[i]
		}
		if child := findSpec(specs[i].Children, target); child != nil {
			return child
		}
	}
	return nil
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
	info := findSpec(t.specs, infoPanelId)
	viewWidth -= info.UI.Layout().Stretch().Left()
	return viewWidth
}

func (f *TraceFile) selectChildCalls(from int) []int {
	children := make([]int, 0)
	to := f.values[from].otherIdx
	for i := from + 1; i < to; i++ {
		children = append(children, i)
		i = f.values[i].otherIdx
	}
	return children
}

func (f *TraceFile) findEndValue(start int) {
	depth := 0
	from := &f.values[start]
	from.isStart = true
	for i := start + 1; i < len(f.values); i++ {
		to := &f.values[i]
		if to.key == from.key {
			depth++
		} else if f.keys[from.key].funcName == f.keys[to.key].funcName {
			depth--
			if depth < 0 {
				from.other = weak.Make(to)
				from.otherIdx = i
				to.other = weak.Make(from)
				to.otherIdx = start
				to.rendered = true
				from.veryTiny = (to.time - from.time) < 1.0/1000.0/100.0
				break
			}
		}
	}
}

func (f *TraceFile) loadFile(file string) error {
	fp, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fp.Close()
	stringBlock, _ := klib.BinaryReadString(fp)
	str := stringBlock
	id := 1
	for i := range len(str) {
		k := EntryKey{}
		offset := strings.Index(str, ":")
		k.file = str[0:offset]
		offset++
		fmt.Sscanf(str[offset:], "%d", &k.line)
		offset += strings.Index(str[offset:], ":") + 1
		k.funcName = str[offset:]
		k.funcHash = klib.HashString(k.funcName)
		k.isStart = true
		// TODO:  This is a bit slow :X
		for _, val := range f.keys {
			if val.funcHash == k.funcHash {
				k.isStart = k.isStart && k.line < val.line
				if k.isStart {
					val.isStart = false
				}
			}
		}
		f.keys[uint16(id)] = k
		strLen := len(str) + 1
		str = str[strLen:]
		i += strLen
		id++
	}
	var valsLen uint64
	klib.BinaryRead(fp, &valsLen)
	r := traceFileRead{}
	f.values = make([]EntryValue,
		(valsLen / uint64(unsafe.Sizeof(r.time)+unsafe.Sizeof(r.key))))
	startTime := 0.0
	frameKeyHash := klib.HashString("main_frame")
	{
		// Read the first frame so we don't need if check every iteration
		klib.BinaryRead(fp, &r.time)
		klib.BinaryRead(fp, &r.key)
		f.values[0].key = r.key
		f.values[0].otherIdx = -1
		startTime = r.time
	}
	for i := 1; i < len(f.values); i++ {
		klib.BinaryRead(fp, &r.time)
		klib.BinaryRead(fp, &r.key)
		f.values[i].key = r.key
		f.values[i].otherIdx = -1
		f.values[i].time = r.time - startTime
	}
	for i := range uint32(len(f.values)) {
		if f.values[i].other.Value() == nil {
			v := &f.values[i]
			f.findEndValue(int(i))
			o := v.other.Value()
			f.maxTime = max(f.maxTime, o.time-v.time)
			if v.isStart && f.keys[v.key].funcHash == frameKeyHash {
				f.maxFrameTime = max(f.maxFrameTime, o.time-v.time)
			}
		}
	}
	return nil
}

func (t *Tracer) updateSelectRangeText() {
	tpl := findSpec(t.trackSpecs, templateBoxId)
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
	areaSelect := findSpec(t.specs, areaSelectId).Children[0].UI.ToLabel()
	areaSelect.SetText(time)
}
