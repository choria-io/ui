// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns

import (
	"strconv"
	"strings"
)

// These helpers compose complex values without a call to fmt.Sprintf, keeping
// the raw form intact for the structured renderers where it is meaningful.

// Annotated renders a main value with a note in parentheses, for example
// "running (pid 1234)". A note that renders empty is omitted. Both arguments may
// themselves be values produced by other helpers.
func Annotated(main, note any) Value {
	m := text(main)
	n := text(note)

	s := m
	if n != "" {
		s = m + " (" + n + ")"
	}

	return fmtValue{display: s, rawv: s}
}

// Join renders several values joined by sep on a single line, for example
// columns.Join(", ", a, b, c). The raw form is the list of the values' raw
// forms.
func Join(sep string, items ...any) Value {
	parts := make([]string, 0, len(items))
	raws := make([]any, 0, len(items))
	for _, it := range items {
		v := toValue(it)
		parts = append(parts, strings.Join(v.lines(&options{}), " "))
		raws = append(raws, v.raw())
	}

	return fmtValue{display: strings.Join(parts, sep), rawv: raws}
}

// Bool renders one of two strings depending on b, for example
// columns.Bool(ok, "up", "down"). The raw form is the boolean.
func Bool(b bool, t, f string) Value {
	if b {
		return fmtValue{display: t, rawv: b}
	}

	return fmtValue{display: f, rawv: b}
}

// YesNo renders "yes" or "no". The raw form is the boolean.
func YesNo(b bool) Value {
	return Bool(b, "yes", "no")
}

// Count renders a number with a singular or plural noun, for example "1 file" or
// "3 files". The raw form is the count.
func Count(n int, singular, plural string) Value {
	word := plural
	if n == 1 {
		word = singular
	}

	return fmtValue{display: strconv.Itoa(n) + " " + word, rawv: n}
}
