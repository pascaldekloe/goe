package el

import (
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"
)

func eval(expr string, root interface{}) []reflect.Value {
	v := []reflect.Value{reflect.ValueOf(root)}
	follow(v)

	expr = path.Clean(expr)
	if expr == "/" || expr == "." {
		return v
	}

	if expr[0] == '/' {
		expr = expr[1:]
	}
	return resolve(strings.Split(expr, "/"), v)
}

// resolve follows track for all path elements and returns the matches.
func resolve(segments []string, track []reflect.Value) []reflect.Value {
	for _, segment := range segments {
		field, key := parseSegment(segment)
		if field != "" {
			track = applyField(field, track)
		}
		if key != "" {
			track = applyKey(key, track)
		}

		if len(track) == 0 {
			break
		}
	}

	return track
}

// parseSegment interprets a path element.
func parseSegment(s string) (field, key string) {
	field = s

	if last := len(s) - 1; s[last] == ']' {
		if i := strings.IndexByte(s, '['); i >= 0 {
			field = s[:i]
			key = s[i+1 : last]
		}
	}

	if field == "." {
		field = ""
	}

	return
}

// applyField returs the field matches from all elements in track.
func applyField(field string, track []reflect.Value) []reflect.Value {
	if field != "*" {
		// Write values back to track with writeIndex.
		writeIndex := 0
		for _, v := range track {
			if v.Kind() == reflect.Struct {
				track[writeIndex] = v.FieldByName(field)
				writeIndex++
			}
		}
		return follow(track[:writeIndex])
	}

	// Filter struct types in track and count values with n.
	writeIndex, n := 0, 0
	for _, v := range track {
		if v.Kind() == reflect.Struct {
			n += v.Type().NumField()
			track[writeIndex] = v
			writeIndex++
		}
	}
	track = track[:writeIndex]

	// Collect all values in found
	found := make([]reflect.Value, n)
	for _, v := range track {
		for i := v.Type().NumField() - 1; i >= 0; i-- {
			n--
			found[n] = v.Field(i)
		}
	}
	return follow(found)
}

// applyField returs the key matches from all elements in track.
func applyKey(key string, track []reflect.Value) []reflect.Value {
	if key != "*" {
		// Write values back to track with writeIndex.
		writeIndex := 0
		for _, v := range track {
			switch v.Kind() {
			case reflect.Array, reflect.Slice, reflect.String:
				i, err := strconv.ParseUint(key, 0, 64)
				if err == nil && i < uint64(v.Len()) {
					track[writeIndex] = v.Index(int(i))
					writeIndex++
				}

			case reflect.Map:
				kt := v.Type().Key()
				kp := reflect.New(kt)
				switch kt.Kind() {
				case reflect.String:
					if s, err := strconv.Unquote(key); err == nil {
						kp.Elem().SetString(s)
					}

				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if i, err := strconv.ParseInt(key, 0, 64); err == nil {
						kp.Elem().SetInt(i)
					}

				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if u, err := strconv.ParseUint(key, 0, 64); err == nil {
						kp.Elem().SetUint(u)
					}

				case reflect.Float32, reflect.Float64:
					if f, err := strconv.ParseFloat(key, 64); err == nil {
						kp.Elem().SetFloat(f)
					}

				default:
					ptr := kp.Interface()
					if n, _ := fmt.Sscan(key, ptr); n == 1 {
						kp = reflect.ValueOf(ptr)
					}

				}
				track[writeIndex] = v.MapIndex(kp.Elem())
				writeIndex++

			}
		}
		return follow(track[:writeIndex])
	}

	// Filter keyed types in track and count values with n.
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

	// Collect all content in found.
	found := make([]reflect.Value, n)
	for _, v := range track {
		switch v.Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
			for i := v.Len() - 1; i >= 0; i-- {
				n--
				found[n] = v.Index(i)
			}
		case reflect.Map:
			for _, k := range v.MapKeys() {
				n--
				found[n] = v.MapIndex(k)
			}
		}
	}
	return follow(found)
}

// folow returns the usable content content in track.
// Pointers are resolved and invalid values are discarded.
func follow(track []reflect.Value) []reflect.Value {
	writeIndex := 0
	for _, v := range track {
		k := v.Kind()
		for k == reflect.Ptr || k == reflect.Interface {
			v = v.Elem()
			k = v.Kind()
		}
		if k != reflect.Invalid {
			track[writeIndex] = v
			writeIndex++
		}
	}
	return track[:writeIndex]
}
