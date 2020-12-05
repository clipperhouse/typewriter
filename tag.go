package typewriter

// +gen slice
type Tag struct {
	Name    string
	Values  []TagValue
	Negated bool
}

type TagValue struct {
	Name           string
	TypeParameters []Type
	typeParameters []item
}

// AddDefaultsIfNeeded adds default tag values if either 1) none are
// specified or 2) if a * is present in the TagValues. The * is removed
// if it is found.
func (tag *Tag) AddDefaultsIfNeeded(typ Type, templates TemplateSlice) {
	if tag.ShouldAddDefaults() {
		defaults := MakeDefaultTagValues(typ, templates)
		tag.Values = append(tag.Values, defaults...)
	}
	tag.RemoveStars()
}

// ShouldAddDefaults tells whether the tag calls for adding
// default methods.
func (tag *Tag) ShouldAddDefaults() bool {
	if len(tag.Values) == 0 {
		return true
	} else {
		for _, tv := range tag.Values {
			if tv.Name == "*" {
				return true
			}
		}
	}
	return false
}

// RemoveStars removes the TagValues with a Name of "*".
func (tag *Tag) RemoveStars() {
	newValues := make([]TagValue, 0)
	for _, tv := range tag.Values {
		if tv.Name != "*" {
			newValues = append(newValues, tv)
		}
	}
	tag.Values = newValues
}

// MakeDefaultTagValues makes an array of TagValues naming templates
// compatible with typ that have no type parameters.
func MakeDefaultTagValues(typ Type, templates TemplateSlice) []TagValue {
	tagValues := make([]TagValue, 0)
	for _, tpl := range templates {
		err := tpl.TypeConstraint.TryType(typ)
		if err == nil && len(tpl.TypeParameterConstraints) == 0 {
			tv := TagValue{}
			tv.Name = tpl.Name
			tagValues = append(tagValues, tv)
		}
	}
	return tagValues
}
