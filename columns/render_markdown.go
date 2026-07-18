// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns

import (
	"io"
	"strings"
)

// Markdown renders the Document as Markdown. Each heading becomes an ATX heading
// whose level tracks the indent depth: a top-level heading is level two and each
// further Indent or Section adds one, clamped at level six. Every row becomes a
// bullet of the form "- **description:** value"; a row that resolves to more than
// one value line (a Map, a Values list or a value containing newlines) lists the
// lines as nested bullets under the description. Markdown metacharacters in
// descriptions and values are escaped so they render literally. Rows added before
// any heading form a leading bullet list with no heading.
func (d *Document) Markdown() ([]byte, error) {
	if d.err != nil {
		return nil, d.err
	}

	plainOpts := d.opts
	plainOpts.plain = true

	var blocks []string
	var list []string

	flushList := func() {
		if len(list) > 0 {
			blocks = append(blocks, strings.Join(list, "\n"))
			list = nil
		}
	}

	for _, r := range d.rows {
		switch r.kind {
		case kindHeading:
			flushList()
			level := min(2+r.indent, 6)
			_, plain := styleHeading(r.desc)
			blocks = append(blocks, strings.Repeat("#", level)+" "+mdInline(plain))

		case kindBlank:
			// Markdown manages its own spacing; explicit blanks are ignored.

		case kindLine:
			flushList()
			_, plain := renderMarkup(r.desc)
			blocks = append(blocks, plain)

		case kindRow:
			lines := d.rowLines(r, &plainOpts)
			if isAllEmpty(lines) {
				if d.opts.omitEmpty {
					continue
				}
				lines = []string{d.opts.emptyText}
			}
			list = append(list, mdBullet(r.desc, lines)...)
		}
	}
	flushList()

	return []byte(strings.Join(blocks, "\n\n") + "\n"), nil
}

// RenderMarkdown writes the Markdown rendering to w.
func (d *Document) RenderMarkdown(w io.Writer) error {
	b, err := d.Markdown()
	if err != nil {
		return err
	}
	_, err = w.Write(b)

	return err
}

// mdBullet renders one row as a Markdown bullet. A single value line becomes
// "- **desc:** value"; more than one line becomes "- **desc:**" followed by one
// nested bullet per line. An empty value emits just "- **desc:**".
func mdBullet(desc string, lines []string) []string {
	label := "- **" + mdInline(desc) + ":**"

	if len(lines) <= 1 {
		val := ""
		if len(lines) == 1 {
			val = mdInline(lines[0])
		}
		if val == "" {
			return []string{label}
		}

		return []string{label + " " + val}
	}

	out := make([]string, 0, len(lines)+1)
	out = append(out, label)
	for _, ln := range lines {
		out = append(out, strings.TrimRight("    - "+mdChildLine(ln), " "))
	}

	return out
}

// mdInline escapes the Markdown metacharacters that would otherwise change the
// rendering of inline text, and collapses newlines so the text stays on one line.
func mdInline(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 8)
	for _, r := range s {
		switch r {
		case '\\', '`', '*', '_', '[', '<':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\r':
			// dropped; a following newline becomes the single space
		case '\n':
			b.WriteByte(' ')
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

// mdChildLine escapes a value line for use as a nested bullet: inline escaping
// plus neutralizing a leading block marker so a value such as "- restart" or ">"
// cannot start a nested list or blockquote.
func mdChildLine(s string) string {
	return escapeLeadingMarker(mdInline(s))
}

// escapeLeadingMarker backslash-escapes a leading character that Markdown treats
// as a list, heading or blockquote marker at the start of a bullet's content.
func escapeLeadingMarker(s string) string {
	if s == "" {
		return s
	}

	switch s[0] {
	case '-', '+', '#', '>':
		return "\\" + s
	}

	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i > 0 && i < len(s) && (s[i] == '.' || s[i] == ')') {
		return s[:i] + "\\" + s[i:]
	}

	return s
}
