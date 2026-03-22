package json

// connection_string_gaps_test.go — internal tests for the new setter/Build methods
// on ConnectionStringBuilder, covering the gaps identified in go-fastreport-e65ss.
//
// Tests are in the "json" package (no _test suffix) so they can access unexported
// helpers like parse/get/set.

import (
	"testing"
)

// ── SetJson / Json round-trip ─────────────────────────────────────────────────

func TestConnectionStringBuilder_SetJson_Roundtrip(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetJson(`[{"a":1}]`)
	if got := b.Json(); got != `[{"a":1}]` {
		t.Errorf("Json() = %q, want [...]", got)
	}
}

func TestConnectionStringBuilder_SetJson_Build(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetJson(`[{"x":1}]`)
	cs := b.Build()
	// Connection string must contain Json= key.
	if cs == "" {
		t.Error("Build() returned empty string after SetJson")
	}
	// Round-trip: parse the built string and read back.
	b2 := NewConnectionStringBuilder(cs)
	if got := b2.Json(); got != `[{"x":1}]` {
		t.Errorf("round-trip Json = %q, want [...]", got)
	}
}

// ── SetJsonSchema ─────────────────────────────────────────────────────────────

func TestConnectionStringBuilder_SetJsonSchema_Roundtrip(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetJsonSchema(`{"type":"array"}`)
	if got := b.JsonSchema(); got != `{"type":"array"}` {
		t.Errorf("JsonSchema() = %q, want {...}", got)
	}
}

// ── SetEncoding ───────────────────────────────────────────────────────────────

func TestConnectionStringBuilder_SetEncoding_Roundtrip(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetEncoding("utf-8")
	if got := b.Encoding(); got != "utf-8" {
		t.Errorf("Encoding() = %q, want utf-8", got)
	}
}

// ── SetSimpleStructure ────────────────────────────────────────────────────────

func TestConnectionStringBuilder_SetSimpleStructure_True(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetSimpleStructure(true)
	if !b.SimpleStructure() {
		t.Error("SimpleStructure() should be true after SetSimpleStructure(true)")
	}
}

func TestConnectionStringBuilder_SetSimpleStructure_False(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetSimpleStructure(false)
	if b.SimpleStructure() {
		t.Error("SimpleStructure() should be false after SetSimpleStructure(false)")
	}
}

// ── SetHeaders ────────────────────────────────────────────────────────────────

func TestConnectionStringBuilder_SetHeaders_Basic(t *testing.T) {
	// Use header values that do not match the base64 regex (contain spaces or
	// hyphens which are not in [A-Za-z0-9+/]) so they are stored and returned
	// verbatim rather than being decoded.
	// C# ref: JsonDataSourceConnectionStringBuilder.Headers getter — values
	// matching the base64 pattern are decoded; values with spaces/hyphens are not.
	b := NewConnectionStringBuilder("")
	b.SetHeaders(map[string]string{
		"Authorization": "Bearer my token here",  // space → not base64
		"X-Custom":      "plain value with space", // space → not base64
	})
	headers := b.Headers()
	if headers["Authorization"] != "Bearer my token here" {
		t.Errorf("Authorization = %q, want 'Bearer my token here'", headers["Authorization"])
	}
	if headers["X-Custom"] != "plain value with space" {
		t.Errorf("X-Custom = %q, want 'plain value with space'", headers["X-Custom"])
	}
}

func TestConnectionStringBuilder_SetHeaders_KeyWithColon(t *testing.T) {
	// Keys containing ':' must be base64-encoded in storage and decoded on read.
	// The value "plain text" has a space so it will NOT be base64-decoded on read.
	b := NewConnectionStringBuilder("")
	b.SetHeaders(map[string]string{
		"X-Weird:Key": "plain text value",
	})
	headers := b.Headers()
	if headers["X-Weird:Key"] != "plain text value" {
		t.Errorf("header with colon key = %v, want 'plain text value'", headers["X-Weird:Key"])
	}
}

func TestConnectionStringBuilder_SetHeaders_ValWithColon(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetHeaders(map[string]string{
		"Authorization": "scheme:credential",
	})
	headers := b.Headers()
	if headers["Authorization"] != "scheme:credential" {
		t.Errorf("header val with colon = %q, want scheme:credential", headers["Authorization"])
	}
}

func TestConnectionStringBuilder_SetHeaders_NilClears(t *testing.T) {
	b := NewConnectionStringBuilder("")
	b.SetHeaders(map[string]string{"Key": "val"})
	b.SetHeaders(nil)
	headers := b.Headers()
	if len(headers) != 0 {
		t.Errorf("SetHeaders(nil) should clear headers, got %v", headers)
	}
}

func TestConnectionStringBuilder_SetHeaders_Roundtrip_ViaBuild(t *testing.T) {
	// Use a header value with a space so it is not base64-decoded on read.
	b := NewConnectionStringBuilder("")
	b.SetHeaders(map[string]string{"Accept": "text plain"}) // space: not base64
	cs := b.Build()
	b2 := NewConnectionStringBuilder(cs)
	h2 := b2.Headers()
	if h2["Accept"] != "text plain" {
		t.Errorf("headers round-trip Accept = %q, want 'text plain'", h2["Accept"])
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
	b.SetJson(`[{"a":1}]`)
	b.SetJsonSchema(`{"type":"array"}`)
	b.SetEncoding("utf-8")
	b.SetSimpleStructure(true)
	// Use a header value with a space so it is not mistaken for base64.
	b.SetHeaders(map[string]string{"Accept": "text plain"})

	cs := b.Build()
	if cs == "" {
		t.Fatal("Build() should not be empty")
	}
	// Parse back and verify each field.
	b2 := NewConnectionStringBuilder(cs)
	if b2.Json() != `[{"a":1}]` {
		t.Errorf("round-trip Json = %q", b2.Json())
	}
	if b2.JsonSchema() != `{"type":"array"}` {
		t.Errorf("round-trip JsonSchema = %q", b2.JsonSchema())
	}
	if b2.Encoding() != "utf-8" {
		t.Errorf("round-trip Encoding = %q", b2.Encoding())
	}
	if !b2.SimpleStructure() {
		t.Error("round-trip SimpleStructure should be true")
	}
	h2 := b2.Headers()
	if h2["Accept"] != "text plain" {
		t.Errorf("round-trip headers Accept = %q", h2["Accept"])
	}
}

// ── Parse existing connection strings ─────────────────────────────────────────

func TestConnectionStringBuilder_ParseUrl(t *testing.T) {
	cs := "Json=https://example.com/data.json;Encoding=utf-8"
	b := NewConnectionStringBuilder(cs)
	if b.Json() != "https://example.com/data.json" {
		t.Errorf("Json() = %q, want URL", b.Json())
	}
	if b.Encoding() != "utf-8" {
		t.Errorf("Encoding() = %q, want utf-8", b.Encoding())
	}
}

func TestConnectionStringBuilder_ParseMultipleHeaders(t *testing.T) {
	// Use header values that don't match the base64 regex (have spaces).
	// C# ref: the base64 regex only matches [A-Za-z0-9+/] with optional padding;
	// spaces are not in the base64 alphabet so these values are returned verbatim.
	cs := "Header0=Authorization:Bearer my token;Header1=X-Custom:plain value"
	b := NewConnectionStringBuilder(cs)
	h := b.Headers()
	if h["Authorization"] != "Bearer my token" {
		t.Errorf("Authorization = %q", h["Authorization"])
	}
	if h["X-Custom"] != "plain value" {
		t.Errorf("X-Custom = %q", h["X-Custom"])
	}
}
