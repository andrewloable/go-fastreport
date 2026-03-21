package preview_test

// blob_store_test.go — tests for file-based caching BlobStore features.
//
// Covers:
//   - NewBlobStoreWithFileCache(false) — in-memory path behaves like NewBlobStore
//   - NewBlobStoreWithFileCache(true)  — file-cache path: Add/Get/Count/GetSource
//   - AddOrUpdate deduplication (in-memory and file-cache)
//   - GetStream returns a readable io.Reader
//   - GetSource returns correct source key
//   - Clear removes all items and resets count
//   - Dispose closes and deletes the temp file; BlobStore is empty afterwards
//   - Close is an alias for Dispose (satisfies io.Closer)

import (
	"bytes"
	"io"
	"testing"

	"github.com/andrewloable/go-fastreport/preview"
)

// ── NewBlobStoreWithFileCache(false) ──────────────────────────────────────────

func TestBlobStore_FileCache_False_AddGet(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(false)
	if err != nil {
		t.Fatalf("NewBlobStoreWithFileCache(false) error: %v", err)
	}
	defer bs.Dispose()

	data := []byte{10, 20, 30}
	idx := bs.Add("key1", data)
	if idx != 0 {
		t.Errorf("first Add idx = %d, want 0", idx)
	}
	got := bs.Get(0)
	if !bytes.Equal(got, data) {
		t.Errorf("Get(0) = %v, want %v", got, data)
	}
}

func TestBlobStore_FileCache_False_DuplicateKey(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer bs.Dispose()

	idx1 := bs.Add("same", []byte{1})
	idx2 := bs.Add("same", []byte{2})
	if idx1 != idx2 {
		t.Errorf("duplicate key should return same idx: got %d and %d", idx1, idx2)
	}
	if bs.Count() != 1 {
		t.Errorf("Count = %d, want 1 after duplicate Add", bs.Count())
	}
}

// ── NewBlobStoreWithFileCache(true) ───────────────────────────────────────────

func TestBlobStore_FileCache_True_AddGet(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(true)
	if err != nil {
		t.Fatalf("NewBlobStoreWithFileCache(true) error: %v", err)
	}
	defer bs.Dispose()

	data1 := []byte{1, 2, 3, 4}
	data2 := []byte{5, 6, 7, 8, 9}
	idx0 := bs.Add("", data1) // anonymous blob
	idx1 := bs.Add("", data2) // anonymous blob

	if idx0 != 0 || idx1 != 1 {
		t.Errorf("indices: got (%d,%d), want (0,1)", idx0, idx1)
	}
	if bs.Count() != 2 {
		t.Errorf("Count = %d, want 2", bs.Count())
	}

	got0 := bs.Get(0)
	if !bytes.Equal(got0, data1) {
		t.Errorf("Get(0) = %v, want %v", got0, data1)
	}
	got1 := bs.Get(1)
	if !bytes.Equal(got1, data2) {
		t.Errorf("Get(1) = %v, want %v", got1, data2)
	}
}

func TestBlobStore_FileCache_True_OutOfRange(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer bs.Dispose()

	if bs.Get(0) != nil {
		t.Error("Get(0) on empty file-cache store should return nil")
	}
	if bs.Get(-1) != nil {
		t.Error("Get(-1) should return nil")
	}
}

func TestBlobStore_FileCache_True_DuplicateKey(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer bs.Dispose()

	idx1 := bs.Add("src", []byte{1, 2})
	idx2 := bs.Add("src", []byte{3, 4})
	if idx1 != idx2 {
		t.Errorf("duplicate source should return same idx: got %d and %d", idx1, idx2)
	}
	if bs.Count() != 1 {
		t.Errorf("Count = %d, want 1", bs.Count())
	}
	// Data should be from the first Add, not the second.
	got := bs.Get(0)
	if !bytes.Equal(got, []byte{1, 2}) {
		t.Errorf("Get(0) = %v, want [1 2]", got)
	}
}

// ── AddOrUpdate ───────────────────────────────────────────────────────────────

func TestBlobStore_AddOrUpdate_InMemory(t *testing.T) {
	bs := preview.NewBlobStore()
	defer bs.Dispose()

	idx0 := bs.AddOrUpdate([]byte{1}, "alpha")
	idx1 := bs.AddOrUpdate([]byte{2}, "beta")
	idx2 := bs.AddOrUpdate([]byte{3}, "alpha") // duplicate source

	if idx0 != 0 || idx1 != 1 {
		t.Errorf("initial indices: got (%d,%d), want (0,1)", idx0, idx1)
	}
	if idx2 != 0 {
		t.Errorf("AddOrUpdate with duplicate source: got idx %d, want 0", idx2)
	}
	if bs.Count() != 2 {
		t.Errorf("Count = %d, want 2", bs.Count())
	}
}

func TestBlobStore_AddOrUpdate_EmptySrc(t *testing.T) {
	bs := preview.NewBlobStore()
	defer bs.Dispose()

	// Empty src means no deduplication — each call creates a new entry.
	idx0 := bs.AddOrUpdate([]byte{1}, "")
	idx1 := bs.AddOrUpdate([]byte{2}, "")
	if idx0 == idx1 {
		t.Errorf("empty src: both calls got same idx %d, want distinct indices", idx0)
	}
	if bs.Count() != 2 {
		t.Errorf("Count = %d, want 2", bs.Count())
	}
}

func TestBlobStore_AddOrUpdate_FileCache(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer bs.Dispose()

	idx0 := bs.AddOrUpdate([]byte{10, 20}, "imgA")
	idx1 := bs.AddOrUpdate([]byte{30, 40}, "imgA") // duplicate
	if idx0 != idx1 {
		t.Errorf("duplicate src in file-cache: got (%d,%d), want same", idx0, idx1)
	}
	if bs.Count() != 1 {
		t.Errorf("Count = %d, want 1", bs.Count())
	}
	got := bs.Get(0)
	if !bytes.Equal(got, []byte{10, 20}) {
		t.Errorf("Get(0) = %v, want [10 20]", got)
	}
}

// ── GetStream ─────────────────────────────────────────────────────────────────

func TestBlobStore_GetStream_InMemory(t *testing.T) {
	bs := preview.NewBlobStore()
	defer bs.Dispose()

	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	bs.Add("blob", data)

	r := bs.GetStream(0)
	if r == nil {
		t.Fatal("GetStream(0) returned nil")
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("GetStream content = %v, want %v", got, data)
	}
}

func TestBlobStore_GetStream_FileCache(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer bs.Dispose()

	data := []byte{0xCA, 0xFE, 0xBA, 0xBE}
	bs.Add("", data)

	r := bs.GetStream(0)
	if r == nil {
		t.Fatal("GetStream(0) returned nil for file-cache store")
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("GetStream content = %v, want %v", got, data)
	}
}

func TestBlobStore_GetStream_OutOfRange(t *testing.T) {
	bs := preview.NewBlobStore()
	if bs.GetStream(0) != nil {
		t.Error("GetStream(0) on empty store should return nil")
	}
	if bs.GetStream(-1) != nil {
		t.Error("GetStream(-1) should return nil")
	}
}

// ── GetSource ─────────────────────────────────────────────────────────────────

func TestBlobStore_GetSource_InMemory(t *testing.T) {
	bs := preview.NewBlobStore()
	defer bs.Dispose()

	bs.Add("http://example.com/img.png", []byte{1})
	bs.Add("", []byte{2}) // anonymous

	if got := bs.GetSource(0); got != "http://example.com/img.png" {
		t.Errorf("GetSource(0) = %q, want source key", got)
	}
	if got := bs.GetSource(1); got != "" {
		t.Errorf("GetSource(1) = %q, want empty string for anonymous blob", got)
	}
}

func TestBlobStore_GetSource_FileCache(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer bs.Dispose()

	bs.Add("src_key", []byte{99})
	if got := bs.GetSource(0); got != "src_key" {
		t.Errorf("GetSource(0) = %q, want %q", got, "src_key")
	}
}

func TestBlobStore_GetSource_OutOfRange(t *testing.T) {
	bs := preview.NewBlobStore()
	if got := bs.GetSource(0); got != "" {
		t.Errorf("GetSource(0) on empty store = %q, want empty string", got)
	}
	if got := bs.GetSource(-1); got != "" {
		t.Errorf("GetSource(-1) = %q, want empty string", got)
	}
}

// ── Clear ─────────────────────────────────────────────────────────────────────

func TestBlobStore_Clear_InMemory(t *testing.T) {
	bs := preview.NewBlobStore()
	bs.Add("x", []byte{1, 2, 3})
	bs.Add("y", []byte{4, 5, 6})
	bs.Clear()

	if bs.Count() != 0 {
		t.Errorf("Count after Clear = %d, want 0", bs.Count())
	}
	if bs.Get(0) != nil {
		t.Error("Get(0) after Clear should return nil")
	}
	// Can still add after Clear.
	idx := bs.Add("z", []byte{7})
	if idx != 0 {
		t.Errorf("first Add after Clear: idx = %d, want 0", idx)
	}
}

func TestBlobStore_Clear_FileCache(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer bs.Dispose()

	bs.Add("", []byte{1, 2})
	bs.Add("", []byte{3, 4})
	bs.Clear()

	if bs.Count() != 0 {
		t.Errorf("Count after Clear = %d, want 0", bs.Count())
	}
	// Can still add after Clear.
	bs.Add("", []byte{5, 6})
	if bs.Count() != 1 {
		t.Errorf("Count after re-Add = %d, want 1", bs.Count())
	}
	got := bs.Get(0)
	if !bytes.Equal(got, []byte{5, 6}) {
		t.Errorf("Get(0) after Clear+Add = %v, want [5 6]", got)
	}
}

// ── Dispose / Close ───────────────────────────────────────────────────────────

func TestBlobStore_Dispose_InMemory(t *testing.T) {
	bs := preview.NewBlobStore()
	bs.Add("k", []byte{1, 2})
	bs.Dispose()
	if bs.Count() != 0 {
		t.Errorf("Count after Dispose = %d, want 0", bs.Count())
	}
}

func TestBlobStore_Dispose_FileCache(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bs.Add("", []byte{42})
	bs.Dispose()
	if bs.Count() != 0 {
		t.Errorf("Count after Dispose = %d, want 0", bs.Count())
	}
}

func TestBlobStore_Close(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bs.Add("", []byte{1, 2, 3})
	if err := bs.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
	if bs.Count() != 0 {
		t.Errorf("Count after Close = %d, want 0", bs.Count())
	}
}

// ── Multiple blobs in file-cache ──────────────────────────────────────────────

func TestBlobStore_FileCache_MultipleBlobs(t *testing.T) {
	bs, err := preview.NewBlobStoreWithFileCache(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer bs.Dispose()

	blobs := [][]byte{
		{0x01, 0x02},
		{0x03, 0x04, 0x05},
		{0x06},
		{0x07, 0x08, 0x09, 0x0A},
	}
	for _, d := range blobs {
		bs.Add("", d)
	}

	if bs.Count() != len(blobs) {
		t.Errorf("Count = %d, want %d", bs.Count(), len(blobs))
	}
	for i, want := range blobs {
		got := bs.Get(i)
		if !bytes.Equal(got, want) {
			t.Errorf("Get(%d) = %v, want %v", i, got, want)
		}
	}
}
