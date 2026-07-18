// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns

import (
	"io"

	"github.com/goccy/go-yaml"
)

// YAML renders the Document as YAML. Headings nest by indent, so a Section
// places its object inside the enclosing heading and the structure mirrors the
// text and Markdown output; rows before any heading become top-level keys. Keys
// keep the order they were added and raw values are preserved. It returns
// ErrDuplicateDescription if two entries would share a key within the same
// object.
func (d *Document) YAML() ([]byte, error) {
	if d.err != nil {
		return nil, d.err
	}

	top, err := d.document()
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(toYAML(top))
}

// RenderYAML writes the YAML rendering to w.
func (d *Document) RenderYAML(w io.Writer) error {
	b, err := d.YAML()
	if err != nil {
		return err
	}
	_, err = w.Write(b)

	return err
}

// toYAML converts the ordered structured model into goccy MapSlice values so
// key order is preserved by the marshaller.
func toYAML(v any) any {
	switch t := v.(type) {
	case *omap:
		ms := make(yaml.MapSlice, 0, t.len())
		for i, k := range t.keys {
			ms = append(ms, yaml.MapItem{Key: k, Value: toYAML(t.vals[i])})
		}

		return ms
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = toYAML(item)
		}

		return out
	default:
		return v
	}
}
