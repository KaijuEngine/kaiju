package rendering

import "testing"

func TestQueuedCommandSubmitterCountsStages(t *testing.T) {
	device := &GPUDevice{}
	device.Painter.writtenCommands = []CommandRecorder{
		{stage: 1},
		{stage: 0},
		{stage: 1},
		{stage: 99},
	}
	submitter := device.queuedCommandSubmitter()

	if submitter.QueuedCommandCount() != 4 {
		t.Fatalf("queued command count = %d, want 4", submitter.QueuedCommandCount())
	}
	counts := submitter.stageCommandCounts()
	if counts[0] != 1 || counts[1] != 3 {
		t.Fatalf("stage counts = %v, want [1 3]", counts)
	}
	if last := lastQueuedCommandStage(counts); last != 1 {
		t.Fatalf("last queued command stage = %d, want 1", last)
	}
}

func TestQueuedCommandSubmitterReadbackNoopsWithoutCommands(t *testing.T) {
	device := &GPUDevice{}
	if !device.queuedCommandSubmitter().SubmitAndWaitForReadback() {
		t.Fatalf("readback submitter should succeed with no queued commands")
	}
	if !device.FlushForReadback() {
		t.Fatalf("FlushForReadback should succeed with no queued commands")
	}
}
