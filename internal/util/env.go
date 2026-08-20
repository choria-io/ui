// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"strconv"
	"strings"
)

// booleans that triggers LLM Format output
var llmFormatBools = []string{"LLMFORMAT", "FISK-AI", "CLAUDECODE"}

// LLMFormatEnabled reports whether output should default to Markdown because the
// process is driving an LLM or agent consumer. LLMFORMAT, when set to a value
// strconv.ParseBool understands, decides outright and overrides everything else.
// Otherwise CLAUDECODE=1, set by the Claude Code harness, enables it.
func LLMFormatEnabled() bool {
	for _, ev := range llmFormatBools {
		v, ok := os.LookupEnv(ev)
		if ok {
			on, err := strconv.ParseBool(strings.TrimSpace(v))
			if err == nil && on {
				return true
			}
		}
	}

	return false
}
