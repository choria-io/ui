// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package table_test

import (
	"math/rand/v2"
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

// expectWordsWhole checks every word of the given cells appears unbroken in a
// cell of the rendered table s.
func expectWordsWhole(s string, cells ...string) {
	GinkgoHelper()

	rendered := make(map[string]bool)
	for _, l := range strings.Split(s, "\n") {
		for _, cell := range strings.Split(l, "│") {
			for _, w := range strings.Fields(cell) {
				rendered[w] = true
			}
		}
	}

	for _, c := range cells {
		for _, w := range strings.Fields(c) {
			Expect(rendered).To(HaveKey(w), "word %q is broken", w)
		}
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

		It("does not narrow columns whose cells cannot wrap", func() {
			Expect(servers().FitWidth(30).Render()).To(Equal(serversText))
		})

		It("disables fitting with a width of 0 or less", func() {
			Expect(servers().FitWidth(10).FitWidth(0).Render()).To(Equal(serversText))
			Expect(servers().FitWidth(-1).Render()).To(Equal(serversText))
		})

		It("renders wider than asked rather than truncating", func() {
			out := refusals().FitWidth(20).Render()

			expectBoxWidth(out, 50)
			expectWordsWhole(out, "items/auth-rework.md", "run planboard item reopen auth-rework then retry the change", "sequence would reference a retired item")
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

	Describe("keeping words whole", func() {
		It("keeps a command in a remedy whole", func() {
			remedy := "planboard api close Child_Project/build_step"
			t := table.NewTableWriter("Refused")
			t.AddHeaders("Target", "Reason", "To get past it")
			t.AddRow("Child_Project/items/build_step.md", "the item is blocked by an open dependency that has not landed yet", remedy)

			out := t.FitWidth(80).Render()
			expectBoxWidth(out, 80)
			expectWordsWhole(out, remedy)
		})

		It("keeps slugs whole and renders wider than asked when the words do not fit", func() {
			t := table.NewTableWriter("Items")
			t.AddHeaders("Item", "Position", "Title", "Status", "Commitment", "Milestone", "Next")
			t.AddRow("Child_Project/base", "1", "Base layer for the build pipeline and its shared tooling", "proposed", "committed", "no", "yes")

			out := t.FitWidth(80).Render()
			expectBoxWidth(out, 87)
			expectWordsWhole(out, "Child_Project/base", "Base layer for the build pipeline and its shared tooling", "Position Commitment Milestone")
		})

		It("does not treat a column as prose because its header has spaces", func() {
			t := table.NewTableWriter("Projects")
			t.AddHeaders("Project", "Title", "Status", "Program or parent", "Claimed by")
			t.AddRow("Child_Project", "A child project inside the program", "active", "Program/one", "rip")

			Expect(t.FitWidth(60).Render()).To(Equal(`╭──────────────────────────────────────────────────────────╮
│                         Projects                         │
├───────────────┬─────────┬────────┬─────────────┬─────────┤
│ Project       │ Title   │ Status │ Program or  │ Claimed │
│               │         │        │ parent      │ by      │
├───────────────┼─────────┼────────┼─────────────┼─────────┤
│ Child_Project │ A child │ active │ Program/one │ rip     │
│               │ project │        │             │         │
│               │ inside  │        │             │         │
│               │ the     │        │             │         │
│               │ program │        │             │         │
╰───────────────┴─────────┴────────┴─────────────┴─────────╯
`))
		})

		It("wraps a header over identifiers before wrapping prose", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("Program or parent", "Title")
			t.AddRow("Program/one", "A child project inside the program")

			out := t.FitWidth(52).Render()
			expectBoxWidth(out, 52)
			Expect(out).To(ContainSubstring("│ Program or  │"))
			Expect(out).To(ContainSubstring("│ Program/one │ A child project inside the program │"))
		})

		It("keeps long slugs whole", func() {
			slug := "Child_Project/a_rather_long_slug"
			t := table.NewTableWriter("")
			t.AddHeaders("Item", "Status", "Title")
			t.AddRow(slug, "in progress", "Rework the build pipeline so that shared tooling is fetched once per run and cached between steps")

			out := t.FitWidth(60).Render()
			expectBoxWidth(out, 60)
			expectWordsWhole(out, slug)
		})

		It("keeps short prose on one line beside long prose", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("Item", "Status", "Title")
			t.AddRow("Child_Project/base", "in progress", "Rework the build pipeline so that shared tooling is fetched once per run and cached between steps")

			out := t.FitWidth(60).Render()
			expectBoxWidth(out, 60)
			Expect(out).To(ContainSubstring("│ in progress │"))
		})

		It("renders wider than asked rather than breaking a long token", func() {
			digest := "sha256:" + strings.Repeat("0123456789abcdef", 4)
			t := table.NewTableWriter("")
			t.AddHeaders("Artifact", "Digest")
			t.AddRow("planboard-linux-amd64", digest)

			out := t.FitWidth(80).Render()
			expectBoxWidth(out, 99)
			expectWordsWhole(out, digest)
		})

		It("wraps text without spaces between wide runes", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("Name", "説明")
			t.AddRow("web1", "これは公開ウェブサイトとAPIを提供するサーバーです")

			Expect(t.FitWidth(30).Render()).To(Equal(`╭──────┬─────────────────────╮
│ Name │ 説明                │
├──────┼─────────────────────┤
│ web1 │ これは公開ウェブサ  │
│      │ イトとAPIを提供する │
│      │ サーバーです        │
╰──────┴─────────────────────╯
`))
		})

		It("keeps a word whole next to wide runes", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("Name", "Note")
			t.AddRow("web1", "ビルドChild_Project/build_stepを実行する")

			out := t.FitWidth(40).Render()
			expectBoxWidth(out, 40)
			Expect(out).To(ContainSubstring("Child_Project/build_step"))
		})

		It("keeps every word whole and fits when the longest words allow it", func() {
			pool := strings.Fields("a to the item is landed and cannot be modified without reopening it first run planboard api close Child_Project/build_step Child_Project/base Program/one 1,610,612,736 items/auth-rework.md Research_Library/attention_is_all")
			headers := []string{"Name", "Program or parent", "To get past it", "Commitment", "Reason", "Id"}
			r := rand.New(rand.NewPCG(1, 2))

			cell := func() string {
				words := make([]string, 1+r.IntN(8))
				for i := range words {
					words[i] = pool[r.IntN(len(pool))]
				}
				return strings.Join(words, " ")
			}

			for range 500 {
				ncols := 1 + r.IntN(6)
				natural := make([]int, ncols)
				longest := make([]int, ncols)
				var cells []string

				measure := func(i int, c string) {
					natural[i] = max(natural[i], len(c))
					for _, w := range strings.Fields(c) {
						longest[i] = max(longest[i], len(w))
					}
				}

				t := table.NewTableWriter("")
				row := make([]any, ncols)
				for i := range row {
					h := headers[r.IntN(len(headers))]
					measure(i, h)
					row[i] = h
				}
				t.AddHeaders(row...)

				for range 1 + r.IntN(5) {
					row := make([]any, ncols)
					for i := range row {
						c := cell()
						measure(i, c)
						cells = append(cells, c)
						row[i] = c
					}
					t.AddRow(row...)
				}

				full, least := 1, 1
				for i := range ncols {
					full += natural[i] + 3
					least += longest[i] + 3
				}

				width := 20 + r.IntN(181)
				want := full
				if full > width {
					want = max(width, least)
				}

				out := t.FitWidth(width).Render()
				expectBoxWidth(out, want)
				expectWordsWhole(out, cells...)
			}
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
