package preview

// blob_store.go — file-based caching BlobStore
//
// Ported from C# FastReport.Utils.BlobStore (FastReport.Base/Utils/BlobStore.cs).
//
// When useFileCache is true, blobs written via Add/AddOrUpdate are
// immediately flushed to a temporary OS file and the in-memory slice is
// cleared, so only the file-position and size are retained.  Get() / GetStream()
// seek-and-read from the temp file on each call, matching the C# behaviour.
// When useFileCache is false (default), blobs are kept only in memory,
// matching the original Go implementation.
//
// The temp file is created with os.CreateTemp and deleted in Dispose() /
// Close().  BlobStore implements io.Closer for convenience.
//
// C# reference lines:
//   BlobStore constructor  → lines 119-129
//   BlobItem constructor   → lines 177-191
//   BlobItem.Stream getter → lines 141-155
//   BlobStore.Add          → lines 26-31
//   BlobStore.AddOrUpdate  → lines 33-48
//   BlobStore.Get          → lines 50-54
//   BlobStore.GetSource    → lines 56-59
//   BlobStore.Clear        → lines 72-80
//   BlobStore.Dispose      → lines 108-117

import (
	"bytes"
	"io"
	"os"
	"sync"
)

// blobItem mirrors C# BlobStore.BlobItem.
// In file-cache mode, data is nil and the content lives in the temp file.
type blobItem struct {
	// source is the optional deduplication key (C# BlobItem.Source).
	source string
	// data holds the blob bytes when useFileCache is false.
	data []byte
	// tempFileOffset and tempFileSize are used when useFileCache is true.
	tempFileOffset int64
	tempFileSize   int64
}

// BlobStore stores binary blobs (e.g. images) referenced by prepared pages.
// It is the Go equivalent of C# FastReport.Utils.BlobStore.
//
// When constructed with NewBlobStoreWithFileCache(true) blobs are persisted to
// a temp file to reduce memory pressure, matching Report.UseFileCache = true in C#.
type BlobStore struct {
	mu           sync.Mutex
	items        []*blobItem
	sourceIndex  map[string]int // source key → list index (for AddOrUpdate dedup)
	useFileCache bool
	tempFile     *os.File // nil when useFileCache is false
}

// NewBlobStore creates an in-memory BlobStore (useFileCache = false).
// This is the default used by PreparedPages.New().
func NewBlobStore() *BlobStore {
	return &BlobStore{sourceIndex: make(map[string]int)}
}

// NewBlobStoreWithFileCache creates a BlobStore that optionally uses a temp file
// backend.  When useFileCache is true a temporary file is created; all blobs are
// written to it and freed from memory after each write.
// Equivalent to C# new BlobStore(useFileCache).
func NewBlobStoreWithFileCache(useFileCache bool) (*BlobStore, error) {
	bs := &BlobStore{
		sourceIndex:  make(map[string]int),
		useFileCache: useFileCache,
	}
	if useFileCache {
		f, err := os.CreateTemp("", "fastreport-blobstore-*")
		if err != nil {
			return nil, err
		}
		bs.tempFile = f
	}
	return bs, nil
}

// Add stores blob data with an optional source key for deduplication and
// returns the integer index.  If name is non-empty and a blob with that source
// was already added, the existing index is returned without storing a duplicate.
//
// This merges C# BlobStore.Add (anonymous, no source) and
// BlobStore.AddOrUpdate (with source dedup).
func (b *BlobStore) Add(name string, data []byte) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.addLocked(data, name)
}

// AddOrUpdate stores data under source key src (may be empty) and returns the
// index.  If src is non-empty and a blob with that source already exists its
// index is returned immediately (no duplicate stored).
// Equivalent to C# BlobStore.AddOrUpdate(byte[] stream, string src).
func (b *BlobStore) AddOrUpdate(data []byte, src string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.addLocked(data, src)
}

// addLocked is the shared implementation; must be called with b.mu held.
func (b *BlobStore) addLocked(data []byte, src string) int {
	if src != "" {
		if idx, ok := b.sourceIndex[src]; ok {
			return idx
		}
	}
	item := &blobItem{source: src}
	if b.useFileCache && b.tempFile != nil {
		// Append to the temp file and clear the in-memory slice.
		// C# BlobItem constructor lines 182-189.
		offset, _ := b.tempFile.Seek(0, io.SeekEnd)
		item.tempFileOffset = offset
		item.tempFileSize = int64(len(data))
		_, _ = b.tempFile.Write(data)
		_ = b.tempFile.Sync()
		// data is intentionally left as nil in item; memory is freed.
	} else {
		item.data = data
	}
	idx := len(b.items)
	b.items = append(b.items, item)
	if src != "" {
		b.sourceIndex[src] = idx
	}
	return idx
}

// Get returns the blob bytes at the given index, or nil if out of range.
// In file-cache mode the bytes are read from the temp file on each call,
// matching C# BlobItem.Stream getter lines 144-154.
func (b *BlobStore) Get(idx int) []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	if idx < 0 || idx >= len(b.items) {
		return nil
	}
	item := b.items[idx]
	if b.useFileCache && b.tempFile != nil {
		buf := make([]byte, item.tempFileSize)
		_, _ = b.tempFile.ReadAt(buf, item.tempFileOffset)
		return buf
	}
	return item.data
}

// GetStream returns an io.Reader over the blob at idx, or nil if out of range.
// In file-cache mode the reader is backed by a slice read from the temp file;
// in memory mode it wraps the in-memory slice with bytes.NewReader.
func (b *BlobStore) GetStream(idx int) io.Reader {
	data := b.Get(idx)
	if data == nil {
		return nil
	}
	return bytes.NewReader(data)
}

// GetSource returns the source key string for the blob at idx.
// Returns "" if the index is out of range or the blob has no source.
// Equivalent to C# BlobStore.GetSource(int index).
func (b *BlobStore) GetSource(idx int) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if idx < 0 || idx >= len(b.items) {
		return ""
	}
	return b.items[idx].source
}

// Count returns the number of stored blobs.
func (b *BlobStore) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.items)
}

// Clear removes all items and frees in-memory data.
// In file-cache mode the underlying temp file is truncated so the
// OS storage is reclaimed; the file itself (and its handle) remain open
// for subsequent Add calls, matching C# BlobStore.Clear lines 72-80.
func (b *BlobStore) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = b.items[:0]
	b.sourceIndex = make(map[string]int)
	if b.tempFile != nil {
		_ = b.tempFile.Truncate(0)
		_, _ = b.tempFile.Seek(0, io.SeekStart)
	}
}

// Dispose releases the temp file and deletes it from disk.
// After Dispose the BlobStore must not be used.
// Equivalent to C# BlobStore.Dispose lines 108-117.
func (b *BlobStore) Dispose() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = b.items[:0]
	b.sourceIndex = make(map[string]int)
	if b.tempFile != nil {
		name := b.tempFile.Name()
		_ = b.tempFile.Close()
		_ = os.Remove(name)
		b.tempFile = nil
	}
}

// Close is an alias for Dispose, satisfying io.Closer.
func (b *BlobStore) Close() error {
	b.Dispose()
	return nil
}
