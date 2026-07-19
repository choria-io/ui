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
