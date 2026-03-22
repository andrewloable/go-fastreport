package xml

// connection_string_gaps_test.go — internal tests for the new setter/Build methods
// on the XML ConnectionStringBuilder.
//
// go-fastreport issue: go-fastreport-g7eo8
// C# ref: FastReport.Base/Data/XmlConnectionStringBuilder.cs

import (
	"testing"
)

// ── SetXmlFile ────────────────────────────────────────────────────────────────

func TestConnectionStringBuilder_SetXmlFile(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetXmlFile("/data/report.xml")
	if b.XmlFile() != "/data/report.xml" {
		t.Errorf("XmlFile = %q, want /data/report.xml", b.XmlFile())
	}
}

// ── SetXsdFile ────────────────────────────────────────────────────────────────

func TestConnectionStringBuilder_SetXsdFile(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetXsdFile("/data/schema.xsd")
	if b.XsdFile() != "/data/schema.xsd" {
		t.Errorf("XsdFile = %q, want /data/schema.xsd", b.XsdFile())
	}
}

// ── SetCodepage ───────────────────────────────────────────────────────────────

func TestConnectionStringBuilder_SetCodepage(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetCodepage(65001)
	if b.Codepage() != 65001 {
		t.Errorf("Codepage = %d, want 65001", b.Codepage())
	}
}

// ── Build ─────────────────────────────────────────────────────────────────────

func TestConnectionStringBuilder_Build_Empty(t *testing.T) {
	b := NewConnectionStringBuilder("")
	if cs := b.Build(); cs != "" {
		t.Errorf("Build on empty builder = %q, want empty", cs)
	}
}

func TestConnectionStringBuilder_Build_AllFields(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetXmlFile("/data/report.xml")
	b.SetXsdFile("/data/schema.xsd")
	b.SetCodepage(65001)
	cs := b.Build()
	if cs == "" {
		t.Fatal("Build() should not be empty")
	}
	// Round-trip parse and verify.
	b2 := NewConnectionStringBuilder(cs)
	if b2.XmlFile() != "/data/report.xml" {
		t.Errorf("round-trip XmlFile = %q", b2.XmlFile())
	}
	if b2.XsdFile() != "/data/schema.xsd" {
		t.Errorf("round-trip XsdFile = %q", b2.XsdFile())
	}
	if b2.Codepage() != 65001 {
		t.Errorf("round-trip Codepage = %d", b2.Codepage())
	}
}

func TestConnectionStringBuilder_Build_PartialFields(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetXmlFile("/data/report.xml")
	cs := b.Build()
	b2 := NewConnectionStringBuilder(cs)
	if b2.XmlFile() != "/data/report.xml" {
		t.Errorf("round-trip XmlFile = %q", b2.XmlFile())
	}
	// XsdFile not set — should be empty.
	if b2.XsdFile() != "" {
		t.Errorf("XsdFile should be empty, got %q", b2.XsdFile())
	}
}

func TestConnectionStringBuilder_Build_Roundtrip_FromParsed(t *testing.T) {
	// Parse an existing connection string, then build it back.
	cs := "XmlFile=/tmp/data.xml;XsdFile=/tmp/schema.xsd;Codepage=1252"
	b := NewConnectionStringBuilder(cs)
	built := b.Build()
	b2 := NewConnectionStringBuilder(built)
	if b2.XmlFile() != "/tmp/data.xml" {
		t.Errorf("XmlFile = %q", b2.XmlFile())
	}
	if b2.XsdFile() != "/tmp/schema.xsd" {
		t.Errorf("XsdFile = %q", b2.XsdFile())
	}
	if b2.Codepage() != 1252 {
		t.Errorf("Codepage = %d", b2.Codepage())
	}
}

func TestConnectionStringBuilder_SetXmlFile_Build_Canonicalkey(t *testing.T) {
	// Set via SetXmlFile → the built string must use "XmlFile" key.
	b := NewConnectionStringBuilder("")
	b.SetXmlFile("/foo.xml")
	cs := b.Build()
	// Should contain "XmlFile=".
	found := false
	for _, part := range splitParts(cs) {
		if len(part) >= 8 && part[:8] == "XmlFile=" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Build() = %q: expected XmlFile= key", cs)
	}
}

func splitParts(cs string) []string {
	var parts []string
	start := 0
	for i, c := range cs {
		if c == ';' {
			parts = append(parts, cs[start:i])
			start = i + 1
		}
	}
	if start < len(cs) {
		parts = append(parts, cs[start:])
	}
	return parts
}
