package typewriter

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"
	"testing"
)

type findDirectiveTest struct {
	text  string
	found bool
}

// dummy type evaluator
type eval struct{}

func (eval) Eval(name string) (Type, error) {
	return Type{
		Name: name,
	}, nil
}

func TestFindDirective(t *testing.T) {
	tests := []findDirectiveTest{
		{`// +test`, true},
		{`// +test foo:"bar,Baz"`, true},
		{`// there's nothing here`, false},
		{`//+test foo:"bar,Baz"`, true},
		{`//+test foo:"bar,Baz"`, true},
		{`// +test * foo:"bar,Baz"`, true},
		{`// +test foo:"bar,Baz" qux:"thing"`, true},
		{`// +tested`, false},
	}

	for i, test := range tests {
		g := &ast.CommentGroup{
			List: []*ast.Comment{{Text: test.text}},
		}
		c := findAnnotation(g, "+test")
		found := c != nil
		if found != test.found {
			t.Errorf("[test %v] found should have been %v for:\n%s", i, test.found, test.text)
		}
	}
}

type parseTest struct {
	comment string
	pointer Pointer
	tags    TagSlice
	valid   bool
}

func TestParse(t *testing.T) {
	tests := []parseTest{
		{`// +test foo`, false, TagSlice{
			{"foo", []TagValue{}, false},
		}, true},
		{`// +test foo bar`, false, TagSlice{
			{"foo", []TagValue{}, false},
			{"bar", []TagValue{}, false},
		}, true},
		{`// +test foo:"bar,Baz"`, false, TagSlice{
			{"foo", []TagValue{
				{"bar", nil, nil},
				{"Baz", nil, nil},
			}, false},
		}, true},
		{`// +test * foo:"bar,Baz"`, true, TagSlice{
			{"foo", []TagValue{
				{"bar", nil, nil},
				{"Baz", nil, nil},
			}, false},
		}, true},
		{`// +test foo:"bar,Baz" qux:"stuff"`, false, TagSlice{
			{"foo", []TagValue{
				{"bar", nil, nil},
				{"Baz", nil, nil},
			}, false},
			{"qux", []TagValue{
				{"stuff", nil, nil},
			}, false},
		}, true},
		{`// +test foo:"-bar,Baz"`, false, TagSlice{
			{"foo", []TagValue{
				{"bar", nil, nil},
				{"Baz", nil, nil},
			}, true},
		}, true},
		{`// +test foo:"bar  ,Baz "  `, false, TagSlice{
			{"foo", []TagValue{
				{"bar", nil, nil},
				{"Baz", nil, nil},
			}, false},
		}, true},
		{`// +test foo:"bar,Baz[qaz], qux"`, false, TagSlice{
			{"foo", []TagValue{
				{"bar", nil, nil},
				{"Baz", nil, []item{{val: "qaz"}}},
				{"qux", nil, nil},
			}, false},
		}, true},
		{`// +test foo:"bar,Baz[[]qaz]"`, false, TagSlice{
			{"foo", []TagValue{
				{"bar", nil, nil},
				{"Baz", nil, []item{{val: "[]qaz"}}},
			}, false},
		}, true},
		{`// +test foo:"bar,Baz[qaz,hey]" qux:"stuff"`, false, TagSlice{
			{"foo", []TagValue{
				{"bar", nil, nil},
				{"Baz", nil, []item{{val: "qaz"}, {val: "hey"}}},
			}, false},
			{"qux", []TagValue{
				{"stuff", nil, nil},
			}, false},
		}, true},
		{`// +test foo:"Baz[qaz],yo[dude]" qux:"stuff[things]"`, false, TagSlice{
			{"foo", []TagValue{
				{"Baz", nil, []item{{val: "qaz"}}},
				{"yo", nil, []item{{val: "dude"}}},
			}, false},
			{"qux", []TagValue{
				{"stuff", nil, []item{{val: "things"}}},
			}, false},
		}, true},
		{`// +test foo:"bar,Baz`, false, nil, false},
		{`// +test foo:"bar,-Baz"`, false, nil, false},
		{`// +test foo:"bar,Baz-"`, false, nil, false},
		{`// +test foo:bar,Baz" qux:"stuff"`, false, nil, false},
		{`// +test foo"bar,Baz" qux:"stuff"`, false, nil, false},
		{`// +test foo:"bar,Baz" 8qux:"stuff"`, false, nil, false},
		{`// +test fo^o:"bar,Baz" qux:"stuff"`, false, nil, false},
		{`// +test foo:"bar,Ba|z" qux:"stuff"`, false, nil, false},
		{`// +test foo:"bar,Baz" qux:"stuff`, false, nil, false},
		{`// +test *foo:"bar,Baz" qux:"stuff"`, false, nil, false},
		{`// +test foo:"bar,Baz" * qux:"stuff"`, false, nil, false},
		{`// +test * foo:"bar,Baz" * qux:"stuff"`, false, nil, false},
		{`// +test foo:"bar,Baz[foo"`, false, nil, false},
		{`// +test foo:"bar,Baz[foo]]"`, false, nil, false},
		{`// +test foo:"bar,Baz[[]foo"`, false, nil, false},
		{`// +test foof:"bar,Baz" foof:"qux"`, false, nil, false},
	}

	fset := token.NewFileSet()

	for i, test := range tests {
		c := &ast.Comment{
			Text: test.comment,
		}
		pointer, tags, err := parse(fset, c, "+test")

		if test.valid != (err == nil) {
			t.Errorf("[test %v] valid should have been %v for: %s\n%s", i, test.valid, test.comment, err)
		}

		if pointer != test.pointer {
			t.Errorf("[test %v] pointer should have been %v for: \n%s", i, bool(test.pointer), test.comment)
		}

		if !tagsEqual(tags, test.tags) {
			t.Fatalf("[test %v] tags should have been \n%v, got \n%v", i, test.tags, tags)
		}
	}
}

func tagsEqual(tags, other TagSlice) bool {
	if len(tags) != len(other) {
		return false
	}

	for i := range tags {
		t := tags[i]
		o := other[i]

		if t.Name != o.Name {
			return false
		}

		if t.Negated != o.Negated {
			return false
		}

		if len(t.Values) != len(o.Values) {
			return false
		}

		for j := range t.Values {
			tv := t.Values[j]
			ov := o.Values[j]

			if tv.Name != ov.Name {
				return false
			}

			if len(tv.TypeParameters) != len(ov.TypeParameters) {
				return false
			}

			for k := range tv.TypeParameters {
				if tv.TypeParameters[k].String() != ov.TypeParameters[k].String() {
					return false
				}
			}
		}
	}

	return true
}

func TestGetTypes(t *testing.T) {
	// app and dummy types are marked up with +test
	pkgs, err := getPackages("+test", DefaultConfig)

	if err != nil {
		t.Error(err)
		return
	}

	typs := pkgs[0].Types

	if len(typs) != 4 {
		t.Errorf("should have found the 4 marked-up types, found %v", len(typs))
	}

	// put 'em into a map for convenience
	m := typeSliceToMap(typs)

	if _, found := m["App"]; !found {
		t.Errorf("should have found the app type")
	}

	if _, found := m["dummy"]; !found {
		t.Errorf("should have found the dummy type")
	}

	if _, found := m["dummy2"]; !found {
		t.Errorf("should have found the dummy2 type")
	}

	if _, found := m["dummy3"]; !found {
		t.Errorf("should have found the dummy3 type")
	}

	if t.Failed() {
		return // get out instead of nil panics below
	}

	dummy := m["dummy"]

	if !dummy.comparable {
		t.Errorf("dummy type should be comparable")
	}

	if !dummy.ordered {
		t.Errorf("dummy type should be ordered")
	}

	if !dummy.numeric {
		t.Errorf("dummy type should be numeric")
	}

	if len(dummy.Tags) != 1 {
		t.Fatalf("typ should have 1 tag, found %v", len(m["dummy"].Tags))
	}

	if len(dummy.Tags[0].Values) != 1 {
		fmt.Println(dummy.Tags[0].Values)
		t.Errorf("Tag should have 1 Item, found %v", len(m["dummy"].Tags[0].Values))
	}

	dummy2 := m["dummy2"]

	if dummy2.comparable {
		t.Errorf("dummy2 type should not be comparable")
	}

	if dummy2.ordered {
		t.Errorf("dummy2 type should not be ordered")
	}

	if dummy2.numeric {
		t.Errorf("dummy2 type should not be numeric")
	}

	dummy3 := m["dummy3"]

	if !dummy3.comparable {
		t.Errorf("dummy3 type should be comparable")
	}

	if !dummy3.ordered {
		t.Errorf("dummy3 type should be ordered")
	}

	if dummy3.numeric {
		t.Errorf("dummy3 type should not be numeric")
	}

	// check tag existence at a high level here, see also tag parsing tests

	app := m["App"]

	if len(app.Tags) != 2 {
		t.Errorf("typ should have 2 TagSlice, found %v", len(app.Tags))
	}

	if len(app.Tags[0].Values) != 1 {
		t.Errorf("Tag should have 1 Item, found %v", len(app.Tags[0].Values))
	}

	if len(app.Tags[1].Values) != 2 {
		t.Errorf("Tag should have 2 Values, found %v", len(app.Tags[1].Values))
	}

	if len(app.Tags[1].Values[0].TypeParameters) != 1 {
		t.Fatalf("TagValue should have 1 TypeParameter, found %v", len(app.Tags[1].Values[0].TypeParameters))
	}

	// filtered types should not show up

	conf := &Config{}
	conf.Filter = func(f os.FileInfo) bool {
		return !strings.HasPrefix(f.Name(), "dummy")
	}

	pkgs2, err2 := getPackages("+test", conf)

	if err2 != nil {
		t.Error(err2)
	}

	typs2 := pkgs2[0].Types

	if len(typs2) != 1 {
		t.Errorf("should have found the 1 marked-up type when filtered, found %v", len(typs2))
	}

	m2 := typeSliceToMap(typs2)

	if _, found := m2["dummy"]; found {
		t.Errorf("should not have found the dummy type")
	}

	if _, found := m2["App"]; !found {
		t.Errorf("should have found the app type")
	}

	// no false positives
	pkgs3, err3 := getPackages("+notreal", DefaultConfig)

	typs3 := pkgs3[0].Types

	if len(typs3) != 0 {
		t.Errorf("should have no marked-up types for +notreal")
	}

	if err3 != nil {
		t.Error(err3)
	}

	// should fail if types can't be evaluated
	// package.go by itself can't compile since it depends on other types

	conf4 := &Config{}
	conf4.Filter = func(f os.FileInfo) bool {
		return f.Name() == "package.go"
	}

	_, err4 := getPackages("+test", conf4)

	if err4 == nil {
		t.Error("should have been unable to evaluate types of incomplete package")
	}
}

func typeSliceToMap(typs []Type) map[string]Type {
	result := make(map[string]Type)
	for _, v := range typs {
		result[v.Name] = v
	}
	return result
}

func tagSliceToMap(typs []Tag) map[string]Tag {
	result := make(map[string]Tag)
	for _, v := range typs {
		result[v.Name] = v
	}
	return result
}
