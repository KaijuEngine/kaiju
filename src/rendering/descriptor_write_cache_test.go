package rendering

import "testing"

func TestDescriptorWriteCacheSkipsUnchangedSignature(t *testing.T) {
	var cache DescriptorWriteCache
	signature := NewDescriptorWriteSignature()
	signature.AddUint64(10)
	signature.AddString("material")

	if !cache.ShouldWrite(0, signature) {
		t.Fatalf("first descriptor signature should write")
	}
	if cache.ShouldWrite(0, signature) {
		t.Fatalf("unchanged descriptor signature should not write")
	}

	changed := signature
	changed.AddUint64(20)
	if !cache.ShouldWrite(0, changed) {
		t.Fatalf("changed descriptor signature should write")
	}
}

func TestDescriptorWriteCacheInvalidates(t *testing.T) {
	var cache DescriptorWriteCache
	signature := NewDescriptorWriteSignature()
	signature.AddUint64(10)
	cache.ShouldWrite(1, signature)

	cache.Invalidate()
	if !cache.ShouldWrite(1, signature) {
		t.Fatalf("invalidated descriptor signature should write again")
	}
}
