package json_test

// json_connection_gaps_test.go — tests for gaps in JsonDataSourceConnection:
// NewFromConnectionString, InitFromConnectionString, HTTP URL detection,
// SetCommandTimeout, SetHTTPHeaders.
//
// go-fastreport issue: go-fastreport-b5vlq
// C# ref: FastReport.Base/Data/JsonConnection/JsonDataSourceConnection.cs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	jsondata "github.com/andrewloable/go-fastreport/data/json"
)

// ── NewFromConnectionString ───────────────────────────────────────────────────

func TestNewFromConnectionString_InlineJSON(t *testing.T) {
	cs := `Json=[{"id":1},{"id":2}]`
	ds := jsondata.NewFromConnectionString("test", cs)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
}

func TestNewFromConnectionString_WithEncoding(t *testing.T) {
	cs := `Json=[{"id":1}];Encoding=utf-8`
	ds := jsondata.NewFromConnectionString("test", cs)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1", ds.RowCount())
	}
}

// ── InitFromConnectionString ──────────────────────────────────────────────────

func TestInitFromConnectionString_OverridesPreviousSource(t *testing.T) {
	ds := jsondata.New("test")
	ds.SetJSON(`[{"a":1},{"a":2},{"a":3}]`)
	// Now override with a new connection string.
	ds.InitFromConnectionString(`Json=[{"b":10}]`)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// Should use the new connection string source (1 row).
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1 (new source)", ds.RowCount())
	}
}

// ── HTTP URL fetching ─────────────────────────────────────────────────────────

// TestJSONDataSource_FetchURL verifies that when the Json field does not begin
// with '{' or '[', it is treated as an HTTP URL and fetched.
// C# ref: JsonDataSourceConnection.InitConnection — if (!(jsonText[0] == '{' || jsonText[0] == '['))
func TestJSONDataSource_FetchURL(t *testing.T) {
	const payload = `[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	ds := jsondata.New("remote")
	ds.SetJSON(srv.URL) // URL string — not JSON text
	if err := ds.Init(); err != nil {
		t.Fatalf("Init from URL: %v", err)
	}
	if ds.RowCount() != 2 {
		t.Errorf("RowCount = %d, want 2", ds.RowCount())
	}
	_ = ds.First()
	name, err := ds.GetValue("name")
	if err != nil {
		t.Fatalf("GetValue(name): %v", err)
	}
	if name != "Alice" {
		t.Errorf("name = %v, want Alice", name)
	}
}

func TestJSONDataSource_FetchURL_WithHeaders(t *testing.T) {
	const payload = `[{"ok":true}]`
	var receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	ds := jsondata.New("remote")
	ds.SetJSON(srv.URL)
	ds.SetHTTPHeaders(map[string]string{"Authorization": "Bearer mytoken"})
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if receivedAuth != "Bearer mytoken" {
		t.Errorf("Authorization header received = %q, want Bearer mytoken", receivedAuth)
	}
}

func TestJSONDataSource_FetchURL_BadURL(t *testing.T) {
	ds := jsondata.New("bad")
	ds.SetJSON("http://127.0.0.1:0/nonexistent") // port 0 — connection refused
	ds.SetCommandTimeout(2)
	err := ds.Init()
	if err == nil {
		t.Error("Init from bad URL should return error")
	}
}

func TestJSONDataSource_FetchURL_ViaConnectionString(t *testing.T) {
	const payload = `[{"val":42}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	// Connection string where Json is a URL (not JSON text).
	cs := "Json=" + srv.URL + ";Encoding=utf-8"
	ds := jsondata.NewFromConnectionString("remote", cs)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1", ds.RowCount())
	}
}

// ── SetCommandTimeout / CommandTimeout ────────────────────────────────────────

func TestJSONDataSource_CommandTimeout_Default(t *testing.T) {
	ds := jsondata.New("test")
	if ds.CommandTimeout() != 30 {
		t.Errorf("CommandTimeout = %d, want 30 (default)", ds.CommandTimeout())
	}
}

func TestJSONDataSource_SetCommandTimeout(t *testing.T) {
	ds := jsondata.New("test")
	ds.SetCommandTimeout(60)
	if ds.CommandTimeout() != 60 {
		t.Errorf("CommandTimeout = %d, want 60", ds.CommandTimeout())
	}
}

// ── SetHTTPHeaders / HTTPHeaders ──────────────────────────────────────────────

func TestJSONDataSource_SetHTTPHeaders(t *testing.T) {
	ds := jsondata.New("test")
	ds.SetHTTPHeaders(map[string]string{"Accept": "application/json"})
	h := ds.HTTPHeaders()
	if h["Accept"] != "application/json" {
		t.Errorf("Accept = %q, want application/json", h["Accept"])
	}
}

// ── isJSONText detection (via behaviour) ─────────────────────────────────────

// When the source string starts with '{' it should be treated as JSON text.
func TestJSONDataSource_SourceStartingWithBrace_TreatedAsJSON(t *testing.T) {
	ds := jsondata.New("obj")
	ds.SetJSON(`{"id":1,"name":"test"}`)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1", ds.RowCount())
	}
}

// When the source string starts with '[' it should be treated as JSON text.
func TestJSONDataSource_SourceStartingWithBracket_TreatedAsJSON(t *testing.T) {
	ds := jsondata.New("arr")
	ds.SetJSON(`[{"id":1}]`)
	if err := ds.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if ds.RowCount() != 1 {
		t.Errorf("RowCount = %d, want 1", ds.RowCount())
	}
}
