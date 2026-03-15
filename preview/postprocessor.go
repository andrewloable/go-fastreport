package preview

import (
	"fmt"
	"strings"
)

// Postprocessor scans all PreparedPages after the engine run completes and
// performs deferred replacements:
//   - [TotalPages] / [PageCount] → actual page count string
//
// This is the Go equivalent of FastReport.Preview.PreparedPagePostprocessor.
// In the C# code the postprocessor also handles text duplicates (Clear/Hide/
// Merge) and merged text objects; those features are added here as stubs and
// can be activated when the engine sets the corresponding flags.
type Postprocessor struct {
	pp *PreparedPages
}

// NewPostprocessor creates a Postprocessor that will operate on pp.
func NewPostprocessor(pp *PreparedPages) *Postprocessor {
	return &Postprocessor{pp: pp}
}

// Process applies all deferred post-processing steps to the prepared pages.
// Call this once after the engine has finished its final pass.
func (p *Postprocessor) Process() {
	total := p.pp.Count()
	totalStr := fmt.Sprintf("%d", total)

	for i := range p.pp.pages {
		pg := p.pp.pages[i]
		// Update the logical page number in case the engine set a placeholder.
		pageStr := fmt.Sprintf("%d", pg.PageNo)
		for j := range pg.Bands {
			band := pg.Bands[j]
			for k := range band.Objects {
				obj := &band.Objects[k]
				if obj.Text == "" {
					continue
				}
				t := obj.Text
				t = strings.ReplaceAll(t, "[TotalPages]", totalStr)
				t = strings.ReplaceAll(t, "[PageCount]", totalStr)
				t = strings.ReplaceAll(t, "[Page]", pageStr)
				obj.Text = t
			}
		}
	}
}
