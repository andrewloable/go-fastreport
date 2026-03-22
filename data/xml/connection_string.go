package xml

import (
	"fmt"
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

// ── Setters ───────────────────────────────────────────────────────────────────

// SetXmlFile sets the path to the XML data file.
func (b *ConnectionStringBuilder) SetXmlFile(v string) { b.vals["xmlfile"] = v }

// SetXsdFile sets the path to the XSD schema file.
func (b *ConnectionStringBuilder) SetXsdFile(v string) { b.vals["xsdfile"] = v }

// SetCodepage sets the code page number.
func (b *ConnectionStringBuilder) SetCodepage(n int) { b.vals["codepage"] = fmt.Sprintf("%d", n) }

// Build serialises the builder state to a connection string.
func (b *ConnectionStringBuilder) Build() string {
	var sb strings.Builder
	write := func(canonical, key string) {
		if v, ok := b.vals[key]; ok {
			if sb.Len() > 0 {
				sb.WriteByte(';')
			}
			sb.WriteString(canonical)
			sb.WriteByte('=')
			sb.WriteString(v)
		}
	}
	write("XmlFile", "xmlfile")
	write("XsdFile", "xsdfile")
	write("Codepage", "codepage")
	return sb.String()
}
