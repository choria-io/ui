// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns

// Option configures a Document.
type Option func(*options)

type options struct {
	margin      int
	indentWidth int
	omitEmpty   bool
	emptyText   string

	// plain requests the unstyled form of values from Style; it is set
	// internally by renderers that cannot carry ANSI escapes.
	plain bool
}

func defaultOptions() options {
	return options{
		margin:      4,
		indentWidth: 2,
	}
}

// WithMargin sets the number of spaces to the left of the description column.
// The default is 4. Negative values are ignored.
func WithMargin(n int) Option {
	return func(o *options) {
		if n >= 0 {
			o.margin = n
		}
	}
}

// WithIndentWidth sets the number of spaces added per indent level. The default
// is 2. Negative values are ignored.
func WithIndentWidth(n int) Option {
	return func(o *options) {
		if n >= 0 {
			o.indentWidth = n
		}
	}
}

// WithOmitEmpty skips any row whose value renders as empty, removing the common
// "if value != \"\" { ... }" guard around optional rows.
func WithOmitEmpty() Option {
	return func(o *options) {
		o.omitEmpty = true
	}
}

// WithEmptyText sets the placeholder rendered for an empty or nil value, for
// example "-" or "n/a". It has no effect on rows removed by WithOmitEmpty.
func WithEmptyText(s string) Option {
	return func(o *options) {
		o.emptyText = s
	}
}
