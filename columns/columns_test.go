// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns_test

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/choria-io/ui/columns"
)

// TestMain clears the environment variables that switch the text renderers to
// Markdown so the golden text assertions in this package and the example tests
// are stable even when the suite runs under a harness that sets CLAUDECODE=1.
// Tests that exercise the env behavior set the variables explicitly per spec.
func TestMain(m *testing.M) {
	os.Unsetenv("CLAUDECODE")
	os.Unsetenv("LLMFORMAT")
	os.Exit(m.Run())
}

func TestColumns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Columns")
}

// specExample builds a Document that reproduces the target layout.
func specExample() *columns.Document {
	d := columns.New()
	d.Heading("Heading 1")
	d.Item("Short", "Value")
	d.Item("Long Heading", "Long Value")
	d.Item("Items from a Map", columns.Map(map[string]string{"X": "Y", "Z": "Z"}))
	d.Values("Items from an Array", []string{"A", "B", "C"})
	d.Fields(map[string]string{"Map Item 1": "Value 1", "Map Item 2": "Value 2"})
	d.Blank()
	d.Heading("Heading 2")
	d.Item("Item 1", "Value 1")
	d.Blank()
	d.Section("Indented Heading", func(d *columns.Document) {
		d.Item("Item 1", "Value 1")
	})

	return d
}

// specText is the golden text rendering that specExample must reproduce. It
// mirrors the target layout from the package design spec and is kept inline so
// the test does not depend on any external file.
const specText = `Heading 1:

                  Short: Value
           Long Heading: Long Value
       Items from a Map: X: Y
                         Z: Z
    Items from an Array: A
                         B
                         C
             Map Item 1: Value 1
             Map Item 2: Value 2

Heading 2:

                 Item 1: Value 1

  Indented Heading:

                   Item 1: Value 1
`

var _ = Describe("Document", func() {
	Describe("Text rendering", func() {
		It("reproduces the target layout exactly", func() {
			Expect(specExample().String()).To(Equal(specText))
		})

		It("aligns colons across all headings", func() {
			d := columns.New()
			d.Heading("A")
			d.Item("Short", "1")
			d.Heading("B")
			d.Item("A Much Longer Description", "2")
			d.Item("x", "3")

			lines := strings.Split(d.String(), "\n")
			colon := func(s string) int { return strings.Index(s, ": ") }
			var positions []int
			for _, l := range lines {
				if c := colon(l); c >= 0 {
					positions = append(positions, c)
				}
			}
			Expect(positions).To(HaveLen(3))
			Expect(positions[0]).To(Equal(positions[1]))
			Expect(positions[1]).To(Equal(positions[2]))
		})

		It("shifts an indented section including its heading", func() {
			d := columns.New()
			d.Item("Top", "1")
			d.Section("Sub", func(d *columns.Document) {
				d.Item("Inner", "2")
			})

			out := d.String()
			Expect(out).To(ContainSubstring("  Sub:"))
			// margin 4 + indent 2 = value column shifted right by 2
			topCol := strings.Index(strings.Split(out, "\n")[0], ": ")
			var innerCol int
			for l := range strings.SplitSeq(out, "\n") {
				if strings.Contains(l, "Inner") {
					innerCol = strings.Index(l, ": ")
				}
			}
			Expect(innerCol).To(Equal(topCol + 2))
		})

		It("inserts a blank line after a heading", func() {
			d := columns.New()
			d.Heading("A")
			d.Item("x", "1")
			Expect(d.String()).To(Equal("A:\n\n    x: 1\n"))
		})

		It("does not double the blank when one is added after a heading", func() {
			d := columns.New()
			d.Heading("A")
			d.Blank()
			d.Item("x", "1")
			Expect(d.String()).To(Equal("A:\n\n    x: 1\n"))
		})

		It("does not add a trailing blank when a heading ends the document", func() {
			d := columns.New()
			d.Heading("A")
			Expect(d.String()).To(Equal("A:\n"))
		})

		It("aligns wide unicode descriptions by display cells", func() {
			d := columns.New()
			d.Item("Items from an Array", "w")
			d.Item("日本語", "v")

			// "日本語" is 6 display cells; right justified in a 19 cell field
			// after a 4 space margin leaves 17 leading spaces.
			Expect(d.String()).To(ContainSubstring("                 日本語: v"))
		})

		It("splits a value containing newlines into continuation lines", func() {
			d := columns.New()
			d.Item("Config", "line1\nline2")

			lines := strings.Split(d.String(), "\n")
			Expect(lines[0]).To(HaveSuffix("Config: line1"))
			first := strings.Index(lines[0], "line1")
			Expect(lines[1]).To(HavePrefix(strings.Repeat(" ", first)))
			Expect(strings.TrimSpace(lines[1])).To(Equal("line2"))
		})

		It("strips control characters and ANSI escapes", func() {
			d := columns.New()
			d.Item("X", "a\x1b[31mb\x07c")
			Expect(d.String()).To(ContainSubstring("X: abc"))
		})

		It("right-trims every line", func() {
			d := columns.New()
			d.Item("X", "")
			for l := range strings.SplitSeq(d.String(), "\n") {
				Expect(l).To(Equal(strings.TrimRight(l, " ")))
			}
		})
	})

	Describe("Section spacing", func() {
		It("puts a blank line before a section that follows content", func() {
			d := columns.New()
			d.Item("Top", 1)
			d.Section("Sub", func(d *columns.Document) {
				d.Item("x", 2)
			})

			lines := strings.Split(d.String(), "\n")
			Expect(lines[0]).To(HaveSuffix("Top: 1"))
			Expect(lines[1]).To(Equal(""))
			Expect(lines[2]).To(Equal("  Sub:"))
		})

		It("does not put a blank before a section that begins the document", func() {
			d := columns.New()
			d.Section("Sub", func(d *columns.Document) {
				d.Item("x", 2)
			})

			Expect(d.String()).To(HavePrefix("  Sub:\n"))
		})

		It("does not double a blank already present before a section", func() {
			d := columns.New()
			d.Item("Top", 1)
			d.Blank()
			d.Section("Sub", func(d *columns.Document) {
				d.Item("x", 2)
			})

			lines := strings.Split(d.String(), "\n")
			Expect(lines[0]).To(HaveSuffix("Top: 1"))
			Expect(lines[1]).To(Equal(""))
			Expect(lines[2]).To(Equal("  Sub:"))
		})

		It("separates consecutive sections with a blank line", func() {
			d := columns.New()
			d.Section("A", func(d *columns.Document) { d.Item("x", 1) })
			d.Section("B", func(d *columns.Document) { d.Item("y", 2) })

			out := d.String()
			Expect(out).To(HavePrefix("  A:\n"))
			Expect(out).To(ContainSubstring("\n\n  B:\n"))
		})
	})

	Describe("LLM output format", func() {
		buildDoc := func() *columns.Document {
			d := columns.New()
			d.Heading("H")
			d.Item("Name", "node-1")

			return d
		}

		It("renders Markdown from String when CLAUDECODE=1", func() {
			GinkgoT().Setenv("CLAUDECODE", "1")

			d := buildDoc()
			md, err := d.Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(d.String()).To(Equal(string(md)))
			Expect(d.String()).To(ContainSubstring("- **Name:** node-1"))
		})

		It("lets LLMFORMAT=0 force text even under CLAUDECODE=1", func() {
			GinkgoT().Setenv("CLAUDECODE", "1")
			GinkgoT().Setenv("LLMFORMAT", "0")

			d := buildDoc()
			Expect(d.String()).To(ContainSubstring("Name: node-1"))
			Expect(d.String()).ToNot(ContainSubstring("|"))
		})

		It("renders Markdown when LLMFORMAT=1 without CLAUDECODE", func() {
			GinkgoT().Setenv("LLMFORMAT", "1")

			d := buildDoc()
			Expect(d.String()).To(ContainSubstring("- **Name:** node-1"))
		})

		It("keeps Bytes and WriteTo consistent with String", func() {
			GinkgoT().Setenv("CLAUDECODE", "1")

			d := buildDoc()
			var buf bytes.Buffer
			_, err := d.WriteTo(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(d.Bytes())).To(Equal(d.String()))
			Expect(buf.String()).To(Equal(d.String()))
		})

		It("falls back to text when Markdown rendering fails", func() {
			GinkgoT().Setenv("CLAUDECODE", "1")

			d := columns.New().Item("A", "1").Fields(42)
			Expect(d.Err()).To(MatchError(columns.ErrInvalidInput))
			Expect(d.String()).To(ContainSubstring("A: 1"))
			Expect(d.String()).ToNot(ContainSubstring("**A:**"))
		})
	})

	Describe("Value formatting", func() {
		It("groups numbers with thousands separators by default", func() {
			d := columns.New()
			d.Item("Total", 1234567)
			d.Item("Big", int64(9876543210))
			d.Item("Float", 1234.5)
			Expect(d.String()).To(ContainSubstring("Total: 1,234,567"))
			Expect(d.String()).To(ContainSubstring("Big: 9,876,543,210"))
			Expect(d.String()).To(ContainSubstring("Float: 1,234.5"))
		})

		It("renders ungrouped identifiers with Plain", func() {
			d := columns.New().Item("Port", columns.Plain(8080))
			Expect(d.String()).To(ContainSubstring("Port: 8080"))
		})

		It("joins a []string", func() {
			d := columns.New().Item("Tags", []string{"a", "b", "c"})
			Expect(d.String()).To(ContainSubstring("Tags: a, b, c"))
		})

		It("formats IEC and SI byte sizes", func() {
			Expect(columns.New().Item("x", columns.IBytes(1536)).String()).To(ContainSubstring("x: 1.5 KiB"))
			Expect(columns.New().Item("x", columns.Bytes(1500)).String()).To(ContainSubstring("x: 1.5 kB"))
			Expect(columns.New().Item("x", columns.IBytes(-1536)).String()).To(ContainSubstring("x: -1.5 KiB"))
		})

		It("formats durations and the never sentinel", func() {
			Expect(columns.New().Item("x", columns.Duration(90*time.Minute)).String()).To(ContainSubstring("x: 1h30m0s"))
			Expect(columns.New().Item("x", 36*time.Hour+12*time.Minute).String()).To(ContainSubstring("x: 1d12h12m0s"))
			Expect(columns.New().Item("x", columns.Duration(time.Duration(math.MaxInt64))).String()).To(ContainSubstring("x: never"))
		})

		It("formats percentages and guards a zero divisor", func() {
			Expect(columns.New().Item("x", columns.Percent(1, 4)).String()).To(ContainSubstring("x: 25%"))
			Expect(columns.New().Item("x", columns.Percent(1, 3)).String()).To(ContainSubstring("x: 33.3%"))
			Expect(columns.New(columns.WithEmptyText("-")).Item("x", columns.Percent(1, 0)).String()).To(ContainSubstring("x: -"))
		})

		It("formats composed values without Sprintf", func() {
			Expect(columns.New().Item("x", columns.Annotated("running", "pid 1234")).String()).To(ContainSubstring("x: running (pid 1234)"))
			Expect(columns.New().Item("x", columns.Join(", ", "a", "b", "c")).String()).To(ContainSubstring("x: a, b, c"))
			Expect(columns.New().Item("x", columns.YesNo(true)).String()).To(ContainSubstring("x: yes"))
			Expect(columns.New().Item("x", columns.Count(1, "file", "files")).String()).To(ContainSubstring("x: 1 file"))
			Expect(columns.New().Item("x", columns.Count(3, "file", "files")).String()).To(ContainSubstring("x: 3 files"))
		})

		It("formats a bare time.Time in local human form", func() {
			ts := time.Date(2026, 7, 17, 10, 30, 0, 0, time.Local)
			Expect(columns.New().Item("When", ts).String()).To(ContainSubstring("When: 2026-07-17 10:30:00"))
		})
	})

	Describe("Free-form lines", func() {
		It("adds a line at the left margin with no column or colon", func() {
			d := columns.New()
			d.Item("Name", "web")
			d.Printf("a note")
			Expect(d.String()).To(Equal("    Name: web\na note\n"))
		})

		It("parses Style markup in text output", func() {
			d := columns.New()
			d.Printf("{bold}hi{/bold}")
			Expect(d.String()).To(Equal("\x1b[1mhi\x1b[0m\n"))
		})

		It("indents a free-form line one level below its depth, with no margin", func() {
			d := columns.New()
			d.Indent().Indent()
			d.Printf("indented note")
			// indent 2, plus one level for the nesting: (2+1) * indentWidth 2 = 6 spaces
			Expect(d.String()).To(Equal("      indented note\n"))
		})

		It("indents below the heading of the section it sits in", func() {
			d := columns.New()
			d.Section("Sub", func(d *columns.Document) {
				d.Printf("under sub")
			})
			// the Section heading renders at "  Sub:", the note sits one level below it
			Expect(d.String()).To(Equal("  Sub:\n\n    under sub\n"))
		})

		It("formats with fmt.Sprintf", func() {
			d := columns.New()
			d.Printf("count=%d name=%s", 3, "web")
			Expect(d.String()).To(Equal("count=3 name=web\n"))
		})

		It("gets a blank line when it follows a heading", func() {
			d := columns.New()
			d.Heading("H")
			d.Printf("note")
			Expect(d.String()).To(Equal("H:\n\nnote\n"))
		})

		It("is ignored by the JSON renderer", func() {
			d := columns.New()
			d.Heading("H")
			d.Item("k", "v")
			d.Printf("a note")
			out, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(Equal("{\n  \"H\": {\n    \"k\": \"v\"\n  }\n}\n"))
		})

		It("renders as a plain paragraph in Markdown", func() {
			d := columns.New()
			d.Printf("{bold}note{/bold}")
			out, err := d.Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(Equal("note\n"))
		})

		It("adds a literal line with Print without interpreting verbs", func() {
			d := columns.New()
			d.Print("90% done: rate is 5/s")
			Expect(d.String()).To(Equal("90% done: rate is 5/s\n"))
		})

		It("parses Style markup in Print", func() {
			d := columns.New()
			d.Print("{bold}hi{/bold}")
			Expect(d.String()).To(Equal("\x1b[1mhi\x1b[0m\n"))
		})

		It("joins Print operands like fmt.Print", func() {
			Expect(columns.New().Print("a", "b").String()).To(Equal("ab\n"))
			Expect(columns.New().Print("n=", 3).String()).To(Equal("n=3\n"))
		})

		It("indents Print below the section heading", func() {
			d := columns.New()
			d.Section("Sub", func(d *columns.Document) {
				d.Print("note")
			})
			Expect(d.String()).To(Equal("  Sub:\n\n    note\n"))
		})
	})

	Describe("Value lists", func() {
		It("stacks a []string with Values", func() {
			d := columns.New()
			d.Values("Tags", []string{"web", "api", "db"})

			lines := strings.Split(strings.TrimRight(d.String(), "\n"), "\n")
			Expect(lines).To(HaveLen(3))
			Expect(lines[0]).To(HaveSuffix("Tags: web"))
			Expect(strings.TrimSpace(lines[1])).To(Equal("api"))
			Expect(strings.TrimSpace(lines[2])).To(Equal("db"))

			col := strings.Index(lines[0], "web")
			Expect(lines[1]).To(HavePrefix(strings.Repeat(" ", col)))
		})

		It("flattens a slice argument with ValuesAny so items stack", func() {
			d := columns.New()
			d.ValuesAny("Tags", []string{"web", "api"})

			lines := strings.Split(strings.TrimRight(d.String(), "\n"), "\n")
			Expect(lines).To(HaveLen(2))
			Expect(lines[0]).To(HaveSuffix("Tags: web"))
			Expect(strings.TrimSpace(lines[1])).To(Equal("api"))
		})

		It("stacks mixed variadic values with ValuesAny and keeps helpers whole", func() {
			d := columns.New()
			d.ValuesAny("Sizes", "small", columns.IBytes(1536))

			lines := strings.Split(strings.TrimRight(d.String(), "\n"), "\n")
			Expect(lines).To(HaveLen(2))
			Expect(lines[0]).To(HaveSuffix("Sizes: small"))
			Expect(strings.TrimSpace(lines[1])).To(Equal("1.5 KiB"))
		})

		It("keeps a []byte as one value rather than stacking bytes", func() {
			d := columns.New()
			d.ValuesAny("Raw", []byte("hi"))

			lines := strings.Split(strings.TrimRight(d.String(), "\n"), "\n")
			Expect(lines).To(HaveLen(1))
		})

		It("renders a flattened ValuesAny slice as a JSON list", func() {
			d := columns.New()
			d.ValuesAny("Tags", []string{"a", "b"})

			out, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(Equal("{\n  \"Tags\": [\n    \"a\",\n    \"b\"\n  ]\n}\n"))
		})
	})

	Describe("Styling", func() {
		It("renders markup as ANSI in text and strips it elsewhere", func() {
			d := columns.New()
			d.Heading("H")
			d.Item("State", columns.Style("{green}{bold}OK{/bold}{/green}"))

			text := d.String()
			Expect(text).To(ContainSubstring("\x1b[32m"))
			Expect(text).To(ContainSubstring("\x1b[1m"))
			Expect(text).To(ContainSubstring("OK"))

			md, err := d.Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(md)).To(ContainSubstring("- **State:** OK"))
			Expect(string(md)).ToNot(ContainSubstring("\x1b"))

			j, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(ContainSubstring(`"State": "OK"`))
			Expect(string(j)).ToNot(ContainSubstring("\x1b"))
		})

		It("passes through unknown tags and strips injected escapes", func() {
			d := columns.New().Item("x", columns.Style("a\x1b[31m{unknown}b"))
			out := d.String()
			Expect(out).To(ContainSubstring("{unknown}"))
			Expect(out).To(ContainSubstring("x: a{unknown}b"))
		})

		It("renders heading markup as ANSI in text and plain elsewhere", func() {
			d := columns.New()
			d.Heading("{bold}Node{/bold}")
			d.Item("Name", "n1")

			text := d.String()
			Expect(text).To(ContainSubstring("\x1b[1mNode\x1b[0m:"))
			Expect(text).ToNot(ContainSubstring("{bold}"))

			md, err := d.Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(md)).To(ContainSubstring("## Node"))
			Expect(string(md)).ToNot(ContainSubstring("\x1b"))
			Expect(string(md)).ToNot(ContainSubstring("{bold}"))

			j, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(ContainSubstring(`"Node": {`))
			Expect(string(j)).ToNot(ContainSubstring("\x1b"))
		})

		It("styles a Section heading through the same markup", func() {
			d := columns.New()
			d.Section("{green}Sub{/green}", func(d *columns.Document) {
				d.Item("k", "v")
			})

			text := d.String()
			Expect(text).To(ContainSubstring("\x1b[32m"))
			Expect(text).To(ContainSubstring("Sub"))

			j, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(ContainSubstring(`"Sub": {`))
		})

		It("strips an ANSI escape injected directly into a heading", func() {
			d := columns.New()
			d.Heading("a\x1b[31mb")
			Expect(d.String()).To(ContainSubstring("ab:"))
			Expect(d.String()).ToNot(ContainSubstring("\x1b[31m"))
		})
	})

	Describe("Value strings", func() {
		It("exposes the display form via String", func() {
			Expect(columns.Style("{bold}OK{/bold}").String()).To(Equal("\x1b[1mOK\x1b[0m"))
			Expect(columns.IBytes(1536).String()).To(Equal("1.5 KiB"))
			Expect(columns.Percent(1, 0).String()).To(Equal(""))
		})

		It("joins a multi-line value with newlines in String", func() {
			s := columns.Map(map[string]string{"a": "1", "b": "2"}).String()
			Expect(s).To(Equal("a: 1\nb: 2"))
		})

		It("satisfies fmt.Stringer", func() {
			var s fmt.Stringer = columns.Style("{green}up{/green}")
			Expect(s.String()).To(Equal("\x1b[32mup\x1b[0m"))
		})
	})

	Describe("Empty values", func() {
		It("substitutes the empty text", func() {
			d := columns.New(columns.WithEmptyText("n/a"))
			d.Item("A", "")
			d.Item("B", nil)
			Expect(d.String()).To(ContainSubstring("A: n/a"))
			Expect(d.String()).To(ContainSubstring("B: n/a"))
		})

		It("omits empty rows when configured", func() {
			d := columns.New(columns.WithOmitEmpty())
			d.Item("A", "kept")
			d.Item("B", "")
			out := d.String()
			Expect(out).To(ContainSubstring("A: kept"))
			Expect(out).ToNot(ContainSubstring("B:"))
		})
	})

	Describe("Conditional items", func() {
		It("adds with ItemIf only when the condition holds", func() {
			shown, hidden := 3, 0
			d := columns.New()
			d.ItemIf("Replicas", shown, shown > 0)
			d.ItemIf("Hidden", hidden, hidden > 0)
			out := d.String()
			Expect(out).To(ContainSubstring("Replicas: 3"))
			Expect(out).ToNot(ContainSubstring("Hidden"))
		})

		It("adds with ItemUnless only when the condition fails", func() {
			d := columns.New()
			d.ItemUnless("Note", "kept", strings.HasPrefix("kept", "foo"))
			d.ItemUnless("Skip", "x", strings.HasPrefix("foobar", "foo"))
			out := d.String()
			Expect(out).To(ContainSubstring("Note: kept"))
			Expect(out).ToNot(ContainSubstring("Skip"))
		})

		It("skips zero values with ItemUnlessZero", func() {
			d := columns.New()
			d.ItemUnlessZero("Count", 0)
			d.ItemUnlessZero("Name", "")
			d.ItemUnlessZero("When", time.Time{})
			d.ItemUnlessZero("Dur", time.Duration(0))
			d.ItemUnlessZero("Missing", nil)
			Expect(d.String()).To(BeEmpty())
		})

		It("adds non-zero values with ItemUnlessZero", func() {
			d := columns.New()
			d.ItemUnlessZero("Count", 3)
			d.ItemUnlessZero("Name", "web")
			out := d.String()
			Expect(out).To(ContainSubstring("Count: 3"))
			Expect(out).To(ContainSubstring("Name: web"))
		})

		It("tests a helper value by its raw form with ItemUnlessZero", func() {
			d := columns.New()
			d.ItemUnlessZero("Zero", columns.IBytes(0))
			d.ItemUnlessZero("Set", columns.IBytes(1536))
			out := d.String()
			Expect(out).ToNot(ContainSubstring("Zero"))
			Expect(out).To(ContainSubstring("Set: 1.5 KiB"))
		})
	})

	Describe("Formatted headings", func() {
		It("formats a heading with Headingf", func() {
			d := columns.New()
			d.Headingf("Node %d", 3)
			d.Item("x", "1")
			Expect(d.String()).To(Equal("Node 3:\n\n    x: 1\n"))
		})
	})

	Describe("Indentation", func() {
		It("clamps Outdent at zero", func() {
			d := columns.New()
			d.Outdent().Outdent()
			d.Item("A", "1")
			Expect(d.String()).To(Equal("    A: 1\n"))
		})
	})

	Describe("Error handling", func() {
		It("records an error for a non-map passed to Fields", func() {
			d := columns.New().Fields(42)
			Expect(d.Err()).To(MatchError(columns.ErrInvalidInput))
		})

		It("does not panic on a nil map", func() {
			var m map[string]string
			d := columns.New().Fields(m)
			Expect(d.Err()).ToNot(HaveOccurred())
			Expect(d.String()).To(BeEmpty())
		})
	})

	Describe("Structured rendering", func() {
		It("produces ordered JSON matching the layout", func() {
			expected := `{
  "Heading 1": {
    "Short": "Value",
    "Long Heading": "Long Value",
    "Items from a Map": {
      "X": "Y",
      "Z": "Z"
    },
    "Items from an Array": [
      "A",
      "B",
      "C"
    ],
    "Map Item 1": "Value 1",
    "Map Item 2": "Value 2"
  },
  "Heading 2": {
    "Item 1": "Value 1",
    "Indented Heading": {
      "Item 1": "Value 1"
    }
  }
}
`
			out, err := specExample().JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(Equal(expected))
		})

		It("preserves raw numeric and byte values in JSON", func() {
			d := columns.New()
			d.Item("Port", 8080)
			d.Item("Memory", columns.IBytes(1536))
			out, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(ContainSubstring(`"Port": 8080`))
			Expect(string(out)).To(ContainSubstring(`"Memory": 1536`))
		})

		It("strips control characters and escapes from string values", func() {
			d := columns.New()
			d.Item("Raw", "a\x1b[31mb\x07c")
			d.Item("Labels", columns.Map(map[string]string{"k": "x\x1b[32my"}))
			d.Values("List", []string{"p\x1b[1mq", "r"})

			j, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(j)).To(ContainSubstring(`"Raw": "abc"`))
			Expect(string(j)).To(ContainSubstring(`"k": "xy"`))
			Expect(string(j)).To(ContainSubstring(`"pq"`))
			Expect(string(j)).ToNot(ContainSubstring(`\u001b`))

			y, err := d.YAML()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(y)).To(ContainSubstring("Raw: abc"))
			Expect(string(y)).ToNot(ContainSubstring("\x1b"))
		})

		It("keeps raw numeric values while sanitizing sibling strings", func() {
			d := columns.New()
			d.Item("Count", 1536)
			d.Item("Name", "a\x1b[31mb")
			out, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(ContainSubstring(`"Count": 1536`))
			Expect(string(out)).To(ContainSubstring(`"Name": "ab"`))
		})

		It("returns an error for duplicate descriptions in JSON and YAML", func() {
			d := columns.New()
			d.Heading("H")
			d.Item("Port", 80)
			d.Item("Port", 443)

			_, err := d.JSON()
			Expect(err).To(MatchError(columns.ErrDuplicateDescription))
			_, err = d.YAML()
			Expect(err).To(MatchError(columns.ErrDuplicateDescription))
		})

		It("keeps rows before any heading at the top level", func() {
			d := columns.New()
			d.Item("a", "b")
			out, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(Equal("{\n  \"a\": \"b\"\n}\n"))
		})

		It("preserves key order in YAML", func() {
			out, err := specExample().YAML()
			Expect(err).ToNot(HaveOccurred())
			s := string(out)
			Expect(strings.Index(s, "Short:")).To(BeNumerically("<", strings.Index(s, "Long Heading:")))
			Expect(strings.Index(s, "Heading 1:")).To(BeNumerically("<", strings.Index(s, "Heading 2:")))
		})

		It("escapes markdown metacharacters in descriptions and values", func() {
			d := columns.New()
			d.Heading("H")
			d.Item("Sub_ject", "orders.*.new")
			out, err := d.Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(ContainSubstring(`- **Sub\_ject:** orders.\*.new`))
		})

		It("renders multi-value rows as nested bullets", func() {
			d := columns.New()
			d.Heading("H")
			d.Values("List", []string{"A", "B", "C"})
			out, err := d.Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("- **List:**\n    - A\n    - B\n    - C"))
		})

		It("escapes a leading block marker on a nested bullet", func() {
			d := columns.New()
			d.Heading("H")
			d.Values("Subjects", []string{"orders.new", ">"})
			out, err := d.Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("    - orders.new\n    - \\>"))
		})

		It("increases heading level with indent depth", func() {
			d := columns.New()
			d.Heading("Top")
			d.Section("Sub", func(d *columns.Document) {
				d.Section("Deeper", func(d *columns.Document) {
					d.Item("k", "v")
				})
			})
			out, err := d.Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("## Top"))
			Expect(string(out)).To(ContainSubstring("### Sub"))
			Expect(string(out)).To(ContainSubstring("#### Deeper"))
		})
	})

	Describe("Nested documents", func() {
		// nestedDoc has three indent levels and a heading (Tier: Default) whose
		// only entries are sub-sections, so it exercises nesting across formats.
		nestedDoc := func() *columns.Document {
			d := columns.New()
			d.Heading("Account Limits")
			d.Item("Max Message Payload", "1.0 MiB")
			d.Section("Tier: Default", func(d *columns.Document) {
				d.Section("Configuration Requirements", func(d *columns.Document) {
					d.Item("Stream Requires Max Bytes Set", false)
				})
				d.Section("Stream Resource Usage Limits", func(d *columns.Document) {
					d.Item("Memory", "0 B of Unlimited")
				})
			})

			return d
		}

		It("nests sections by indent in JSON", func() {
			expected := `{
  "Account Limits": {
    "Max Message Payload": "1.0 MiB",
    "Tier: Default": {
      "Configuration Requirements": {
        "Stream Requires Max Bytes Set": false
      },
      "Stream Resource Usage Limits": {
        "Memory": "0 B of Unlimited"
      }
    }
  }
}
`
			out, err := nestedDoc().JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(Equal(expected))
			// a heading holding only sub-sections is not an empty object
			Expect(string(out)).ToNot(ContainSubstring(`"Tier: Default": {}`))
		})

		It("nests sections by indent in YAML and quotes a colon key", func() {
			out, err := nestedDoc().YAML()
			Expect(err).ToNot(HaveOccurred())
			s := string(out)
			Expect(s).To(ContainSubstring("Account Limits:\n  Max Message Payload: 1.0 MiB"))
			Expect(s).To(ContainSubstring("\n  \"Tier: Default\":\n    Configuration Requirements:\n      Stream Requires Max Bytes Set: false"))
			Expect(s).To(ContainSubstring("\n    Stream Resource Usage Limits:\n      Memory: 0 B of Unlimited"))
		})

		It("nests headings by depth in Markdown", func() {
			expected := "## Account Limits\n\n" +
				"- **Max Message Payload:** 1.0 MiB\n\n" +
				"### Tier: Default\n\n" +
				"#### Configuration Requirements\n\n" +
				"- **Stream Requires Max Bytes Set:** false\n\n" +
				"#### Stream Resource Usage Limits\n\n" +
				"- **Memory:** 0 B of Unlimited\n"
			out, err := nestedDoc().Markdown()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(Equal(expected))
		})

		It("keeps headings at the same indent as siblings", func() {
			d := columns.New()
			d.Heading("A")
			d.Item("x", 1)
			d.Heading("B")
			d.Item("y", 2)

			out, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(Equal("{\n  \"A\": {\n    \"x\": 1\n  },\n  \"B\": {\n    \"y\": 2\n  }\n}\n"))
		})

		It("places a row added after a section as a sibling of it", func() {
			d := columns.New()
			d.Heading("Root")
			d.Section("Sub", func(d *columns.Document) {
				d.Item("inner", 1)
			})
			d.Item("after", 2)

			out, err := d.JSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(out)).To(Equal("{\n  \"Root\": {\n    \"Sub\": {\n      \"inner\": 1\n    },\n    \"after\": 2\n  }\n}\n"))
		})

		It("errors on two sibling headings with the same name", func() {
			d := columns.New()
			d.Heading("Peer")
			d.Item("a", 1)
			d.Heading("Peer")

			_, err := d.JSON()
			Expect(err).To(MatchError(columns.ErrDuplicateDescription))
		})

		It("errors when a sub-heading collides with a sibling row", func() {
			d := columns.New()
			d.Heading("H")
			d.Item("Config", 1)
			d.Section("Config", func(d *columns.Document) {
				d.Item("x", 2)
			})

			_, err := d.JSON()
			Expect(err).To(MatchError(columns.ErrDuplicateDescription))
		})

		It("errors when a headless row collides with a top-level heading", func() {
			d := columns.New()
			d.Item("A", 1)
			d.Heading("A")

			_, err := d.YAML()
			Expect(err).To(MatchError(columns.ErrDuplicateDescription))
		})
	})

	Describe("Writers", func() {
		It("WriteTo writes the text rendering", func() {
			var buf bytes.Buffer
			n, err := specExample().WriteTo(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(int(n)).To(Equal(buf.Len()))
			Expect(buf.String()).To(Equal(specExample().String()))
		})
	})
})
