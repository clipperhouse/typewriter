package typewriter

import (
	"go/ast"
	"go/token"
	"strings"

	// gcimporter implements Import for gc-generated files
	_ "golang.org/x/tools/go/gcimporter"
	"golang.org/x/tools/go/types"
)

type evaluator interface {
	Eval(string) (Type, error)
}

type Package struct {
	*types.Package
	Types []Type
}

func NewPackage(path, name string) *Package {
	return &Package{
		types.NewPackage(path, name),
		[]Type{},
	}
}

type TypeCheckError struct {
	err error
}

func (t *TypeCheckError) Error() string {
	return t.err.Error()
}

func getPackage(fset *token.FileSet, a *ast.Package, conf *Config) (*Package, error) {
	// pull map into a slice
	var files []*ast.File
	for _, f := range a.Files {
		files = append(files, f)
	}

	config := types.Config{}

	if conf.IgnoreTypeCheckErrors {
		// no-op allows type checking to proceed in presence of errors
		// https://godoc.org/golang.org/x/tools/go/types#Config
		config.Error = func(err error) {}
	}

	typesPkg, err := config.Check(a.Name, fset, files, nil)

	p := &Package{typesPkg, []Type{}}

	// type-check error is the only error this func can return
	if err != nil {
		return p, &TypeCheckError{err}
	}

	return p, nil
}

func (p *Package) Eval(name string) (Type, error) {
	var result Type

	t, err := types.Eval(name, p.Package, p.Scope())
	if err != nil {
		return result, err
	}

	result = Type{
		Pointer:    isPointer(t.Type),
		Name:       strings.TrimLeft(name, Pointer(true).String()), // trims the * if it exists
		comparable: isComparable(t.Type),
		numeric:    isNumeric(t.Type),
		ordered:    isOrdered(t.Type),
	}

	return result, nil
}
