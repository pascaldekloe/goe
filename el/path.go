package el

import (
	"path"
	"reflect"
	"strconv"
	"strings"
)

func eval(expr string, root interface{}) []reflect.Value {
	v := reflect.ValueOf(root)

	expr = path.Clean(expr)
	if expr == "/" || expr == "." {
		return resolve(nil, v)
	}

	if expr[0] == '/' {
		expr = expr[1:]
	}
	return resolve(strings.Split(expr, "/"), v)
}

// resolve follows track for all path elements and returns the matches.
func resolve(segments []string, track ...reflect.Value) []reflect.Value {
	for _, segment := range segments {
		field, key := parseSegment(segment)
		if field != "" {
			track = follow(track)
			track = applyField(field, track)
		}
		if key != "" {
			track = follow(track)
			track = applyKey(key, track)
		}

		if len(track) == 0 {
			break
		}
	}

	return follow(track)
}

// parseSegment interprets a path element.
func parseSegment(s string) (field, key string) {
	last := len(s) - 1
	if s[last] != ']' {
		return s, ""
	}

	i := strings.IndexByte(s, '[')
	if i < 0 {
		return s, "" // matches nothing
	}

	return s[:i], s[i+1 : last]
}

func applyField(field string, track []reflect.Value) []reflect.Value {
	var gained []reflect.Value
	writeIndex := 0
	for _, v := range track {
		switch v.Kind() {
		case reflect.Struct:
			if field == "*" {
				for i := v.Type().NumField() - 1; i >= 0; i-- {
					gained = append(gained, v.Field(i))
				}
			} else {
				track[writeIndex] = v.FieldByName(field)
				writeIndex++
			}

		default:
		}
	}
	return append(track[:writeIndex], gained...)
}

func applyKey(key string, track []reflect.Value) []reflect.Value {
	var gained []reflect.Value
	writeIndex := 0
	for _, v := range track {
		switch v.Kind() {
		case reflect.Slice, reflect.Array, reflect.String:
			if key == "*" {
				for i, n := 0, v.Len(); i < n; i++ {
					gained = append(gained, v.Index(i))
				}
			} else {
				i, err := strconv.Atoi(key)
				if err == nil && i >= 0 && i < v.Len() {
					track[writeIndex] = v.Index(i)
					writeIndex++
				}
			}

		default:
			// BUG(pascaldekloe): Maps are ignored / unsupported.

		}
	}
	return append(track[:writeIndex], gained...)
}

func follow(matches []reflect.Value) []reflect.Value {
	writeIndex := 0
	for _, v := range matches {
		k := v.Kind()
		for k == reflect.Ptr || k == reflect.Interface {
			v = v.Elem()
			k = v.Kind()
		}

		if !v.IsValid() {
			continue
		}

		matches[writeIndex] = v
		writeIndex++
	}
	return matches[:writeIndex]
}
