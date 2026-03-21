package csv

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

func (b *ConnectionStringBuilder) CsvFile() string {
	if v, ok := b.get("CsvFile"); ok {
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

func (b *ConnectionStringBuilder) Separator() string {
	if v, ok := b.get("Separator"); ok {
		return v
	}
	return ";"
}

func (b *ConnectionStringBuilder) FieldNamesInFirstString() bool {
	if v, ok := b.get("FieldNamesInFirstString"); ok {
		return strings.ToLower(v) == "true"
	}
	return false
}

func (b *ConnectionStringBuilder) RemoveQuotationMarks() bool {
	if v, ok := b.get("RemoveQuotationMarks"); ok {
		return strings.ToLower(v) == "true"
	}
	return true
}

func (b *ConnectionStringBuilder) ConvertFieldTypes() bool {
	if v, ok := b.get("ConvertFieldTypes"); ok {
		return strings.ToLower(v) == "true"
	}
	return true
}

func (b *ConnectionStringBuilder) NumberFormat() string {
	if v, ok := b.get("NumberFormat"); ok {
		return v
	}
	return ""
}

func (b *ConnectionStringBuilder) CurrencyFormat() string {
	if v, ok := b.get("CurrencyFormat"); ok {
		return v
	}
	return ""
}

func (b *ConnectionStringBuilder) DateTimeFormat() string {
	if v, ok := b.get("DateTimeFormat"); ok {
		return v
	}
	return ""
}
