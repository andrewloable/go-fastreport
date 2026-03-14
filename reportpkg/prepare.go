package reportpkg

import (
	"errors"
	"fmt"

	"github.com/andrewloable/go-fastreport/preview"
)

// PreparedPages returns the prepared pages from the last Prepare() call,
// or nil if Prepare() has not been called.
func (r *Report) PreparedPages() *preview.PreparedPages {
	return r.preparedPages
}

// PrepareFunc is the signature of the function that executes the engine.
// It is set by the engine package via SetPrepareFunc to avoid an import cycle.
type PrepareFunc func(r *Report) (*preview.PreparedPages, error)

var globalPrepareFunc PrepareFunc

// SetPrepareFunc registers the engine's Prepare implementation.
// This is called from engine's init() to break the reportpkg→engine import cycle.
func SetPrepareFunc(f PrepareFunc) {
	globalPrepareFunc = f
}

// Prepare executes the report and populates PreparedPages.
// It is the primary user-facing API for report generation.
//
// An engine implementation must be registered via SetPrepareFunc before
// calling this method. Import "github.com/andrewloable/go-fastreport/engine"
// (or the high-level "github.com/andrewloable/go-fastreport" package) as a
// blank import to ensure the registration runs.
func (r *Report) Prepare() error {
	if globalPrepareFunc == nil {
		return errors.New("report.Prepare: no engine registered — import the engine package")
	}

	// Evaluate expression-based parameters before running.
	if r.dictionary != nil {
		if err := r.dictionary.EvaluateAll(); err != nil {
			return fmt.Errorf("report.Prepare: parameter evaluation: %w", err)
		}
	}

	pp, err := globalPrepareFunc(r)
	if err != nil {
		return fmt.Errorf("report.Prepare: %w", err)
	}
	r.preparedPages = pp
	return nil
}
