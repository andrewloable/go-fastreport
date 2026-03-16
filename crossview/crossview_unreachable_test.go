package crossview

// crossview_unreachable_test.go — documents structurally unreachable error branches
// in serial.go that cannot be covered by tests.
//
// UNREACHABLE ANALYSIS
// ====================
//
// The following error-propagation branches in serial.go are structurally
// unreachable because the upstream Deserialize implementations always return nil:
//
//   serial.go:165 CrossViewHeader.Deserialize — 80% coverage
//     Lines 174–178: `if err := d.Deserialize(r); err != nil { ... }`
//     Reason: HeaderDescriptor.Deserialize (lines 48–60) only calls r.ReadBool,
//     r.ReadStr, r.ReadInt — none of which return errors — so it always returns nil.
//     The inner FinishChild-error break (line 176) and continue (line 178) are
//     therefore unreachable.
//
//   serial.go:232 CrossViewCells.Deserialize — 80% coverage
//     Lines 241–245: `if err := d.Deserialize(r); err != nil { ... }`
//     Reason: CellDescriptor.Deserialize (lines 108–120) only calls r.ReadBool,
//     r.ReadStr, r.ReadInt — always returns nil. Same reasoning as above.
//
//   serial.go:321 CrossViewDataSerial.Deserialize — 66.7% coverage
//     Lines 333–338 ("Columns" case): `if err := s.columnHeader.Deserialize(r); err != nil { ... }`
//     Lines 340–345 ("Rows" case):    `if err := s.rowHeader.Deserialize(r); err != nil { ... }`
//     Lines 347–352 ("Cells" case):   `if err := s.cells.Deserialize(r); err != nil { ... }`
//     Reason: CrossViewHeader.Deserialize and CrossViewCells.Deserialize always
//     return nil (their own error paths are also unreachable, as shown above).
//
// These branches exist as defensive error-handling boilerplate. They cannot be
// exercised through any public or internal API without modifying source code.
// Coverage for these functions is bounded at 80% / 66.7% respectively.
