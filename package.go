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

func getPackage(fset *token.FileSet, a *ast.Package) (*Package, error) {
	// pull map into a slice
	var files []*ast.File
	for _, f := range a.Files {
		files = append(files, f)
	}

	typesPkg, err := types.Check(a.Name, fset, files)

	if err != nil {
		return nil, err
	}

	return &Package{typesPkg, []Type{}}, nil
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
