package el

import (
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"
)

// resolve follows expr on root.
func resolve(expr string, root interface{}, buildCallbacks *[]finisher) []reflect.Value {
	x := follow(reflect.ValueOf(root), buildCallbacks != nil)
	if x == nil {
		return nil
	}
	track := []reflect.Value{*x}

	segments := strings.Split(path.Clean(expr), "/")[1:]
	if segments[0] == "" { // root selection
		return track
	}

	for _, selection := range segments {
		var key string
		if last := len(selection) - 1; selection[last] == ']' {
			if i := strings.IndexByte(selection, '['); i >= 0 {
				key = selection[i+1 : last]
				if key != "" {
					selection = selection[:i]
				}
			}
		}

		if selection != "." {
			track = followField(track, selection, buildCallbacks != nil)
		}
		if key != "" {
			track = followKey(track, key, buildCallbacks)
		}
	}

	return track
}

// followField returns all fields matching s from track.
func followField(track []reflect.Value, s string, doBuild bool) []reflect.Value {
	if s == "*" {
		// Count fields with n and filter struct types in track while we're at it.
		writeIndex, n := 0, 0
		for _, v := range track {
			if v.Kind() == reflect.Struct {
				n += v.Type().NumField()
				track[writeIndex] = v
				writeIndex++
			}
		}
		track = track[:writeIndex]

		dst := make([]reflect.Value, n)
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

// followKey returns all elements matching s from track.
func followKey(track []reflect.Value, s string, buildCallbacks *[]finisher) []reflect.Value {
	if s == "*" {
		// Count elements with n and filter keyed types in track while we're at it.
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

		dst := make([]reflect.Value, n)
		writeIndex = 0
		for _, v := range track {
			switch v.Kind() {
			case reflect.Array, reflect.Slice, reflect.String:
				for i, n := 0, v.Len(); i < n; i++ {
					if x := follow(v.Index(i), buildCallbacks != nil); x != nil {
						dst[writeIndex] = *x
						writeIndex++
					}
				}

			case reflect.Map:
				for _, key := range v.MapKeys() {
					if x := followMap(v, key, buildCallbacks); x != nil {
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
			if k, err := strconv.ParseUint(s, 0, 64); err == nil && k < (1<<31) {
				i := int(k)
				if i >= v.Len() {
					if v.Kind() != reflect.Slice || !v.CanSet() {
						continue
					}
					n := i - v.Len() + 1
					v.Set(reflect.AppendSlice(v, reflect.MakeSlice(v.Type(), n, n)))
				}
				if x := follow(v.Index(i), buildCallbacks != nil); x != nil {
					track[writeIndex] = *x
					writeIndex++
				}
			}

		case reflect.Map:
			key := parseLiteral(s, v.Type().Key())
			if key != nil {
				if x := followMap(v, *key, buildCallbacks); x != nil {
					track[writeIndex] = *x
					writeIndex++
				}
			}

		}
	}
	return track[:writeIndex]
}

// follow returns content when possible.
func follow(v reflect.Value, doBuild bool) *reflect.Value {
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

// mapWrap re-SetMapIndex elements because modifications on elements won't apply without it.
type mapWrap struct{ m, k, v *reflect.Value }

func (w *mapWrap) Finish() {
	w.m.SetMapIndex(*w.k, *w.v)
}

func followMap(m reflect.Value, key reflect.Value, buildCallbacks *[]finisher) *reflect.Value {
	v := m.MapIndex(key)

	if buildCallbacks != nil {
		if !m.CanInterface() {
			return nil
		}

		if v.IsValid() {
			// Make addressable
			pv := reflect.New(v.Type()).Elem()
			pv.Set(v)
			v = pv
		} else {
			v = reflect.New(m.Type().Elem()).Elem()
		}

		*buildCallbacks = append(*buildCallbacks, &mapWrap{m: &m, k: &key, v: &v})
	}

	return follow(v, buildCallbacks != nil)
}

// parseLiteral returns the interpretation of s for t or nil on failure.
func parseLiteral(s string, t reflect.Type) *reflect.Value {
	var v reflect.Value

	switch t.Kind() {
	case reflect.String:
		if s, err := strconv.Unquote(s); err == nil {
			v = reflect.ValueOf(s)
		} else {
			return nil
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(s, 0, 64); err == nil {
			v = reflect.ValueOf(i)
		} else if s[0] == '\'' {
			r, _, tail, err := strconv.UnquoteChar(s[1:], '\'')
			if err != nil || tail != "'" {
				return nil
			}
			v = reflect.ValueOf(r)
		} else {
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, err := strconv.ParseUint(s, 0, 64); err == nil {
			v = reflect.ValueOf(u)
		} else if s[0] == '\'' {
			r, _, tail, err := strconv.UnquoteChar(s[1:], '\'')
			if err != nil || tail != "'" {
				return nil
			}
			v = reflect.ValueOf(r)
		} else {
			return nil
		}

	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			v = reflect.ValueOf(f)
		} else {
			return nil
		}

	default:
		p := reflect.New(t)
		if n, _ := fmt.Sscan(s, p.Interface()); n == 1 {
			v = p.Elem()
		} else {
			return nil
		}

	}

	if v.Type() != t {
		v = v.Convert(t)
	}
	return &v
}
