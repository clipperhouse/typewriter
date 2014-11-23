package typewriter

// ImportSpec describes the name and path of an import.
// The name is often omitted.
//
// +gen slice:"Select[string]"
type ImportSpec struct {
	Name, Path string
}
