package reportpkg

// ReportLoader loads a Report from a path.
// It is supplied by the caller so the engine is not tied to disk I/O.
type ReportLoader interface {
	Load(path string) (*Report, error)
}

// ReportLoaderFunc allows using a plain function as a ReportLoader.
type ReportLoaderFunc func(path string) (*Report, error)

// Load implements ReportLoader.
func (f ReportLoaderFunc) Load(path string) (*Report, error) { return f(path) }

// ApplyBase merges properties and pages from base into r (the child report).
//
// Merge rules (mirrors FastReport .NET behaviour):
//   - Scalar properties (Info, flags, etc.) in the child keep their values;
//     unset (zero) values fall back to base values.
//   - Pages from the base that are NOT present in the child (matched by Name)
//     are prepended to the child's page list.
//   - Pages present in both base and child are merged: bands from the base page
//     that are NOT present in the child page (matched by Name) are added.
func (r *Report) ApplyBase(base *Report) {
	// Merge scalar Info fields — child wins if non-zero.
	if r.Info.Name == "" {
		r.Info.Name = base.Info.Name
	}
	if r.Info.Author == "" {
		r.Info.Author = base.Info.Author
	}
	if r.Info.Description == "" {
		r.Info.Description = base.Info.Description
	}
	if r.Info.Version == "" {
		r.Info.Version = base.Info.Version
	}

	// Merge flags — child keeps true if set; otherwise inherits base.
	if !r.Compressed {
		r.Compressed = base.Compressed
	}
	if !r.ConvertNulls {
		r.ConvertNulls = base.ConvertNulls
	}
	if !r.DoublePass {
		r.DoublePass = base.DoublePass
	}
	if r.InitialPageNumber == 1 && base.InitialPageNumber != 1 {
		r.InitialPageNumber = base.InitialPageNumber
	}
	if r.MaxPages == 0 {
		r.MaxPages = base.MaxPages
	}
	if r.StartReportEvent == "" {
		r.StartReportEvent = base.StartReportEvent
	}
	if r.FinishReportEvent == "" {
		r.FinishReportEvent = base.FinishReportEvent
	}

	// Build lookup of child pages by name.
	childPageByName := make(map[string]*ReportPage, len(r.pages))
	for _, p := range r.pages {
		childPageByName[p.Name()] = p
	}

	// For each base page: merge into matching child page, or prepend if absent.
	var toInsert []*ReportPage
	for _, bp := range base.pages {
		if cp, ok := childPageByName[bp.Name()]; ok {
			// Merge: add base bands that the child page doesn't have.
			cp.mergeFromBase(bp)
		} else {
			// Child has no page with this name → clone base page.
			clone := bp.Clone()
			clone.SetInherited(true)
			toInsert = append(toInsert, clone)
		}
	}

	// Prepend base-only pages in original order.
	if len(toInsert) > 0 {
		r.pages = append(toInsert, r.pages...)
	}
}

// LoadBase loads the base report via loader and applies it to r.
// It is a convenience wrapper around ApplyBase.
func (r *Report) LoadBase(loader ReportLoader) error {
	if r.BaseReportPath == "" {
		return nil // nothing to do
	}
	base, err := loader.Load(r.BaseReportPath)
	if err != nil {
		return err
	}
	r.ApplyBase(base)
	return nil
}
