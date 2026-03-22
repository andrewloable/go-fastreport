package json

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

var base64Re = regexp.MustCompile(`^([A-Za-z0-9+\/]{4})*([A-Za-z0-9+\/]{3}=|[A-Za-z0-9+\/]{2}==)?$`)

type ConnectionStringBuilder struct {
	vals map[string]string
}

func NewConnectionStringBuilder(connectionString string) *ConnectionStringBuilder {
	b := &ConnectionStringBuilder{vals: make(map[string]string)}
	b.parse(connectionString)
	return b
}

func (b *ConnectionStringBuilder) parse(cs string) {
	for _, part := range strings.Split(cs, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.IndexByte(part, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(part[:idx])
		val := strings.TrimSpace(part[idx+1:])
		b.vals[strings.ToLower(key)] = val
	}
}

func (b *ConnectionStringBuilder) get(key string) (string, bool) {
	v, ok := b.vals[strings.ToLower(key)]
	return v, ok
}

func decodeIfBase64(s string) string {
	if base64Re.MatchString(s) {
		if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
			return string(decoded)
		}
	}
	return s
}

func (b *ConnectionStringBuilder) Json() string {
	if v, ok := b.get("Json"); ok && v != "" {
		return decodeIfBase64(v)
	}
	return ""
}

func (b *ConnectionStringBuilder) JsonSchema() string {
	if v, ok := b.get("JsonSchema"); ok && v != "" {
		return decodeIfBase64(v)
	}
	return ""
}

func (b *ConnectionStringBuilder) Encoding() string {
	if v, ok := b.get("Encoding"); ok {
		return v
	}
	return ""
}

func (b *ConnectionStringBuilder) Headers() map[string]string {
	headers := make(map[string]string)
	for i := 0; ; i++ {
		key := fmt.Sprintf("header%d", i)
		v, ok := b.vals[key]
		if !ok {
			break
		}
		if strings.TrimSpace(v) == "" {
			continue
		}
		parts := strings.SplitN(v, ":", 2)
		if len(parts) != 2 {
			continue
		}
		headerKey := decodeIfBase64(parts[0])
		headerVal := decodeIfBase64(parts[1])
		headers[headerKey] = headerVal
	}
	return headers
}

func (b *ConnectionStringBuilder) SimpleStructure() bool {
	if v, ok := b.get("SimpleStructure"); ok {
		return strings.ToLower(v) == "true"
	}
	return false
}

// ── Setters ───────────────────────────────────────────────────────────────────

// SetJson sets the JSON data (stores as base64 if it contains semicolons/equals).
func (b *ConnectionStringBuilder) SetJson(v string) {
	b.vals["json"] = base64.StdEncoding.EncodeToString([]byte(v))
}

// SetJsonSchema sets the JSON schema.
func (b *ConnectionStringBuilder) SetJsonSchema(v string) {
	b.vals["jsonschema"] = base64.StdEncoding.EncodeToString([]byte(v))
}

// SetEncoding sets the character encoding name.
func (b *ConnectionStringBuilder) SetEncoding(v string) {
	b.vals["encoding"] = v
}

// SetSimpleStructure sets the SimpleStructure flag.
func (b *ConnectionStringBuilder) SetSimpleStructure(v bool) {
	if v {
		b.vals["simplestructure"] = "true"
	} else {
		b.vals["simplestructure"] = "false"
	}
}

// SetHeaders replaces the HTTP header collection.
func (b *ConnectionStringBuilder) SetHeaders(headers map[string]string) {
	// Remove existing header keys.
	for k := range b.vals {
		if strings.HasPrefix(k, "header") {
			delete(b.vals, k)
		}
	}
	i := 0
	for k, v := range headers {
		key := fmt.Sprintf("header%d", i)
		b.vals[key] = base64.StdEncoding.EncodeToString([]byte(k)) + ":" + base64.StdEncoding.EncodeToString([]byte(v))
		i++
	}
}

// Build serialises the builder state to a connection string.
func (b *ConnectionStringBuilder) Build() string {
	var sb strings.Builder
	// Write in deterministic order: json, jsonschema, encoding, simplestructure, headers.
	keys := []string{"json", "jsonschema", "encoding", "simplestructure"}
	for _, k := range keys {
		if v, ok := b.vals[k]; ok {
			if sb.Len() > 0 {
				sb.WriteByte(';')
			}
			// Use canonical capitalised key names.
			canonical := map[string]string{
				"json": "Json", "jsonschema": "JsonSchema",
				"encoding": "Encoding", "simplestructure": "SimpleStructure",
			}
			sb.WriteString(canonical[k])
			sb.WriteByte('=')
			sb.WriteString(v)
		}
	}
	// Append headers.
	for k, v := range b.vals {
		if strings.HasPrefix(k, "header") {
			if sb.Len() > 0 {
				sb.WriteByte(';')
			}
			sb.WriteString(k)
			sb.WriteByte('=')
			sb.WriteString(v)
		}
	}
	return sb.String()
}
