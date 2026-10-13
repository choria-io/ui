// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package table

import (
	"slices"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/choria-io/ui/internal/util"
)

const (
	// minWrapWidth is the narrowest a column is made by FitWidth, unless its
	// content is narrower still or a header word is wider.
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
// go-pretty, at width display cells in the text output. Calling it again for a
// column replaces its width, and a width of 0 or less removes the limit. Column
// numbers below 1 or beyond the last column are ignored. Markdown, JSON and YAML
// are unaffected.
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
// display cells wide, borders included, when that is possible. Columns whose
// cells contain spaces are narrowed first, widest first, so prose wraps while
// paths, numbers and identifiers stay whole. No column is narrowed below the
// smaller of its content and 10 cells, or its longest header word, so a very
// small width can still produce a wider table; nothing is ever truncated. A
// width of 0 or less disables fitting. The width is applied when the table is
// rendered, and Markdown, JSON and YAML are unaffected. FitWidth replaces an
// earlier FitWidth or FitTerminal.
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

	cells := t.textCells()
	ncols := 0
	for _, row := range cells {
		ncols = max(ncols, len(row))
	}

	if ncols == 0 {
		return nil, 0
	}

	natural := make([]int, ncols)
	prose := make([]bool, ncols)
	for _, row := range cells {
		for i, c := range row {
			natural[i] = max(natural[i], text.LongestLineLen(c))
			if strings.Contains(c, " ") {
				prose[i] = true
			}
		}
	}

	limits := make([]int, ncols)
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
		floors := t.columnFloors(widths)
		available := width - rowOverhead - columnOverhead*ncols

		var proseCols, solidCols []int
		solidWidth := 0
		for i := range widths {
			if prose[i] {
				proseCols = append(proseCols, i)
				continue
			}
			solidCols = append(solidCols, i)
			solidWidth += widths[i]
		}

		fitColumns(widths, floors, proseCols, available-solidWidth)

		proseWidth := 0
		for _, i := range proseCols {
			proseWidth += widths[i]
		}

		fitColumns(widths, floors, solidCols, available-proseWidth)

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

// columnFloors is the narrowest each column may be made by FitWidth: its
// current width when that is below the minimum, else the larger of the minimum
// and its longest header word.
func (t *Table) columnFloors(widths []int) []int {
	floors := make([]int, len(widths))
	for i, w := range widths {
		floor := minWrapWidth
		if i < len(t.headers) {
			for _, word := range strings.Fields(cellText(t.headers[i])) {
				floor = max(floor, text.StringWidthWithoutEscSequences(word))
			}
		}
		floors[i] = min(w, floor)
	}

	return floors
}

// fitColumns narrows the columns listed in cols so their widths sum to at most
// budget, never below their floors. It finds the largest cap that fits, limits
// every wider column to it, then hands any cells left over to the capped
// columns from left to right.
func fitColumns(widths []int, floors []int, cols []int, budget int) {
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

// textCells returns the display text of the header, every row and the footer.
func (t *Table) textCells() [][]string {
	var cells [][]string

	add := func(row []any) {
		if len(row) == 0 {
			return
		}
		s := make([]string, len(row))
		for i, c := range row {
			s[i] = cellText(c)
		}
		cells = append(cells, s)
	}

	add(t.headers)
	for _, r := range t.rows {
		if r.sep {
			continue
		}
		add(r.cells)
	}
	add(t.footer)

	return cells
}

// wrapCell soft-wraps each line of s to width display cells, keeping the line
// breaks already in s. A word longer than width is broken across lines, and
// wide runes never push a line past width.
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

		for _, l := range strings.Split(text.WrapSoft(line, width), "\n") {
			out = append(out, breakLine(l, width)...)
		}
	}

	return strings.Join(out, "\n")
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
