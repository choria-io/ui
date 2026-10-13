// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// TerminalWidth reports the width of the terminal in display cells. COLUMNS
// decides when it holds a positive integer, even when stdout is not a terminal;
// otherwise the size of the terminal attached to stdout is used. It returns 0
// when neither is available, for example when stdout is piped.
func TerminalWidth() int {
	v, ok := os.LookupEnv("COLUMNS")
	if ok {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil && n > 0 {
			return n
		}
	}

	return stdoutWidth()
}

// stdoutWidth is replaced in tests to simulate a terminal on stdout.
var stdoutWidth = func() int {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		return 0
	}

	w, _, err := term.GetSize(fd)
	if err != nil || w < 0 {
		return 0
	}

	return w
}
