package typewriter

import (
	"reflect"
	"testing"
)

func TestShouldAddDefaultsEmptyCase(t *testing.T) {
	tag := &Tag{}
	if !tag.ShouldAddDefaults() {
		t.Error("tag should add defaults if empty")
	}
}

func TestShouldAddDefaultsNonEmtpyWithoutStar(t *testing.T) {
	tag := &Tag{
		Values: []TagValue{
			TagValue{Name: "foo"},
		},
	}
	if tag.ShouldAddDefaults() {
		t.Error("tag should not add defaults for non-empty case without *")
	}
}

func TestShouldAddDefaultsNonEmptyWithStar(t *testing.T) {
	tag := &Tag{
		Values: []TagValue{
			TagValue{Name: "*"},
		},
	}
	if !tag.ShouldAddDefaults() {
		t.Error("tag should add defaults for non-empty case with *")
	}
}

func TestMakeDefaultTagValuesEmptyCase(t *testing.T) {
	tagvals := MakeDefaultTagValues(Type{}, TemplateSlice{})
	if len(tagvals) != 0 {
		t.Error("Should get nothing but got something for empty case:", tagvals)
	}
}

func TestMakeDefaultTagValuesCaseOneYes(t *testing.T) {
	templates := TemplateSlice([]*Template{&Template{Name: "foo"}})
	typ := Type{}
	defaults := MakeDefaultTagValues(typ, templates)
	expectedDefaults := []TagValue{
		TagValue{Name: "foo"},
	}
	if !reflect.DeepEqual(expectedDefaults, defaults) {
		t.Errorf("\nexpectedDefaults:\n%#v\ndefaults:\n%#v\n", expectedDefaults, defaults)
	}
}

func TestMakeDefaultTagValuesCaseOneNo(t *testing.T) {
	templates := TemplateSlice([]*Template{&Template{Name: "foo", TypeConstraint: Constraint{Ordered: true}}})
	typ := Type{}
	defaults := MakeDefaultTagValues(typ, templates)
	if len(defaults) != 0 {
		t.Errorf("Expected no defaults. Got", defaults)
	}
}
