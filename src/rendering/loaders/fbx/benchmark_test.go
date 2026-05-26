/******************************************************************************/
/* benchmark_test.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"os"
	"testing"
)

var benchmarkDocument Document
var benchmarkMeshCount int

func BenchmarkParseMonkeyFBX(b *testing.B) {
	data := readBenchmarkMonkeyFBX(b)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		doc, err := Parse(data)
		if err != nil {
			b.Fatalf("Parse(monkey.fbx) returned error: %v", err)
		}
		benchmarkDocument = doc
	}
}

func BenchmarkImportMonkeyFBX(b *testing.B) {
	data := readBenchmarkMonkeyFBX(b)
	doc, err := Parse(data)
	if err != nil {
		b.Fatalf("Parse(monkey.fbx) returned error: %v", err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := ToLoadResult(doc)
		if err != nil {
			b.Fatalf("ToLoadResult(monkey.fbx) returned error: %v", err)
		}
		benchmarkMeshCount = len(res.Meshes)
	}
}

func readBenchmarkMonkeyFBX(b *testing.B) []byte {
	b.Helper()
	data, err := os.ReadFile("../../../editor/editor_embedded_content/editor_content/meshes/monkey.fbx")
	if err != nil {
		b.Skipf("monkey fixture not available: %v", err)
	}
	return data
}
