package preview

import (
	"testing"
)

func buildPP(n int) *PreparedPages {
	pp := New()
	for i := 0; i < n; i++ {
		pp.AddPage(595, 842, i+1)
	}
	return pp
}

func TestNewPageCache_DefaultLimit(t *testing.T) {
	pp := buildPP(5)
	c := NewPageCache(pp, 0) // 0 → default 50
	if c.Limit != 50 {
		t.Errorf("Limit = %d, want 50", c.Limit)
	}
}

func TestNewPageCache_CustomLimit(t *testing.T) {
	pp := buildPP(5)
	c := NewPageCache(pp, 10)
	if c.Limit != 10 {
		t.Errorf("Limit = %d, want 10", c.Limit)
	}
}

func TestPageCache_Get_CachesMRU(t *testing.T) {
	pp := buildPP(5)
	c := NewPageCache(pp, 50)

	// Initial Get — fetches from pp.
	p0 := c.Get(0)
	if p0 == nil {
		t.Fatal("Get(0) returned nil")
	}
	if c.Len() != 1 {
		t.Errorf("Len = %d, want 1 after first get", c.Len())
	}

	// Get again — should be a cache hit (no change in Len).
	p0Again := c.Get(0)
	if p0Again != p0 {
		t.Error("second Get(0) should return same page")
	}
}

func TestPageCache_Get_PromotesToMRU(t *testing.T) {
	pp := buildPP(5)
	c := NewPageCache(pp, 50)

	// Load pages 0,1,2 into cache.
	c.Get(0)
	c.Get(1)
	c.Get(2)

	// Get(0) again — should promote index 0 to front.
	c.Get(0)
	if c.items[0].index != 0 {
		t.Errorf("MRU item index = %d, want 0", c.items[0].index)
	}
}

func TestPageCache_Get_OutOfRange(t *testing.T) {
	pp := buildPP(3)
	c := NewPageCache(pp, 50)
	if got := c.Get(99); got != nil {
		t.Error("Get(99) should return nil for out-of-range index")
	}
}

func TestPageCache_Get_EvictsLRU(t *testing.T) {
	pp := buildPP(5)
	c := NewPageCache(pp, 3) // cache holds at most 3

	c.Get(0)
	c.Get(1)
	c.Get(2)
	c.Get(3) // should evict LRU (0)

	if c.Len() > 3 {
		t.Errorf("Len = %d, want <= 3", c.Len())
	}
}

func TestPageCache_Get_SmallPP_NoCache(t *testing.T) {
	// If pp.Count() <= len(items), don't add to cache.
	pp := buildPP(1)
	c := NewPageCache(pp, 50)
	c.Get(0)         // first get: adds to cache
	_ = c.Get(0)     // second: cache hit, still 1 item
	if c.Len() > 1 {
		t.Errorf("Len = %d, want 1", c.Len())
	}
}

func TestPageCache_Get_SmallPP_ForcedNoCache(t *testing.T) {
	// Force pp.Count() <= len(c.items) by directly populating items with stale entries.
	// This covers the "don't add to cache" branch when cache has >= pages count entries.
	pp := buildPP(1)
	c := NewPageCache(pp, 50)
	// Inject 2 stale items so len(items)=2 while pp.Count()=1 → 1 <= 2
	fakePage := &PreparedPage{PageNo: 99}
	c.items = []cacheItem{{index: 5, page: fakePage}, {index: 6, page: fakePage}}
	p := c.Get(0)
	if p == nil {
		t.Fatal("Get(0) should return valid page")
	}
	if c.Len() != 2 {
		t.Errorf("Len = %d, want 2 (no new entry added)", c.Len())
	}
}

func TestPageCache_Remove(t *testing.T) {
	pp := buildPP(3)
	c := NewPageCache(pp, 50)
	c.Get(0)
	c.Get(1)
	if c.Len() != 2 {
		t.Fatalf("before Remove: Len = %d, want 2", c.Len())
	}
	c.Remove(0)
	if c.Len() != 1 {
		t.Errorf("after Remove(0): Len = %d, want 1", c.Len())
	}
	// Remove of missing index should be a no-op.
	c.Remove(99)
	if c.Len() != 1 {
		t.Errorf("after Remove(missing): Len = %d, want 1", c.Len())
	}
}

func TestPageCache_Clear(t *testing.T) {
	pp := buildPP(3)
	c := NewPageCache(pp, 50)
	c.Get(0)
	c.Get(1)
	c.Clear()
	if c.Len() != 0 {
		t.Errorf("after Clear: Len = %d, want 0", c.Len())
	}
}
