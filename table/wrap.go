// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package table

import (
	"slices"
	"strings"
	"unicode"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/choria-io/ui/internal/util"
)

const (
	// minWrapWidth is the narrowest FitWidth makes a column whose cells wrap
	// before it narrows it further, down to its longest word.
	minWrapWidth = 10

	// rowOverhead and columnOverhead are the cells StyleRounded adds around the
	// content: the left border, then a padding space either side of each column
	// and the separator or right border after it.
	rowOverhead    = 1
	columnOverhead = 3

	// titleOverhead is the border and padding either side of the title text.
	titleOverhead = 4
)

// terminalWidth is replaced in tests to simulate a terminal.
var terminalWidth = util.TerminalWidth

// TerminalWidth reports the width of the terminal in display cells: COLUMNS when
// it holds a positive integer, else the width of the terminal attached to stdout,
// else 0. Its result can be passed straight to FitWidth, where 0 disables
// fitting, and adjusted first when the table is not written to stdout or is
// indented inside a columns.Document.
func TerminalWidth() int {
	return terminalWidth()
}

// WrapColumns soft-wraps the cells of the given columns, numbered from 1 like
// go-pretty, at width display cells in the text output. A word wider than width
// is broken across lines. Calling it again for a column replaces its width, and
// a width of 0 or less removes the limit. Column numbers below 1 or beyond the
// last column are ignored. Markdown, JSON and YAML are unaffected.
func (t *Table) WrapColumns(width int, columns ...int) *Table {
	for _, c := range columns {
		if c < 1 {
			continue
		}

		if width <= 0 {
			delete(t.wrapWidths, c)
			continue
		}

		if t.wrapWidths == nil {
			t.wrapWidths = make(map[int]int)
		}
		t.wrapWidths[c] = width
	}

	return t
}

// FitWidth soft-wraps cells so each line of the text output is at most width
// display cells wide, borders included, when that is possible. No column is
// narrowed below its longest word, header included, so names, paths, numbers
// and commands stay whole and only text wraps, at spaces. Headers wider than
// their column's cells wrap first, then columns whose cells can wrap are
// narrowed, widest first, to 10 cells, and then to their longest word. When
// that is still too wide the table renders wider than asked; words are never
// broken. A width of 0 or less disables fitting. The width is applied when the
// table is rendered, and Markdown, JSON and YAML are unaffected. FitWidth
// replaces an earlier FitWidth or FitTerminal.
func (t *Table) FitWidth(width int) *Table {
	t.fitWidth = width
	t.fitTerminal = false

	return t
}

// FitTerminal is FitWidth with the width from TerminalWidth, read when the table
// is rendered. It measures stdout, so when the table is written elsewhere or
// embedded in an indented columns.Document, use FitWidth with an adjusted
// TerminalWidth instead. FitTerminal replaces an earlier FitWidth or
// FitTerminal.
func (t *Table) FitTerminal() *Table {
	t.fitWidth = 0
	t.fitTerminal = true

	return t
}

// applyWrapping configures w to wrap columns according to WrapColumns and
// FitWidth or FitTerminal, and soft-wraps the title to the resulting table
// width. It is only used for the text output.
func (t *Table) applyWrapping(w table.Writer) {
	limits, rowWidth := t.columnLimits()
	if len(limits) == 0 {
		return
	}

	configs := make([]table.ColumnConfig, 0, len(limits))
	for i, l := range limits {
		if l <= 0 {
			continue
		}
		configs = append(configs, table.ColumnConfig{Number: i + 1, WidthMax: l, WidthMaxEnforcer: wrapCell})
	}

	if len(configs) == 0 {
		return
	}

	w.SetColumnConfigs(configs)

	if t.title != "" {
		w.SetTitle("%s", wrapCell(util.Sanitize(t.title), rowWidth-titleOverhead))
	}
}

// columnLimits computes the width limit of each column, 0 meaning unlimited,
// and the width of the resulting table. It returns no limits when no wrapping is
// configured or the table has no columns.
func (t *Table) columnLimits() ([]int, int) {
	width := t.fitWidth
	if t.fitTerminal {
		width = terminalWidth()
	}

	if width <= 0 && len(t.wrapWidths) == 0 {
		return nil, 0
	}

	header, body := t.textCells()
	ncols := len(header)
	for _, row := range body {
		ncols = max(ncols, len(row))
	}

	if ncols == 0 {
		return nil, 0
	}

	natural := make([]int, ncols)
	headerWord := make([]int, ncols)
	for i, c := range header {
		natural[i] = text.LongestLineLen(c)
		headerWord[i] = longestWord(c)
	}

	bodyWidth := make([]int, ncols)
	bodyWord := make([]int, ncols)
	for _, row := range body {
		for i, c := range row {
			bodyWidth[i] = max(bodyWidth[i], text.LongestLineLen(c))
			bodyWord[i] = max(bodyWord[i], longestWord(c))
		}
	}

	limits := make([]int, ncols)
	for i := range natural {
		natural[i] = max(natural[i], bodyWidth[i])
	}

	widths := slices.Clone(natural)
	for c, l := range t.wrapWidths {
		i := c - 1
		if i >= ncols {
			continue
		}
		if l < widths[i] {
			widths[i] = l
			limits[i] = l
		}
	}

	if width > 0 {
		available := width - rowOverhead - columnOverhead*ncols

		var wrapping, whole []int
		words := make([]int, ncols)
		readable := make([]int, ncols)
		for i, w := range widths {
			words[i] = min(w, max(bodyWord[i], headerWord[i]))
			readable[i] = max(words[i], min(w, minWrapWidth))

			if bodyWord[i] < bodyWidth[i] {
				wrapping = append(wrapping, i)
			} else {
				whole = append(whole, i)
			}
		}

		fitColumns(widths, words, whole, available)
		fitColumns(widths, readable, wrapping, available)
		fitColumns(widths, words, wrapping, available)

		for i, w := range widths {
			if w < natural[i] {
				limits[i] = w
			}
		}
	}

	rowWidth := rowOverhead
	for _, w := range widths {
		rowWidth += w + columnOverhead
	}

	return limits, rowWidth
}

// fitColumns narrows the columns listed in cols, never below their floors, so
// that all the columns together fit available. It finds the largest cap that
// fits, limits every wider column to it, then hands any cells left over to the
// capped columns from left to right.
func fitColumns(widths []int, floors []int, cols []int, available int) {
	budget := available
	for i, w := range widths {
		if !slices.Contains(cols, i) {
			budget -= w
		}
	}

	total := 0
	floorTotal := 0
	widest := 0
	for _, i := range cols {
		total += widths[i]
		floorTotal += floors[i]
		widest = max(widest, widths[i])
	}

	if total <= budget {
		return
	}

	if floorTotal >= budget {
		for _, i := range cols {
			widths[i] = floors[i]
		}
		return
	}

	capped := func(i int, c int) int {
		return max(floors[i], min(widths[i], c))
	}

	sumAt := func(c int) int {
		s := 0
		for _, i := range cols {
			s += capped(i, c)
		}
		return s
	}

	lo, hi := 0, widest
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if sumAt(mid) <= budget {
			lo = mid
		} else {
			hi = mid - 1
		}
	}

	spare := budget - sumAt(lo)
	for _, i := range cols {
		w := capped(i, lo)
		if spare > 0 && w < widths[i] {
			w++
			spare--
		}
		widths[i] = w
	}
}

// textCells returns the display text of the header, and of every row and the
// footer.
func (t *Table) textCells() ([]string, [][]string) {
	var body [][]string

	for _, r := range t.rows {
		if r.sep || len(r.cells) == 0 {
			continue
		}
		body = append(body, textRow(r.cells))
	}

	if len(t.footer) > 0 {
		body = append(body, textRow(t.footer))
	}

	return textRow(t.headers), body
}

// textRow returns the display text of each cell in row.
func textRow(row []any) []string {
	s := make([]string, len(row))
	for i, c := range row {
		s[i] = cellText(c)
	}

	return s
}

// longestWord is the width of the widest piece of s that wrapLine keeps on one
// line: a run of runes between whitespace, or a single wide rune.
func longestWord(s string) int {
	longest := 0
	run := 0

	for _, r := range s {
		w := text.RuneWidth(r)

		switch {
		case unicode.IsSpace(r):
			run = 0
		case w > 1:
			run = 0
			longest = max(longest, w)
		default:
			run += w
			longest = max(longest, run)
		}
	}

	return longest
}

// wrapCell soft-wraps each line of s to width display cells, keeping the line
// breaks already in s. Lines that fit are kept as they are, others are wrapped
// by wrapLine.
func wrapCell(s string, width int) string {
	if width <= 0 {
		return s
	}

	var out []string
	for _, line := range strings.Split(s, "\n") {
		if text.StringWidthWithoutEscSequences(line) <= width {
			out = append(out, line)
			continue
		}

		out = append(out, wrapLine(line, width)...)
	}

	return strings.Join(out, "\n")
}

// wrapLine soft-wraps line to width display cells. It breaks at whitespace,
// which becomes a single space within a line, and on either side of a wide
// rune, so text without spaces in scripts such as Chinese and Japanese wraps
// too. A word wider than width is broken by breakLine, and wide runes never
// push a line past width.
func wrapLine(line string, width int) []string {
	var (
		out   []string
		cur   strings.Builder
		curW  int
		word  strings.Builder
		wordW int
		gap   bool
	)

	place := func(s string, w int, spaced bool) {
		sep := 0
		if spaced && cur.Len() > 0 {
			sep = 1
		}

		if curW+sep+w <= width {
			if sep > 0 {
				cur.WriteByte(' ')
			}
			cur.WriteString(s)
			curW += sep + w
			return
		}

		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}

		pieces := breakLine(s, width)
		last := pieces[len(pieces)-1]
		out = append(out, pieces[:len(pieces)-1]...)
		cur.WriteString(last)
		curW = text.StringWidthWithoutEscSequences(last)
	}

	flush := func() {
		if word.Len() == 0 {
			return
		}
		place(word.String(), wordW, gap)
		word.Reset()
		wordW = 0
		gap = false
	}

	for _, r := range line {
		w := text.RuneWidth(r)

		switch {
		case unicode.IsSpace(r):
			flush()
			gap = true
		case w > 1:
			flush()
			place(string(r), w, gap)
			gap = false
		default:
			word.WriteRune(r)
			wordW += w
		}
	}

	flush()

	if cur.Len() > 0 || len(out) == 0 {
		out = append(out, cur.String())
	}

	return out
}

// breakLine splits line into pieces of at most width display cells, by rune.
func breakLine(line string, width int) []string {
	if text.StringWidthWithoutEscSequences(line) <= width {
		return []string{line}
	}

	var (
		out []string
		b   strings.Builder
		n   int
	)

	for _, r := range line {
		rw := text.RuneWidth(r)
		if n > 0 && n+rw > width {
			out = append(out, b.String())
			b.Reset()
			n = 0
		}
		b.WriteRune(r)
		n += rw
	}

	if b.Len() > 0 {
		out = append(out, b.String())
	}

	return out
}
