package el

import (
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"
)

// value defines the API from reflect.Value used over here.
// See mapWrap for the reason.
type value interface {
	Kind() reflect.Kind
	Type() reflect.Type
	IsValid() bool
	IsNil() bool
	Elem() reflect.Value
	CanInterface() bool
	Interface() interface{}
	CanSet() bool
	Set(reflect.Value)
	NumField() int
	Field(int) reflect.Value
	FieldByName(string) reflect.Value
	Len() int
	Index(int) reflect.Value
	MapIndex(reflect.Value) reflect.Value
	SetMapIndex(k, v reflect.Value)
	MapKeys() []reflect.Value
	Bool() bool
	Int() int64
	Uint() uint64
	Float() float64
	Complex() complex128
	String() string
}

// resolve follows expr on root.
// When doBuild is true the path will be created where possible.
func resolve(expr string, root interface{}, doBuild bool) []value {
	x := follow(reflect.ValueOf(root), doBuild)
	if x == nil {
		return nil
	}
	track := []value{*x}

	segments := strings.Split(path.Clean(expr), "/")[1:]
	if segments[0] == "" { // root selection
		return track
	}

	for _, selection := range segments {
		var key string
		if last := len(selection) - 1; selection[last] == ']' {
			if i := strings.IndexByte(selection, '['); i >= 0 {
				key = selection[i+1 : last]
				selection = selection[:i]
			}
		}

		if selection != "." {
			track = followField(track, selection, doBuild)
		}
		if key != "" {
			track = followKey(track, key, doBuild)
		}
	}
	return track
}

// followField returns all fields matching s from track.
func followField(track []value, s string, doBuild bool) []value {
	if s == "*" {
		// Count values with n and filter struct types in track while we're at it.
		writeIndex, n := 0, 0
		for _, v := range track {
			if v.Kind() == reflect.Struct {
				n += v.Type().NumField()
				track[writeIndex] = v
				writeIndex++
			}
		}
		track = track[:writeIndex]

		dst := make([]value, n)
		writeIndex = 0
		for _, v := range track {
			for i := v.Type().NumField() - 1; i >= 0; i-- {
				if x := follow(v.Field(i), doBuild); x != nil {
					dst[writeIndex] = *x
					writeIndex++
				}
			}
		}
		return dst[:writeIndex]
	}

	// Write result back to track with writeIndex to safe memory.
	writeIndex := 0
	for _, v := range track {
		if v.Kind() == reflect.Struct {
			field := v.FieldByName(s)
			if field.IsValid() { // exists
				if x := follow(field, doBuild); x != nil {
					track[writeIndex] = *x
					writeIndex++
				}
			}
		}
	}
	return track[:writeIndex]
}

// followField returns all fields matching s from track.
func followKey(track []value, s string, doBuild bool) []value {
	if s == "*" {
		// Count values with n and filter keyed types in track while we're at it.
		writeIndex, n := 0, 0
		for _, v := range track {
			switch v.Kind() {
			case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
				n += v.Len()
				track[writeIndex] = v
				writeIndex++
			}
		}
		track = track[:writeIndex]

		dst := make([]value, n)
		writeIndex = 0
		for _, v := range track {
			switch v.Kind() {
			case reflect.Array, reflect.Slice, reflect.String:
				for i, n := 0, v.Len(); i < n; i++ {
					if x := follow(v.Index(i), doBuild); x != nil {
						dst[writeIndex] = *x
						writeIndex++
					}
				}

			case reflect.Map:
				for _, key := range v.MapKeys() {
					if x := followMap(v, key, doBuild); x != nil {
						dst[writeIndex] = *x
						writeIndex++
					}
				}

			}
		}
		return dst[:writeIndex]
	}

	// Write result back to track with writeIndex to safe memory.
	writeIndex := 0
	for _, v := range track {
		switch v.Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
			i, err := strconv.ParseUint(s, 0, 64)
			if err == nil && i < uint64(v.Len()) {
				if x := follow(v.Index(int(i)), doBuild); x != nil {
					track[writeIndex] = *x
					writeIndex++
				}
			}

		case reflect.Map:
			key := parseLiteral(s, v.Type().Key())
			if key != nil {
				if x := followMap(v, *key, doBuild); x != nil {
					track[writeIndex] = *x
					writeIndex++
				}
			}

		}
	}
	return track[:writeIndex]
}

// follow returns content when possible.
func follow(v value, doBuild bool) *value {
	for {
		k := v.Kind()
		for k == reflect.Interface {
			if v.IsNil() {
				return nil
			}
			v = v.Elem()
			k = v.Kind()
		}

		if k == reflect.Ptr {
			if doBuild && v.IsNil() {
				if !v.CanSet() {
					return nil
				}
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
			k = v.Kind()
			continue
		}

		if k == reflect.Map && v.IsNil() {
			if !doBuild || !v.CanSet() {
				return nil
			}
			v.Set(reflect.MakeMap(v.Type()))
		}

		if !v.IsValid() {
			return nil
		}

		return &v
	}
}

// folowMap applies mapWrap for follow.
func followMap(m value, key reflect.Value, doBuild bool) *value {
	v := m.MapIndex(key)
	if !v.IsValid() {
		if !doBuild {
			return nil
		}
		v = reflect.New(m.Type().Elem()).Elem()
		m.SetMapIndex(key, v)
	}

	return follow(&mapWrap{m: m, k: key, v: v}, doBuild)
}

// mapWrap wraps map elements because changing the value (Set) requires a SetMapIndex.
type mapWrap struct {
	m value
	k reflect.Value
	v value
}

func (w *mapWrap) Set(v reflect.Value) {
	w.v = v
	w.m.SetMapIndex(w.k, v)
}

func (w *mapWrap) CanSet() bool {
	return w.m.CanInterface()
}

func (w *mapWrap) Kind() reflect.Kind                     { return w.v.Kind() }
func (w *mapWrap) Type() reflect.Type                     { return w.v.Type() }
func (w *mapWrap) IsValid() bool                          { return w.v.IsValid() }
func (w *mapWrap) IsNil() bool                            { return w.v.IsNil() }
func (w *mapWrap) Elem() reflect.Value                    { return w.v.Elem() }
func (w *mapWrap) CanInterface() bool                     { return w.v.CanInterface() }
func (w *mapWrap) Interface() interface{}                 { return w.v.Interface() }
func (w *mapWrap) NumField() int                          { return w.v.NumField() }
func (w *mapWrap) Field(i int) reflect.Value              { return w.v.Field(i) }
func (w *mapWrap) FieldByName(name string) reflect.Value  { return w.v.FieldByName(name) }
func (w *mapWrap) Len() int                               { return w.v.Len() }
func (w *mapWrap) Index(i int) reflect.Value              { return w.v.Index(i) }
func (w *mapWrap) MapIndex(k reflect.Value) reflect.Value { return w.v.MapIndex(k) }
func (w *mapWrap) SetMapIndex(k, v reflect.Value)         { w.v.SetMapIndex(k, v) }
func (w *mapWrap) MapKeys() []reflect.Value               { return w.v.MapKeys() }
func (w *mapWrap) Bool() bool                             { return w.v.Bool() }
func (w *mapWrap) Int() int64                             { return w.v.Int() }
func (w *mapWrap) Uint() uint64                           { return w.v.Uint() }
func (w *mapWrap) Float() float64                         { return w.v.Float() }
func (w *mapWrap) Complex() complex128                    { return w.v.Complex() }
func (w *mapWrap) String() string                         { return w.v.String() }

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
