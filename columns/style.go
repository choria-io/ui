// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns

import (
	"strings"

	"github.com/choria-io/ui/internal/util"
)

// styleCodes maps a markup tag name to its SGR (Select Graphic Rendition)
// parameter. Text values are otherwise stripped of escape sequences, so styling
// is only ever produced through Style.
var styleCodes = map[string]string{
	"bold":      "1",
	"dim":       "2",
	"italic":    "3",
	"underline": "4",
	"reverse":   "7",
	"black":     "30",
	"red":       "31",
	"green":     "32",
	"yellow":    "33",
	"blue":      "34",
	"magenta":   "35",
	"cyan":      "36",
	"white":     "37",
	"gray":      "90",
	"grey":      "90",
}

// styledValue holds a value rendered with ANSI escapes and its plain visible
// form. The escapes survive to the text renderer, but the plain form is used for
// Markdown, JSON and YAML and for width sensitive contexts.
type styledValue struct {
	ansi  string
	plain string
}

// Style renders simple markup into terminal styling, for example
// "{green}{bold}OK{/bold}{/green}" or "load {yellow}high{/yellow}". Supported
// tags are bold, dim, italic, underline, reverse and the colors black, red,
// green, yellow, blue, magenta, cyan, white and gray. A tag is closed with
// {/name}, and {reset} or {/} clears all styling. Unknown tags and stray braces
// are passed through as literal text.
//
// Unlike ordinary values, the escape sequences Style produces are preserved in
// text output; every other value has control characters and escapes stripped.
// The plain, unstyled text is used for Markdown and the structured renderers.
func Style(markup string) Value {
	ansi, plain := renderMarkup(markup)

	return styledValue{ansi: ansi, plain: plain}
}

// styleHeading renders a heading's text through the same markup vocabulary as
// Style: the ansi form carries the terminal escapes for text output and the
// plain form is the visible text used by Markdown and the structured renderers.
// Headings therefore accept markup such as {bold}Node{/bold} without a separate
// styled-heading API. Plain headings pass through unchanged.
func styleHeading(text string) (ansi, plain string) {
	return renderMarkup(text)
}

func (s styledValue) String() string {
	return stringOf(s)
}

func (s styledValue) lines(o *options) []string {
	body := s.ansi
	if o != nil && o.plain {
		body = s.plain
	}
	if body == "" {
		return []string{""}
	}

	return strings.Split(body, "\n")
}

func (s styledValue) raw() any {
	return s.plain
}

// renderMarkup parses the markup vocabulary into an ANSI string and the plain
// visible text. Literal runs are stripped of control characters, including raw
// escapes, so styling can only come from recognized tags.
func renderMarkup(s string) (ansiText, plainText string) {
	rs := []rune(s)
	var ansi, plain strings.Builder
	var stack []string

	reapply := func() {
		ansi.WriteString("\x1b[0m")
		for _, c := range stack {
			ansi.WriteString("\x1b[")
			ansi.WriteString(c)
			ansi.WriteByte('m')
		}
	}

	writeLiteral := func(r rune) {
		switch {
		case r == '\n':
			ansi.WriteByte('\n')
			plain.WriteByte('\n')
		case r == '\t':
			ansi.WriteByte(' ')
			plain.WriteByte(' ')
		case r < 0x20, r >= 0x7f && r < 0xa0:
			// drop control characters, including raw escapes
		default:
			ansi.WriteRune(r)
			plain.WriteRune(r)
		}
	}

	for i := 0; i < len(rs); {
		if rs[i] == 0x1b {
			i = util.SkipEscape(rs, i) + 1
			continue
		}

		if rs[i] != '{' {
			writeLiteral(rs[i])
			i++
			continue
		}

		j := i + 1
		for j < len(rs) && rs[j] != '}' {
			j++
		}
		if j >= len(rs) {
			writeLiteral('{')
			i++
			continue
		}

		name := string(rs[i+1 : j])
		if applyTag(name, &stack, &ansi, reapply) {
			i = j + 1
			continue
		}

		writeLiteral('{')
		for _, r := range name {
			writeLiteral(r)
		}
		writeLiteral('}')
		i = j + 1
	}

	if len(stack) > 0 {
		ansi.WriteString("\x1b[0m")
	}

	return ansi.String(), plain.String()
}

// applyTag handles one {tag} token, updating the active style stack and emitting
// the corresponding escapes. It reports whether name was a recognized tag.
func applyTag(name string, stack *[]string, ansi *strings.Builder, reapply func()) bool {
	switch {
	case name == "reset", name == "/":
		*stack = (*stack)[:0]
		ansi.WriteString("\x1b[0m")
		return true

	case strings.HasPrefix(name, "/"):
		code, ok := styleCodes[name[1:]]
		if !ok {
			return false
		}
		for k := len(*stack) - 1; k >= 0; k-- {
			if (*stack)[k] == code {
				*stack = append((*stack)[:k], (*stack)[k+1:]...)
				break
			}
		}
		reapply()
		return true

	default:
		code, ok := styleCodes[name]
		if !ok {
			return false
		}
		*stack = append(*stack, code)
		ansi.WriteString("\x1b[")
		ansi.WriteString(code)
		ansi.WriteByte('m')
		return true
	}
}
