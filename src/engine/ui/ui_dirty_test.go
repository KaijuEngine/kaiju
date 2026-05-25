package ui

import (
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func newDirtyTestPanel(renderCount *int) *UI {
	u := &UI{
		elmType: ElementTypePanel,
		elmData: &panelData{},
		shaderData: &ShaderData{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Scissor: matrix.Vec4{
				-matrix.FloatMax,
				-matrix.FloatMax,
				matrix.FloatMax,
				matrix.FloatMax,
			},
		},
	}
	u.entity.Init(nil)
	u.entity.AddNamedData(EntityDataName, u)
	u.layout.initialize(u)
	u.postLayoutUpdate = func() {}
	u.render = func() {
		if renderCount != nil {
			*renderCount = *renderCount + 1
		}
	}
	return u
}

func parentDirtyTestChild(parent, child *UI) {
	child.entity.SetParent(&parent.entity)
}

func TestScopedCleanRenderDirtySkipsUnrelatedBranches(t *testing.T) {
	rootCount := 0
	leftCount := 0
	rightCount := 0
	leafCount := 0
	root := newDirtyTestPanel(&rootCount)
	left := newDirtyTestPanel(&leftCount)
	right := newDirtyTestPanel(&rightCount)
	leaf := newDirtyTestPanel(&leafCount)
	parentDirtyTestChild(root, left)
	parentDirtyTestChild(root, right)
	parentDirtyTestChild(left, leaf)

	leaf.SetDirty(DirtyTypeColorChange)
	root.cleanScopedIfNeeded()

	if leafCount != 1 {
		t.Fatalf("expected dirty leaf to render once, got %d", leafCount)
	}
	if leftCount != 0 || rightCount != 0 || rootCount != 0 {
		t.Fatalf("render-only dirty should not render ancestors or siblings; root=%d left=%d right=%d",
			rootCount, leftCount, rightCount)
	}
	if root.hasDirty() || left.hasDirty() || right.hasDirty() || leaf.hasDirty() {
		t.Fatalf("expected scoped clean to clear all dirty flags")
	}
}

func TestScopedCleanLayoutDirtyUsesNearestFixedParent(t *testing.T) {
	rootCount := 0
	leftCount := 0
	rightCount := 0
	leafCount := 0
	root := newDirtyTestPanel(&rootCount)
	left := newDirtyTestPanel(&leftCount)
	right := newDirtyTestPanel(&rightCount)
	leaf := newDirtyTestPanel(&leafCount)
	parentDirtyTestChild(root, left)
	parentDirtyTestChild(root, right)
	parentDirtyTestChild(left, leaf)

	leaf.SetDirty(DirtyTypeLayout)

	if (left.dirtyFlags & uiDirtyLayoutChildren) == 0 {
		t.Fatalf("expected direct parent to be marked for child layout")
	}
	if (root.dirtyFlags & uiDirtyLayoutChildren) != 0 {
		t.Fatalf("fixed-size ancestor should not be marked for child layout")
	}

	root.cleanScopedIfNeeded()

	if leftCount != 1 || leafCount != 1 {
		t.Fatalf("expected nearest layout scope to render parent and child once; left=%d leaf=%d",
			leftCount, leafCount)
	}
	if rootCount != 0 || rightCount != 0 {
		t.Fatalf("layout dirty in one branch should not render root or sibling branch; root=%d right=%d",
			rootCount, rightCount)
	}
}

func TestScopedCleanLayoutDirtyReachesFitContentAncestor(t *testing.T) {
	root := newDirtyTestPanel(nil)
	left := newDirtyTestPanel(nil)
	right := newDirtyTestPanel(nil)
	leaf := newDirtyTestPanel(nil)
	left.ToPanel().PanelData().fitContent = ContentFitBoth
	parentDirtyTestChild(root, left)
	parentDirtyTestChild(root, right)
	parentDirtyTestChild(left, leaf)

	leaf.SetDirty(DirtyTypeLayout)

	if (root.dirtyFlags & uiDirtyLayoutChildren) == 0 {
		t.Fatalf("fit-content parent should propagate child layout dirtiness to its parent")
	}
}

func TestScopedCleanLayoutDirtyBubblesThroughControlInternals(t *testing.T) {
	root := newDirtyTestPanel(nil)
	textarea := newDirtyTestPanel(nil)
	content := newDirtyTestPanel(nil)
	leaf := newDirtyTestPanel(nil)
	textarea.elmType = ElementTypeTextArea
	textarea.elmData = &textareaData{}
	parentDirtyTestChild(root, textarea)
	parentDirtyTestChild(textarea, content)
	parentDirtyTestChild(content, leaf)

	leaf.SetDirty(DirtyTypeLayout)

	if (content.dirtyFlags & uiDirtyLayoutChildren) == 0 {
		t.Fatalf("expected direct internal parent to be marked for child layout")
	}
	if (textarea.dirtyFlags & uiDirtyLayoutChildren) == 0 {
		t.Fatalf("expected control owner to be marked when an internal child changes layout")
	}
	if (root.dirtyFlags & uiDirtyLayoutChildren) != 0 {
		t.Fatalf("fixed parent above control owner should not be marked for child layout")
	}
}

func TestScopedCleanScissorDirtyIncludesDescendants(t *testing.T) {
	rootCount := 0
	leftCount := 0
	rightCount := 0
	leafCount := 0
	root := newDirtyTestPanel(&rootCount)
	left := newDirtyTestPanel(&leftCount)
	right := newDirtyTestPanel(&rightCount)
	leaf := newDirtyTestPanel(&leafCount)
	parentDirtyTestChild(root, left)
	parentDirtyTestChild(root, right)
	parentDirtyTestChild(left, leaf)

	left.SetDirty(DirtyTypeScissor)
	root.cleanScopedIfNeeded()

	if leftCount != 1 || leafCount != 1 {
		t.Fatalf("expected scissor dirty scope to render parent and descendants; left=%d leaf=%d",
			leftCount, leafCount)
	}
	if rootCount != 0 || rightCount != 0 {
		t.Fatalf("scissor dirty branch should not render root or unrelated sibling; root=%d right=%d",
			rootCount, rightCount)
	}
}

func TestCleanFullKeepsExplicitFullTreeFallback(t *testing.T) {
	leftCount := 0
	leafCount := 0
	grandchildCount := 0
	container := newDirtyTestPanel(nil)
	left := newDirtyTestPanel(&leftCount)
	leaf := newDirtyTestPanel(&leafCount)
	grandchild := newDirtyTestPanel(&grandchildCount)
	parentDirtyTestChild(container, left)
	parentDirtyTestChild(left, leaf)
	parentDirtyTestChild(leaf, grandchild)

	grandchild.SetDirty(DirtyTypeColorChange)
	left.cleanFull()

	if leftCount != 1 || leafCount != 1 || grandchildCount != 1 {
		t.Fatalf("expected full clean fallback to render entire requested tree; left=%d leaf=%d grandchild=%d",
			leftCount, leafCount, grandchildCount)
	}
}
