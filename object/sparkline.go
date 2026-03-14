package object

import "github.com/andrewloable/go-fastreport/report"

// SparklineObject renders a small chart (sparkline) embedded in a band.
// The chart definition is stored as a base64-encoded XML string in ChartData.
//
// It is the Go equivalent of FastReport.SparklineObject.
type SparklineObject struct {
	report.ReportComponentBase

	// ChartData holds the base64-encoded chart XML.
	ChartData string
	// Dock specifies the docking style within the parent band.
	Dock string
}

// NewSparklineObject creates a SparklineObject with default settings.
func NewSparklineObject() *SparklineObject {
	return &SparklineObject{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (s *SparklineObject) BaseName() string { return "Sparkline" }

// TypeName returns "SparklineObject".
func (s *SparklineObject) TypeName() string { return "SparklineObject" }

// Serialize writes SparklineObject properties that differ from defaults.
func (s *SparklineObject) Serialize(w report.Writer) error {
	if err := s.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if s.ChartData != "" {
		w.WriteStr("ChartData", s.ChartData)
	}
	if s.Dock != "" {
		w.WriteStr("Dock", s.Dock)
	}
	return nil
}

// Deserialize reads SparklineObject properties from an FRX reader.
func (s *SparklineObject) Deserialize(r report.Reader) error {
	if err := s.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	s.ChartData = r.ReadStr("ChartData", "")
	s.Dock = r.ReadStr("Dock", "")
	return nil
}
