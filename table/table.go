// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package table renders tabular CLI output on top of go-pretty.
//
// A Table accumulates headers, rows, an optional footer and separators, then
// renders them as an aligned text table, a Markdown table, JSON or YAML. The
// default rendering is text, switching to Markdown for an LLM or agent consumer
// using the same LLMFORMAT and CLAUDECODE signals as the columns package.
//
//	t := table.NewTableWriter("Servers")
//	t.AddHeaders("Name", "Cores", "Memory")
//	t.AddRow("web1", 8, columns.IBytes(1610612736))
//	fmt.Print(t.Render())
//
// Cells are formatted the way the columns package formats values (thousands
// separators for numbers, humanized durations and times) and every cell has
// terminal escape sequences and control characters stripped. A *Table satisfies
// the columns.Embeddable interface, so it can be passed to columns.Document.Item
// to nest a table inside a document.
package table

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/choria-io/ui/internal/util"
)

// Table is an in-memory representation of a table. Build it with AddHeaders,
// AddRow, AddSeparator and AddFooter, then render it with Render, String, Bytes,
// WriteTo, JSON, YAML or Markdown. A Table is not safe for concurrent use.
type Table struct {
	title   string
	headers []any
	footer  []any
	rows    []tableRow
}

type tableRow struct {
	cells []any
	sep   bool
}

// NewTableWriter creates a Table with an optional centered title.
func NewTableWriter(title string) *Table {
	return &Table{title: title}
}

// NewTableWriterf is NewTableWriter with the title formatted by fmt.Sprintf.
func NewTableWriterf(format string, a ...any) *Table {
	return NewTableWriter(fmt.Sprintf(format, a...))
}

// AddHeaders sets the column headers. Calling it again replaces them.
func (t *Table) AddHeaders(items ...any) *Table {
	t.headers = append([]any(nil), items...)

	return t
}

// AddFooter sets a footer row, typically holding totals or a summary. Calling it
// again replaces it.
func (t *Table) AddFooter(items ...any) *Table {
	t.footer = append([]any(nil), items...)

	return t
}

// AddSeparator adds a horizontal rule between rows in the text and Markdown
// output. It is ignored by the JSON and YAML renderers.
func (t *Table) AddSeparator() *Table {
	t.rows = append(t.rows, tableRow{sep: true})

	return t
}

// AddRow appends a data row.
func (t *Table) AddRow(items ...any) *Table {
	t.rows = append(t.rows, tableRow{cells: append([]any(nil), items...)})

	return t
}

// String renders the table as text, or as Markdown when util.LLMFormatEnabled
// reports an LLM or agent consumer, matching the columns package.
func (t *Table) String() string {
	if util.LLMFormatEnabled() {
		b, err := t.Markdown()
		if err == nil {
			return string(b)
		}
	}

	return t.renderText()
}

// Render is an alias for String, provided for callers used to go-pretty.
func (t *Table) Render() string {
	return t.String()
}

// Bytes renders the table with the same format selection as String.
func (t *Table) Bytes() []byte {
	return []byte(t.String())
}

// WriteTo renders the table with the same format selection as String and writes
// it to w. It satisfies io.WriterTo, so it works as defer t.WriteTo(os.Stdout).
func (t *Table) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, t.String())

	return int64(n), err
}

func (t *Table) renderText() string {
	s := t.writer().Render()
	if s == "" {
		return ""
	}

	return s + "\n"
}

// Markdown renders the table as a Markdown table. The error is always nil; it is
// present so Table satisfies the columns.Embeddable interface.
func (t *Table) Markdown() ([]byte, error) {
	s := t.writer().RenderMarkdown()
	if s == "" {
		return nil, nil
	}

	return []byte(s + "\n"), nil
}

// RenderMarkdown writes the Markdown rendering to w.
func (t *Table) RenderMarkdown(w io.Writer) error {
	b, err := t.Markdown()
	if err != nil {
		return err
	}
	_, err = w.Write(b)

	return err
}

// JSON renders the table as indented JSON with the shape
// {title, headers, rows, footer}, omitting title, headers and footer when unset.
// Plain scalar cells keep their native type; other values use their display form.
func (t *Table) JSON() ([]byte, error) {
	b, err := json.MarshalIndent(t.structured(), "", "  ")
	if err != nil {
		return nil, err
	}

	return append(b, '\n'), nil
}

// RenderJSON writes the JSON rendering to w.
func (t *Table) RenderJSON(w io.Writer) error {
	b, err := t.JSON()
	if err != nil {
		return err
	}
	_, err = w.Write(b)

	return err
}

// YAML renders the table as YAML with the same shape as JSON.
func (t *Table) YAML() ([]byte, error) {
	return yaml.Marshal(t.structured())
}

// RenderYAML writes the YAML rendering to w.
func (t *Table) RenderYAML(w io.Writer) error {
	b, err := t.YAML()
	if err != nil {
		return err
	}
	_, err = w.Write(b)

	return err
}

// writer builds a go-pretty writer from the captured model, formatting and
// sanitizing every cell so untrusted content cannot inject terminal escapes.
func (t *Table) writer() table.Writer {
	w := table.NewWriter()
	w.SuppressTrailingSpaces()
	w.SetStyle(table.StyleRounded)
	w.Style().Title.Align = text.AlignCenter
	w.Style().Format.Header = text.FormatDefault
	w.Style().Format.Footer = text.FormatDefault

	if t.title != "" {
		w.SetTitle("%s", util.Sanitize(t.title))
	}

	if len(t.headers) > 0 {
		w.AppendHeader(toRow(t.headers))
	}

	for _, r := range t.rows {
		if r.sep {
			w.AppendSeparator()
			continue
		}
		w.AppendRow(toRow(r.cells))
	}

	if len(t.footer) > 0 {
		w.AppendFooter(toRow(t.footer))
	}

	return w
}

// toRow formats and sanitizes each cell for display in the go-pretty table.
func toRow(cells []any) table.Row {
	row := make(table.Row, len(cells))
	for i, c := range cells {
		row[i] = util.Sanitize(util.Format(c))
	}

	return row
}

// tableDoc is the structured form shared by the JSON and YAML renderers. Both
// encoders preserve the field order.
type tableDoc struct {
	Title   string  `json:"title,omitempty" yaml:"title,omitempty"`
	Headers []any   `json:"headers,omitempty" yaml:"headers,omitempty"`
	Rows    [][]any `json:"rows" yaml:"rows"`
	Footer  []any   `json:"footer,omitempty" yaml:"footer,omitempty"`
}

// structured builds the JSON/YAML model. Plain scalars keep their native type so
// numbers stay numbers; other values, including formatting helpers, use their
// sanitized display string. Separators are dropped.
func (t *Table) structured() tableDoc {
	doc := tableDoc{Title: util.Sanitize(t.title), Rows: [][]any{}}

	if len(t.headers) > 0 {
		doc.Headers = make([]any, len(t.headers))
		for i, h := range t.headers {
			doc.Headers[i] = util.Sanitize(util.Format(h))
		}
	}

	for _, r := range t.rows {
		if r.sep {
			continue
		}
		cells := make([]any, len(r.cells))
		for i, c := range r.cells {
			cells[i] = rawCell(c)
		}
		doc.Rows = append(doc.Rows, cells)
	}

	if len(t.footer) > 0 {
		doc.Footer = make([]any, len(t.footer))
		for i, c := range t.footer {
			doc.Footer[i] = rawCell(c)
		}
	}

	return doc
}

// rawCell keeps JSON-native scalars as their own type and renders anything else
// through the shared formatter, sanitizing string leaves.
func rawCell(c any) any {
	switch x := c.(type) {
	case nil:
		return nil
	case string:
		return util.Sanitize(x)
	case bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return x
	case []byte:
		return util.Sanitize(string(x))
	default:
		return util.Sanitize(util.Format(x))
	}
}
