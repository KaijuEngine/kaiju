package ui

import "testing"

func benchmarkDirtyPanelTree(man *Manager, width, depth int, batched bool) *UI {
	root := newDirtyTestPanelWithManager(man, nil)
	buildLevel := func(parent *UI, level int) {}
	buildLevel = func(parent *UI, level int) {
		if level >= depth {
			return
		}
		for i := 0; i < width; i++ {
			child := newDirtyTestPanelWithManager(man, nil)
			parent.ToPanel().AddChild(child)
			buildLevel(child, level+1)
		}
	}
	if batched {
		man.beginDirtyBatch()
		defer man.endDirtyBatch()
	}
	buildLevel(root, 0)
	return root
}

func BenchmarkUIDirtyPanelLabelAddChild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		man := &Manager{}
		root := newDirtyTestPanelWithManager(man, nil)
		man.beginDirtyBatch()
		for j := 0; j < 100; j++ {
			child := newDirtyTestPanelWithManager(man, nil)
			root.ToPanel().AddChild(child)
		}
		man.endDirtyBatch()
	}
}

func BenchmarkUIDirtyCompositeInputLike(b *testing.B) {
	for i := 0; i < b.N; i++ {
		man := &Manager{}
		root := newDirtyTestPanelWithManager(man, nil)
		man.beginDirtyBatch()
		label := newDirtyTestPanelWithManager(man, nil)
		placeholder := newDirtyTestPanelWithManager(man, nil)
		cursor := newDirtyTestPanelWithManager(man, nil)
		highlight := newDirtyTestPanelWithManager(man, nil)
		root.ToPanel().AddChild(label)
		root.ToPanel().AddChild(placeholder)
		root.ToPanel().AddChild(cursor)
		root.ToPanel().AddChild(highlight)
		label.SetDirty(DirtyTypeGenerated)
		placeholder.SetDirty(DirtyTypeGenerated)
		cursor.SetDirty(DirtyTypeGenerated)
		highlight.SetDirty(DirtyTypeGenerated)
		man.endDirtyBatch()
	}
}

func BenchmarkUIDirtyDocumentLikeBuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		man := &Manager{}
		benchmarkDirtyPanelTree(man, 4, 4, true)
	}
}

func BenchmarkUIDirtyRepeatedDeepChildAlreadyDirty(b *testing.B) {
	man := &Manager{}
	root := benchmarkDirtyPanelTree(man, 1, 32, false)
	leaf := root
	for len(leaf.entity.Children) > 0 {
		leaf = FirstOnEntity(leaf.entity.Children[0])
	}
	leaf.SetDirty(DirtyTypeLayout)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		leaf.SetDirty(DirtyTypeLayout)
	}
}
