// Package verify offers convenience routenes for content verification.
package verify

import (
	"bytes"
	"fmt"
	"path"
	"runtime"
)

// travel is the verification state
type travel struct {
	diffs []differ
}

// differ is a verification failure.
type differ struct {
	// path is the expression to the content.
	path string
	// msg has a reason.
	msg string
}

// segment is a differ.path component used for lazy formatting.
type segment struct {
	format string
	x      interface{}
}

func (t *travel) differ(path []*segment, msg string, args ...interface{}) {
	var buf bytes.Buffer
	for _, s := range path {
		buf.WriteString(fmt.Sprintf(s.format, s.x))
	}

	t.diffs = append(t.diffs, differ{
		msg:  fmt.Sprintf(msg, args...),
		path: buf.String(),
	})
}

func (t *travel) report(test, name string) string {
	if len(t.diffs) == 0 {
		return ""
	}

	var buf bytes.Buffer

	if _, file, lineno, ok := runtime.Caller(2); ok {
		fmt.Fprintf(&buf, "%s at %s:%d: ", test, path.Base(file), lineno)
	}

	buf.WriteString("verification for ")
	buf.WriteString(name)
	buf.WriteByte(':')

	for _, d := range t.diffs {
		buf.WriteByte('\n')
		if d.path != "" {
			buf.WriteString(d.path)
			buf.WriteString(": ")
		}
		buf.WriteString(d.msg)
	}

	return buf.String()
}
