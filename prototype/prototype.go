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

	"github.com/pascaldekloe/goe/el"
)

// Fatalf stops further execution when called.
// From the standard library, testing.T, testing.B and log.Logger provide such
// functionality.
var Fatalf = log.New(os.Stderr, "goe: ", log.LstdFlags).Fatalf

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

	// Assign gets the transformation options for the specified GoEL path.
	Assign(path string) Transform
}

// Transform builds new templates.
type Transform interface {
	// To applies value as a new template.
	To(value interface{}) Template

	// Each applies each value as a new template.
	Each(values ...interface{}) Collection
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

// gobt is a gob template.
type gobt struct {
	typ    reflect.Type
	serial []byte
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

func (t *gobt) Assign(path string) Transform {
	return elt{path: path, t: t}
}

type elt struct {
	path string
	t    Template
}

func (e elt) To(value interface{}) Template {
	x := e.t.Build()

	var notPtr bool
	if t := reflect.TypeOf(x); t.Kind() != reflect.Ptr {
		notPtr = true
		// Make it addressable
		v := reflect.New(t)
		v.Elem().Set(reflect.ValueOf(x))
		x = v.Interface()
	}

	if n := el.Assign(x, e.path, value); n == 0 {
		Fatalf("prototype: can't apply %s on %s", e.path, e.t)
	}

	if notPtr {
		x = reflect.ValueOf(x).Elem().Interface()
	}

	return New(x)
}

func (e elt) Each(values ...interface{}) Collection {
	c := make(Collection, len(values))
	for i, v := range values {
		c[i] = e.To(v)
	}
	return c
}

func (t *gobt) String() string {
	return t.typ.String()
}

type Collection []Template

// Add appends the entry to c.
func (c *Collection) Add(entry Template) {
	*c = append(*c, entry)
}

// AddAll appends all templates from entries to c.
func (c *Collection) AddAll(entries Collection) {
	*c = append(*c, entries...)
}

// BuildAll instantiates the prototypes as in Template.Build.
func (c Collection) BuildAll() []interface{} {
	builds := make([]interface{}, len(c))
	for i, t := range c {
		builds[i] = t.Build()
	}
	return builds
}
