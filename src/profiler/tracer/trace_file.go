package tracer

import (
	"fmt"
	"kaiju/klib"
	"kaiju/markup/document"
	"os"
	"strings"
	"unsafe"
)

type TraceFile struct {
	keys         map[uint16]EntryKey
	values       []EntryValue
	frames       []int
	maxTime      float64
	maxFrameTime float64
}

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
	other        int
	timelineSpec *document.Element
	time         float64
	y            float32
	isStart      bool
	rendered     bool
	veryTiny     bool
}

type traceFileRead struct {
	time float64
	key  uint16
}

func (f *TraceFile) selectChildCalls(from int) []int {
	children := make([]int, 0)
	to := f.values[from].other
	for i := from + 1; i < to; i++ {
		children = append(children, i)
		i = f.values[i].other
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
				from.other = i
				to.other = start
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
		f.values[0].other = -1
		startTime = r.time
	}
	for i := 1; i < len(f.values); i++ {
		klib.BinaryRead(fp, &r.time)
		klib.BinaryRead(fp, &r.key)
		f.values[i].key = r.key
		f.values[i].other = -1
		f.values[i].time = r.time - startTime
	}
	for i := range uint32(len(f.values)) {
		if f.values[i].other == -1 {
			v := &f.values[i]
			f.findEndValue(int(i))
			o := &f.values[v.other]
			f.maxTime = max(f.maxTime, o.time-v.time)
			if v.isStart && f.keys[v.key].funcHash == frameKeyHash {
				f.maxFrameTime = max(f.maxFrameTime, o.time-v.time)
			}
		}
	}
	return nil
}
