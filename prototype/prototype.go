// Package prototype simplifies test object composition.
package prototype

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"log"
	"os"
	"reflect"

	"el"
)

// Template is an immutable prototype definition.
type Template interface {
	// Build instantiates the prototype.
	// Modifications on the returned content won't affect the template.
	Build() interface{}

	// BuildJSON serializes the prototype as UTF-8 JSON.
	BuildJSON() []byte

	// BuildXML serializes the prototype as UTF-8 XML.
	BuildXML() []byte

	// String gets a description for messaging purposes.
	String() string

	// Path gets the transformation options for the GoEL path.
	Path(string) Transform
}

// Fatalf stops further execution when called.
// From the standard library, testing.T, testing.B and log.Logger provide such
// functionality.
var Fatalf = log.New(os.Stderr, "goe: ", log.LstdFlags).Fatalf

// gobt is a gob template.
type gobt struct {
	typ    reflect.Type
	serial []byte
}

// New creates a template out of x.
// Modifications on x after this call won't affect the template.
func New(x interface{}) Template {
	t := &gobt{typ: reflect.TypeOf(x)}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(x); err != nil {
		Fatalf("prototype: can't serialize %s: %s", t, err)
	}

	t.serial = buf.Bytes()
	return t
}

func (t *gobt) Build() interface{} {
	v := reflect.New(t.typ)
	if err := gob.NewDecoder(bytes.NewReader(t.serial)).Decode(v.Interface()); err != nil {
		Fatalf("prototype: can't deserialize %s: %s", t, err)
	}
	return v.Elem().Interface()
}

func (t *gobt) BuildJSON() []byte {
	x := t.Build()
	bytes, err := json.Marshal(&x)
	if err != nil {
		Fatalf("prototype: can't serialize %s to JSON: %s", t, err)
	}
	return bytes
}

func (t *gobt) BuildXML() []byte {
	x := t.Build()
	bytes, err := xml.Marshal(&x)
	if err != nil {
		Fatalf("prototype: can't serialize %s to XML: %s", t, err)
	}
	return bytes
}

func (t *gobt) String() string {
	return t.typ.String()
}

type Collection []Template

func (c Collection) Run(f func(Template)) {
	for _, t := range c {
		f(t)
	}
}

// Transform creates new template collections.
type Transform interface {
	To(v interface{}) Collection
	In(v ...interface{}) Collection
}

type elt struct {
	src  Collection
	path string
}

// Path gets the transformation options for the GoEL path.
func (c Collection) Path(p string) Transform {
	return elt{
		src:  c,
		path: p,
	}
}

func (t *gobt) Path(p string) Transform {
	return elt{
		src:  Collection{t},
		path: p,
	}
}

func (e elt) To(value interface{}) Collection {
	dst := make(Collection, len(e.src))

	var n int
	for i, t := range e.src {
		x := t.Build()
		n += el.Have(&x, e.path, value)
		dst[i] = New(x)
	}

	if n == 0 {
		Fatalf("prototype: can't apply %#v on %s", value, e.path)
	}

	return dst
}

func (e elt) In(values ...interface{}) Collection {
	if len(values) < 2 {
		Fatalf("Transformation In requires at least two values.")
	}

	dst := e.To(values[0])
	for _, v := range values[1:] {
		dst = append(dst, e.To(v)...)
	}

	return dst
}
