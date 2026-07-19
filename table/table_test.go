// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package table_test

import (
	"bytes"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/choria-io/ui/table"
)

// TestMain clears the environment variables that switch the renderer to Markdown
// so the golden text assertions are stable even under a harness that sets
// CLAUDECODE=1. Tests that exercise the env behavior set the variables per spec.
func TestMain(m *testing.M) {
	os.Unsetenv("CLAUDECODE")
	os.Unsetenv("LLMFORMAT")
	os.Exit(m.Run())
}

func TestTable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Table")
}

func servers() *table.Table {
	t := table.NewTableWriter("Servers")
	t.AddHeaders("Name", "Cores", "Memory")
	t.AddRow("web1", 8, 1610612736)
	t.AddRow("web2", 4, 1073741824)
	t.AddSeparator()
	t.AddFooter("Total", 12, "")

	return t
}

const serversText = `╭───────────────────────────────╮
│            Servers            │
├───────┬───────┬───────────────┤
│ Name  │ Cores │ Memory        │
├───────┼───────┼───────────────┤
│ web1  │ 8     │ 1,610,612,736 │
│ web2  │ 4     │ 1,073,741,824 │
├───────┼───────┼───────────────┤
│ Total │ 12    │               │
╰───────┴───────┴───────────────╯
`

const serversJSON = `{
  "title": "Servers",
  "headers": [
    "Name",
    "Cores",
    "Memory"
  ],
  "rows": [
    [
      "web1",
      8,
      1610612736
    ],
    [
      "web2",
      4,
      1073741824
    ]
  ],
  "footer": [
    "Total",
    12,
    ""
  ]
}
`

const serversYAML = `title: Servers
headers:
- Name
- Cores
- Memory
rows:
- - web1
  - 8
  - 1610612736
- - web2
  - 4
  - 1073741824
footer:
- Total
- 12
- ""
`

var _ = Describe("Table", func() {
	Describe("text", func() {
		It("renders an aligned box table ending in a single newline", func() {
			Expect(servers().Render()).To(Equal(serversText))
			Expect(servers().String()).To(Equal(serversText))
		})

		It("formats numbers with thousands separators like columns", func() {
			Expect(servers().Render()).To(ContainSubstring("1,610,612,736"))
		})

		It("renders an empty table as an empty string", func() {
			Expect(table.NewTableWriter("").Render()).To(Equal(""))
			Expect(table.NewTableWriter("Title only").Render()).To(Equal(""))
		})
	})

	Describe("markdown", func() {
		It("renders a Markdown table with an H1 title", func() {
			md, err := servers().Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(md)).To(Equal("# Servers\n| Name | Cores | Memory |\n| --- | --- | --- |\n| web1 | 8 | 1,610,612,736 |\n| web2 | 4 | 1,073,741,824 |\n| Total | 12 |  |\n"))
		})

		It("returns Markdown from String when LLMFORMAT is set", func() {
			GinkgoT().Setenv("LLMFORMAT", "1")
			Expect(servers().String()).To(HavePrefix("# Servers\n|"))
		})
	})

	Describe("json", func() {
		It("renders title, headers, rows and footer with native scalars", func() {
			j, err := servers().JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(Equal(serversJSON))
		})

		It("omits headers and title when unset and keeps rows as arrays", func() {
			t := table.NewTableWriter("")
			t.AddRow("a", 1)
			j, err := t.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(Equal("{\n  \"rows\": [\n    [\n      \"a\",\n      1\n    ]\n  ]\n}\n"))
		})

		It("renders an empty rows array for a table with no rows", func() {
			j, err := table.NewTableWriter("").JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(Equal("{\n  \"rows\": []\n}\n"))
		})

		It("drops separators", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("A")
			t.AddRow("x")
			t.AddSeparator()
			t.AddRow("y")
			j, err := t.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(ContainSubstring("\"x\""))
			Expect(string(j)).To(ContainSubstring("\"y\""))
		})
	})

	Describe("yaml", func() {
		It("renders the same shape as JSON", func() {
			y, err := servers().YAML()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(y)).To(Equal(serversYAML))
		})
	})

	Describe("safety", func() {
		It("strips ANSI escapes and control characters from cells", func() {
			t := table.NewTableWriter("")
			t.AddHeaders("Val")
			t.AddRow("\x1b[31mred\x1b[0m\x07")

			Expect(t.Render()).ToNot(ContainSubstring("\x1b"))
			Expect(t.Render()).To(ContainSubstring("red"))

			j, err := t.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).ToNot(ContainSubstring("\\u001b"))
			Expect(string(j)).To(ContainSubstring("\"red\""))
		})

		It("treats a title containing a percent verb literally", func() {
			t := table.NewTableWriterf("%d%% done", 50)
			t.AddRow("a column wide enough for the title")
			Expect(t.Render()).To(ContainSubstring("50% done"))
			Expect(t.Render()).ToNot(ContainSubstring("NOVERB"))
		})
	})

	Describe("output helpers", func() {
		It("WriteTo writes the same bytes as String", func() {
			var buf bytes.Buffer
			n, err := servers().WriteTo(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.String()).To(Equal(serversText))
			Expect(n).To(Equal(int64(len(serversText))))
		})

		It("Bytes matches String", func() {
			Expect(servers().Bytes()).To(Equal([]byte(serversText)))
		})

		It("builders return the table for chaining", func() {
			t := table.NewTableWriter("t").AddHeaders("A").AddRow("x").AddSeparator().AddFooter("f")
			Expect(t).ToNot(BeNil())
			Expect(t.Render()).To(ContainSubstring("x"))
		})
	})
})
