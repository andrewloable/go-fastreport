package preview

// PageCache is an LRU cache of PreparedPage objects for large reports.
// It keeps at most Limit pages in memory; Get() fetches pages from
// the backing PreparedPages and evicts the least-recently-used page when
// the cache is full.
//
// It is the Go equivalent of FastReport.Preview.PageCache.
//
// Usage:
//
//	cache := preview.NewPageCache(preparedPages, 50)
//	page := cache.Get(42)   // loads page 42 from preparedPages if not cached
type PageCache struct {
	// pp is the backing PreparedPages store.
	pp *PreparedPages
	// Limit is the maximum number of pages to keep in the cache (default 50).
	Limit int
	// items holds cached entries in most-recently-used order (index 0 = MRU).
	items []cacheItem
}

// cacheItem stores one cached page.
type cacheItem struct {
	index int
	page  *PreparedPage
}

// NewPageCache creates a PageCache backed by pp with the given capacity limit.
// If limit <= 0 the default of 50 is used (matching the C# implementation).
func NewPageCache(pp *PreparedPages, limit int) *PageCache {
	if limit <= 0 {
		limit = 50
	}
	return &PageCache{pp: pp, Limit: limit}
}

// Get returns the PreparedPage at index. If it is already in the cache the
// entry is promoted to MRU. Otherwise the page is fetched from the backing
// PreparedPages and inserted at the front; the LRU entry is evicted if the
// cache is at capacity.
// Returns nil if index is out of range.
func (c *PageCache) Get(index int) *PreparedPage {
	// Search for an existing cached entry.
	for i, item := range c.items {
		if item.index == index {
			if i != 0 {
				// Promote to MRU (move to front).
				copy(c.items[1:], c.items[:i])
				c.items[0] = item
			}
			return item.page
		}
	}

	// Not in cache: fetch from PreparedPages.
	page := c.pp.GetPage(index)
	if page == nil {
		return nil
	}

	// Only cache when total pages > current cache size (matches C# condition).
	if c.pp.Count() <= len(c.items) {
		return page
	}

	// Insert at front.
	c.items = append(c.items, cacheItem{}) // grow slice
	copy(c.items[1:], c.items)
	c.items[0] = cacheItem{index: index, page: page}

	// Evict LRU entries beyond capacity.
	if len(c.items) > c.Limit {
		c.items = c.items[:c.Limit]
	}

	return page
}

// Remove evicts the entry for index from the cache, if present.
func (c *PageCache) Remove(index int) {
	for i, item := range c.items {
		if item.index == index {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return
		}
	}
}

// Clear removes all entries from the cache.
func (c *PageCache) Clear() {
	c.items = c.items[:0]
}

// Len returns the number of pages currently held in the cache.
func (c *PageCache) Len() int { return len(c.items) }
