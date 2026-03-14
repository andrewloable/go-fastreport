package data

import "fmt"

// TotalEngine manages a set of AggregateTotals and accumulates values during
// report execution. It acts as the runtime registry for AggregateTotal
// instances and is the Go equivalent of FastReport's runtime total collection.
type TotalEngine struct {
	totals map[string]*AggregateTotal
	order  []string // preserve registration order
}

// NewTotalEngine creates a new, empty TotalEngine.
func NewTotalEngine() *TotalEngine {
	return &TotalEngine{
		totals: make(map[string]*AggregateTotal),
	}
}

// Register adds an AggregateTotal to the engine.
// If a total with the same name already exists, it is replaced.
func (te *TotalEngine) Register(t *AggregateTotal) {
	if _, exists := te.totals[t.Name]; !exists {
		te.order = append(te.order, t.Name)
	}
	te.totals[t.Name] = t
}

// Accumulate adds value to the named total.
// Returns an error when the total is not registered or value cannot be
// accumulated (e.g. non-numeric value for a Sum total).
func (te *TotalEngine) Accumulate(name string, value any) error {
	t, ok := te.totals[name]
	if !ok {
		return fmt.Errorf("TotalEngine: total %q not registered", name)
	}
	return t.Add(value)
}

// Value returns the current computed aggregate value for the named total.
// Returns nil when the total is not registered.
func (te *TotalEngine) Value(name string) any {
	t, ok := te.totals[name]
	if !ok {
		return nil
	}
	return t.Value()
}

// Reset resets the named total to its zero state.
// Returns an error when the total is not registered.
func (te *TotalEngine) Reset(name string) error {
	t, ok := te.totals[name]
	if !ok {
		return fmt.Errorf("TotalEngine: total %q not registered", name)
	}
	t.Reset()
	return nil
}

// ResetAll resets every registered total.
func (te *TotalEngine) ResetAll() {
	for _, t := range te.totals {
		t.Reset()
	}
}

// All returns all registered AggregateTotal objects in registration order.
func (te *TotalEngine) All() []*AggregateTotal {
	result := make([]*AggregateTotal, 0, len(te.order))
	for _, name := range te.order {
		result = append(result, te.totals[name])
	}
	return result
}

// Len returns the number of registered totals.
func (te *TotalEngine) Len() int { return len(te.totals) }

// Find returns the AggregateTotal with the given name, or nil if not found.
func (te *TotalEngine) Find(name string) *AggregateTotal {
	return te.totals[name]
}
