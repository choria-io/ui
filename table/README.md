# table

Package `table` renders tabular CLI output on top of
[go-pretty](https://github.com/jedib0t/go-pretty). You add headers, rows, an
optional footer and separators to an in-memory table and render it at the end.
The default rendering is an aligned text table, and the same table can also be
rendered to Markdown, JSON and YAML.

It shares the value formatting and format selection of the sibling `columns`
package, and a `*Table` can be embedded directly into a `columns.Document`.

## Install

```
go get github.com/choria-io/ui/table
```

## Quick start

```go
t := table.NewTableWriter("Servers")
t.AddHeaders("Name", "Cores", "Memory")
t.AddRow("web1", 8, 1610612736)
t.AddRow("web2", 4, 1073741824)
t.AddSeparator()
t.AddFooter("Total", 12, "")

fmt.Print(t.Render())
```

```
╭───────────────────────────────╮
│            Servers            │
├───────┬───────┬───────────────┤
│ Name  │ Cores │ Memory        │
├───────┼───────┼───────────────┤
│ web1  │ 8     │ 1,610,612,736 │
│ web2  │ 4     │ 1,073,741,824 │
├───────┼───────┼───────────────┤
│ Total │ 12    │               │
╰───────┴───────┴───────────────╯
```

Every builder returns the `*Table` so calls can be chained, but a statement per
line reads well when loops and conditionals sit between them.

| Method                        | Purpose                                                                        |
|-------------------------------|--------------------------------------------------------------------------------|
| `NewTableWriter(title)`       | A new table with an optional centered title.                                   |
| `NewTableWriterf(format, a…)` | `NewTableWriter` with the title formatted by `fmt.Sprintf`.                    |
| `AddHeaders(items…)`          | Set the column headers. Calling it again replaces them.                        |
| `AddRow(items…)`              | Append a data row.                                                             |
| `AddSeparator()`              | A horizontal rule in text and Markdown; ignored by JSON and YAML.              |
| `AddFooter(items…)`           | Set a footer row, typically totals or a summary. Calling it again replaces it. |
| `FitTerminal()`               | Wrap cells so the text output fits the terminal. See below.                    |
| `FitWidth(width)`             | Wrap cells so the text output is at most `width` cells wide.                   |
| `WrapColumns(width, cols…)`   | Wrap the given columns, numbered from 1, at `width` cells.                     |

## Value formatting

Cells are formatted by `columns.Format`: strings as is, numbers with thousands
separators, durations and times in human form, and `[]string` joined. The
`columns` helpers such as `columns.IBytes`, `columns.Duration` and
`columns.Percent` work as cells too. Because cells are formatted to strings, the
text table left-aligns them.

Every cell has ANSI escape sequences and control characters stripped, so
untrusted values cannot break the table layout or inject terminal escapes.

## Rendering

The default rendering is text. When `LLMFORMAT` (parsed as a bool) or
`CLAUDECODE=1` marks an LLM or agent consumer, `String`, `Render`, `Bytes` and
`WriteTo` return the Markdown rendering instead, matching `columns`.

Text (`String`, `Render`, `Bytes`, `WriteTo`):

```
╭──────────────╮
│     Nodes    │
├──────┬───────┤
│ Name │ Cores │
├──────┼───────┤
│ web1 │ 8     │
╰──────┴───────╯
```

Markdown (`Markdown`, `RenderMarkdown`):

```markdown
# Nodes
| Name | Cores |
| --- | --- |
| web1 | 8 |
```

JSON (`JSON`, `RenderJSON`) has the shape `{title, headers, rows, footer}`, with
`title`, `headers` and `footer` omitted when unset. Plain scalar cells keep their
native type, so numbers stay numeric; other values, including formatting helpers,
use their display string. Separators are dropped.

```json
{
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
```

YAML (`YAML`, `RenderYAML`) has the same shape:

```yaml
title: Nodes
headers:
- Name
- Cores
rows:
- - web1
  - 8
```

`String`, `Render`, `Bytes` and `WriteTo` render text (or Markdown per the
environment); `Markdown`, `JSON` and `YAML` return bytes and an error, and each
has a `Render*` variant that writes to an `io.Writer`. `Render` is an alias for
`String` for callers used to go-pretty.

## Fitting the terminal

Long prose in a cell makes a text table wider than the terminal, which then
breaks every line and garbles the box. `FitTerminal` soft-wraps cells so the
table fits:

```go
t := table.NewTableWriter("Refused")
t.AddHeaders("Target", "Reason", "To get past it")
t.AddRow("items/auth-rework.md", "the item has landed and cannot be modified without reopening it", "run planboard item reopen auth-rework, then retry the change")
t.AddRow("project.md", "the sequence would reference a retired item", "remove it from the sequence")
t.FitTerminal()

fmt.Print(t.Render())
```

Unfitted this table is 153 cells wide. In an 80 column terminal it renders as:

```
╭──────────────────────────────────────────────────────────────────────────────╮
│                                    Refused                                   │
├──────────────────────┬───────────────────────────┬───────────────────────────┤
│ Target               │ Reason                    │ To get past it            │
├──────────────────────┼───────────────────────────┼───────────────────────────┤
│ items/auth-rework.md │ the item has landed and   │ run planboard item reopen │
│                      │ cannot be modified        │ auth-rework, then retry   │
│                      │ without reopening it      │ the change                │
│ project.md           │ the sequence would        │ remove it from the        │
│                      │ reference a retired item  │ sequence                  │
╰──────────────────────┴───────────────────────────┴───────────────────────────╯
```

| Method                      | Purpose                                                                                    |
|-----------------------------|--------------------------------------------------------------------------------------------|
| `FitWidth(width)`           | Wrap cells so each line is at most `width` cells wide, borders included. 0 disables it.    |
| `FitTerminal()`             | `FitWidth` with the width from `TerminalWidth`, read when the table renders.               |
| `WrapColumns(width, cols…)` | Wrap the given columns at `width` cells. Columns are numbered from 1, like go-pretty.      |
| `TerminalWidth()`           | `COLUMNS` when it is a positive integer, else the width of the terminal on stdout, else 0. |

No column is narrowed below its longest word, header included, so names, paths,
numbers and commands stay whole and only text wraps, at spaces. A word is a run
of characters between whitespace; in text such as Chinese or Japanese, which has
no spaces, a line can also break either side of each character. Footer cells
count like the other rows.

Until the table fits, columns are narrowed in this order, each step only as far
as needed:

1. Columns whose cells cannot wrap, but whose header is wider than the cells,
   wrap the header at spaces. "Program or parent" above `Program/one` becomes
   two header lines.
2. Columns whose cells can wrap are narrowed, widest first, down to 10 cells or
   their longest word, whichever is wider.
3. Those columns are narrowed further, down to their longest word.

When the table still does not fit, it renders wider than asked and the terminal
wraps every line. Words are never broken. That happens when the longest word of
each column, plus 3 cells per column and 1 for the border, exceeds the width.
Shorter headers, fewer columns in narrow terminals, or `WrapColumns` on a column
where breaking words is acceptable avoid it.

- Line breaks already in a cell are kept, and a wrapped line loses repeated
  spaces and indentation.
- The title is wrapped at word boundaries to the width of the table.
- `WrapColumns` sets a hard width: a word longer than it is broken across
  lines. Its limit acts as the natural width of those columns when fitting.

Settings are applied when the table renders, so they can be set before or after
adding rows. `FitWidth` and `FitTerminal` replace each other, the last call
winning. Only the text output wraps: Markdown, JSON and YAML, including the
Markdown that `String` returns for an LLM or agent consumer, are unchanged, so a
Markdown table never gains line breaks in its cells.

`TerminalWidth` honors `COLUMNS` even when output is piped, and otherwise
measures stdout, returning 0 when stdout is not a terminal so fitting is skipped.
When the table is written to stderr or a file, or embedded in a document, measure
or adjust the width yourself and use `FitWidth`.

## Embedding in a Document

A `*Table` satisfies `columns.Embeddable`, so it can be passed to
`columns.Document.Item` and is nested in whichever format the document is
rendered to:

```go
d := columns.New()
d.Item("Count", 3)
d.Item("Detail", t)

fmt.Print(d.String())
```

```
     Count: 3
    Detail:
            ╭──────────────╮
            │     Nodes    │
            ├──────┬───────┤
            │ Name │ Cores │
            ├──────┼───────┤
            │ web1 │ 8     │
            ╰──────┴───────╯
```

The same document rendered to JSON or YAML embeds the table's own JSON or YAML at
the correct indentation.

An embedded table is indented by the document, which `FitTerminal` does not know
about. Subtract the indent instead, `t.FitWidth(table.TerminalWidth() - 12)` in
the example above. When the width is unknown the result is negative, which
disables fitting just as 0 does.

To embed a table with no description, use `Embed` rather than `Item`. It places
the table as a standalone block, like the document's `Print`, and is ignored by
the JSON and YAML renderers since it has no key to nest under.

```go
d.Embed(t)
```

## Notes

- A table with a title but no headers or rows renders as empty output.
- Markdown, JSON and YAML escaping is handled by go-pretty and the standard
  encoders; the text and Markdown output uses go-pretty's rounded style.
- A `Table` is not safe for concurrent use.

## License

Apache-2.0, part of the Choria project.
