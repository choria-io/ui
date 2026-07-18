// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// JSON renders the Document as indented JSON. Headings nest by indent, so a
// Section places its object inside the enclosing heading and the structure
// mirrors the text and Markdown output; rows before any heading become top-level
// keys. Keys keep the order they were added. Raw values are preserved, so numbers
// stay numbers and byte counts stay integers. It returns ErrDuplicateDescription
// if two entries would share a key within the same object.
func (d *Document) JSON() ([]byte, error) {
	if d.err != nil {
		return nil, d.err
	}

	top, err := d.document()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := encodeJSON(&buf, top, 0); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

// RenderJSON writes the JSON rendering to w.
func (d *Document) RenderJSON(w io.Writer) error {
	b, err := d.JSON()
	if err != nil {
		return err
	}
	_, err = w.Write(b)

	return err
}

func encodeJSON(buf *bytes.Buffer, v any, depth int) error {
	switch t := v.(type) {
	case *omap:
		return encodeJSONObject(buf, t, depth)
	case []any:
		return encodeJSONArray(buf, t, depth)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		buf.Write(b)

		return nil
	}
}

func encodeJSONObject(buf *bytes.Buffer, o *omap, depth int) error {
	if o.len() == 0 {
		buf.WriteString("{}")
		return nil
	}

	buf.WriteString("{\n")
	inner := strings.Repeat("  ", depth+1)
	for i, k := range o.keys {
		buf.WriteString(inner)

		kb, err := json.Marshal(k)
		if err != nil {
			return err
		}
		buf.Write(kb)
		buf.WriteString(": ")

		if err := encodeJSON(buf, o.vals[i], depth+1); err != nil {
			return err
		}

		if i < len(o.keys)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString(strings.Repeat("  ", depth))
	buf.WriteByte('}')

	return nil
}

func encodeJSONArray(buf *bytes.Buffer, list []any, depth int) error {
	if len(list) == 0 {
		buf.WriteString("[]")
		return nil
	}

	buf.WriteString("[\n")
	inner := strings.Repeat("  ", depth+1)
	for i, item := range list {
		buf.WriteString(inner)

		if err := encodeJSON(buf, item, depth+1); err != nil {
			return err
		}

		if i < len(list)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString(strings.Repeat("  ", depth))
	buf.WriteByte(']')

	return nil
}
