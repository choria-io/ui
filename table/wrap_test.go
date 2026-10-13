// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package table_test

import (
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/choria-io/ui/table"
)

func refusals() *table.Table {
	t := table.NewTableWriter("Refused: the change could not be applied to the plan")
	t.AddHeaders("Target", "Reason", "To get past it")
	t.AddRow("items/auth-rework.md", "the item is already landed and cannot be modified without reopening it first", "run planboard item reopen auth-rework then retry the change")
	t.AddRow("project.md", "sequence would reference a retired item", "remove it from the sequence")

	return t
}

// lineWidths measures each rendered line the way go-pretty does, so box drawing
// runes count as one cell regardless of the locale.
func lineWidths(s string) []int {
	var widths []int
	for _, l := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
		widths = append(widths, text.StringWidthWithoutEscSequences(l))
	}

	return widths
}

func expectBoxWidth(s string, width int) {
	GinkgoHelper()

	Expect(s).ToNot(BeEmpty())
	Expect(s).ToNot(ContainSubstring("≈"))
	for _, w := range lineWidths(s) {
		Expect(w).To(Equal(width))
	}
}

var _ = Describe("Wrapping", func() {
	Describe("WrapColumns", func() {
		It("soft-wraps the given columns, header included", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("Name", "Long description")
			t.AddRow("web1", "serves the public website and the api")
			t.WrapColumns(12, 2)

			Expect(t.Render()).To(Equal(`╭──────┬──────────────╮
│ Name │ Long         │
│      │ description  │
├──────┼──────────────┤
│ web1 │ serves the   │
│      │ public       │
│      │ website and  │
│      │ the api      │
╰──────┴──────────────╯
`))
		})

		It("keeps line breaks already in a cell", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("Values")
			t.AddRow("- active\n- landed after review by the team")
			t.WrapColumns(14, 1)

			Expect(t.Render()).To(Equal(`╭────────────────╮
│ Values         │
├────────────────┤
│ - active       │
│ - landed after │
│ review by the  │
│ team           │
╰────────────────╯
`))
		})

		It("keeps wide runes within the column", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("A", "B")
			t.AddRow("x", "a漢字漢字漢字漢字漢字漢字")
			t.WrapColumns(10, 2)

			expectBoxWidth(t.Render(), 18)
		})

		It("removes a limit with a width of 0", func() {
			t := servers().WrapColumns(5, 3).WrapColumns(0, 3)
			Expect(t.Render()).To(Equal(serversText))
		})

		It("ignores columns out of range", func() {
			t := servers().WrapColumns(5, 0, -1, 4)
			Expect(t.Render()).To(Equal(serversText))
		})
	})

	Describe("FitWidth", func() {
		It("narrows prose columns and keeps paths whole", func() {
			out := refusals().FitWidth(70).Render()

			expectBoxWidth(out, 70)
			Expect(out).To(ContainSubstring("│ items/auth-rework.md │"))
			Expect(out).To(ContainSubstring("Refused: the change could not be applied to the plan"))
		})

		It("leaves a table that already fits unchanged", func() {
			Expect(servers().FitWidth(200).Render()).To(Equal(serversText))
			Expect(servers().FitWidth(33).Render()).To(Equal(serversText))
		})

		It("narrows columns without spaces only as a last resort", func() {
			out := servers().FitWidth(30).Render()

			expectBoxWidth(out, 30)
			Expect(out).To(ContainSubstring("│ Memory     │"))
			Expect(out).To(ContainSubstring("│ 1,610,612, │"))
		})

		It("disables fitting with a width of 0 or less", func() {
			Expect(servers().FitWidth(10).FitWidth(0).Render()).To(Equal(serversText))
			Expect(servers().FitWidth(-1).Render()).To(Equal(serversText))
		})

		It("renders wider than asked rather than truncating", func() {
			out := refusals().FitWidth(20).Render()

			expectBoxWidth(out, 40)
			for _, word := range strings.Fields("reopening retired sequence planboard") {
				Expect(out).To(ContainSubstring(word))
			}
		})

		It("does not split header words", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("Name", "Descriptions")
			t.AddRow("web1", "serves the public website and the api")

			out := t.FitWidth(20).Render()
			Expect(out).To(ContainSubstring("│ Descriptions │"))
		})

		It("wraps the title at word boundaries", func() {
			out := refusals().FitWidth(50).Render()

			expectBoxWidth(out, 50)
			Expect(out).To(ContainSubstring("Refused: the change could not be applied to"))
			Expect(out).To(MatchRegexp(`│ +the plan +│`))
		})

		It("considers rows added after it is called", func() {
			t := table.NewTableWriter("")
			t.FitWidth(40)
			t.AddHeaders("Name", "Reason")
			t.AddRow("web1", "the item is already landed and cannot be modified")

			expectBoxWidth(t.Render(), 40)
		})

		It("handles rows with more cells than headers", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("A")
			t.AddRow("x", "the item is already landed and cannot be modified")

			expectBoxWidth(t.FitWidth(30).Render(), 30)
		})

		It("renders an empty or title only table as an empty string", func() {
			Expect(table.NewTableWriter("").FitWidth(40).Render()).To(Equal(""))
			Expect(table.NewTableWriter("Title only").FitWidth(40).WrapColumns(5, 1).Render()).To(Equal(""))
		})
	})

	Describe("FitTerminal", func() {
		It("uses the terminal width when the table renders", func() {
			width := 0
			DeferCleanup(table.SetTerminalWidth(func() int { return width }))

			t := refusals().FitTerminal()
			width = 60
			expectBoxWidth(t.Render(), 60)
			Expect(table.TerminalWidth()).To(Equal(60))
		})

		It("does nothing when the width is unknown", func() {
			DeferCleanup(table.SetTerminalWidth(func() int { return 0 }))

			Expect(servers().FitTerminal().Render()).To(Equal(serversText))
		})

		It("is replaced by a later FitWidth and replaces an earlier one", func() {
			DeferCleanup(table.SetTerminalWidth(func() int { return 60 }))

			expectBoxWidth(refusals().FitTerminal().FitWidth(70).Render(), 70)
			expectBoxWidth(refusals().FitWidth(70).FitTerminal().Render(), 60)
		})
	})

	Describe("other formats", func() {
		wrapped := func() *table.Table {
			return refusals().FitWidth(40).WrapColumns(10, 1, 2, 3)
		}

		It("does not change Markdown, JSON or YAML", func() {
			for _, render := range []func(*table.Table) ([]byte, error){(*table.Table).Markdown, (*table.Table).JSON, (*table.Table).YAML} {
				plain, err := render(refusals())
				Expect(err).ToNot(HaveOccurred())

				got, err := render(wrapped())
				Expect(err).ToNot(HaveOccurred())
				Expect(got).To(Equal(plain))
			}

			md, err := wrapped().Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(md)).ToNot(ContainSubstring("<br/>"))
		})

		It("does not change String in LLM mode", func() {
			GinkgoT().Setenv("LLMFORMAT", "1")
			Expect(wrapped().String()).To(Equal(refusals().String()))
			Expect(wrapped().String()).To(HavePrefix("# Refused"))
		})
	})
})
