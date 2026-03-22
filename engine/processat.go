package engine

import "github.com/andrewloable/go-fastreport/object"

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
// senderPred, when non-nil, is evaluated against the sender argument of
// OnStateChanged; the handler only fires when senderPred returns true.
// This mirrors the C# ProcessInfo.Process(sender, state) sender-check logic
// for DataFinished / GroupFinished state events (ProcessAt.cs lines 115-137).
type deferredItem struct {
	state      EngineState
	handler    func()
	repeat     bool
	senderPred func(any) bool // nil = always match sender
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
		senderOK := item.senderPred == nil || item.senderPred(sender)
		if item.state == state && senderOK {
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

// RegisterCustomObject registers a TextObject with ProcessAtCustom for later
// manual processing via ProcessObject. The fn closure evaluates the object's
// text expression and updates the corresponding PreparedObject.
// Called from populateBandObjects2 when ProcessAt == ProcessAtCustom.
// Mirrors C# AddObjectToProcess (ProcessAt.cs line 199-205).
func (e *ReportEngine) RegisterCustomObject(txt *object.TextObject, fn func()) {
	if e.customObjects == nil {
		e.customObjects = make(map[*object.TextObject]func())
	}
	e.customObjects[txt] = fn
}

// ProcessObject manually triggers deferred evaluation for a TextObject with
// ProcessAt == Custom. This is the Go equivalent of C# ReportEngine.ProcessObject
// (ProcessAt.cs lines 216-228). After processing, the object is removed from
// the custom queue so it is not processed again.
func (e *ReportEngine) ProcessObject(txt *object.TextObject) {
	if e.customObjects == nil {
		return
	}
	fn, ok := e.customObjects[txt]
	if !ok {
		return
	}
	fn()
	delete(e.customObjects, txt)
}

// AddSenderDeferredHandler registers a one-shot deferred handler that only
// fires when the state matches AND senderPred returns true for the event's
// sender. Used for ProcessAtDataFinished and ProcessAtGroupFinished to match
// the specific DataBand or GroupHeaderBand that triggered the state transition.
// Mirrors C# ProcessInfo.Process(object sender, EngineState state) lines 110-154.
func (e *ReportEngine) AddSenderDeferredHandler(state EngineState, senderPred func(any) bool, fn func()) {
	e.deferredObjects = append(e.deferredObjects, deferredItem{
		state: state, handler: fn, senderPred: senderPred,
	})
}

// AddMultipleHandler registers a one-shot handler that fires when any of the
// supplied states is reached. After the first matching state event fires, the
// handler is removed (one-shot semantics across all listed states).
// Mirrors the C# concept of ProcessInfo where a TextObject with ProcessAt set
// can match multiple engine states before deciding to fire. In Go, this is
// implemented by registering separate one-shot deferredItems for each state
// and using a shared fired flag so only the first match triggers the handler.
// Passing an empty states slice is a no-op.
func (e *ReportEngine) AddMultipleHandler(states []EngineState, fn func()) {
	if len(states) == 0 {
		return
	}
	fired := false
	wrapper := func() {
		if fired {
			return
		}
		fired = true
		fn()
	}
	for _, s := range states {
		e.deferredObjects = append(e.deferredObjects, deferredItem{state: s, handler: wrapper, repeat: false})
	}
}
