package serial_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/loabletech/go-fastreport/report"
	"github.com/loabletech/go-fastreport/serial"
)

// --- stub report.Base implementation ---

type stubObject struct {
	name string
}

func (s *stubObject) Name() string             { return s.name }
func (s *stubObject) SetName(n string)         { s.name = n }
func (s *stubObject) BaseName() string         { return "Stub" }
func (s *stubObject) Parent() report.Parent    { return nil }
func (s *stubObject) SetParent(report.Parent)  {}
func (s *stubObject) Serialize(report.Writer) error  { return nil }
func (s *stubObject) Deserialize(report.Reader) error { return nil }

func newStub() report.Base { return &stubObject{} }

// --- Tests ---

func TestNewRegistry_Empty(t *testing.T) {
	r := serial.NewRegistry()
	if names := r.Names(); len(names) != 0 {
		t.Errorf("new registry should have no names, got %v", names)
	}
}

func TestRegister_AndHas(t *testing.T) {
	r := serial.NewRegistry()
	if err := r.Register("Stub", newStub); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if !r.Has("Stub") {
		t.Error("Has should return true after Register")
	}
	if r.Has("Unknown") {
		t.Error("Has should return false for unregistered name")
	}
}

func TestRegister_DuplicateReturnsError(t *testing.T) {
	r := serial.NewRegistry()
	if err := r.Register("Stub", newStub); err != nil {
		t.Fatalf("first Register failed: %v", err)
	}
	err := r.Register("Stub", newStub)
	if err == nil {
		t.Fatal("second Register should return an error")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMustRegister_Success(t *testing.T) {
	r := serial.NewRegistry()
	// Should not panic.
	r.MustRegister("Stub", newStub)
	if !r.Has("Stub") {
		t.Error("Has should return true after MustRegister")
	}
}

func TestMustRegister_PanicsOnDuplicate(t *testing.T) {
	r := serial.NewRegistry()
	r.MustRegister("Stub", newStub)
	defer func() {
		if rec := recover(); rec == nil {
			t.Error("MustRegister should panic on duplicate")
		}
	}()
	r.MustRegister("Stub", newStub)
}

func TestCreate_Success(t *testing.T) {
	r := serial.NewRegistry()
	r.MustRegister("Stub", newStub)
	obj, err := r.Create("Stub")
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if obj == nil {
		t.Fatal("Create returned nil object")
	}
	if obj.BaseName() != "Stub" {
		t.Errorf("BaseName = %q, want Stub", obj.BaseName())
	}
}

func TestCreate_UnknownReturnsError(t *testing.T) {
	r := serial.NewRegistry()
	obj, err := r.Create("NonExistent")
	if err == nil {
		t.Fatal("Create should return error for unknown type")
	}
	if obj != nil {
		t.Error("Create should return nil object on error")
	}
	if !strings.Contains(err.Error(), "unknown type") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNames_Sorted(t *testing.T) {
	r := serial.NewRegistry()
	r.MustRegister("Zebra", newStub)
	r.MustRegister("Apple", newStub)
	r.MustRegister("Mango", newStub)

	names := r.Names()
	if len(names) != 3 {
		t.Fatalf("Names length = %d, want 3", len(names))
	}
	expected := []string{"Apple", "Mango", "Zebra"}
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("Names[%d] = %q, want %q", i, names[i], want)
		}
	}
}

func TestCreate_ReturnsFreshInstance(t *testing.T) {
	r := serial.NewRegistry()
	r.MustRegister("Stub", newStub)

	obj1, _ := r.Create("Stub")
	obj2, _ := r.Create("Stub")
	obj1.SetName("first")
	if obj2.Name() != "" {
		t.Error("Create should return independent instances")
	}
}

func TestDefaultRegistry_Exists(t *testing.T) {
	// DefaultRegistry should be non-nil and usable.
	if serial.DefaultRegistry == nil {
		t.Fatal("DefaultRegistry should not be nil")
	}
	// Registering a unique name to avoid conflicts with parallel test runs.
	_ = serial.DefaultRegistry.Register("__test_default__", newStub)
	// Either it succeeds or it was already registered; both are acceptable here.
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := serial.NewRegistry()
	// Pre-register a type that all goroutines will read.
	r.MustRegister("Shared", newStub)

	var wg sync.WaitGroup
	const n = 50

	// Concurrent reads.
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !r.Has("Shared") {
				t.Errorf("Has returned false under concurrency")
			}
			obj, err := r.Create("Shared")
			if err != nil || obj == nil {
				t.Errorf("Create failed under concurrency: %v", err)
			}
			_ = r.Names()
		}()
	}

	// Concurrent writes (unique names to avoid duplicate-error noise).
	for i := 0; i < n; i++ {
		wg.Add(1)
		name := strings.Repeat("x", i+1)
		go func(n string) {
			defer wg.Done()
			_ = r.Register(n, newStub)
		}(name)
	}

	wg.Wait()
}
