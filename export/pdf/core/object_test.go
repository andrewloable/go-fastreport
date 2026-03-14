package core

import "testing"

func TestObjectTypeConstants(t *testing.T) {
	cases := []struct {
		name string
		got  ObjectType
		want ObjectType
	}{
		{"indirect", TypeIndirect, "indirect"},
		{"dictionary", TypeDictionary, "dictionary"},
		{"array", TypeArray, "array"},
		{"stream", TypeStream, "stream"},
		{"string", TypeString, "string"},
		{"name", TypeName, "name"},
		{"numeric", TypeNumeric, "numeric"},
		{"boolean", TypeBoolean, "boolean"},
		{"null", TypeNull, "null"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q want %q", tc.got, tc.want)
			}
		})
	}
}

// Verify all concrete types satisfy the Object interface at compile time.
var (
	_ Object = (*IndirectObject)(nil)
	_ Object = (*Dictionary)(nil)
	_ Object = (*Array)(nil)
	_ Object = (*Stream)(nil)
	_ Object = (*String)(nil)
	_ Object = (*Name)(nil)
	_ Object = (*Numeric)(nil)
	_ Object = (*Boolean)(nil)
	_ Object = (*Null)(nil)
)
