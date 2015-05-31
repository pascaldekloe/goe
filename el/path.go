package el

import (
	"path"
	"reflect"
	"strconv"
	"strings"
)

// BUG(pascaldekloe): Relative paths are not supported yet.
func eval(expr string, root interface{}) (result []reflect.Value) {
	expr = path.Clean(expr)
	if !path.IsAbs(expr) {
		return // TODO
	}

	return resolve(strings.Split(expr, "/")[1:], root)
}

// BUG(pascaldekloe): Maps are ignored.
func resolve(path []string, root interface{}) (matches []reflect.Value) {
	matches = []reflect.Value{reflect.ValueOf(root)}
	for _, segment := range path {
		if segment == "" {
			continue
		}

		field, key := parseSegment(segment)

		if field != "" {
			matches = follow(matches)
			matches = applyField(field, matches)
		}

		if key != "" {
			matches = follow(matches)
			matches = applyKey(key, matches)
		}

		if len(matches) == 0 {
			return
		}
	}

	matches = follow(matches)
	return
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

func applyField(field string, matches []reflect.Value) []reflect.Value {
	var gained []reflect.Value
	writeIndex := 0
	for _, v := range matches {
		switch v.Kind() {
		case reflect.Struct:
			if field == "*" {
				for i := v.Type().NumField() - 1; i >= 0; i-- {
					gained = append(gained, v.Field(i))
				}
			} else {
				matches[writeIndex] = v.FieldByName(field)
				writeIndex++
			}
		default:
		}
	}
	return append(matches[:writeIndex], gained...)
}

func applyKey(key string, matches []reflect.Value) []reflect.Value {
	var gained []reflect.Value
	writeIndex := 0
	for _, v := range matches {
		switch v.Kind() {
		case reflect.Slice, reflect.Array, reflect.String:
			if key == "*" {
				for i, n := 0, v.Len(); i < n; i++ {
					gained = append(gained, v.Index(i))
				}
			} else {
				i, err := strconv.Atoi(key)
				if err == nil && i >= 0 && i < v.Len() {
					matches[writeIndex] = v.Index(i)
					writeIndex++
				}
			}
		default:
		}
	}
	return append(matches[:writeIndex], gained...)
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
