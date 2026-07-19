// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns

import (
	"fmt"
	"slices"
	"strings"

	"github.com/choria-io/ui/internal/util"
)

// omap is an insertion-ordered string-keyed map used to build structured output
// so that JSON and YAML preserve the order entries were added in.
type omap struct {
	keys []string
	vals []any
}

func newOmap() *omap {
	return &omap{}
}

func (o *omap) set(k string, v any) {
	o.keys = append(o.keys, k)
	o.vals = append(o.vals, v)
}

func (o *omap) has(k string) bool {
	return slices.Contains(o.keys, k)
}

func (o *omap) len() int {
	return len(o.keys)
}

// rowStructured returns the raw structured value for a row: a single value for
// one entry, or a list for several. String leaves are sanitized so JSON and YAML
// cannot carry terminal escapes, matching the text and Markdown renderers.
func rowStructured(r row) any {
	if len(r.vals) == 1 {
		return sanitizeRaw(r.vals[0].raw())
	}

	list := make([]any, 0, len(r.vals))
	for _, v := range r.vals {
		list = append(list, sanitizeRaw(v.raw()))
	}

	return list
}

// sanitizeRaw strips control characters and escape sequences from the string
// leaves of a raw structured value, recursing through lists and Map objects.
// Non-string values keep their type, so numbers stay numbers and byte counts
// stay integers. Slices are copied so a caller's value is never mutated.
func sanitizeRaw(v any) any {
	switch t := v.(type) {
	case string:
		return util.Sanitize(t)

	case []string:
		out := make([]string, len(t))
		for i, s := range t {
			out[i] = util.Sanitize(s)
		}

		return out

	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = sanitizeRaw(e)
		}

		return out

	case *omap:
		for i := range t.vals {
			t.vals[i] = sanitizeRaw(t.vals[i])
		}

		return t

	default:
		return v
	}
}

// document builds the top-level ordered object shared by JSON and YAML. Headings
// nest by indent: a Section or Indent places a heading inside its enclosing
// heading, so the structure mirrors the text and Markdown output. Rows attach to
// the heading at their indent, and rows added before any heading become top-level
// keys. Two entries that would share a key within the same object, whether rows
// or sub-headings, return ErrDuplicateDescription.
func (d *Document) document() (*omap, error) {
	top := newOmap()

	type frame struct {
		indent int
		name   string
		obj    *omap
	}
	stack := []frame{{indent: -1, obj: top}}

	containerPath := func() string {
		if len(stack) <= 1 {
			return ""
		}

		names := make([]string, 0, len(stack)-1)
		for _, f := range stack[1:] {
			names = append(names, f.name)
		}

		return strings.Join(names, " > ")
	}

	insert := func(o *omap, k string, v any) error {
		if o.has(k) {
			if path := containerPath(); path != "" {
				return fmt.Errorf("%w: %q in %q", ErrDuplicateDescription, k, path)
			}

			return fmt.Errorf("%w: %q", ErrDuplicateDescription, k)
		}
		o.set(k, v)

		return nil
	}

	for _, r := range d.rows {
		switch r.kind {
		case kindHeading:
			for len(stack) > 1 && stack[len(stack)-1].indent >= r.indent {
				stack = stack[:len(stack)-1]
			}

			_, name := styleHeading(r.desc)
			obj := newOmap()
			if err := insert(stack[len(stack)-1].obj, name, obj); err != nil {
				return nil, err
			}
			stack = append(stack, frame{indent: r.indent, name: name, obj: obj})

		case kindRow:
			for len(stack) > 1 && stack[len(stack)-1].indent > r.indent {
				stack = stack[:len(stack)-1]
			}

			if err := insert(stack[len(stack)-1].obj, r.desc, rowStructured(r)); err != nil {
				return nil, err
			}

		case kindEmbed:
			for len(stack) > 1 && stack[len(stack)-1].indent > r.indent {
				stack = stack[:len(stack)-1]
			}

			if err := insert(stack[len(stack)-1].obj, r.desc, r.embed); err != nil {
				return nil, err
			}

		case kindBlank, kindLine:
			// no structural effect
		}
	}

	return top, nil
}
