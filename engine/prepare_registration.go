package engine

import (
	"github.com/andrewloable/go-fastreport/data"
	"github.com/andrewloable/go-fastreport/preview"
	"github.com/andrewloable/go-fastreport/reportpkg"
)

func init() {
	reportpkg.SetPrepareFunc(func(r *reportpkg.Report) (*preview.PreparedPages, error) {
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
		if err := eng.Run(opts); err != nil {
			return nil, err
		}
		return eng.PreparedPages(), nil
	})
}
