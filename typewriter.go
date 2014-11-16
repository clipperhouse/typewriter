package typewriter

import (
	"io"
)

// TypeWriter is the interface to be implemented for code generation via gen
type TypeWriter interface {
	Name() string
	// Imports is a slice of imports required for the type; each will be written into the imports declaration.
	Imports(t Type) []ImportSpec
	// Write writes to the body of the generated code, following package declaration and imports.
	Write(w io.Writer, t Type) error
}
