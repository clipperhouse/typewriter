package typewriter

import "testing"

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
