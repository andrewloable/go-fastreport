package crossview

import (
	"fmt"
	"strings"
)

// ── SliceCubeSource ───────────────────────────────────────────────────────────

// SliceCubeSource is a concrete CubeSourceBase that takes flat tabular data
// (rows of map[string]any) and pivots it into a cross-tab (cube) form.
//
// It is the Go equivalent of FastReport.Data.SliceCubeSource + BaseCubeLink.
//
// Usage:
//
//	src := crossview.NewSliceCubeSource()
//	src.AddXAxisField("Category")
//	src.AddYAxisField("Region")
//	src.AddMeasure("Sales")
//	for _, row := range rows {
//	    src.AddRow(row)
//	}
//	src.Build()
//
//	cv := crossview.NewCrossViewObject()
//	cv.SetSource(src)
//	grid, err := cv.Build()
type SliceCubeSource struct {
	// Field names that form the X (column) axis.
	xFields []string
	// Field names that form the Y (row) axis.
	yFields []string
	// Measure field names (value cells).
	measures []string
	// measuresInX controls where measures appear (default: X axis when > 1).
	measuresInX bool
	// measuresLevel is the nesting depth for measure headers (-1 = innermost).
	measuresLevel int

	// Raw input rows.
	rows []map[string]any

	// Built index structures (populated by Build()).
	xKeys    [][]string  // unique keys per X-axis level (flattened tuples)
	yKeys    [][]string  // unique keys per Y-axis level (flattened tuples)
	xTuples  [][]string  // all unique X-axis tuples (one per data column)
	yTuples  [][]string  // all unique Y-axis tuples (one per data row)
	cellData map[string]map[string][]any // xKey → yKey → measure values
}

// NewSliceCubeSource creates an empty SliceCubeSource.
// Call AddXAxisField, AddYAxisField, AddMeasure, AddRow, then Build.
func NewSliceCubeSource() *SliceCubeSource {
	return &SliceCubeSource{
		measuresInX:   true,
		measuresLevel: -1,
	}
}

// AddXAxisField adds a field name to the X (column) axis.
func (s *SliceCubeSource) AddXAxisField(name string) { s.xFields = append(s.xFields, name) }

// AddYAxisField adds a field name to the Y (row) axis.
func (s *SliceCubeSource) AddYAxisField(name string) { s.yFields = append(s.yFields, name) }

// AddMeasure adds a measure (value) field name.
func (s *SliceCubeSource) AddMeasure(name string) { s.measures = append(s.measures, name) }

// SetMeasuresInXAxis controls whether measure headers appear on the X axis.
// Default is true. When false, measures appear on the Y axis.
func (s *SliceCubeSource) SetMeasuresInXAxis(v bool) { s.measuresInX = v }

// SetMeasuresLevel sets the nesting depth at which measure headers are inserted.
// -1 (default) means innermost.
func (s *SliceCubeSource) SetMeasuresLevel(level int) { s.measuresLevel = level }

// AddRow appends a data row. Call Build after all rows are added.
func (s *SliceCubeSource) AddRow(row map[string]any) { s.rows = append(s.rows, row) }

// AddRows appends multiple data rows.
func (s *SliceCubeSource) AddRows(rows []map[string]any) { s.rows = append(s.rows, rows...) }

// Build computes the cross-tab index from the loaded rows.
// Must be called before using the source with a CrossViewObject.
func (s *SliceCubeSource) Build() {
	// Collect unique X-axis and Y-axis tuples preserving insertion order.
	seenX := map[string]bool{}
	seenY := map[string]bool{}
	s.xTuples = nil
	s.yTuples = nil
	s.cellData = map[string]map[string][]any{}

	for _, row := range s.rows {
		xKey := tupleKey(row, s.xFields)
		yKey := tupleKey(row, s.yFields)

		if !seenX[xKey] {
			seenX[xKey] = true
			s.xTuples = append(s.xTuples, tupleValues(row, s.xFields))
		}
		if !seenY[yKey] {
			seenY[yKey] = true
			s.yTuples = append(s.yTuples, tupleValues(row, s.yFields))
		}

		if _, ok := s.cellData[xKey]; !ok {
			s.cellData[xKey] = map[string][]any{}
		}
		// Accumulate measure values.
		vals := s.cellData[xKey][yKey]
		if vals == nil {
			vals = make([]any, len(s.measures))
		}
		for i, m := range s.measures {
			vals[i] = aggregateAdd(vals[i], row[m])
		}
		s.cellData[xKey][yKey] = vals
	}
}

// ── CubeSourceBase interface implementation ───────────────────────────────────

// XAxisFieldsCount returns the number of X-axis fields.
func (s *SliceCubeSource) XAxisFieldsCount() int { return len(s.xFields) }

// YAxisFieldsCount returns the number of Y-axis fields.
func (s *SliceCubeSource) YAxisFieldsCount() int { return len(s.yFields) }

// MeasuresCount returns the number of measures.
func (s *SliceCubeSource) MeasuresCount() int { return len(s.measures) }

// GetXAxisFieldName returns the field name at the given X-axis level.
func (s *SliceCubeSource) GetXAxisFieldName(i int) string {
	if i < 0 || i >= len(s.xFields) {
		return ""
	}
	return s.xFields[i]
}

// GetYAxisFieldName returns the field name at the given Y-axis level.
func (s *SliceCubeSource) GetYAxisFieldName(i int) string {
	if i < 0 || i >= len(s.yFields) {
		return ""
	}
	return s.yFields[i]
}

// GetMeasureName returns the measure name at index j.
func (s *SliceCubeSource) GetMeasureName(j int) string {
	if j < 0 || j >= len(s.measures) {
		return ""
	}
	return s.measures[j]
}

// DataColumnCount returns the number of data columns (X-axis data cells).
// When MeasuresInXAxis and measures > 1, this is xTuples × measures.
func (s *SliceCubeSource) DataColumnCount() int {
	if s.measuresInX && len(s.measures) > 1 {
		return len(s.xTuples) * len(s.measures)
	}
	return len(s.xTuples)
}

// DataRowCount returns the number of data rows (Y-axis data cells).
func (s *SliceCubeSource) DataRowCount() int {
	if !s.measuresInX && len(s.measures) > 1 {
		return len(s.yTuples) * len(s.measures)
	}
	return len(s.yTuples)
}

// MeasuresInXAxis returns true when measure headers belong on the X axis.
func (s *SliceCubeSource) MeasuresInXAxis() bool { return s.measuresInX }

// MeasuresInYAxis returns true when measure headers belong on the Y axis.
func (s *SliceCubeSource) MeasuresInYAxis() bool { return !s.measuresInX }

// MeasuresLevel returns the nesting level at which measure headers are inserted.
func (s *SliceCubeSource) MeasuresLevel() int { return s.measuresLevel }

// SourceAssigned returns true when Build() has been called and the source has
// data. Mirrors IBaseCubeLink.SourceAssigned (BaseCubeLink.cs).
func (s *SliceCubeSource) SourceAssigned() bool {
	return s.cellData != nil
}

// TraverseXAxis calls fn for each X-axis header cell in level-major order.
// When MeasuresInXAxis and MeasuresCount > 1, an extra level is emitted for
// measure names. The Cell coordinate is the 0-based data-column index.
// MeasureIndex is set to the measure index for measure-level cells, 0 elsewhere.
func (s *SliceCubeSource) TraverseXAxis(fn AxisTraverseFunc) {
	if len(s.xTuples) == 0 {
		return
	}
	nFields := len(s.xFields)
	measuresInX := s.measuresInX && len(s.measures) > 1

	if nFields == 0 && !measuresInX {
		return
	}

	// Compute total levels and the position of the measures level.
	totalLevels := nFields
	measLevel := -1
	if measuresInX {
		totalLevels++
		measLevel = s.measuresLevel
		if measLevel < 0 {
			measLevel = totalLevels - 1 // innermost
		}
	}

	fieldIdx := 0 // which xFields index we are processing
	for level := 0; level < totalLevels; level++ {
		if measuresInX && level == measLevel {
			// Emit one cell per measure for each xTuple group.
			// Data column index: xTupleIdx*len(measures) + measureIdx.
			for col := 0; col < len(s.xTuples); col++ {
				for j, mName := range s.measures {
					fn(AxisDrawCell{
						Text:         mName,
						Cell:         col*len(s.measures) + j,
						Level:        level,
						SizeCell:     1,
						SizeLevel:    1,
						MeasureIndex: j,
					})
				}
			}
		} else {
			// Emit span cells for the field at fieldIdx.
			// When measures are in the X axis, each xTuple occupies len(measures) data columns.
			col := 0
			for col < len(s.xTuples) {
				val := s.xTuples[col][fieldIdx]
				span := 1
				for col+span < len(s.xTuples) {
					if samePrefixUpTo(s.xTuples[col], s.xTuples[col+span], fieldIdx) {
						span++
					} else {
						break
					}
				}
				var dataCol, dataSpan int
				if measuresInX {
					dataCol = col * len(s.measures)
					dataSpan = span * len(s.measures)
				} else {
					dataCol = col
					dataSpan = span
				}
				fn(AxisDrawCell{
					Text:      val,
					Cell:      dataCol,
					Level:     level,
					SizeCell:  dataSpan,
					SizeLevel: 1,
				})
				col += span
			}
			fieldIdx++
		}
	}
}

// TraverseYAxis calls fn for each Y-axis header cell in level-major order.
// When MeasuresInYAxis and MeasuresCount > 1, an extra level is emitted for
// measure names. The Cell coordinate is the 0-based data-row index.
// MeasureIndex is set to the measure index for measure-level cells, 0 elsewhere.
func (s *SliceCubeSource) TraverseYAxis(fn AxisTraverseFunc) {
	if len(s.yTuples) == 0 {
		return
	}
	nFields := len(s.yFields)
	measuresInY := !s.measuresInX && len(s.measures) > 1

	if nFields == 0 && !measuresInY {
		return
	}

	totalLevels := nFields
	measLevel := -1
	if measuresInY {
		totalLevels++
		measLevel = s.measuresLevel
		if measLevel < 0 {
			measLevel = totalLevels - 1 // innermost
		}
	}

	fieldIdx := 0
	for level := 0; level < totalLevels; level++ {
		if measuresInY && level == measLevel {
			// Emit one cell per measure for each yTuple group.
			for row := 0; row < len(s.yTuples); row++ {
				for j, mName := range s.measures {
					fn(AxisDrawCell{
						Text:         mName,
						Cell:         row*len(s.measures) + j,
						Level:        level,
						SizeCell:     1,
						SizeLevel:    1,
						MeasureIndex: j,
					})
				}
			}
		} else {
			row := 0
			for row < len(s.yTuples) {
				val := s.yTuples[row][fieldIdx]
				span := 1
				for row+span < len(s.yTuples) {
					if samePrefixUpTo(s.yTuples[row], s.yTuples[row+span], fieldIdx) {
						span++
					} else {
						break
					}
				}
				var dataRow, dataSpan int
				if measuresInY {
					dataRow = row * len(s.measures)
					dataSpan = span * len(s.measures)
				} else {
					dataRow = row
					dataSpan = span
				}
				fn(AxisDrawCell{
					Text:      val,
					Cell:      dataRow,
					Level:     level,
					SizeCell:  dataSpan,
					SizeLevel: 1,
				})
				row += span
			}
			fieldIdx++
		}
	}
}

// GetMeasureCell returns the cell value at data coordinates (x, y).
// x is the 0-based data column, y is the 0-based data row.
func (s *SliceCubeSource) GetMeasureCell(x, y int) MeasureCell {
	// When measures are in the X axis, x encodes (xTupleIdx, measureIdx).
	var xTupleIdx, measureIdx int
	if s.measuresInX && len(s.measures) > 1 {
		measureIdx = x % len(s.measures)
		xTupleIdx = x / len(s.measures)
	} else {
		xTupleIdx = x
		measureIdx = 0
	}

	var yTupleIdx int
	if !s.measuresInX && len(s.measures) > 1 {
		measureIdx = y % len(s.measures)
		yTupleIdx = y / len(s.measures)
	} else {
		yTupleIdx = y
	}

	if xTupleIdx >= len(s.xTuples) || yTupleIdx >= len(s.yTuples) {
		return MeasureCell{}
	}

	xKey := strings.Join(s.xTuples[xTupleIdx], "\x00")
	yKey := strings.Join(s.yTuples[yTupleIdx], "\x00")

	yMap, ok := s.cellData[xKey]
	if !ok {
		return MeasureCell{}
	}
	vals, ok := yMap[yKey]
	if !ok || measureIdx >= len(vals) {
		return MeasureCell{}
	}
	return MeasureCell{Text: fmt.Sprintf("%v", vals[measureIdx])}
}

// ── Helper functions ──────────────────────────────────────────────────────────

// tupleKey returns a unique string key for the given fields in a row.
func tupleKey(row map[string]any, fields []string) string {
	return strings.Join(tupleValues(row, fields), "\x00")
}

// tupleValues returns the string values of the given fields from row.
func tupleValues(row map[string]any, fields []string) []string {
	vals := make([]string, len(fields))
	for i, f := range fields {
		v := row[f]
		if v == nil {
			vals[i] = ""
		} else {
			vals[i] = fmt.Sprintf("%v", v)
		}
	}
	return vals
}

// samePrefixUpTo returns true when two tuples have the same values at all
// positions 0..level (inclusive).
func samePrefixUpTo(a, b []string, level int) bool {
	for i := 0; i <= level && i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// aggregateAdd adds v2 to v1 using numeric addition when both are numeric,
// otherwise concatenates strings.
func aggregateAdd(v1, v2 any) any {
	if v2 == nil {
		return v1
	}
	switch val := v2.(type) {
	case int:
		if v1 == nil {
			return val
		}
		if n, ok := v1.(int); ok {
			return n + val
		}
		if f, ok := v1.(float64); ok {
			return f + float64(val)
		}
		return val
	case int64:
		if v1 == nil {
			return val
		}
		if n, ok := v1.(int64); ok {
			return n + val
		}
		if n, ok := v1.(int); ok {
			return int64(n) + val
		}
		return val
	case float64:
		if v1 == nil {
			return val
		}
		if f, ok := v1.(float64); ok {
			return f + val
		}
		if n, ok := v1.(int); ok {
			return float64(n) + val
		}
		return val
	case float32:
		return aggregateAdd(v1, float64(val))
	default:
		// Non-numeric: return the first non-nil value.
		if v1 == nil {
			return v2
		}
		return v1
	}
}
