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
