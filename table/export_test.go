// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package table

// SetTerminalWidth replaces the terminal width probe until the returned function
// restores it.
func SetTerminalWidth(f func() int) func() {
	orig := terminalWidth
	terminalWidth = f

	return func() { terminalWidth = orig }
}
