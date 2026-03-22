// Package json provides a JSON data source for go-fastreport.
// It is the Go equivalent of FastReport.Data.JsonConnection.
package json

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andrewloable/go-fastreport/data"
)

// JSONDataSource is a DataSource backed by a JSON file or string.
// It supports flat arrays of objects as rows, with nested objects
// exposed as column values.
//
// If the root JSON is an array, each element becomes a row.
// If the root JSON is an object, it is wrapped in a single-element array.
type JSONDataSource struct {
	data.BaseDataSource

	// sourcePath is the file path to read JSON from (if set).
	sourcePath string
	// sourceString is the raw JSON string, URL, or file path (if set).
	sourceString string
	// rootPath is a dot-separated path to the sub-array within the JSON.
	// e.g. "data.items" navigates obj["data"]["items"].
	rootPath string
	// commandTimeout is the HTTP request timeout in seconds (default 30).
	commandTimeout int
	// httpHeaders are extra headers sent with HTTP URL requests.
	httpHeaders map[string]string
}

// New creates a JSONDataSource with the given name.
func New(name string) *JSONDataSource {
	return &JSONDataSource{
		BaseDataSource: *data.NewBaseDataSource(name),
		commandTimeout: 30,
	}
}

// NewFromConnectionString creates a JSONDataSource populated from a
// FastReport JSON connection string.
// C# ref: JsonDataSourceConnection constructor / InitConnection.
func NewFromConnectionString(name, connectionString string) *JSONDataSource {
	j := New(name)
	j.InitFromConnectionString(connectionString)
	return j
}

// InitFromConnectionString re-initialises the data source from a FastReport
// JSON connection string, overriding any previously configured source.
func (j *JSONDataSource) InitFromConnectionString(cs string) {
	b := NewConnectionStringBuilder(cs)
	if v := b.Json(); v != "" {
		j.sourceString = v
		j.sourcePath = ""
	}
	// Encoding is informational only for now (JSON is typically UTF-8).
}

// SetFilePath sets the path to a JSON file as the data source.
func (j *JSONDataSource) SetFilePath(path string) { j.sourcePath = path }

// FilePath returns the JSON file path.
func (j *JSONDataSource) FilePath() string { return j.sourcePath }

// SetJSON sets a raw JSON string as the data source.
func (j *JSONDataSource) SetJSON(s string) { j.sourceString = s }

// JSON returns the raw JSON string.
func (j *JSONDataSource) JSON() string { return j.sourceString }

// SetRootPath sets the dot-separated path to the JSON sub-array.
// e.g. "data.items" navigates {"data":{"items":[...]}}
func (j *JSONDataSource) SetRootPath(path string) { j.rootPath = path }

// RootPath returns the dot-separated root path.
func (j *JSONDataSource) RootPath() string { return j.rootPath }

// SetCommandTimeout sets the HTTP request timeout in seconds.
func (j *JSONDataSource) SetCommandTimeout(seconds int) { j.commandTimeout = seconds }

// CommandTimeout returns the HTTP request timeout in seconds (default 30).
func (j *JSONDataSource) CommandTimeout() int { return j.commandTimeout }

// SetHTTPHeaders sets extra headers to include in HTTP URL requests.
func (j *JSONDataSource) SetHTTPHeaders(headers map[string]string) {
	j.httpHeaders = headers
}

// HTTPHeaders returns the configured HTTP headers.
func (j *JSONDataSource) HTTPHeaders() map[string]string { return j.httpHeaders }

// Init loads and parses the JSON, populating the in-memory row store.
func (j *JSONDataSource) Init() error {
	raw, err := j.readSource()
	if err != nil {
		return fmt.Errorf("JSONDataSource %q: %w", j.Name(), err)
	}

	var root any
	if err := json.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("JSONDataSource %q: parse error: %w", j.Name(), err)
	}

	// Navigate to rootPath if specified.
	if j.rootPath != "" {
		root, err = navigate(root, j.rootPath)
		if err != nil {
			return fmt.Errorf("JSONDataSource %q: rootPath %q: %w", j.Name(), j.rootPath, err)
		}
	}

	// Normalise to []any.
	var items []any
	switch v := root.(type) {
	case []any:
		items = v
	case map[string]any:
		items = []any{v}
	default:
		return fmt.Errorf("JSONDataSource %q: unsupported root type %T", j.Name(), root)
	}

	// Build columns from the union of all keys found in the first row.
	colSet := make(map[string]bool)
	var colOrder []string
	if len(items) > 0 {
		if firstObj, ok := items[0].(map[string]any); ok {
			for k := range firstObj {
				if !colSet[k] {
					colSet[k] = true
					colOrder = append(colOrder, k)
				}
			}
		}
	}
	j.BaseDataSource = *data.NewBaseDataSource(j.Name())
	for _, col := range colOrder {
		j.AddColumn(data.Column{Name: col, Alias: col, DataType: "any"})
	}

	// Load rows.
	for _, item := range items {
		row := make(map[string]any)
		if obj, ok := item.(map[string]any); ok {
			for k, v := range obj {
				row[k] = v
			}
		} else {
			// scalar value: expose as "_value" column.
			row["_value"] = item
		}
		j.AddRow(row)
	}

	return j.BaseDataSource.Init()
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// isJSONText returns true if s looks like inline JSON (starts with '{' or '[').
// C# ref: JsonDataSourceConnection.InitConnection — if (!(jsonText[0] == '{' || jsonText[0] == '['))
func isJSONText(s string) bool {
	t := strings.TrimSpace(s)
	return len(t) > 0 && (t[0] == '{' || t[0] == '[')
}

func (j *JSONDataSource) readSource() ([]byte, error) {
	if j.sourceString != "" {
		if isJSONText(j.sourceString) {
			return []byte(j.sourceString), nil
		}
		// Treat as URL if not inline JSON.
		client := &http.Client{Timeout: time.Duration(j.commandTimeout) * time.Second}
		req, err := http.NewRequest(http.MethodGet, j.sourceString, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid URL %q: %w", j.sourceString, err)
		}
		for k, v := range j.httpHeaders {
			req.Header.Set(k, v)
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("HTTP fetch %q: %w", j.sourceString, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("HTTP fetch %q: status %d", j.sourceString, resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	if j.sourcePath != "" {
		f, err := os.Open(j.sourcePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return io.ReadAll(f)
	}
	return nil, fmt.Errorf("no source configured (set FilePath or JSON)")
}

// navigate traverses a decoded JSON value following a dot-separated path.
func navigate(v any, path string) (any, error) {
	parts := strings.Split(path, ".")
	cur := v
	for _, part := range parts {
		if part == "" {
			continue
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected object at %q, got %T", part, cur)
		}
		next, exists := obj[part]
		if !exists {
			return nil, fmt.Errorf("key %q not found", part)
		}
		cur = next
	}
	return cur, nil
}
