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

type TypeCheckError struct {
	err     error
	ignored bool
}

func (t *TypeCheckError) Error() string {
	var result string
	if t.ignored {
		result += "[ignored] "
	}
	return result + t.err.Error()
}

func (t *TypeCheckError) addPos(fset *token.FileSet, pos token.Pos) {
	// some errors come with empty pos
	err := strings.TrimLeft(t.err.Error(), ":- ")
	// prepend position information (file name, line, column)
	t.err = fmt.Errorf("%s: %s", fset.Position(pos), err)
}

func combine(ts []*TypeCheckError) error {
	if len(ts) == 0 {
		return nil
	}

	var errs []string
	for _, t := range ts {
		errs = append(errs, t.Error())
	}
	return fmt.Errorf(strings.Join(errs, "\n"))
}

func getPackage(fset *token.FileSet, a *ast.Package, conf *Config) (*Package, *TypeCheckError) {
	// pull map into a slice
	var files []*ast.File
	for _, f := range a.Files {
		files = append(files, f)
	}

	config := types.Config{
		DisableUnusedImportCheck: true,
		IgnoreFuncBodies:         true,
	}

	if conf.IgnoreTypeCheckErrors {
		// no-op allows type checking to proceed in presence of errors
		// https://godoc.org/golang.org/x/tools/go/types#Config
		config.Error = func(err error) {}
	}

	typesPkg, err := config.Check(a.Name, fset, files, nil)

	p := &Package{typesPkg, []Type{}}

	if err != nil {
		return p, &TypeCheckError{err, conf.IgnoreTypeCheckErrors}
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
		Type:       t.Type,
	}

	if isInvalid(t.Type) {
		err := fmt.Errorf("invalid type: %s", name)
		return result, &TypeCheckError{err, false}
	}

	return result, nil
}
