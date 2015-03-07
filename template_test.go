package typewriter

import (
	"bytes"
	"strings"
	"testing"
)

func TestTryTypeAndValue(t *testing.T) {
	// no constraints
	tmpl1 := Template{
		Name: "TestTemplate",
	}

	typ1 := Type{
		Name: "TestType",
	}

	v1 := TagValue{
		Name: "TestValue",
	}

	err1 := tmpl1.TryTypeAndValue(typ1, v1)

	if err1 != nil {
		t.Error(err1)
	}

	// type constraint ok
	tmpl2 := Template{
		Name:           "TestTemplate",
		TypeConstraint: Constraint{Numeric: true},
	}

	typ2 := Type{
		Name:    "TestType",
		numeric: true,
	}

	v2 := TagValue{
		Name: "TestValue",
	}

	err2 := tmpl2.TryTypeAndValue(typ2, v2)

	if err2 != nil {
		t.Error(err2)
	}

	// type constraint not ok
	tmpl3 := Template{
		Name:           "TestTemplate",
		TypeConstraint: Constraint{Numeric: true},
	}

	typ3 := Type{
		Name:    "TestType",
		ordered: true,
	}

	v3 := TagValue{
		Name: "TestValue",
	}

	err3 := tmpl3.TryTypeAndValue(typ3, v3)

	if err3 == nil {
		t.Error("TryTypeAndValue should have returned an error")
	}

	// type parameter constraints ok
	tmpl4 := Template{
		Name: "TestTemplate",
		TypeParameterConstraints: []Constraint{
			{Numeric: true},
			{Ordered: true},
		}}

	typ4 := Type{
		Name: "TestType",
	}

	v4 := TagValue{
		Name: "TestValue",
		TypeParameters: []Type{
			{Name: "foo", numeric: true},
			{Name: "bar", ordered: true},
		},
	}

	err4 := tmpl4.TryTypeAndValue(typ4, v4)

	if err4 != nil {
		t.Error(err4)
	}

	// type parameter constraints not ok (by number)
	tmpl5 := Template{
		Name: "TestTemplate",
		TypeParameterConstraints: []Constraint{
			{Numeric: true},
		}}

	typ5 := Type{
		Name: "TestType",
	}

	v5 := TagValue{
		Name: "TestValue",
		TypeParameters: []Type{
			{Name: "foo", numeric: false},
			{Name: "bar", ordered: true},
		},
	}

	err5 := tmpl5.TryTypeAndValue(typ5, v5)

	if err5 == nil {
		t.Error("TryTypeAndValue should have returned an error")
	}

	// type parameter constraints not ok (by constraints)
	tmpl6 := Template{
		Name: "TestTemplate",
		TypeParameterConstraints: []Constraint{
			{Numeric: true},
			{Comparable: true},
		}}

	typ6 := Type{
		Name: "TestType",
	}

	v6 := TagValue{
		Name: "TestValue",
		TypeParameters: []Type{
			{Name: "foo", numeric: true},
			{Name: "bar", ordered: true},
		},
	}

	err6 := tmpl6.TryTypeAndValue(typ6, v6)

	if err6 == nil {
		t.Error("TryTypeAndValue should have returned an error")
	}
}

func TestByTag(t *testing.T) {
	// template exists, no constraints
	slice1 := TemplateSlice{
		{
			Name: "TestTag",
			Text: "This should compile.",
		},
		{
			Name: "SomethingElse",
			Text: "This should compile.",
		},
	}

	typ1 := Type{
		Name: "TestType",
	}

	tag1 := Tag{
		Name: "TestTag",
	}

	_, err1 := slice1.ByTag(typ1, tag1)

	if err1 != nil {
		t.Error(err1)
	}

	// template doesn't exist
	slice2 := TemplateSlice{
		{
			Name: "SomethingElse",
			Text: "This should compile.",
		},
		{
			Name: "AnotherSomethingElse",
			Text: "This should compile.",
		},
	}

	typ2 := Type{
		Name: "TestType",
	}

	tag2 := Tag{
		Name: "TestTag",
	}

	_, err2 := slice2.ByTag(typ2, tag2)

	if err2 == nil {
		t.Error("should return an error for not found")
	}

	// template name exists but type constraint is wrong
	slice3 := TemplateSlice{
		{
			Name:           "TestTag",
			Text:           "This should compile.",
			TypeConstraint: Constraint{Numeric: true},
		},
		{
			Name: "AnotherSomethingElse",
			Text: "This should compile.",
		},
	}

	typ3 := Type{
		Name:    "TestType",
		ordered: true,
	}

	tag3 := Tag{
		Name: "TestTag",
	}

	_, err3 := slice3.ByTag(typ3, tag3)

	if err3 == nil {
		t.Error("should return an error for type constraint")
	}

	// multiple exist, only one matches type constraint
	slice4 := TemplateSlice{
		{
			Name:           "TestTag",
			Text:           "This should compile.",
			TypeConstraint: Constraint{Numeric: true},
		},
		{
			Name: "AnotherSomethingElse",
			Text: "This should compile.",
		},
		{
			Name:           "TestTag",
			Text:           "This should be found.",
			TypeConstraint: Constraint{Ordered: true},
		},
	}

	typ4 := Type{
		Name:    "TestType",
		ordered: true,
	}

	tag4 := Tag{
		Name: "TestTag",
	}

	tmpl4, err4 := slice4.ByTag(typ4, tag4)

	if err4 != nil {
		t.Error(err4)
	}

	// ensure that we got the right template
	var b bytes.Buffer
	tmpl4.Execute(&b, nil)

	if b.String() != slice4[2].Text { // "This should be found."
		t.Error("should have picked the template which matches type constraints")
	}
}

func TestByTagValue(t *testing.T) {
	// template exists, no constraints
	slice1 := TemplateSlice{
		{
			Name: "TestValue",
			Text: "This should compile.",
		},
		{
			Name: "SomethingElse",
			Text: "This should compile.",
		},
	}

	typ1 := Type{
		Name: "TestType",
	}

	v1 := TagValue{
		Name: "TestValue",
	}

	_, err1 := slice1.ByTagValue(typ1, v1)

	if err1 != nil {
		t.Error(err1)
	}

	// template doesn't exist
	slice2 := TemplateSlice{
		{
			Name: "SomethingElse",
			Text: "This should compile.",
		},
		{
			Name: "AnotherSomethingElse",
			Text: "This should compile.",
		},
	}

	typ2 := Type{
		Name: "TestType",
	}

	v2 := TagValue{
		Name: "TestValue",
	}

	_, err2 := slice2.ByTagValue(typ2, v2)

	if err2 == nil {
		t.Error("should return an error for not found")
	}

	// template name exists but type parameters are wrong
	slice3 := TemplateSlice{
		{
			Name: "TestValue",
			Text: "This should compile.",
			TypeParameterConstraints: []Constraint{
				{Numeric: true},
			},
		},
		{
			Name: "AnotherSomethingElse",
			Text: "This should compile.",
		},
	}

	typ3 := Type{
		Name: "TestType",
	}

	v3 := TagValue{
		Name: "TestValue",
	}

	_, err3 := slice3.ByTagValue(typ3, v3)

	if err3 == nil {
		t.Error("should return an error for not found")
	}

	// multiple exist, only one matches type constraints
	slice4 := TemplateSlice{
		{
			Name: "TestValue",
			Text: "This should compile.",
			TypeParameterConstraints: []Constraint{
				{Numeric: true},
			},
		},
		{
			Name: "AnotherSomethingElse",
			Text: "This should compile.",
		},
		{
			Name: "TestValue",
			Text: "This should be found.",
			TypeParameterConstraints: []Constraint{
				{Ordered: true},
			},
		},
	}

	typ4 := Type{
		Name: "TestType",
	}

	v4 := TagValue{
		Name: "TestValue",
		TypeParameters: []Type{
			{Name: "bar", ordered: true},
		},
	}

	tmpl4, err4 := slice4.ByTagValue(typ4, v4)

	if err4 != nil {
		t.Error(err4)
	}

	// ensure that we got the right template
	var b bytes.Buffer
	tmpl4.Execute(&b, nil)

	if b.String() != slice4[2].Text { // "This should be found."
		t.Error("should have picked the template which matches type constraints")
	}

	// template funcs
	slice5 := TemplateSlice{
		{
			Name: "TestValue",
			Text: "This {{ToLower}} should compile.",
		},
		{
			Name: "SomethingElse",
			Text: "This should compile.",
		},
	}

	typ5 := Type{
		Name: "TestType",
	}

	v5 := TagValue{
		Name: "TestValue",
	}

	slice5.Funcs(map[string]interface{}{
		"ToLower": func(t Type) string {
			return strings.ToLower(t.Name)
		},
	})

	_, err5 := slice5.ByTagValue(typ5, v5)

	if err5 != nil {
		t.Error(err5)
	}
}
