package engine

// outline.go provides AddOutline / OutlineUp / AddBookmark helpers that wrap
// the PreparedPages outline and bookmark collections.

// AddOutline adds an outline entry at the current outline position with the
// current page and Y position. After adding, the outline cursor descends into
// the new entry (use OutlineUp to return to the parent level).
func (e *ReportEngine) AddOutline(text string) {
	if e.preparedPages == nil {
		return
	}
	e.preparedPages.Outline.Add(text, e.preparedPages.CurPage(), e.curY)
}

// OutlineRoot resets the outline cursor to the root level.
func (e *ReportEngine) OutlineRoot() {
	if e.preparedPages == nil {
		return
	}
	e.preparedPages.Outline.LevelRoot()
}

// OutlineUp moves the outline cursor one level up toward the root.
func (e *ReportEngine) OutlineUp() {
	if e.preparedPages == nil {
		return
	}
	e.preparedPages.Outline.LevelUp()
}

// AddBookmark adds a named navigation bookmark at the current page and Y position.
func (e *ReportEngine) AddBookmark(name string) {
	if e.preparedPages == nil || name == "" {
		return
	}
	e.preparedPages.AddBookmark(name, e.curY)
}

// GetBookmarkPage returns the 1-based page number of the named bookmark, or 0
// if no bookmark with that name has been registered.
func (e *ReportEngine) GetBookmarkPage(name string) int {
	if e.preparedPages == nil {
		return 0
	}
	return e.preparedPages.Bookmarks.GetPageNo(name)
}
