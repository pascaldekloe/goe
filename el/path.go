package el

import (
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"
)

// pathEval represents an evaluation.
type pathEval struct {
	// track holds the values to follow.
	track []reflect.Value
	// next buffers the following track.
	next []reflect.Value
	// factory is for write operations.
	factory haveAs
	// last flags the final path component.
	last bool
}

// resolve follows expr on root and applies facory if not nil.
func resolve(expr string, root interface{}, factory haveAs) []reflect.Value {
	segments := strings.Split(path.Clean(expr), "/")[1:]

	e := pathEval{
		factory: factory,
		next:    make([]reflect.Value, 0, 1),
		last:    segments[0] == "", // root selection
	}

	v := reflect.ValueOf(root)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	} else if factory != nil {
		return nil
	}

	e.applyTo(v)
	if e.last {
		return e.next
	}

	for i, selection := range segments {
		var key string
		if last := len(selection) - 1; selection[last] == ']' {
			if i := strings.IndexByte(selection, '['); i >= 0 {
				key = selection[i+1 : last]
				selection = selection[:i]
			}
		}

		if selection != "." {
			if key == "" && i == len(segments)-1 {
				e.last = true
			}
			e.track = e.next
			e.applyField(selection)
		}

		if key != "" {
			if i == len(segments)-1 {
				e.last = true
			}
			e.track = e.next
			e.applyKey(key)
		}
	}
	return e.next
}

// applyField evaluates the field on e.track and writes the matches into e.next.
func (e *pathEval) applyField(s string) {
	if s == "*" {
		e.next = make([]reflect.Value, 0, len(e.track))
		for _, v := range e.track {
			for i := v.Type().NumField() - 1; i >= 0; i-- {
				e.applyTo(v.Field(i))
			}
		}
		return
	}

	// Write result back to e.track to safe memory.
	e.next = e.track[:0]
	for _, v := range e.track {
		switch v.Kind() {
		case reflect.Struct:
			field := v.FieldByName(s)
			if field.IsValid() { // exists
				e.applyTo(field)
			}

		}
	}
}

// applyKey evaluates the key on e.track and writes the matches into e.next.
func (e *pathEval) applyKey(s string) {
	if s == "*" {
		e.next = make([]reflect.Value, 0, len(e.track))
		for _, v := range e.track {
			switch v.Kind() {
			case reflect.Array, reflect.Slice, reflect.String:
				for i, n := 0, v.Len(); i < n; i++ {
					e.applyTo(v.Index(i))
				}

			case reflect.Map:
				for _, key := range v.MapKeys() {
					e.applyToMap(v, key)
				}

			}
		}
		return
	}

	// Write result back to e.track to safe memory.
	e.next = e.track[:0]
	for _, v := range e.track {
		switch v.Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
			i, err := strconv.ParseUint(s, 0, 64)
			if err == nil && i < uint64(v.Len()) {
				e.applyTo(v.Index(int(i)))
			}

		case reflect.Map:
			key := parseLiteral(s, v.Type().Key())
			if key != nil {
				e.applyToMap(v, *key)
			}

		}
	}
}

func (e *pathEval) applyTo(v reflect.Value) {
	if e.factory != nil && e.last {
		if v.CanSet() {
			v.Set(e.build(v.Type()))
		}
	} else if !e.forNext(v) && e.factory != nil && v.CanSet() {
		v.Set(e.build(v.Type()))
	}
}

func (e *pathEval) applyToMap(v, key reflect.Value) {
	if e.factory != nil && e.last {
		if v.CanInterface() {
			v.SetMapIndex(key, e.build(v.Type().Elem()))
		}
	} else if !e.forNext(v.MapIndex(key)) && e.factory != nil && v.CanInterface() {
		v.SetMapIndex(key, e.build(v.Type().Elem()))
	}
}

// forNext appends any content in v to e.next.
func (e *pathEval) forNext(v reflect.Value) bool {
	k := v.Kind()
	for k == reflect.Ptr || k == reflect.Interface {
		v = v.Elem()
		k = v.Kind()
	}

	if !v.IsValid() {
		return false
	}

	e.next = append(e.next, v)
	return true
}

// build instantiates a value of type t including pointers.
func (e *pathEval) build(t reflect.Type) reflect.Value {
	// Resolve destination type
	dt := t
	ptrCount := 0
	for dt.Kind() == reflect.Ptr {
		dt = dt.Elem()
		ptrCount++
	}

	result := e.new(dt)

	// Construct path and apply to entry
	for n := ptrCount; n > 0; n-- {
		ptr := reflect.New(result.Type())
		ptr.Elem().Set(result)
		result = ptr
	}

	e.forNext(result)
	return result
}

// new instatiates a value of type t.
func (e *pathEval) new(t reflect.Type) reflect.Value {
	if e.last {
		if v := e.factory(t); v != nil {
			return *v
		}
	}

	switch t.Kind() {
	case reflect.Map:
		return reflect.MakeMap(t)
	default:
		return reflect.New(t).Elem()
	}
}

// parseLiteral returns the interpretation of s for t or nil on failure.
func parseLiteral(s string, t reflect.Type) *reflect.Value {
	p := reflect.New(t)
	switch t.Kind() {
	case reflect.String:
		if s, err := strconv.Unquote(s); err == nil {
			p.Elem().SetString(s)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(s, 0, 64); err == nil {
			p.Elem().SetInt(i)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, err := strconv.ParseUint(s, 0, 64); err == nil {
			p.Elem().SetUint(u)
		}

	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			p.Elem().SetFloat(f)
		}

	default:
		ptr := p.Interface()
		if n, _ := fmt.Sscan(s, ptr); n == 1 {
			p = reflect.ValueOf(ptr)
		}

	}

	p = p.Elem()
	return &p
}
