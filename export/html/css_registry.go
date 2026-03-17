package html

import (
	"fmt"
	"strings"
)

// cssRegistry collects unique CSS rule strings and assigns short class names
// (.s0, .s1, …) to each.  The generated <style> block can be retrieved via
// StyleBlock() after all objects have been rendered.
type cssRegistry struct {
	index        map[string]string // css content → class name
	namedClasses map[string]string // explicit key → css content (for RegisterClass)
	order        []string          // insertion order of css content (for deterministic output)
}

func newCSSRegistry() *cssRegistry {
	return &cssRegistry{
		index:        make(map[string]string),
		namedClasses: make(map[string]string),
	}
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

// HasClass returns true if a named class with the given key has already been
// registered via RegisterClass.
func (r *cssRegistry) HasClass(key string) bool {
	_, ok := r.namedClasses[key]
	return ok
}

// RegisterClass registers css under an explicit class name (key).
// If key is already registered it is a no-op.
// The class name used in HTML will be exactly key.
func (r *cssRegistry) RegisterClass(key, css string) {
	if _, ok := r.namedClasses[key]; ok {
		return
	}
	r.namedClasses[key] = css
	r.index[css] = key
	r.order = append(r.order, css)
}

// StyleBlock returns a complete <style> element with all registered classes,
// or an empty string if no classes have been registered.
// The format matches C# FastReport output: type="text/css" with HTML comment wrapping
// and a leading p { margin-block-start/end: initial; } reset rule.
func (r *cssRegistry) StyleBlock() string {
	if len(r.order) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("<style type=\"text/css\"><!-- \n")
	sb.WriteString("p { margin-block-start: initial; margin-block-end: initial; }\n")
	for _, css := range r.order {
		name := r.index[css]
		sb.WriteString(fmt.Sprintf(".%s { %s}\n", name, css))
	}
	sb.WriteString("--></style>\n")
	return sb.String()
}
