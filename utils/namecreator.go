package utils

import "strconv"

// FastNameCreator generates unique component names efficiently.
// It works by tracking the highest numbered suffix seen for each base name.
//
// Usage:
//
//	creator := NewFastNameCreator(existingNames)
//	creator.CreateUniqueName(obj)
type FastNameCreator struct {
	baseNames map[string]int
}

// ObjectNamer is the minimal interface required by FastNameCreator.
type ObjectNamer interface {
	// BaseName returns the base name prefix (e.g. "Text" for "Text1").
	BaseName() string
	// Name returns the current object name.
	Name() string
	// SetName sets the object name.
	SetName(name string)
}

// NewFastNameCreator creates a FastNameCreator pre-seeded with the names of
// the provided objects so that generated names do not clash with existing ones.
func NewFastNameCreator(objects []ObjectNamer) *FastNameCreator {
	nc := &FastNameCreator{
		baseNames: make(map[string]int),
	}
	for _, obj := range objects {
		name := obj.Name()
		if name == "" {
			continue
		}
		baseName, num := splitBaseName(name)
		if num > 0 {
			if existing, ok := nc.baseNames[baseName]; !ok || num > existing {
				nc.baseNames[baseName] = num
			}
		} else if _, ok := nc.baseNames[name]; !ok {
			nc.baseNames[name] = 0
		}
	}
	return nc
}

// CreateUniqueName assigns a unique name to obj based on its BaseName().
// For example, if BaseName() == "Text" and names "Text1" and "Text2" exist,
// it will set the name to "Text3".
func (nc *FastNameCreator) CreateUniqueName(obj ObjectNamer) {
	base := obj.BaseName()
	num := nc.baseNames[base] + 1
	obj.SetName(base + strconv.Itoa(num))
	nc.baseNames[base] = num
}

// splitBaseName splits a name like "Text42" into ("Text", 42).
// If no trailing digits are found, returns (name, 0).
func splitBaseName(name string) (string, int) {
	i := len(name) - 1
	for i > 0 && name[i] >= '0' && name[i] <= '9' {
		i--
	}
	if i >= 0 && i < len(name)-1 {
		numStr := name[i+1:]
		num, err := strconv.Atoi(numStr)
		if err == nil && num > 0 {
			return name[:i+1], num
		}
	}
	return name, 0
}
