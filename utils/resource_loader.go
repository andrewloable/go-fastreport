package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"
)

// ResourceLoader is the Go equivalent of FastReport.Utils.ResourceLoader.
//
// In .NET, ResourceLoader retrieves embedded assembly resources by name using
// Assembly.GetManifestResourceStream. In Go there are no DLL-level embedded
// resources, so ResourceLoader instead maintains a process-wide registry:
// packages register named byte-slice providers via RegisterResource, and
// GetStream / UnpackStream look them up.
//
// Registration key format: "assembly:name" (e.g. "FastReport:en.xml").
// The convenience functions GetStream(name) and UnpackStream(name) default
// the assembly to "FastReport".
//
// C# source: FastReport.Base/Utils/ResourceLoader.cs

// resourceEntry is a factory that returns a fresh io.ReadCloser over the
// resource bytes. Using a factory (rather than storing raw []byte) avoids
// holding an extra copy of large resources in memory.
type resourceEntry func() (io.ReadCloser, error)

var resourceRegistry struct {
	mu      sync.RWMutex
	entries map[string]resourceEntry
}

func init() {
	resourceRegistry.entries = make(map[string]resourceEntry)
}

func resourceKey(assembly, name string) string {
	return assembly + ":" + name
}

// RegisterResource registers a named resource under the given assembly and
// resource name. The provider function is called each time GetStream is
// invoked for that resource; it must return a fresh, independently readable
// io.ReadCloser.
//
// Calling RegisterResource with an already-registered key replaces the
// existing provider.
func RegisterResource(assembly, name string, provider func() (io.ReadCloser, error)) {
	resourceRegistry.mu.Lock()
	defer resourceRegistry.mu.Unlock()
	resourceRegistry.entries[resourceKey(assembly, name)] = provider
}

// RegisterResourceBytes is a convenience wrapper around RegisterResource that
// registers a static byte slice as a resource. Each call to GetStream returns
// a fresh reader over a copy of the bytes.
func RegisterResourceBytes(assembly, name string, data []byte) {
	RegisterResource(assembly, name, func() (io.ReadCloser, error) {
		cp := make([]byte, len(data))
		copy(cp, data)
		return io.NopCloser(bytes.NewReader(cp)), nil
	})
}

// GetStream returns a stream for the named resource in the given assembly.
// Returns (nil, nil) when the resource is not registered, mirroring the
// C# behaviour where GetManifestResourceStream returns null for unknown names.
//
// C# equivalent: ResourceLoader.GetStream(assembly, resource)
func GetStream(assembly, name string) (io.ReadCloser, error) {
	resourceRegistry.mu.RLock()
	provider, ok := resourceRegistry.entries[resourceKey(assembly, name)]
	resourceRegistry.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return provider()
}

// GetStreamFR returns a stream for a named resource in the default "FastReport"
// assembly. Returns (nil, nil) when the resource is not registered.
//
// C# equivalent: ResourceLoader.GetStream(resource)
func GetStreamFR(name string) (io.ReadCloser, error) {
	return GetStream("FastReport", name)
}

// UnpackStream returns the gzip-decompressed content of the named resource in
// the given assembly as an in-memory reader. The decompressed bytes are fully
// buffered in memory, matching the C# implementation which copies to a
// MemoryStream before returning.
//
// Returns (nil, nil) when the resource is not registered.
//
// C# equivalent: ResourceLoader.UnpackStream(assembly, resource)
func UnpackStream(assembly, name string) (io.ReadCloser, error) {
	rc, err := GetStream(assembly, name)
	if err != nil {
		return nil, fmt.Errorf("resourceloader: GetStream %s:%s: %w", assembly, name, err)
	}
	if rc == nil {
		return nil, nil
	}
	defer rc.Close()

	gr, err := gzip.NewReader(rc)
	if err != nil {
		return nil, fmt.Errorf("resourceloader: gzip.NewReader %s:%s: %w", assembly, name, err)
	}
	defer gr.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, gr); err != nil {
		return nil, fmt.Errorf("resourceloader: decompress %s:%s: %w", assembly, name, err)
	}

	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// UnpackStreamFR returns the gzip-decompressed content of a named resource in
// the default "FastReport" assembly.
//
// C# equivalent: ResourceLoader.UnpackStream(resource)
func UnpackStreamFR(name string) (io.ReadCloser, error) {
	return UnpackStream("FastReport", name)
}
