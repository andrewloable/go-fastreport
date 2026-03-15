package engine

import (
	"context"

	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func init() {
	reportpkg.SetPrepareFunc(func(r *reportpkg.Report) (*preview.PreparedPages, error) {
		return runEngine(context.Background(), r)
	})
	reportpkg.SetPrepareFuncCtx(func(ctx context.Context, r *reportpkg.Report) (*preview.PreparedPages, error) {
		return runEngine(ctx, r)
	})
}

// runEngine is the shared implementation for both Prepare and PrepareWithContext.
func runEngine(ctx context.Context, r *reportpkg.Report) (*preview.PreparedPages, error) {
	eng := New(r)

	// Register BaseDataSource-backed data sources from the dictionary
	// so the engine initialises them during Phase 1.
	if r.Dictionary() != nil {
		for _, ds := range r.Dictionary().DataSources() {
			if bds, ok := ds.(*data.BaseDataSource); ok {
				eng.RegisterDataSource(bds)
			}
		}
	}

	opts := DefaultRunOptions()
	opts.Context = ctx
	if err := eng.Run(opts); err != nil {
		return nil, err
	}
	return eng.PreparedPages(), nil
}
