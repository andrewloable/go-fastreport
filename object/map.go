package object

import "github.com/andrewloable/go-fastreport/report"

// MapLayer represents a single data layer within a MapObject.
// Each layer binds geographic shapes (from a Shapefile) to a data source,
// applying color or size ranges to visualize analytical values.
type MapLayer struct {
	report.ReportComponentBase

	// Shapefile is the name of the shapefile resource for this layer.
	Shapefile string
	// Type is the layer type ("Choropleth", "Bubble", etc.).
	Type string
	// DataSource is the bound data source name.
	DataSource string
	// Filter is an expression used to filter rows from the data source.
	Filter string
	// SpatialColumn is the data column containing geographic identifiers.
	SpatialColumn string
	// SpatialValue is an expression evaluating to the geographic identifier.
	SpatialValue string
	// AnalyticalValue is an expression evaluating to the data value for coloring/sizing.
	AnalyticalValue string
	// LabelColumn is the data column used for map labels.
	LabelColumn string
}

// NewMapLayer creates a MapLayer with defaults.
func NewMapLayer() *MapLayer {
	return &MapLayer{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (l *MapLayer) BaseName() string { return "MapLayer" }

// TypeName returns "MapLayer".
func (l *MapLayer) TypeName() string { return "MapLayer" }

// Serialize writes MapLayer properties that differ from defaults.
func (l *MapLayer) Serialize(w report.Writer) error {
	if err := l.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if l.Shapefile != "" {
		w.WriteStr("Shapefile", l.Shapefile)
	}
	if l.Type != "" {
		w.WriteStr("Type", l.Type)
	}
	if l.DataSource != "" {
		w.WriteStr("DataSource", l.DataSource)
	}
	if l.Filter != "" {
		w.WriteStr("Filter", l.Filter)
	}
	if l.SpatialColumn != "" {
		w.WriteStr("SpatialColumn", l.SpatialColumn)
	}
	if l.SpatialValue != "" {
		w.WriteStr("SpatialValue", l.SpatialValue)
	}
	if l.AnalyticalValue != "" {
		w.WriteStr("AnalyticalValue", l.AnalyticalValue)
	}
	if l.LabelColumn != "" {
		w.WriteStr("LabelColumn", l.LabelColumn)
	}
	return nil
}

// Deserialize reads MapLayer properties from an FRX reader.
func (l *MapLayer) Deserialize(r report.Reader) error {
	if err := l.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	l.Shapefile = r.ReadStr("Shapefile", "")
	l.Type = r.ReadStr("Type", "")
	l.DataSource = r.ReadStr("DataSource", "")
	l.Filter = r.ReadStr("Filter", "")
	l.SpatialColumn = r.ReadStr("SpatialColumn", "")
	l.SpatialValue = r.ReadStr("SpatialValue", "")
	l.AnalyticalValue = r.ReadStr("AnalyticalValue", "")
	l.LabelColumn = r.ReadStr("LabelColumn", "")
	// Drain additional dot-notation attributes (BoxAsString, Palette, color ranges, etc.)
	_ = r.ReadStr("BoxAsString", "")
	_ = r.ReadStr("Palette", "")
	return nil
}

// MapObject renders a geographic map visualization. It contains one or more
// MapLayer objects that bind data to geographic shapes.
//
// It is the Go equivalent of FastReport.Map.MapObject.
// This stub supports FRX loading; actual map rendering is not yet implemented.
type MapObject struct {
	report.ReportComponentBase

	// Layers holds the ordered list of map layers.
	Layers []*MapLayer
	// OffsetX is the horizontal pan offset.
	OffsetX float32
	// OffsetY is the vertical pan offset.
	OffsetY float32
}

// NewMapObject creates a MapObject with defaults.
func NewMapObject() *MapObject {
	return &MapObject{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (m *MapObject) BaseName() string { return "Map" }

// TypeName returns "MapObject".
func (m *MapObject) TypeName() string { return "MapObject" }

// Serialize writes MapObject properties that differ from defaults.
func (m *MapObject) Serialize(w report.Writer) error {
	if err := m.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if m.OffsetX != 0 {
		w.WriteFloat("OffsetX", m.OffsetX)
	}
	if m.OffsetY != 0 {
		w.WriteFloat("OffsetY", m.OffsetY)
	}
	return nil
}

// Deserialize reads MapObject properties from an FRX reader.
func (m *MapObject) Deserialize(r report.Reader) error {
	if err := m.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	m.OffsetX = r.ReadFloat("OffsetX", 0)
	m.OffsetY = r.ReadFloat("OffsetY", 0)
	return nil
}

// DeserializeChild handles MapLayer child elements.
func (m *MapObject) DeserializeChild(childType string, r report.Reader) bool {
	if childType == "MapLayer" {
		layer := NewMapLayer()
		_ = layer.Deserialize(r)
		// Drain any grandchildren.
		for {
			_, ok := r.NextChild()
			if !ok {
				break
			}
			_ = r.FinishChild()
		}
		m.Layers = append(m.Layers, layer)
		return true
	}
	return false
}
