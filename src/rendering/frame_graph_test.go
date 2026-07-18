package rendering

import (
	"errors"
	"reflect"
	"testing"
)

func TestFrameGraphCompileTracksAllResourceHazards(t *testing.T) {
	graph := NewFrameGraph()
	image, err := graph.AddResource(FrameGraphResourceDescription{Name: "lighting", Kind: FrameGraphResourceImage})
	if err != nil {
		t.Fatal(err)
	}
	writeA, _ := graph.AddPass(FrameGraphPassDescription{Name: "write-a", Uses: []FrameGraphResourceUse{{Resource: image, Access: FrameGraphAccessWrite}}})
	readB, _ := graph.AddPass(FrameGraphPassDescription{Name: "read-b", Uses: []FrameGraphResourceUse{{Resource: image, Access: FrameGraphAccessRead}}})
	readC, _ := graph.AddPass(FrameGraphPassDescription{Name: "read-c", Uses: []FrameGraphResourceUse{{Resource: image, Access: FrameGraphAccessRead}}})
	writeD, _ := graph.AddPass(FrameGraphPassDescription{Name: "write-d", Uses: []FrameGraphResourceUse{{Resource: image, Access: FrameGraphAccessWrite}}})

	schedule, err := graph.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if got := schedule.Passes[1].Dependencies; !reflect.DeepEqual(got, []FrameGraphPassID{writeA}) {
		t.Fatalf("read-b dependencies = %v", got)
	}
	if got := schedule.Passes[2].Dependencies; !reflect.DeepEqual(got, []FrameGraphPassID{writeA}) {
		t.Fatalf("read-c dependencies = %v", got)
	}
	if got := schedule.Passes[3].Dependencies; !reflect.DeepEqual(got, []FrameGraphPassID{writeA, readB, readC}) {
		t.Fatalf("write-d dependencies = %v", got)
	}
	if schedule.Passes[3].ID != writeD {
		t.Fatalf("write-d id = %d, want %d", schedule.Passes[3].ID, writeD)
	}
}

func TestFrameGraphCompileReadWriteDependsOnPreviousWriter(t *testing.T) {
	graph := NewFrameGraph()
	buffer, _ := graph.AddResource(FrameGraphResourceDescription{Name: "history", Kind: FrameGraphResourceBuffer})
	first, _ := graph.AddPass(FrameGraphPassDescription{Name: "seed", Uses: []FrameGraphResourceUse{{Resource: buffer, Access: FrameGraphAccessWrite}}})
	graph.AddPass(FrameGraphPassDescription{Name: "accumulate", Uses: []FrameGraphResourceUse{{Resource: buffer, Access: FrameGraphAccessReadWrite}}})
	schedule, err := graph.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if got := schedule.Passes[1].Dependencies; !reflect.DeepEqual(got, []FrameGraphPassID{first}) {
		t.Fatalf("dependencies = %v", got)
	}
}

func TestFrameGraphRejectsInvalidDeclarations(t *testing.T) {
	graph := NewFrameGraph()
	if _, err := graph.AddResource(FrameGraphResourceDescription{}); err == nil {
		t.Fatal("expected empty resource name error")
	}
	resource, _ := graph.AddResource(FrameGraphResourceDescription{Name: "depth"})
	if _, err := graph.AddResource(FrameGraphResourceDescription{Name: "depth"}); err == nil {
		t.Fatal("expected duplicate resource error")
	}
	if _, err := graph.AddPass(FrameGraphPassDescription{Name: "bad", Uses: []FrameGraphResourceUse{
		{Resource: resource, Access: FrameGraphAccessRead},
		{Resource: resource, Access: FrameGraphAccessWrite},
	}}); err == nil {
		t.Fatal("expected conflicting access error")
	}
}

func TestFrameGraphExecuteWrapsPassError(t *testing.T) {
	graph := NewFrameGraph()
	graph.AddPass(FrameGraphPassDescription{Name: "failure", Execute: func(*FrameGraphExecutionContext) error {
		return errors.New("boom")
	}})
	schedule, err := graph.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if err := schedule.Execute(nil); err == nil || err.Error() != `frame graph pass "failure" failed: boom` {
		t.Fatalf("unexpected error: %v", err)
	}
}
