package visuals2d

import (
	"encoding/json"
	"kaiju/klib"
	"strconv"
	"strings"
	"unicode"
)

type spriteSheetFrameDataRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type spriteSheetFrameDataSize struct {
	W int `json:"w"`
	H int `json:"h"`
}

type spriteSheetFrameDataPivot struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type spriteSheetFrameData struct {
	Frame            spriteSheetFrameDataRect  `json:"frame"`
	Rotated          bool                      `json:"rotated"`
	Trimmed          bool                      `json:"trimmed"`
	SpriteSourceSize spriteSheetFrameDataRect  `json:"spriteSourceSize"`
	SourceSize       spriteSheetFrameDataSize  `json:"sourceSize"`
	Pivot            spriteSheetFrameDataPivot `json:"pivot"`
}

type spriteSheetData struct {
	ClipStart int                             `json:"clipStart"`
	MirrorX   bool                            `json:"mirrorX"`
	Frames    map[string]spriteSheetFrameData `json:"frames"`
}

type spriteSheet struct {
	data  spriteSheetData
	clips map[string][]spriteSheetFrameData
}

func ReadSpriteSheetData(jsonStr string) (spriteSheet, error) {
	var data spriteSheetData
	err := klib.JsonDecode(json.NewDecoder(strings.NewReader(jsonStr)), &data)
	sheet := spriteSheet{
		data:  data,
		clips: make(map[string][]spriteSheetFrameData),
	}
	if err == nil {
		for k, v := range data.Frames {
			k = strings.TrimSuffix(k, ".png")
			parts := strings.Split(k, "_")
			last := parts[len(parts)-1]
			isClip := true
			for _, r := range last {
				isClip = isClip && unicode.IsDigit(r)
			}
			if isClip {
				idx, _ := strconv.Atoi(last)
				idx -= data.ClipStart
				clipName := strings.Join(parts[:len(parts)-1], "_")
				if _, ok := sheet.clips[clipName]; !ok {
					sheet.clips[clipName] = make([]spriteSheetFrameData, 0)
				}
				for len(sheet.clips[clipName]) <= idx {
					sheet.clips[clipName] = append(sheet.clips[clipName], spriteSheetFrameData{})
				}
				sheet.clips[clipName][idx] = v
			} else {
				sheet.clips[k] = []spriteSheetFrameData{v}
			}
		}
	}
	if sheet.data.MirrorX {
		for k, v := range sheet.clips {
			if strings.HasSuffix(k, "left") {
				cpy := make([]spriteSheetFrameData, len(v))
				for i := range v {
					cpy[i] = v[i]
					cpy[i].Frame.X += cpy[i].Frame.W
					cpy[i].Frame.W *= -1
				}
				sheet.clips[strings.TrimSuffix(k, "left")+"right"] = cpy
			}
		}
	}
	return sheet, err
}