package rendering

import "testing"

func TestInstanceBufferCapacityGrowsGeometrically(t *testing.T) {
	var capacity InstanceBufferCapacity
	next, changed := capacity.Ensure(1)
	if !changed || next != minInstanceBufferCapacity {
		t.Fatalf("first capacity = %d/%v, want %d/true", next, changed, minInstanceBufferCapacity)
	}

	next, changed = capacity.Ensure(4)
	if changed || next != minInstanceBufferCapacity {
		t.Fatalf("capacity should not grow at current capacity: %d/%v", next, changed)
	}

	next, changed = capacity.Ensure(5)
	if !changed || next != 8 {
		t.Fatalf("capacity should grow geometrically to 8: %d/%v", next, changed)
	}
}

func TestInstanceBufferCapacityDoesNotShrink(t *testing.T) {
	var capacity InstanceBufferCapacity
	capacity.Ensure(9)
	next, changed := capacity.Ensure(3)
	if changed || next != 16 {
		t.Fatalf("capacity shrank or changed unexpectedly: %d/%v", next, changed)
	}
}

func TestInstanceBufferCapacityNextDoesNotCommit(t *testing.T) {
	var capacity InstanceBufferCapacity
	next, changed := capacity.Next(7)
	if !changed || next != 8 {
		t.Fatalf("next capacity = %d/%v, want 8/true", next, changed)
	}
	if capacity.Capacity() != 0 {
		t.Fatalf("Next committed capacity = %d, want 0", capacity.Capacity())
	}
	capacity.Commit(next)
	if capacity.Capacity() != 8 {
		t.Fatalf("committed capacity = %d, want 8", capacity.Capacity())
	}
}
