package engine

// processat.go implements a simplified deferred-processing mechanism.
// In the full FastReport engine, TextObjects with ProcessAt != Default are
// queued and processed when the matching EngineState event fires (e.g.
// PageFinished, ReportFinished).  Here we expose the EngineState type and a
// lightweight queue so that host code can hook into state transitions.

// EngineState identifies a lifecycle event emitted by the report engine.
type EngineState int

const (
	EngineStateReportStarted    EngineState = iota
	EngineStateReportFinished                // end of the whole report
	EngineStateReportPageStarted             // start of a report page template
	EngineStateReportPageFinished
	EngineStatePageStarted   // start of a physical page
	EngineStatePageFinished  // end of a physical page
	EngineStateColumnStarted // start of a column
	EngineStateColumnFinished
	EngineStateBlockStarted  // start of a data block
	EngineStateBlockFinished // end of a data block
	EngineStateGroupStarted
	EngineStateGroupFinished
)

// EngineStateHandler is a callback invoked when an EngineState event fires.
// sender is the engine object that triggered the state change (e.g. a BandBase).
type EngineStateHandler func(sender any, state EngineState)

// deferredItem holds a deferred handler and its processing trigger state.
// When repeat is true, the handler re-queues itself after firing so it fires
// on every occurrence of state (e.g. every PageFinished). One-shot handlers
// (repeat=false) are removed after the first matching event.
type deferredItem struct {
	state   EngineState
	handler func()
	repeat  bool
}

// OnStateChanged fires all registered deferred handlers that match state.
// It is called internally at lifecycle boundaries.
// Repeating handlers (repeat=true) remain in the queue after firing;
// one-shot handlers (repeat=false) are removed.
func (e *ReportEngine) OnStateChanged(sender any, state EngineState) {
	// Snapshot the current list so handlers added during execution
	// are not triggered in the same pass.
	current := e.deferredObjects
	e.deferredObjects = e.deferredObjects[:0]
	for _, item := range current {
		if item.state == state {
			item.handler()
			if item.repeat {
				// Re-queue repeating handler for next occurrence.
				e.deferredObjects = append(e.deferredObjects, item)
			}
		} else {
			e.deferredObjects = append(e.deferredObjects, item)
		}
	}

	// Notify external state listeners.
	for _, h := range e.stateHandlers {
		h(sender, state)
	}
}

// AddDeferredHandler registers a one-shot function to be called when a specific
// EngineState fires. The handler is removed after it fires once.
func (e *ReportEngine) AddDeferredHandler(state EngineState, fn func()) {
	e.deferredObjects = append(e.deferredObjects, deferredItem{state: state, handler: fn, repeat: false})
}

// AddRepeatingDeferredHandler registers a function that fires on every occurrence
// of state (e.g. every PageFinished or ColumnFinished). The handler persists in the
// queue until ClearDeferredHandlers or the report ends.
// Use this for ProcessAt=PageFinished and ProcessAt=ColumnFinished where the same
// handler must evaluate on every page/column boundary.
func (e *ReportEngine) AddRepeatingDeferredHandler(state EngineState, fn func()) {
	e.deferredObjects = append(e.deferredObjects, deferredItem{state: state, handler: fn, repeat: true})
}

// AddStateHandler registers a persistent callback invoked on every state change.
// Unlike AddDeferredHandler, it is not removed after firing.
func (e *ReportEngine) AddStateHandler(h EngineStateHandler) {
	e.stateHandlers = append(e.stateHandlers, h)
}

// ClearDeferredHandlers removes all pending deferred handlers.
func (e *ReportEngine) ClearDeferredHandlers() {
	e.deferredObjects = nil
}
