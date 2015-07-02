package el

import (
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"
)

// resolve follows expr on root.
func resolve(expr string, root interface{}, buildCallbacks *[]finisher) (track []reflect.Value) {
	if v, k := follow(reflect.ValueOf(root), buildCallbacks != nil); k != reflect.Invalid {
		track = []reflect.Value{v}
	}

	segments := strings.Split(path.Clean(expr), "/")[1:]
	if segments[0] == "" { // root selection
		return track
	}

	for _, selection := range segments {
		if len(track) == 0 {
			return nil
		}

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

	writeIndex := 0
	if buildCallbacks == nil {
		for _, v := range track {
			v, kind := follow(v, buildCallbacks != nil)
			if kind != reflect.Invalid {
				track[writeIndex] = v
				writeIndex++
			}
		}
	} else {
		for _, v := range track {
			for {
				if v.Kind() != reflect.Ptr {
					track[writeIndex] = v
					writeIndex++
					break
				}
				if v.IsNil() {
					if !v.CanSet() {
						break
					}
					v.Set(reflect.New(v.Type().Elem()))
				}
				v = v.Elem()
			}
		}
	}
	return track[:writeIndex]
}

// followField returns all fields matching s from track.
func followField(track []reflect.Value, s string, doBuild bool) []reflect.Value {
	if s == "*" {
		// Count fields with n and filter struct types in track while we're at it.
		writeIndex, n := 0, 0
		for _, v := range track {
			v, kind := follow(v, doBuild)
			if kind == reflect.Struct {
				n += v.Type().NumField()
				track[writeIndex] = v
				writeIndex++
			}
		}
		track = track[:writeIndex]

		dst := make([]reflect.Value, n)
		for _, v := range track {
			for i := v.Type().NumField() - 1; i >= 0; i-- {
				n--
				dst[n] = v.Field(i)
				writeIndex++
			}
		}
		return dst
	}

	// Write result back to track with writeIndex to safe memory.
	writeIndex := 0
	for _, v := range track {
		v, kind := follow(v, doBuild)
		if kind == reflect.Struct {
			track[writeIndex] = v.FieldByName(s)
			writeIndex++
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
			v, kind := follow(v, buildCallbacks != nil)
			switch kind {
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
					dst[writeIndex] = v.Index(i)
					writeIndex++
				}

			case reflect.Map:
				for _, key := range v.MapKeys() {
					if e := followMap(v, key, buildCallbacks); e != nil {
						dst[writeIndex] = *e
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
		v, kind := follow(v, buildCallbacks != nil)
		switch kind {
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
				track[writeIndex] = v.Index(i)
				writeIndex++
			}

		case reflect.Map:
			if key := parseLiteral(s, v.Type().Key()); key != nil {
				if e := followMap(v, *key, buildCallbacks); e != nil {
					track[writeIndex] = *e
					writeIndex++
				}
			}

		}
	}
	return track[:writeIndex]
}

// follow tracks content. If the returned kind is invalid the value must be skipped.
func follow(v reflect.Value, doBuild bool) (reflect.Value, reflect.Kind) {
	for {
		k := v.Kind()
		for k == reflect.Interface {
			if v.IsNil() {
				return v, reflect.Invalid
			}
			v = v.Elem()
			k = v.Kind()
		}

		if k == reflect.Ptr {
			if v.IsNil() {
				if !doBuild || !v.CanSet() {
					return v, reflect.Invalid
				}
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
			k = v.Kind()
			continue
		}

		if k == reflect.Map && v.IsNil() {
			if !doBuild || !v.CanSet() {
				return v, reflect.Invalid
			}
			v.Set(reflect.MakeMap(v.Type()))
		}

		return v, k
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

	return &v
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
