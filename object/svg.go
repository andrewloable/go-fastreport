package object

import "github.com/andrewloable/go-fastreport/report"

// SVGObject renders a scalable vector graphic. The SVG content is stored
// as a base64-encoded string in SvgData, matching the FastReport .NET
// SVGObject (FastReport.SVGObject).
type SVGObject struct {
	report.ReportComponentBase

	// SvgData holds the base64-encoded SVG XML.
	SvgData string
}

// NewSVGObject creates an SVGObject with defaults.
func NewSVGObject() *SVGObject {
	return &SVGObject{
		ReportComponentBase: *report.NewReportComponentBase(),
	}
}

// BaseName returns the base name prefix for auto-generated names.
func (s *SVGObject) BaseName() string { return "SVG" }

// TypeName returns "SVGObject".
func (s *SVGObject) TypeName() string { return "SVGObject" }

// Serialize writes SVGObject properties that differ from defaults.
func (s *SVGObject) Serialize(w report.Writer) error {
	if err := s.ReportComponentBase.Serialize(w); err != nil {
		return err
	}
	if s.SvgData != "" {
		w.WriteStr("SvgData", s.SvgData)
	}
	return nil
}

// Deserialize reads SVGObject properties from an FRX reader.
func (s *SVGObject) Deserialize(r report.Reader) error {
	if err := s.ReportComponentBase.Deserialize(r); err != nil {
		return err
	}
	s.SvgData = r.ReadStr("SvgData", "")
	return nil
}
