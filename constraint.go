package typewriter

import "fmt"

// Constraint describes type requirements
type Constraint struct {
	// A numeric type is one that supports supports arithmetic operations.
	Numeric bool
	// A comparable type is one that supports the == operator. Map keys must be comparable, for example.
	Comparable bool
	// An ordered type is one where greater-than and less-than are supported
	Ordered bool
}

func (c Constraint) tryType(t Type) error {
	if c.Comparable && !t.Comparable() {
		return fmt.Errorf("%s must be comparable (i.e. supports == and != operators)", t)
	}

	if c.Numeric && !t.Numeric() {
		return fmt.Errorf("%s must be numeric", t)
	}

	if c.Ordered && !t.Ordered() {
		return fmt.Errorf("%s must be ordered (i.e. supports > and < operators)", t)
	}

	return nil
}
