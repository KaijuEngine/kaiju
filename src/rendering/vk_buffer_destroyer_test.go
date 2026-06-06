/******************************************************************************/
/* vk_buffer_destroyer_test.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"fmt"
	"testing"
	"unsafe"
)

var bufferDestroyerTestHandles [32]byte

type recordingBufferTrashReleaser struct {
	events []string
}

func (r *recordingBufferTrashReleaser) FreeDescriptorSets(pool GPUDescriptorPool, sets []GPUDescriptorSet) {
	r.events = append(r.events, fmt.Sprintf("free-descriptor-sets:%d", len(sets)))
}

func (r *recordingBufferTrashReleaser) DestroyBuffer(GPUBuffer) {
	r.events = append(r.events, "destroy-buffer")
}

func (r *recordingBufferTrashReleaser) FreeMemory(GPUDeviceMemory) {
	r.events = append(r.events, "free-memory")
}

func (r *recordingBufferTrashReleaser) RemoveDebug(unsafe.Pointer) {
	r.events = append(r.events, "remove-debug")
}

func TestBufferDestroyerFreesDescriptorSetsBeforeBuffers(t *testing.T) {
	recorder := &recordingBufferTrashReleaser{}
	trash := bufferTrash{
		delay: 1,
		pool:  testDescriptorPoolHandle(0),
	}
	trash.sets[0] = testDescriptorSetHandle(1)
	trash.sets[3] = testDescriptorSetHandle(2)
	trash.buffers[0] = testBufferHandle(3)
	trash.memories[0] = testMemoryHandle(4)
	trash.namedBuffers[0] = []GPUBuffer{testBufferHandle(5)}
	trash.namedMemories[0] = []GPUDeviceMemory{testMemoryHandle(6)}

	destroyer := bufferDestroyer{releaser: recorder}
	destroyer.Add(trash)
	destroyer.Cycle()

	if len(destroyer.trash) != 0 {
		t.Fatalf("released trash remained queued")
	}
	if len(recorder.events) == 0 {
		t.Fatalf("expected resources to be released")
	}
	if recorder.events[0] != "free-descriptor-sets:2" {
		t.Fatalf("first release event = %q, want descriptor sets before buffers; events=%v",
			recorder.events[0], recorder.events)
	}
	firstBuffer := indexString(recorder.events, "destroy-buffer")
	if firstBuffer < 0 {
		t.Fatalf("expected buffer destruction event, got %v", recorder.events)
	}
	if firstBuffer == 0 {
		t.Fatalf("buffer was destroyed before descriptor sets were freed: %v", recorder.events)
	}
}

func TestBufferDestroyerHonorsDelayBeforeReleasing(t *testing.T) {
	recorder := &recordingBufferTrashReleaser{}
	trash := bufferTrash{
		delay:   2,
		pool:    testDescriptorPoolHandle(7),
		buffers: [maxFramesInFlight]GPUBuffer{testBufferHandle(8)},
	}
	trash.sets[0] = testDescriptorSetHandle(9)

	destroyer := bufferDestroyer{releaser: recorder}
	destroyer.Add(trash)
	destroyer.Cycle()

	if len(recorder.events) != 0 {
		t.Fatalf("resources released before delay elapsed: %v", recorder.events)
	}
	if len(destroyer.trash) != 1 {
		t.Fatalf("trash removed before delay elapsed")
	}

	destroyer.Cycle()
	if len(recorder.events) == 0 {
		t.Fatalf("resources were not released after delay elapsed")
	}
}

func testDescriptorPoolHandle(index int) GPUDescriptorPool {
	return GPUDescriptorPool{GPUHandle{handle: testBufferDestroyerHandle(index)}}
}

func testDescriptorSetHandle(index int) GPUDescriptorSet {
	return GPUDescriptorSet{GPUHandle{handle: testBufferDestroyerHandle(index)}}
}

func testDescriptorSetLayoutHandle(index int) GPUDescriptorSetLayout {
	return GPUDescriptorSetLayout{GPUHandle{handle: testBufferDestroyerHandle(index)}}
}

func testBufferHandle(index int) GPUBuffer {
	return GPUBuffer{GPUHandle{handle: testBufferDestroyerHandle(index)}}
}

func testMemoryHandle(index int) GPUDeviceMemory {
	return GPUDeviceMemory{GPUHandle{handle: testBufferDestroyerHandle(index)}}
}

func testBufferDestroyerHandle(index int) unsafe.Pointer {
	return unsafe.Pointer(&bufferDestroyerTestHandles[index])
}

func indexString(values []string, target string) int {
	for i := range values {
		if values[i] == target {
			return i
		}
	}
	return -1
}
