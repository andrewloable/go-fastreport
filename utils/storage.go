package utils

import (
	"io"
	"os"
	"path/filepath"
)

// StorageService is the report storage abstraction. Implementations may load
// and save report data from local files, databases, object stores (S3, GCS),
// or any other backend.
//
// This is the Go equivalent of FastReport.Utils.StorageService.
type StorageService interface {
	// Load reads the resource at the given path and returns its bytes.
	Load(path string) ([]byte, error)
	// Save writes data to the given path.
	Save(path string, data []byte) error
	// Exists reports whether the resource at path exists.
	Exists(path string) bool
	// Reader opens the resource at path for streaming reads.
	// Callers must close the returned ReadCloser.
	Reader(path string) (io.ReadCloser, error)
	// Writer opens the resource at path for streaming writes.
	// Callers must close the returned WriteCloser.
	Writer(path string) (io.WriteCloser, error)
}

// ── FileStorageService ────────────────────────────────────────────────────────

// FileStorageService is a StorageService backed by the local filesystem.
// The optional BaseDir is prepended to all relative paths.
type FileStorageService struct {
	// BaseDir is prepended to relative paths. If empty, paths are used as-is.
	BaseDir string
}

// NewFileStorageService creates a FileStorageService rooted at baseDir.
func NewFileStorageService(baseDir string) *FileStorageService {
	return &FileStorageService{BaseDir: baseDir}
}

func (f *FileStorageService) abs(path string) string {
	if f.BaseDir == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(f.BaseDir, path)
}

// Load reads the file at path (relative to BaseDir) and returns its bytes.
func (f *FileStorageService) Load(path string) ([]byte, error) {
	return os.ReadFile(f.abs(path))
}

// Save writes data to path (relative to BaseDir), creating parent directories
// as needed.
func (f *FileStorageService) Save(path string, data []byte) error {
	abs := f.abs(path)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, data, 0o644)
}

// Exists reports whether the file at path exists.
func (f *FileStorageService) Exists(path string) bool {
	_, err := os.Stat(f.abs(path))
	return err == nil
}

// Reader opens the file at path for reading.
func (f *FileStorageService) Reader(path string) (io.ReadCloser, error) {
	return os.Open(f.abs(path))
}

// Writer opens the file at path for writing, creating it or truncating it.
func (f *FileStorageService) Writer(path string) (io.WriteCloser, error) {
	abs := f.abs(path)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, err
	}
	return os.Create(abs)
}

// ── MemoryStorageService ──────────────────────────────────────────────────────

// MemoryStorageService is a StorageService backed by an in-memory map.
// Useful for testing and embedded scenarios.
type MemoryStorageService struct {
	data map[string][]byte
}

// NewMemoryStorageService creates an empty MemoryStorageService.
func NewMemoryStorageService() *MemoryStorageService {
	return &MemoryStorageService{data: make(map[string][]byte)}
}

// Put adds or replaces the resource at path with data.
func (m *MemoryStorageService) Put(path string, data []byte) {
	cp := make([]byte, len(data))
	copy(cp, data)
	m.data[path] = cp
}

// Load returns the bytes stored at path, or an error if not found.
func (m *MemoryStorageService) Load(path string) ([]byte, error) {
	d, ok := m.data[path]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}
	cp := make([]byte, len(d))
	copy(cp, d)
	return cp, nil
}

// Save stores data under path.
func (m *MemoryStorageService) Save(path string, data []byte) error {
	m.Put(path, data)
	return nil
}

// Exists reports whether path exists in the memory store.
func (m *MemoryStorageService) Exists(path string) bool {
	_, ok := m.data[path]
	return ok
}

// Reader returns an io.ReadCloser over the stored bytes.
func (m *MemoryStorageService) Reader(path string) (io.ReadCloser, error) {
	d, err := m.Load(path)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(readerFromBytes(d)), nil
}

// Writer returns an io.WriteCloser that commits to the memory store when closed.
func (m *MemoryStorageService) Writer(path string) (io.WriteCloser, error) {
	return &memWriter{store: m, path: path}, nil
}

// ── internal helpers ──────────────────────────────────────────────────────────

// readerFromBytes returns an io.Reader over b.
func readerFromBytes(b []byte) io.Reader {
	return &bytesReader{data: b}
}

type bytesReader struct {
	data []byte
	pos  int
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// memWriter accumulates writes and commits to MemoryStorageService on Close.
type memWriter struct {
	store *MemoryStorageService
	path  string
	buf   []byte
}

func (w *memWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *memWriter) Close() error {
	w.store.Put(w.path, w.buf)
	return nil
}
