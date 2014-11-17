package typewriter

import "testing"

func TestLongName(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"Foo", "Foo"},
		{"*Foo", "Foo"},
		{"map[Foo]Bar", "MapFooBar"},
		{"[]map[Foo]Bar", "SliceMapFooBar"},
		{"[]map[Foo]struct{}", "SliceMapFooStruct"},
	}

	for _, test := range tests {
		typ := Type{
			Name: test.input,
		}

		if typ.LongName() != test.expected {
			t.Errorf("expected %q, got %q", test.expected, typ.LongName())
		}
	}

}
