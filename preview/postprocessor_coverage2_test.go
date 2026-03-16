package preview

// postprocessor_coverage2_test.go — documents the structurally-unreachable
// branch in processDuplicates (postprocessor.go:99-101) and adds supplementary
// behavioural tests to maximise reachable coverage.
//
// Dead-code analysis:
//
//   postprocessor.go:99  `if len(entries) == 0 { continue }`
//
//   The `groups` map is populated only via:
//
//     groups[obj.Name] = append(groups[obj.Name], entry)
//
//   Since `append` always produces a slice of length ≥ 1, every value in
//   `groups` has len ≥ 1 by construction.  Therefore the `len(entries) == 0`
//   guard is dead code: it can never be true when iterating over the map.
//
//   The `continue` statement on line 100 is therefore unreachable.

import "testing"

// TestProcessDuplicates_LenEntriesZero_IsDeadCode is a documentation test
// confirming that the `if len(entries) == 0 { continue }` guard in
// processDuplicates (postprocessor.go:99-101) is structurally unreachable.
//
// The `groups` map is filled exclusively through:
//
//	groups[name] = append(groups[name], entry)
//
// The append call always produces a slice of length ≥ 1, so every map entry
// obtained via `for _, entries := range groups` has len(entries) ≥ 1.
// The dead guard was added defensively; it cannot fire in practice.
func TestProcessDuplicates_LenEntriesZero_IsDeadCode(t *testing.T) {
	t.Log("processDuplicates:99 `if len(entries) == 0 { continue }` is dead code " +
		"— groups map entries are always non-empty (built via append)")
}

// TestProcessDuplicates_MixedKindAndNameObjects exercises processDuplicates
// with a band that contains objects of various types:
//   - non-text kind with a name → skipped by `obj.Kind != ObjectTypeText`
//   - text kind without a name  → skipped by `obj.Name == ""`
//   - text kind, DuplicatesShow → skipped by `obj.Duplicates == DuplicatesShow`
//   - text kind, named, DuplicatesClear → added to groups → processGroup called
//
// This test confirms all three continue-paths within the inner loop are
// exercised in combination, and that only the qualifying object ends up in
// the groups map.
func TestProcessDuplicates_MixedKindAndNameObjects(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:   "mixed",
		Top:    0,
		Height: 80,
		Objects: []PreparedObject{
			// non-text kind, named — must NOT be added to groups.
			{Name: "pic", Kind: ObjectTypePicture, Top: 0, Height: 20,
				Duplicates: DuplicatesClear},
			// text kind, no name — must NOT be added to groups.
			{Name: "", Kind: ObjectTypeText, Top: 20, Height: 20, Text: "anon",
				Duplicates: DuplicatesClear},
			// text kind, named, DuplicatesShow — must NOT be added to groups.
			{Name: "shown", Kind: ObjectTypeText, Top: 40, Height: 20, Text: "v",
				Duplicates: DuplicatesShow},
			// Two qualifying duplicate objects.
			{Name: "dup", Kind: ObjectTypeText, Top: 0, Height: 20, Text: "X",
				Duplicates: DuplicatesClear},
			{Name: "dup", Kind: ObjectTypeText, Top: 20, Height: 20, Text: "X",
				Duplicates: DuplicatesClear},
		},
	})

	proc := &Postprocessor{pp: pp}
	proc.processDuplicates()

	objs := pp.GetPage(0).Bands[0].Objects
	// pic object unchanged.
	if objs[0].Kind != ObjectTypePicture {
		t.Error("picture object should be unchanged")
	}
	// anonymous text unchanged.
	if objs[1].Text != "anon" {
		t.Errorf("unnamed text = %q, want anon", objs[1].Text)
	}
	// shown object unchanged.
	if objs[2].Text != "v" {
		t.Errorf("DuplicatesShow text = %q, want v", objs[2].Text)
	}
	// duplicate pair: first kept, second cleared.
	if objs[3].Text != "X" {
		t.Errorf("first dup text = %q, want X", objs[3].Text)
	}
	if objs[4].Text != "" {
		t.Errorf("second dup text = %q, want empty (cleared)", objs[4].Text)
	}
}

// TestProcessDuplicates_MultipleGroups exercises processDuplicates when there
// are multiple distinct object names that each form independent duplicate groups.
// This confirms the outer `for _, entries := range groups` loop is exercised
// with multiple iterations (each with len(entries) ≥ 2).
func TestProcessDuplicates_MultipleGroups(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 80,
		Objects: []PreparedObject{
			// Group "alpha" — two adjacent same-text objects.
			{Name: "alpha", Kind: ObjectTypeText, Top: 0, Height: 20, Text: "A",
				Duplicates: DuplicatesClear},
			{Name: "alpha", Kind: ObjectTypeText, Top: 20, Height: 20, Text: "A",
				Duplicates: DuplicatesClear},
			// Group "beta" — two adjacent same-text objects.
			{Name: "beta", Kind: ObjectTypeText, Top: 40, Height: 20, Text: "B",
				Duplicates: DuplicatesClear},
			{Name: "beta", Kind: ObjectTypeText, Top: 60, Height: 20, Text: "B",
				Duplicates: DuplicatesClear},
		},
	})

	proc := &Postprocessor{pp: pp}
	proc.processDuplicates()

	objs := pp.GetPage(0).Bands[0].Objects
	// First of each group kept; second cleared.
	if objs[0].Text != "A" {
		t.Errorf("alpha[0].Text = %q, want A", objs[0].Text)
	}
	if objs[1].Text != "" {
		t.Errorf("alpha[1].Text = %q, want empty", objs[1].Text)
	}
	if objs[2].Text != "B" {
		t.Errorf("beta[0].Text = %q, want B", objs[2].Text)
	}
	if objs[3].Text != "" {
		t.Errorf("beta[1].Text = %q, want empty", objs[3].Text)
	}
}

// TestProcessDuplicates_SingletonGroup exercises processDuplicates when a
// named text object appears only once in the groups map (singleton group,
// len(entries) == 1).  processGroup is still called; no clearing should occur
// because a run of length 1 is never treated as a duplicate.
func TestProcessDuplicates_SingletonGroup(t *testing.T) {
	pp := New()
	pp.AddPage(595, 842, 1)
	_ = pp.AddBand(&PreparedBand{
		Name:   "b",
		Top:    0,
		Height: 20,
		Objects: []PreparedObject{
			{Name: "unique", Kind: ObjectTypeText, Top: 0, Height: 20, Text: "solo",
				Duplicates: DuplicatesClear},
		},
	})

	proc := &Postprocessor{pp: pp}
	proc.processDuplicates()

	if txt := pp.GetPage(0).Bands[0].Objects[0].Text; txt != "solo" {
		t.Errorf("singleton text = %q, want solo", txt)
	}
}
