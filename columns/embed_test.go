// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/choria-io/ui/columns"
	"github.com/choria-io/ui/table"
)

// fakeEmbed is a minimal Embeddable used to exercise error propagation without
// depending on the table package's behavior.
type fakeEmbed struct {
	err error
}

func (f fakeEmbed) String() string { return "TEXT" }

func (f fakeEmbed) Markdown() ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []byte("MD\n"), nil
}

func (f fakeEmbed) JSON() ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []byte("{}\n"), nil
}

func (f fakeEmbed) YAML() ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []byte("k: v\n"), nil
}

func nodes() *table.Table {
	t := table.NewTableWriter("Nodes")
	t.AddHeaders("Name", "Cores")
	t.AddRow("web1", 8)

	return t
}

const embedText = `     Count: 3
    Detail:
            ╭──────────────╮
            │     Nodes    │
            ├──────┬───────┤
            │ Name │ Cores │
            ├──────┼───────┤
            │ web1 │ 8     │
            ╰──────┴───────╯
`

const embedMarkdown = `- **Count:** 3

- **Detail:**

# Nodes
| Name | Cores |
| --- | --- |
| web1 | 8 |
`

const embedJSON = `{
  "Count": 3,
  "Detail": {
    "title": "Nodes",
    "headers": [
      "Name",
      "Cores"
    ],
    "rows": [
      [
        "web1",
        8
      ]
    ]
  }
}
`

const embedYAML = `Count: 3
Detail:
  title: Nodes
  headers:
  - Name
  - Cores
  rows:
  - - web1
    - 8
`

const bareEmbedText = `╭──────────────╮
│     Nodes    │
├──────┬───────┤
│ Name │ Cores │
├──────┼───────┤
│ web1 │ 8     │
╰──────┴───────╯
`

var _ = Describe("Embedding", func() {
	build := func() *columns.Document {
		d := columns.New()
		d.Item("Count", 3)
		d.Item("Detail", nodes())

		return d
	}

	It("embeds a table in text with the block aligned to the value column", func() {
		Expect(build().String()).To(Equal(embedText))
	})

	It("embeds a table in Markdown as a labeled block", func() {
		md, err := build().Markdown()
		Expect(err).ToNot(HaveOccurred())
		Expect(string(md)).To(Equal(embedMarkdown))
	})

	It("embeds a table in JSON nested and re-indented", func() {
		j, err := build().JSON()
		Expect(err).ToNot(HaveOccurred())
		Expect(string(j)).To(Equal(embedJSON))
	})

	It("embeds a table in YAML spliced at the correct indentation", func() {
		y, err := build().YAML()
		Expect(err).ToNot(HaveOccurred())
		Expect(string(y)).To(Equal(embedYAML))
	})

	It("nests inside a Section", func() {
		d := columns.New()
		d.Section("Cluster", func(d *columns.Document) {
			d.Item("Detail", nodes())
		})

		j, err := d.JSON()
		Expect(err).ToNot(HaveOccurred())
		Expect(string(j)).To(ContainSubstring("\"Cluster\": {"))
		Expect(string(j)).To(ContainSubstring("\"Detail\": {"))
		Expect(string(j)).To(ContainSubstring("\"web1\""))
	})

	It("renders a nil Embeddable as an empty value without panicking", func() {
		d := columns.New()
		d.Item("Detail", (*table.Table)(nil))

		Expect(d.String()).To(ContainSubstring("Detail:"))

		j, err := d.JSON()
		Expect(err).ToNot(HaveOccurred())
		Expect(string(j)).To(ContainSubstring("\"Detail\": null"))
	})

	It("embeds one Document inside another", func() {
		sub := columns.New()
		sub.Item("Inner", "value")

		d := columns.New()
		d.Item("Sub", sub)

		j, err := d.JSON()
		Expect(err).ToNot(HaveOccurred())
		Expect(string(j)).To(ContainSubstring("\"Sub\": {"))
		Expect(string(j)).To(ContainSubstring("\"Inner\": \"value\""))
	})

	It("returns ErrDuplicateDescription for two embeds sharing a key", func() {
		d := columns.New()
		d.Item("Detail", nodes())
		d.Item("Detail", nodes())

		_, err := d.JSON()
		Expect(err).To(MatchError(columns.ErrDuplicateDescription))
	})

	It("propagates an embed render error for each structured format", func() {
		boom := errors.New("boom")

		_, jerr := columns.New().Item("X", fakeEmbed{err: boom}).JSON()
		Expect(jerr).To(MatchError(ContainSubstring("boom")))

		_, yerr := columns.New().Item("X", fakeEmbed{err: boom}).YAML()
		Expect(yerr).To(MatchError(ContainSubstring("boom")))

		_, merr := columns.New().Item("X", fakeEmbed{err: boom}).Markdown()
		Expect(merr).To(MatchError(ContainSubstring("boom")))
	})

	Describe("Embed", func() {
		It("adds a description-less block at the current indent", func() {
			d := columns.New()
			d.Embed(nodes())

			Expect(d.String()).To(Equal(bareEmbedText))
		})

		It("omits the Markdown label", func() {
			md, err := columns.New().Embed(nodes()).Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(md)).To(HavePrefix("# Nodes\n|"))
			Expect(string(md)).ToNot(ContainSubstring("**"))
		})

		It("is ignored by JSON and YAML, like Print", func() {
			d := columns.New()
			d.Item("Count", 3)
			d.Embed(nodes())

			j, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(Equal("{\n  \"Count\": 3\n}\n"))
			Expect(string(j)).ToNot(ContainSubstring("Nodes"))
		})

		It("falls back to Print for a non-Embeddable value", func() {
			d := columns.New()
			d.Item("A", "b")
			d.Embed("just a note")

			Expect(d.String()).To(ContainSubstring("just a note"))

			j, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).ToNot(ContainSubstring("just a note"))
		})

		It("skips a nil Embeddable", func() {
			d := columns.New()
			d.Embed((*table.Table)(nil))

			Expect(d.String()).To(Equal(""))

			j, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(Equal("{}\n"))
		})
	})
})
