// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package util

import "strings"

// Sanitize keeps newlines, turns tabs into a single space, removes ANSI escape
// sequences whole and drops other C0 and C1 control characters, so untrusted
// values cannot break column alignment or inject terminal escape sequences.
func Sanitize(s string) string {
	if s == "" {
		return s
	}

	rs := []rune(s)
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch {
		case r == 0x1b:
			i = SkipEscape(rs, i)
		case r == '\n':
			b.WriteRune(r)
		case r == '\t':
			b.WriteByte(' ')
		case r < 0x20, r >= 0x7f && r < 0xa0:
			continue
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

// SkipEscape consumes an ANSI escape sequence that begins at rs[i] (an ESC) and
// returns the index of its final rune, so the caller's loop advances past it. It
// handles CSI (ESC [ ... final) and OSC (ESC ] ... BEL or ST) sequences and
// falls back to dropping a two-rune escape.
func SkipEscape(rs []rune, i int) int {
	n := len(rs)
	if i+1 >= n {
		return i
	}

	switch rs[i+1] {
	case '[':
		j := i + 2
		for j < n && !(rs[j] >= 0x40 && rs[j] <= 0x7e) {
			j++
		}
		if j >= n {
			return n
		}
		return j

	case ']':
		j := i + 2
		for j < n {
			if rs[j] == 0x07 {
				return j
			}
			if rs[j] == 0x1b && j+1 < n && rs[j+1] == '\\' {
				return j + 1
			}
			j++
		}
		return n

	default:
		return i + 1
	}
}
