package typewriter

import (
	"fmt"

	"text/template"
)

// Template includes the text of a template as well as requirements for the types to which it can be applied.
// +gen * slice:"Where"
type Template struct {
	Name, Text     string
	TypeConstraint Constraint
	// Indicates both the number of required type parameters, and the constraints of each (if any)
	TypeParameterConstraints []Constraint
}

// TryTypeAndValue verifies that a given Type and TagValue satisfy a Template's type constraints.
func (tmpl *Template) TryTypeAndValue(t Type, v TagValue) error {
	if err := tmpl.TypeConstraint.tryType(t); err != nil {
		return fmt.Errorf("cannot implement %s on %s: %s", v, t, err)
	}

	if len(tmpl.TypeParameterConstraints) != len(v.TypeParameters) {
		return fmt.Errorf("%s requires %d type parameters", v.Name, len(tmpl.TypeParameterConstraints))
	}

	for i := range v.TypeParameters {
		c := tmpl.TypeParameterConstraints[i]
		tp := v.TypeParameters[i]
		if err := c.tryType(tp); err != nil {
			return fmt.Errorf("cannot implement %s on %s: %s", v, t, err)
		}
	}

	return nil
}

// Get attempts to locate a template which meets type constraints, and parses it.
func (ts TemplateSlice) Get(t Type, v TagValue) (*template.Template, error) {
	// a bit of poor-man's type resolution here

	// templates which might work
	candidates := ts.Where(func(tmpl *Template) bool {
		return tmpl.Name == v.Name
	})

	if len(candidates) == 0 {
		err := fmt.Errorf("%s is unknown", v.Name)
		return nil, err
	}

	// try to find one that meets type constraints
	for _, tmpl := range candidates {
		if err := tmpl.TryTypeAndValue(t, v); err == nil {
			// eagerly return on success
			return template.New(v.String()).Parse(tmpl.Text)
		}
	}

	// send back the first error message; not great but OK most of the time
	return nil, candidates[0].TryTypeAndValue(t, v)
}
