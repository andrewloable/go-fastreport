package export_test

// base_extra_test.go — coverage for the remaining branch in ExportBase.Export.
// Specifically: the `if pg == nil { continue }` guard (line 199 in base.go).
//
// Strategy: create an exporter whose Start() trims the PreparedPages down to
// zero pages AFTER preparePageIndices has already resolved the page indices.
// That causes GetPage(idx) to return nil for every idx in e.pages.

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/andrewloable/go-fastreport/export"
	"github.com/andrewloable/go-fastreport/preview"
)

// nilPageExporter trims the PreparedPages in Start() so every subsequent
// GetPage call returns nil, exercising the `if pg == nil { continue }` branch.
type nilPageExporter struct {
	export.NoopExporter
	pp *preview.PreparedPages
}

func (n *nilPageExporter) Start() error {
	// Remove all pages so GetPage returns nil for every resolved index.
	n.pp.TrimTo(0)
	return nil
}

func TestExport_GetPageReturnsNil(t *testing.T) {
	// Build a PreparedPages with 2 pages so preparePageIndices resolves indices [0,1].
	pp := preview.New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 30})
	pp.AddPage(595, 842, 2)
	_ = pp.AddBand(&preview.PreparedBand{Name: "B", Top: 0, Height: 30})

	base := export.NewExportBase()
	exp := &nilPageExporter{pp: pp}

	// Export must succeed (nil pages are silently skipped).
	if err := base.Export(pp, new(bytes.Buffer), exp); err != nil {
		t.Fatalf("Export: unexpected error: %v", err)
	}

	// No pages were visited because all GetPage calls returned nil.
	// Pages() should still reflect the indices that were resolved before Start().
	pages := base.Pages()
	if len(pages) != 2 {
		t.Errorf("Pages() len = %d, want 2 (preparePageIndices ran before Start)", len(pages))
	}
}

// Ensure errExporter's Finish error path is covered separately (already exists,
// but include a companion test that also exercises Start error recovery).

type startErrExporter2 struct {
	export.NoopExporter
}

func (e *startErrExporter2) Start() error { return fmt.Errorf("injected start error") }

func TestExport_StartError_Message(t *testing.T) {
	pp := buildPreparedPages(1, []string{"B"})
	base := export.NewExportBase()
	err := base.Export(pp, new(bytes.Buffer), &startErrExporter2{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The error must be wrapped with the "export: start:" prefix.
	if got := err.Error(); got != "export: start: injected start error" {
		t.Errorf("unexpected error message: %q", got)
	}
}
