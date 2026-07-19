// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns

// Embeddable is a value that can render itself in each of the output formats a
// Document supports. When a value satisfying Embeddable is passed to Item, the
// Document nests its rendering under the description, deferring to the value's own
// method for whichever format the Document is being rendered to: String for text,
// Markdown, JSON and YAML for the others. The result stays correctly indented
// within the surrounding output.
//
// A *Table from the table package satisfies Embeddable, as does a *Document, so a
// document can embed a sub-document. The interface is checked structurally, so an
// implementer needs no dependency on this package.
type Embeddable interface {
	// String returns the text form, matching the Document's own String.
	String() string
	// Markdown returns the Markdown form and any render error.
	Markdown() ([]byte, error)
	// JSON returns the indented JSON form and any render error.
	JSON() ([]byte, error)
	// YAML returns the YAML form and any render error.
	YAML() ([]byte, error)
}
