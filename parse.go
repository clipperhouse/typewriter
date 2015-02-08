package typewriter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// unlike the go build tool, the parser does not ignore . and _ files
var ignored = func(f os.FileInfo) bool {
	return !strings.HasPrefix(f.Name(), "_") && !strings.HasPrefix(f.Name(), ".")
}

func getPackages(directive string, conf *Config) ([]*Package, error) {
	// wrap filter with default filter
	filt := func(f os.FileInfo) bool {
		if conf.Filter != nil {
			return ignored(f) && conf.Filter(f)
		}
		return ignored(f)
	}

	// get the AST
	fset := token.NewFileSet()
	astPkgs, err := parser.ParseDir(fset, "./", filt, parser.ParseComments)

	if err != nil {
		return nil, err
	}

	var pkgs []*Package
	var typeCheckError *TypeCheckError

	for _, a := range astPkgs {
		pkg, err := getPackage(fset, a, conf)

		if err != nil {
			// under normal circumstances, err means bail
			// however, if caller chooses to ignore TypeCheckErrors, continue
			// (getPackage returns only TypeCheckErrors)
			if !conf.IgnoreTypeCheckErrors {
				return nil, err
			}

			// store the first error for return at bottom
			if typeCheckError == nil {
				typeCheckError = err
			}
		}

		pkgs = append(pkgs, pkg)

		specs := getTaggedComments(a, directive)

		for s, c := range specs {
			pointer, tags, err := parse(c.Text, directive)

			if err != nil {
				// error should have Pos relative to the whole AST
				err.Pos += c.Slash
				// Go-style syntax error with filename, line number, column
				err.msg = fset.Position(err.Pos).String() + ": " + err.Error()
				return nil, err
			}

			typ, evalErr := pkg.Eval(pointer.String() + s.Name.Name)

			if evalErr != nil {
				// possibly attributable to type check error
				if typeCheckError != nil {
					// combine messages
					combinedErr := fmt.Errorf("%s\n%s", typeCheckError, evalErr)
					return pkgs, combinedErr
				}
				return pkgs, evalErr
			}

			// do type evaluation on type parameters
			for _, tag := range tags {
				for i, val := range tag.Values {
					for _, name := range val.typeParameterNames {
						tp, evalErr := pkg.Eval(name)
						if evalErr != nil {
							// possibly attributable to type check error
							if typeCheckError != nil {
								// combine messages
								combinedErr := fmt.Errorf("%s\n%s", typeCheckError, evalErr)
								return pkgs, combinedErr
							}
							return pkgs, evalErr
						}
						val.TypeParameters = append(val.TypeParameters, tp)
					}
					tag.Values[i] = val // mutate the original
				}
				typ.Tags = append(typ.Tags, tag)
			}

			typ.test = test(strings.HasSuffix(fset.Position(s.Pos()).Filename, "_test.go"))

			pkg.Types = append(pkg.Types, typ)
		}
	}

	if typeCheckError != nil {
		return pkgs, typeCheckError
	}

	return pkgs, nil
}

// getTaggedComments walks the AST and returns types which have directive comment
// returns a map of TypeSpec to directive
func getTaggedComments(pkg *ast.Package, directive string) map[*ast.TypeSpec]*ast.Comment {
	specs := make(map[*ast.TypeSpec]*ast.Comment)

	ast.Inspect(pkg, func(n ast.Node) bool {
		g, ok := n.(*ast.GenDecl)

		// is it a type?
		// http://golang.org/pkg/go/ast/#GenDecl
		if !ok || g.Tok != token.TYPE {
			// never mind, move on
			return true
		}

		if g.Lparen == 0 {
			// not parenthesized, copy GenDecl.Doc into TypeSpec.Doc
			g.Specs[0].(*ast.TypeSpec).Doc = g.Doc
		}

		for _, s := range g.Specs {
			t := s.(*ast.TypeSpec)

			if c := findAnnotation(t.Doc, directive); c != nil {
				specs[t] = c
			}
		}

		// no need to keep walking, we don't care about TypeSpec's children
		return false
	})

	return specs
}

// findDirective return the first line of a doc which contains a directive
// the directive and '//' are removed
func findAnnotation(doc *ast.CommentGroup, directive string) *ast.Comment {
	if doc == nil {
		return nil
	}

	// check lines of doc for directive
	for _, c := range doc.List {
		l := c.Text
		// does the line start with the directive?
		t := strings.TrimLeft(l, "/ ")
		if !strings.HasPrefix(t, directive) {
			continue
		}

		// remove the directive from the line
		t = strings.TrimPrefix(t, directive)

		// must be eof or followed by a space
		if len(t) > 0 && t[0] != ' ' {
			continue
		}

		return c
	}

	return nil
}

type parsr struct {
	lex       *lexer
	token     [2]item // two-token lookahead for parser.
	peekCount int
}

// next returns the next token.
func (p *parsr) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

// backup backs the input stream up one token.
func (p *parsr) backup() {
	p.peekCount++
}

// peek returns but does not consume the next token.
func (p *parsr) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

func parse(input, directive string) (Pointer, TagSlice, *SyntaxError) {
	var pointer Pointer
	var tags TagSlice
	p := &parsr{
		lex: lex(input),
	}

	// to ensure no duplicate tags
	exists := make(map[string]struct{})

Loop:
	for {
		item := p.next()
		switch item.typ {
		case itemEOF:
			break Loop
		case itemError:
			err := NewSyntaxError(item, item.val)
			return false, nil, err
		case itemCommentPrefix:
			// don't care, move on
			continue
		case itemDirective:
			// is it the directive we care about?
			if item.val != directive {
				return false, nil, nil
			}
			continue
		case itemPointer:
			// have we already seen a pointer?
			if pointer {
				err := NewSyntaxError(item, "second pointer declaration")
				return false, nil, err
			}

			// have we already seen tags? pointer must be first
			if len(tags) > 0 {
				err := NewSyntaxError(item, "pointer declaration must precede tags")
				return false, nil, err
			}

			pointer = true
		case itemTag:
			// we have an identifier, start a tag
			tag := Tag{
				Name: item.val,
			}

			// check for duplicate
			if _, seen := exists[tag.Name]; seen {
				err := NewSyntaxError(item, "duplicate tag %q", tag.Name)
				return pointer, nil, err
			}

			// mark tag as previously seen
			exists[tag.Name] = struct{}{}

			// tag has values
			if p.peek().typ == itemColonQuote {
				p.next() // absorb the colonQuote
				negated, vals, err := parseTagValues(p)

				if err != nil {
					return false, nil, err
				}

				tag.Negated = negated
				tag.Values = vals

			}

			tags = append(tags, tag)
		default:
			return false, nil, unexpected(item)
		}
	}

	return pointer, tags, nil
}

func parseTagValues(p *parsr) (bool, []TagValue, *SyntaxError) {
	var negated bool
	var vals []TagValue

	for {
		item := p.next()

		switch item.typ {
		case itemError:
			err := NewSyntaxError(item, item.val)
			return false, nil, err
		case itemEOF:
			// shouldn't happen within a tag
			err := NewSyntaxError(item, "expected a close quote")
			return false, nil, err
		case itemMinus:
			if len(vals) > 0 {
				err := NewSyntaxError(item, "negation must precede tag values")
				return false, nil, err
			}
			negated = true
		case itemTagValue:
			val := TagValue{
				Name: item.val,
			}

			if p.peek().typ == itemTypeParameter {
				typs, err := parseTypeParameters(p)
				if err != nil {
					return false, nil, err
				}
				val.typeParameterNames = typs
			}

			vals = append(vals, val)
		case itemCloseQuote:
			// we're done
			return negated, vals, nil
		default:
			return false, nil, unexpected(item)
		}
	}
}

func parseTypeParameters(p *parsr) ([]string, *SyntaxError) {
	var typs []string

	for {
		item := p.next()

		switch item.typ {
		case itemTypeParameter:
			typs = append(typs, item.val)
		default:
			p.backup()
			return typs, nil
		}
	}
}

func unexpected(item item) *SyntaxError {
	return NewSyntaxError(item, "unexpected '%v'", item.val)
}

type SyntaxError struct {
	msg string
	Pos token.Pos
}

func (e *SyntaxError) Error() string {
	return strings.TrimLeft(e.msg, ":- ") // some errors try to add empty Pos()
}

func NewSyntaxError(item item, format string, a ...interface{}) *SyntaxError {
	var msg string
	if len(a) > 0 {
		msg = fmt.Sprintf(format, a)
	} else {
		msg = format
	}
	return &SyntaxError{
		msg: msg,
		Pos: item.pos,
	}
}
