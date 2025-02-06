package context_menu

type ContextMenuSet struct {
	ctxMenu *ContextMenu
	Entries []ContextMenuEntry
}

func NewSet(ctxMenu *ContextMenu, entries []ContextMenuEntry) ContextMenuSet {
	return ContextMenuSet{
		ctxMenu: ctxMenu,
		Entries: entries,
	}
}

func (s *ContextMenuSet) Show() { s.ctxMenu.Show(s.Entries) }

func (s *ContextMenuSet) ShowWithTarget(target any) {
	s.ctxMenu.Show(s.Entries)
}
