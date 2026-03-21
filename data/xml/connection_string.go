package xml

import (
	"strconv"
	"strings"
)

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

func (b *ConnectionStringBuilder) XmlFile() string {
	if v, ok := b.get("XmlFile"); ok {
		return v
	}
	return ""
}

func (b *ConnectionStringBuilder) XsdFile() string {
	if v, ok := b.get("XsdFile"); ok {
		return v
	}
	return ""
}

func (b *ConnectionStringBuilder) Codepage() int {
	if v, ok := b.get("Codepage"); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return 0
}
