package data

import (
	"fmt"
	"strings"

	"github.com/andrewloable/go-fastreport/report"
)

// CommandParameterCollection is an ordered, named collection of CommandParameter
// objects. It is the Go equivalent of FastReport.Data.CommandParameterCollection.
type CommandParameterCollection struct {
	items []*CommandParameter
}

// NewCommandParameterCollection creates an empty collection.
func NewCommandParameterCollection() *CommandParameterCollection {
	return &CommandParameterCollection{}
}

// Add appends a parameter to the collection.
func (c *CommandParameterCollection) Add(p *CommandParameter) {
	c.items = append(c.items, p)
}

// Remove removes a parameter by reference.
func (c *CommandParameterCollection) Remove(p *CommandParameter) {
	for i, v := range c.items {
		if v == p {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return
		}
	}
}

// Count returns the number of parameters.
func (c *CommandParameterCollection) Count() int { return len(c.items) }

// Get returns the parameter at index i.
func (c *CommandParameterCollection) Get(i int) *CommandParameter { return c.items[i] }

// All returns a copy of the internal slice.
func (c *CommandParameterCollection) All() []*CommandParameter {
	out := make([]*CommandParameter, len(c.items))
	copy(out, c.items)
	return out
}

// FindByName returns the parameter with the given name, or nil if not found.
func (c *CommandParameterCollection) FindByName(name string) *CommandParameter {
	for _, v := range c.items {
		if strings.EqualFold(v.Name, name) {
			return v
		}
	}
	return nil
}

// CreateUniqueName returns a unique parameter name based on name. If a
// parameter named name already exists, a numeric suffix is appended until
// the name is unique.
func (c *CommandParameterCollection) CreateUniqueName(name string) string {
	base := name
	for i := 1; c.FindByName(name) != nil; i++ {
		name = fmt.Sprintf("%s%d", base, i)
	}
	return name
}

// Serialize writes all parameters to w as child elements named "Parameter".
func (c *CommandParameterCollection) Serialize(w report.Writer) error {
	for _, p := range c.items {
		if err := w.WriteObjectNamed("Parameter", p); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads all "Parameter" children from r into the collection.
func (c *CommandParameterCollection) Deserialize(r report.Reader) error {
	for {
		typeName, ok := r.NextChild()
		if !ok {
			break
		}
		if typeName == "Parameter" {
			p := NewCommandParameter("")
			if err := p.Deserialize(r); err != nil {
				return err
			}
			c.Add(p)
		}
		if err := r.FinishChild(); err != nil {
			return err
		}
	}
	return nil
}
