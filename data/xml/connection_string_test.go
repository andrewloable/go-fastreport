package xml

// connection_string_test.go — tests for ConnectionStringBuilder.
// Uses the xml package directly (no _test suffix) to access unexported parse/get.

import (
	"testing"
)

func TestNewConnectionStringBuilder_Empty(t *testing.T) {
	b := NewConnectionStringBuilder("")
	if b.XmlFile() != "" {
		t.Errorf("XmlFile = %q, want empty", b.XmlFile())
	}
	if b.XsdFile() != "" {
		t.Errorf("XsdFile = %q, want empty", b.XsdFile())
	}
	if b.Codepage() != 0 {
		t.Errorf("Codepage = %d, want 0", b.Codepage())
	}
}

func TestNewConnectionStringBuilder_FullString(t *testing.T) {
	cs := "XmlFile=/data/test.xml;XsdFile=/data/schema.xsd;Codepage=65001"
	b := NewConnectionStringBuilder(cs)

	if b.XmlFile() != "/data/test.xml" {
		t.Errorf("XmlFile = %q, want /data/test.xml", b.XmlFile())
	}
	if b.XsdFile() != "/data/schema.xsd" {
		t.Errorf("XsdFile = %q, want /data/schema.xsd", b.XsdFile())
	}
	if b.Codepage() != 65001 {
		t.Errorf("Codepage = %d, want 65001", b.Codepage())
	}
}

func TestConnectionStringBuilder_XmlCaseInsensitiveKeys(t *testing.T) {
	cs := "XMLFILE=/tmp/data.xml;XSDFILE=/tmp/schema.xsd;CODEPAGE=1252"
	b := NewConnectionStringBuilder(cs)
	if b.XmlFile() != "/tmp/data.xml" {
		t.Errorf("XmlFile = %q, want /tmp/data.xml", b.XmlFile())
	}
	if b.XsdFile() != "/tmp/schema.xsd" {
		t.Errorf("XsdFile = %q, want /tmp/schema.xsd", b.XsdFile())
	}
	if b.Codepage() != 1252 {
		t.Errorf("Codepage = %d, want 1252", b.Codepage())
	}
}

func TestConnectionStringBuilder_XmlCodepage_NonNumeric(t *testing.T) {
	cs := "Codepage=notanumber"
	b := NewConnectionStringBuilder(cs)
	if b.Codepage() != 0 {
		t.Errorf("Codepage for non-numeric = %d, want 0", b.Codepage())
	}
}

func TestConnectionStringBuilder_XmlParse_SkipsEmptyParts(t *testing.T) {
	cs := ";;XmlFile=/tmp/x.xml;;"
	b := NewConnectionStringBuilder(cs)
	if b.XmlFile() != "/tmp/x.xml" {
		t.Errorf("XmlFile = %q, want /tmp/x.xml", b.XmlFile())
	}
}

func TestConnectionStringBuilder_XmlParse_SkipsNoEquals(t *testing.T) {
	cs := "NoEqualsSign;XmlFile=/tmp/y.xml"
	b := NewConnectionStringBuilder(cs)
	if b.XmlFile() != "/tmp/y.xml" {
		t.Errorf("XmlFile = %q, want /tmp/y.xml", b.XmlFile())
	}
}

func TestConnectionStringBuilder_XmlFile_Only(t *testing.T) {
	cs := "XmlFile=C:\\reports\\data.xml"
	b := NewConnectionStringBuilder(cs)
	if b.XmlFile() != `C:\reports\data.xml` {
		t.Errorf("XmlFile = %q, want C:\\reports\\data.xml", b.XmlFile())
	}
	// XsdFile not in string, should be empty.
	if b.XsdFile() != "" {
		t.Errorf("XsdFile = %q, want empty", b.XsdFile())
	}
}
