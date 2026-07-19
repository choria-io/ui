// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns

import (
	"io"
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/choria-io/ui/internal/util"
)

// String renders the Document. By default this is aligned text where every
// description colon lines up across all headings, a blank line separates each
// heading from its body and an indented section shifts its whole block to the
// right. When llmFormatEnabled reports an LLM or agent consumer, the Markdown
// rendering is returned instead; see render.
func (d *Document) String() string {
	return d.render()
}

// Bytes renders the Document with the same format selection as String and
// returns the bytes.
func (d *Document) Bytes() []byte {
	return []byte(d.render())
}

// WriteTo renders the Document with the same format selection as String and
// writes it to w. It satisfies io.WriterTo.
func (d *Document) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, d.render())

	return int64(n), err
}

// render returns the Document's default rendering. This is aligned text unless
// llmFormatEnabled reports an LLM or agent consumer, in which case the Markdown
// rendering is returned. If Markdown rendering fails, such as when a builder
// method recorded an error, it falls back to text so no output is lost.
func (d *Document) render() string {
	if util.LLMFormatEnabled() {
		md, err := d.Markdown()
		if err == nil {
			return string(md)
		}
	}

	return d.renderText()
}

// visibleRow is a row with its value lines resolved, ready to measure and emit.
type visibleRow struct {
	kind   rowKind
	indent int
	desc   string
	lines  []string
}

func (d *Document) renderText() string {
	visible := d.visibleRows()

	descWidth := 0
	for _, r := range visible {
		if r.kind != kindRow && r.kind != kindEmbed {
			continue
		}
		if w := runewidth.StringWidth(r.desc); w > descWidth {
			descWidth = w
		}
	}

	iw := d.opts.indentWidth
	margin := d.opts.margin

	var b strings.Builder
	for i, r := range visible {
		switch r.kind {
		case kindHeading:
			ansi, _ := styleHeading(r.desc)
			b.WriteString(strings.Repeat(" ", r.indent*iw))
			b.WriteString(ansi)
			b.WriteString(":")
			b.WriteByte('\n')
			if headingWantsBlank(visible, i) {
				b.WriteByte('\n')
			}

		case kindBlank:
			b.WriteByte('\n')

		case kindLine:
			ansi, _ := renderMarkup(r.desc)
			// Inside a section a free-form line sits one level below the heading
			// so it reads as part of the section; at the top level it stays at
			// the left margin.
			pad := r.indent * iw
			if r.indent > 0 {
				pad += iw
			}
			b.WriteString(strings.Repeat(" ", pad))
			b.WriteString(ansi)
			b.WriteByte('\n')

		case kindRow:
			leftPad := strings.Repeat(" ", margin+r.indent*iw)
			descField := padLeft(r.desc, descWidth)
			contPad := strings.Repeat(" ", margin+r.indent*iw+descWidth+len(sep))

			for i, ln := range r.lines {
				if i == 0 {
					b.WriteString(rtrim(leftPad + descField + sep + ln))
				} else {
					b.WriteString(rtrim(contPad + ln))
				}
				b.WriteByte('\n')
			}

		case kindEmbed:
			// The embedded block is placed on the lines below the description,
			// aligned to the value column, since it is generally too wide to sit
			// beside the description.
			leftPad := strings.Repeat(" ", margin+r.indent*iw)
			descField := padLeft(r.desc, descWidth)
			contPad := strings.Repeat(" ", margin+r.indent*iw+descWidth+len(sep))

			b.WriteString(rtrim(leftPad + descField + ":"))
			b.WriteByte('\n')
			for _, ln := range r.lines {
				b.WriteString(rtrim(contPad + ln))
				b.WriteByte('\n')
			}
		}
	}

	return b.String()
}

// headingWantsBlank reports whether a blank line should follow the heading at
// visible[i]. Text output separates every heading from its body with a blank
// line, but it is omitted when the heading ends the document or the next row is
// already a blank, so an explicit Blank is never doubled.
func headingWantsBlank(visible []visibleRow, i int) bool {
	if i+1 >= len(visible) {
		return false
	}

	return visible[i+1].kind != kindBlank
}

// visibleRows resolves each row's value lines and applies the omit-empty and
// empty-text options, dropping rows that should not appear.
func (d *Document) visibleRows() []visibleRow {
	out := make([]visibleRow, 0, len(d.rows))
	for _, r := range d.rows {
		switch r.kind {
		case kindHeading, kindBlank, kindLine:
			out = append(out, visibleRow{kind: r.kind, indent: r.indent, desc: r.desc})

		case kindRow:
			lines := d.rowLines(r, &d.opts)
			if isAllEmpty(lines) {
				if d.opts.omitEmpty {
					continue
				}
				if d.opts.emptyText != "" {
					lines = []string{d.opts.emptyText}
				}
			}
			out = append(out, visibleRow{kind: kindRow, indent: r.indent, desc: r.desc, lines: lines})

		case kindEmbed:
			lines := trimTrailingBlank(strings.Split(util.Sanitize(r.embed.String()), "\n"))
			out = append(out, visibleRow{kind: kindEmbed, indent: r.indent, desc: r.desc, lines: lines})
		}
	}

	return out
}

// trimTrailingBlank drops trailing empty lines, such as the one produced by a
// renderer that ends its output with a newline.
func trimTrailingBlank(lines []string) []string {
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// rowLines gathers the display lines for every value on a row in order, using
// the given options so callers can request the plain, unstyled form.
func (d *Document) rowLines(r row, o *options) []string {
	var out []string
	for _, v := range r.vals {
		out = append(out, v.lines(o)...)
	}
	if len(out) == 0 {
		out = []string{""}
	}

	return out
}

func isAllEmpty(lines []string) bool {
	for _, l := range lines {
		if l != "" {
			return false
		}
	}

	return true
}

// valueLines sanitizes a display string and splits it into physical lines so
// each embedded newline becomes a continuation line at the value column.
func valueLines(s string) []string {
	s = util.Sanitize(s)
	if s == "" {
		return []string{""}
	}

	return strings.Split(s, "\n")
}

// padLeft right-justifies s in a field of the given display-cell width.
func padLeft(s string, width int) string {
	pad := width - runewidth.StringWidth(s)
	if pad <= 0 {
		return s
	}

	return strings.Repeat(" ", pad) + s
}

// padRight left-justifies s in a field of the given display-cell width.
func padRight(s string, width int) string {
	pad := width - runewidth.StringWidth(s)
	if pad <= 0 {
		return s
	}

	return s + strings.Repeat(" ", pad)
}

func rtrim(s string) string {
	return strings.TrimRight(s, " ")
}
