package typewriter

import (
	"fmt"
	"regexp"
	"strings"
)

type Type struct {
	Pointer                      Pointer
	Name                         string
	Tags                         TagSlice
	comparable, numeric, ordered bool
	test                         test
}

type test bool

// a convenience for using bool in file name, see WriteAll
func (t test) String() string {
	if t {
		return "_test"
	}
	return ""
}

func (t Type) String() (result string) {
	return fmt.Sprintf("%s%s", t.Pointer.String(), t.Name)
}

func (t Type) LongName() string {
	name := ""
	r := regexp.MustCompile(`[\[\]{}*]`)
	els := r.Split(t.String(), -1)
	for _, s := range els {
		name += strings.Title(s)
	}
	return name
}

// Pointer exists as a type to allow simple use as bool or as String, which returns *
type Pointer bool

func (p Pointer) String() string {
	if p {
		return "*"
	}
	return ""
}
