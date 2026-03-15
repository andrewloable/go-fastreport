package html

import (
	"fmt"
	"strings"
)

// cssRegistry collects unique CSS rule strings and assigns short class names
// (.s0, .s1, …) to each.  The generated <style> block can be retrieved via
// StyleBlock() after all objects have been rendered.
type cssRegistry struct {
	index map[string]string // css content → class name
	order []string          // insertion order (for deterministic output)
}

func newCSSRegistry() *cssRegistry {
	return &cssRegistry{index: make(map[string]string)}
}

// Register returns the CSS class name for css.  If css has not been seen
// before it is registered and a new name is allocated.  Empty css returns "".
func (r *cssRegistry) Register(css string) string {
	if css == "" {
		return ""
	}
	if name, ok := r.index[css]; ok {
		return name
	}
	name := fmt.Sprintf("s%d", len(r.order))
	r.index[css] = name
	r.order = append(r.order, css)
	return name
}

// StyleBlock returns a complete <style> element with all registered classes,
// or an empty string if no classes have been registered.
func (r *cssRegistry) StyleBlock() string {
	if len(r.order) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("<style>\n")
	for _, css := range r.order {
		name := r.index[css]
		sb.WriteString(fmt.Sprintf(".%s{%s}\n", name, css))
	}
	sb.WriteString("</style>\n")
	return sb.String()
}
