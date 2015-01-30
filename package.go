package typewriter

import (
	"fmt"
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

	cfg := &types.Config{
		IgnoreFuncBodies: true,
		Error: func(err error){ fmt.Println("gen: "+err.Error()) },
	}

	// TODO: Is there a way to distinguish between errors we care
	// about and errors for undefined types that gen hasn't generated
	// yet?
	typesPkg, _ := cfg.Check(a.Name, fset, files, nil)

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
