// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package columns builds aligned, columnar CLI output.
//
// Rather than hand tuning a series of fmt.Printf calls whose padding all has to
// be recalculated whenever a description changes, a Document stores the raw
// values you add to it and defers formatting until render time. The default
// rendering is aligned text where every description colon lines up across all
// headings; the same Document can also be rendered to JSON, YAML and Markdown.
//
// The pattern is: create a Document, add entries line by line, then render.
//
//	d := columns.New()
//	d.Heading("Node")
//	d.Item("Name", "web1.example.net")
//	d.Item("Memory", columns.IBytes(1610612736))
//	fmt.Print(d.String())
//
// Values passed to Item, Values and Fields are formatted automatically by
// Format: strings are used as is, booleans become true/false, times and
// durations render in human form and numbers get thousands separators. Because a
// bare int64 is ambiguous, the helper functions (IBytes, Bytes, Duration,
// Percent, Plain and others) let the caller state the intent explicitly while
// preserving the raw value for the structured renderers. Style opts a value into
// terminal markup such as {bold}...{/bold}.
package columns

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

// sep separates a description from its value in text output. The documented
// column positions ("the colon lines up across headings") assume this value.
const sep = ": "

// Sentinel errors returned by the structured renderers and recorded on the
// Document by builder methods that receive invalid input.
var (
	// ErrInvalidInput is recorded when a builder method is given a value it
	// cannot use, such as a non-map passed to Fields.
	ErrInvalidInput = errors.New("columns: invalid input")

	// ErrDuplicateDescription is returned by the JSON and YAML renderers when two
	// entries would share a key within the same object, whether two rows, two
	// sub-headings or a row and a sub-heading, since a key cannot be repeated in
	// an object.
	ErrDuplicateDescription = errors.New("columns: duplicate description")
)

// Document is an in-memory representation of columnar output. Build it up with
// the Heading, Item, Values, Fields, Blank and Section methods, then render it
// with String, WriteTo, JSON, YAML or Markdown. A Document is not safe for
// concurrent use.
type Document struct {
	opts   options
	rows   []row
	indent int
	err    error
}

type rowKind int

const (
	kindHeading rowKind = iota
	kindRow
	kindBlank
	kindLine
	kindEmbed
)

type row struct {
	kind   rowKind
	indent int
	desc   string
	vals   []Value
	embed  Embeddable
}

// New creates an empty Document. See the With* functions for available options.
func New(opts ...Option) *Document {
	d := &Document{opts: defaultOptions()}
	for _, o := range opts {
		o(&d.opts)
	}

	return d
}

// Heading adds a section heading at the current indent level. Descriptions added
// after it are grouped under the heading by the structured renderers. Text output
// places a blank line between the heading and its body. The text accepts the same
// markup vocabulary as Style, such as {bold}Node{/bold}, rendered as ANSI in text
// output and as the plain visible text everywhere else.
func (d *Document) Heading(text string) *Document {
	d.rows = append(d.rows, row{kind: kindHeading, indent: d.indent, desc: text})

	return d
}

// Headingf is Heading with the text formatted by fmt.Sprintf, saving a separate
// Sprintf call at the call site.
func (d *Document) Headingf(format string, args ...any) *Document {
	return d.Heading(fmt.Sprintf(format, args...))
}

// Item adds a single description and value. The value is formatted per the rules
// described on the package; wrap it with a helper such as IBytes or Duration to
// control the formatting.
//
// A value that satisfies Embeddable, such as a *Table, is embedded under the
// description and rendered in whichever format the Document is rendered to. A nil
// Embeddable renders as an empty value rather than panicking.
func (d *Document) Item(desc string, value any) *Document {
	if e, ok := value.(Embeddable); ok {
		if isNil(e) {
			d.rows = append(d.rows, row{kind: kindRow, indent: d.indent, desc: desc, vals: []Value{emptyValue{}}})
			return d
		}
		d.rows = append(d.rows, row{kind: kindEmbed, indent: d.indent, desc: desc, embed: e})
		return d
	}

	d.rows = append(d.rows, row{kind: kindRow, indent: d.indent, desc: desc, vals: []Value{toValue(value)}})

	return d
}

// isNil reports whether v is nil or holds a nil pointer, map, slice, channel,
// function or interface, so an Embeddable backed by a typed nil is detected
// before any of its methods are called.
func isNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// ItemIf adds an Item only when cond is true, replacing the common
// "if cond { d.Item(...) }" guard around optional rows.
func (d *Document) ItemIf(desc string, value any, cond bool) *Document {
	if cond {
		return d.Item(desc, value)
	}

	return d
}

// ItemUnless adds an Item unless cond is true, the inverse of ItemIf.
func (d *Document) ItemUnless(desc string, value any, cond bool) *Document {
	if !cond {
		return d.Item(desc, value)
	}

	return d
}

// ItemUnlessZero adds an Item unless value is the zero value for its type, the
// common guard for optional fields that should be hidden when unset, such as an
// empty string, a zero number or a zero time.Time. A nil is treated as zero, and
// a value produced by a helper such as IBytes is tested by its raw form, so
// IBytes(0) is zero.
func (d *Document) ItemUnlessZero(desc string, value any) *Document {
	if isZeroValue(value) {
		return d
	}

	return d.Item(desc, value)
}

// isZeroValue reports whether value is the zero value for its type. A nil is
// zero; a helper Value is unwrapped to its raw form first; everything else is
// compared with reflect.Value.IsZero.
func isZeroValue(value any) bool {
	if value == nil {
		return true
	}

	if v, ok := value.(Value); ok {
		return isZeroValue(v.raw())
	}

	return reflect.ValueOf(value).IsZero()
}

// Values adds a static description followed by a list of string values. The
// first value shares the description line and the rest are listed below it,
// aligned to the value column. For formatted helpers or a mix of types, use
// ValuesAny.
func (d *Document) Values(desc string, values []string) *Document {
	vals := make([]Value, 0, len(values))
	for _, v := range values {
		vals = append(vals, toValue(v))
	}
	d.rows = append(d.rows, row{kind: kindRow, indent: d.indent, desc: desc, vals: vals})

	return d
}

// ValuesAny is Values for values that are not plain strings, such as formatted
// helpers or a mix of types. Each argument occupies its own line; a slice
// argument, other than a []byte, is expanded element by element so its items
// stack rather than rendering as one joined line.
func (d *Document) ValuesAny(desc string, values ...any) *Document {
	vals := make([]Value, 0, len(values))
	for _, v := range values {
		vals = append(vals, expandValue(v)...)
	}
	d.rows = append(d.rows, row{kind: kindRow, indent: d.indent, desc: desc, vals: vals})

	return d
}

// expandValue turns one ValuesAny argument into one or more Values. An argument
// that is already a Value is kept whole. A slice that is not a string and not a
// []byte is expanded element by element so its items stack on their own lines;
// anything else becomes a single automatically formatted Value.
func expandValue(v any) []Value {
	if _, ok := v.(Value); ok {
		return []Value{toValue(v)}
	}

	rv := reflect.ValueOf(v)
	if rv.IsValid() && rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() != reflect.Uint8 {
		out := make([]Value, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out = append(out, toValue(rv.Index(i).Interface()))
		}

		return out
	}

	return []Value{toValue(v)}
}

// Fields adds one row per entry of a map, using each key as the description and
// each value as the value. Keys are sorted by their string form so the output is
// deterministic; pass an already ordered sequence of Item calls if you need a
// specific order. A non-map argument records ErrInvalidInput, retrievable via
// Err.
func (d *Document) Fields(m any) *Document {
	rv := reflect.ValueOf(m)
	if !rv.IsValid() || rv.Kind() != reflect.Map {
		d.setErr(fmt.Errorf("%w: Fields requires a map, got %T", ErrInvalidInput, m))
		return d
	}

	keys := rv.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
	})

	for _, k := range keys {
		desc := fmt.Sprint(k.Interface())
		val := rv.MapIndex(k).Interface()
		d.rows = append(d.rows, row{kind: kindRow, indent: d.indent, desc: desc, vals: []Value{toValue(val)}})
	}

	return d
}

// Blank adds an empty line. Text output already places a blank line after each
// heading; use this to add further separation, such as before a heading.
func (d *Document) Blank() *Document {
	d.rows = append(d.rows, row{kind: kindBlank, indent: d.indent})

	return d
}

// Printf adds a free-form line with no description column and no trailing colon.
// Inside a Section the line is indented one level below the section heading so it
// reads as part of the section; at the top level it starts at the left margin.
// The text is formatted with fmt.Sprintf and always parsed for Style markup such
// as {bold}, rendered as ANSI in text output and as the plain visible text in
// Markdown; the JSON and YAML renderers ignore it. Use it for notes, separators
// or pre-formatted content that does not fit the description and value columns.
func (d *Document) Printf(format string, args ...any) *Document {
	d.rows = append(d.rows, row{kind: kindLine, indent: d.indent, desc: fmt.Sprintf(format, args...)})

	return d
}

// Print adds a free-form line from its operands, formatted like fmt.Print: each
// operand uses its default format and spaces are inserted between operands that
// are not strings. Unlike Printf no format string is interpreted, so pre-rendered
// content containing a '%' is safe to pass directly. The line is placed and
// parsed for Style markup exactly as Printf describes.
func (d *Document) Print(a ...any) *Document {
	d.rows = append(d.rows, row{kind: kindLine, indent: d.indent, desc: fmt.Sprint(a...)})

	return d
}

// Embed adds an Embeddable, such as a *Table, as a block with no description,
// the counterpart to Item for content that stands on its own. In text and
// Markdown the block is placed at the current indent, like Print; the JSON and
// YAML renderers ignore it, since it has no key to nest under. Use Item when the
// embedded value should appear in the structured output. A value that is not an
// Embeddable is added with Print, and a nil Embeddable is skipped.
func (d *Document) Embed(value any) *Document {
	e, ok := value.(Embeddable)
	if !ok {
		return d.Print(value)
	}
	if isNil(e) {
		return d
	}

	d.rows = append(d.rows, row{kind: kindEmbed, indent: d.indent, embed: e})

	return d
}

// Indent increases the indent level for subsequent rows by one. The indent
// shifts the whole row, including the description column, to the right.
func (d *Document) Indent() *Document {
	d.indent++

	return d
}

// Outdent decreases the indent level by one. It never goes below zero.
func (d *Document) Outdent() *Document {
	if d.indent > 0 {
		d.indent--
	}

	return d
}

// Section adds an indented heading and body. It increases the indent, adds the
// heading, runs fn (which adds the body at the deeper indent) and then restores
// the previous indent. It is the convenient way to produce an indented
// sub-section. As with Heading, text output places a blank line between the
// heading and its body, and the heading text accepts the Style markup vocabulary.
// A blank line is also placed before the section to separate it from preceding
// content, unless the section begins the document or a blank already precedes it.
func (d *Document) Section(heading string, fn func(d *Document)) *Document {
	// Separate the section from preceding content with a blank line, unless it
	// begins the document or a blank already precedes it.
	n := len(d.rows)
	if n > 0 && d.rows[n-1].kind != kindBlank {
		d.Blank()
	}

	d.Indent()
	d.Heading(heading)
	if fn != nil {
		fn(d)
	}
	d.Outdent()

	return d
}

// Err returns the first error recorded while building the Document, or nil. The
// structured render methods also return this error.
func (d *Document) Err() error {
	return d.err
}

func (d *Document) setErr(err error) {
	if d.err == nil {
		d.err = err
	}
}
