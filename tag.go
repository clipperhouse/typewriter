package typewriter

import "strings"

// +gen slice
type Tag struct {
	Name    string
	Values  []TagValue
	Negated bool
}

func (t Tag) String() string {
	return t.Name
}

type TagValue struct {
	Name           string
	TypeParameters []Type
}

func (v TagValue) String() string {
	if len(v.TypeParameters) == 0 {
		return v.Name
	}

	var a []string
	for i := 0; i < len(v.TypeParameters); i++ {
		a = append(a, v.TypeParameters[i].Name)
	}

	return v.Name + "[" + strings.Join(a, ",") + "]"
}
