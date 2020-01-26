package typewriter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

var fw = &fooWriter{}
var bw = &barWriter{}
var ew = &errWriter{}

func TestRegister(t *testing.T) {
	if err := Register(&fooWriter{}); err != nil {
		t.Error(err)
	}

	if err := Register(&barWriter{}); err != nil {
		t.Error(err)
	}

	if err := Register(&fooWriter{}); err == nil {
		t.Error("registering the same typewriter twice should be an error")
	}

	if len(typeWriters) != 2 {
		t.Errorf("should have 2 typewriters registered, found %v", len(typeWriters))
	}

	// clear 'em out for later tests
	typeWriters = make([]Interface, 0)
}

func TestNewApp(t *testing.T) {
	// set up some registered typewriters for this app
	// no error checking here, see TestRegister
	Register(&fooWriter{})
	Register(&barWriter{})

	a1, err1 := NewApp("+test")

	if err1 != nil {
		t.Error(err1)
		return
	}

	p1 := a1.Packages[0]

	// app and dummy types
	if len(p1.Types) != 4 {
		t.Errorf("should have found 4 types, found %v", len(p1.Types))
	}

	// this merely tests that they've been assigned to the app
	if len(a1.TypeWriters) != 2 {
		t.Errorf("should have found 2 typewriters, found %v", len(a1.TypeWriters))
	}

	// clear 'em out for later tests
	typeWriters = make([]Interface, 0)
}

func TestNewAppFiltered(t *testing.T) {
	filter := func(f os.FileInfo) bool {
		return !strings.HasPrefix(f.Name(), "dummy")
	}

	a1, err1 := NewAppFiltered("+test", filter)

	if err1 != nil {
		t.Error(err1)
		return
	}

	p1 := a1.Packages[0]

	// dummy is filtered out
	if len(p1.Types) != 1 {
		t.Errorf("should have found 1 types, found %v", len(p1.Types))
	}

	// should fail if types can't be evaluated
	// package.go by itself can't compile since it depends on other types
	filter2 := func(f os.FileInfo) bool {
		return f.Name() == "package.go"
	}

	_, err2 := NewAppFiltered("+test", filter2)

	if err2 == nil {
		t.Error("should have been unable to create app for an incomplete package")
	}
}

func TestWrite(t *testing.T) {
	a := &App{
		Directive: "+test",
	}

	p := NewPackage("dummy", "somepkg")

	typ := Type{
		Name: "sometype",
	}

	var b bytes.Buffer
	write(&b, a, p, []Type{typ}, &fooWriter{})

	// make sure the critical bits actually get written

	s := b.String()

	if !strings.Contains(s, "package somepkg") {
		t.Errorf("package declaration did not get written")
	}

	if !strings.Contains(s, "import") || !strings.Contains(s, `"fmt"`) {
		t.Errorf("imports declaration or package did not get written")
	}

	if !strings.Contains(s, "func pointlesssometype()") {
		t.Errorf("Write did not write func as expected")
	}
}

func TestWriteAllOneFile(t *testing.T) {
	var written []string
	var err error

	// set up some registered typewriters for this app
	fw1 := &fooWriter{}

	// no error checking here, see TestRegister
	Register(fw1)

	var c Config
	c.OneFile = true

	a1, err := c.NewApp("+test")

	if err != nil {
		t.Error(err)
		return
	}

	p1 := a1.Packages[0]

	written, err = a1.WriteAll()
	cleanup(written) // we don't need the written files

	if err != nil {
		t.Error(err)
	}

	if len(written) != 1 {
		t.Errorf(".WriteAll() should have written 1 file, wrote %d files", len(written))
	}

	if fw1.writeCalls != len(p1.Types) {
		t.Errorf(".Write() should have been called %v times (once for each type); was called %v", len(p1.Types), fw.writeCalls)
	}

	// clear 'em out
	typeWriters = make([]Interface, 0)

	// new set of writers for this test

	fw2 := &fooWriter{}
	bw2 := &barWriter{}
	ew2 := &errWriter{}

	Register(fw2)
	Register(bw2)
	Register(ew2)

	a2, _ := NewApp("+test") // error checked above, ignore here

	written, err = a2.WriteAll()
	cleanup(written) // we don't need the written files

	if err != nil {
		t.Error(err)
	}

	if fw.writeCalls != 0 {
		t.Errorf(".Write() should have been called no times due to error in validation; was called %v", fw.writeCalls)
	}

	if bw.writeCalls != 0 {
		t.Errorf(".Write() should have been called no times due to error in validation; was called %v", bw.writeCalls)
	}

	if ew.writeCalls != 0 {
		t.Errorf(".Write() should have been called no times due to error in validation; was called %v", ew.writeCalls)
	}

	// clear 'em out
	typeWriters = make([]Interface, 0)

	// new set of writers for this test

	fw3 := &fooWriter{}
	jw3 := &junkWriter{}
	bw3 := &barWriter{}

	Register(fw3)
	Register(jw3)
	Register(bw3)

	a3, _ := NewApp("+test") // error checked above, ignore here

	written, err = a3.WriteAll()
	cleanup(written) // we don't need the written files

	if err == nil {
		t.Errorf("writer producing invalid Go code should return an error")
	}

	// clear 'em out for later tests
	typeWriters = make([]Interface, 0)
}

func cleanup(files []string) {
	for _, f := range files {
		os.Remove(f)
	}
}

func TestWriteAll(t *testing.T) {
	var written []string
	var err error

	// set up some registered typewriters for this app
	fw1 := &fooWriter{}

	// no error checking here, see TestRegister
	Register(fw1)

	a1, err := NewApp("+test")

	if err != nil {
		t.Error(err)
		return
	}

	p1 := a1.Packages[0]

	written, err = a1.WriteAll()
	cleanup(written) // we don't need the written files

	if err != nil {
		t.Error(err)
	}

	if fw1.writeCalls != len(p1.Types) {
		t.Errorf(".Write() should have been called %v times (once for each type); was called %v", len(p1.Types), fw.writeCalls)
	}

	// clear 'em out
	typeWriters = make([]Interface, 0)

	// new set of writers for this test

	fw2 := &fooWriter{}
	bw2 := &barWriter{}
	ew2 := &errWriter{}

	Register(fw2)
	Register(bw2)
	Register(ew2)

	a2, _ := NewApp("+test") // error checked above, ignore here

	written, err = a2.WriteAll()
	cleanup(written) // we don't need the written files

	if err != nil {
		t.Error(err)
	}

	if fw.writeCalls != 0 {
		t.Errorf(".Write() should have been called no times due to error in validation; was called %v", fw.writeCalls)
	}

	if bw.writeCalls != 0 {
		t.Errorf(".Write() should have been called no times due to error in validation; was called %v", bw.writeCalls)
	}

	if ew.writeCalls != 0 {
		t.Errorf(".Write() should have been called no times due to error in validation; was called %v", ew.writeCalls)
	}

	// clear 'em out
	typeWriters = make([]Interface, 0)

	// new set of writers for this test

	fw3 := &fooWriter{}
	jw3 := &junkWriter{}
	bw3 := &barWriter{}

	Register(fw3)
	Register(jw3)
	Register(bw3)

	a3, _ := NewApp("+test") // error checked above, ignore here

	written, err = a3.WriteAll()
	cleanup(written) // we don't need the written files

	if err == nil {
		t.Errorf("writer producing invalid Go code should return an error")
	}

	// clear 'em out for later tests
	typeWriters = make([]Interface, 0)
}

type fooWriter struct {
	writeCalls int
}

func (f *fooWriter) Name() string {
	return "foo"
}

func (f *fooWriter) Imports(t Type) []ImportSpec {
	imports := []ImportSpec{
		{Path: "fmt"},
		{Path: "qux"}, // this is intentionally spurious and should be removed by imports.Process in WriteAll
	}
	return imports
}

func (f *fooWriter) Write(w io.Writer, t Type) error {
	f.writeCalls++
	w.Write([]byte(fmt.Sprintf(`func pointless%s(){
		fmt.Println("pointless!")
		}`, t.String())))
	return nil
}

type barWriter struct {
	writeCalls int
}

func (f *barWriter) Name() string {
	return "bar"
}

func (f *barWriter) Imports(t Type) (result []ImportSpec) {
	return result
}

func (f *barWriter) Write(w io.Writer, t Type) error {
	f.writeCalls++
	return nil
}

type errWriter struct {
	writeCalls int
}

func (f *errWriter) Name() string {
	return "err"
}

func (f *errWriter) Imports(t Type) (result []ImportSpec) {
	return result
}

func (f *errWriter) Write(w io.Writer, t Type) error {
	f.writeCalls++
	return nil
}

type junkWriter struct{}

func (f *junkWriter) Name() string {
	return "junk"
}

func (f *junkWriter) Imports(t Type) (result []ImportSpec) {
	return result
}

func (f *junkWriter) Write(w io.Writer, t Type) error {
	w.Write([]byte("this is invalid Go code, innit?"))
	return nil
}
