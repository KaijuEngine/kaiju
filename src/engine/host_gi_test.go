package engine

import "testing"

func TestHostBuildsGIFrameGraphPerView(t *testing.T) {
	host := NewHost("test", nil, nil)
	views := host.RenderViews.FrameViews()
	schedule, context := host.captureGlobalIlluminationFrame(views, 2)
	if len(schedule.Passes) != 2 {
		t.Fatalf("GI frame graph pass count = %d, want null resolve + composite", len(schedule.Passes))
	}
	if err := schedule.Execute(context); err != nil {
		t.Fatal(err)
	}
	if _, ok := context.Values["gi.null.diffuse.1"]; !ok {
		t.Fatalf("null GI output missing: %#v", context.Values)
	}
}
