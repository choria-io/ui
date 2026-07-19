# columns

Package `columns` builds aligned, columnar CLI output. You add entries to an
in-memory document line by line and render it at the end. The default rendering
is aligned text where every description colon lines up across all headings; the
same document can also be rendered to JSON, YAML and Markdown.

It exists to replace hand-tuned blocks like this, where the padding of every
line has to be recalculated whenever a description changes:

```go
fmt.Printf("        Item 1: %s\n", value)
fmt.Printf("   Long Item 2: %s\n", value2)
```

## Install

```
go get github.com/choria-io/ui/columns
```

## Quick start

```go
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

fmt.Print(d.String())
```

```
Heading 1:

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
```

The description column width is computed across every row, so the colons line up
across all headings. An indent shifts a whole section, including its heading, to
the right. Text output places a blank line after every heading, and before every
`Section` unless it begins the document; call `Blank` to add further separation.

## Building a document

Every builder method returns the `*Document` so calls can be chained, but the
natural style is a statement per line so you can drop conditionals and loops
between them.

| Method                          | Purpose                                                                                              |
|---------------------------------|------------------------------------------------------------------------------------------------------|
| `Heading(text)`                 | A section heading. Descriptions after it are grouped under it by the structured renderers.           |
| `Headingf(format, args...)`     | `Heading` with the text formatted by `fmt.Sprintf`.                                                  |
| `Item(desc, value)`             | A description and a single value.                                                                    |
| `ItemIf(desc, value, cond)`     | `Item`, but only when `cond` is true.                                                                |
| `ItemUnless(desc, value, cond)` | `Item`, but only when `cond` is false.                                                               |
| `ItemUnlessZero(desc, value)`   | `Item`, but only when `value` is not the zero value for its type.                                    |
| `Values(desc, []string)`        | A static description with a list of string values; first on its line, the rest below.                |
| `ValuesAny(desc, values...)`    | Like `Values` for non-string or mixed values; a slice argument (not `[]byte`) is expanded to stack.  |
| `Fields(m)`                     | One row per map entry, key as description and value as value. Keys are sorted by their string form.  |
| `Blank()`                       | An explicit blank line, in addition to the one text output adds after each heading.                  |
| `Printf(format, args...)`       | Free-form line, indented below the section heading; parsed for `{bold}`; ignored by JSON/YAML.       |
| `Print(a...)`                   | Like `Printf` but with no format string, so `%` is safe; operands joined like `fmt.Print`.           |
| `Embed(value)`                  | Embed an `Embeddable` such as a `*Table` with no description; ignored by JSON/YAML like `Print`.     |
| `Indent()` / `Outdent()`        | Change the indent level. `Outdent` never goes below zero.                                            |
| `Section(heading, fn)`          | Indented sub-section: indent, add heading, run `fn`, outdent; a blank precedes it unless first.      |
| `Err()`                         | The first error recorded while building, or nil.                                                     |

### Conditional rows

`ItemIf` and `ItemUnless` remove the `if` guards around optional rows:

```go
d.Item("Name", name)
d.ItemIf("Replicas", n, n > 0)                       // added only when n > 0
d.ItemUnless("Note", s, strings.HasPrefix(s, "foo")) // added unless it starts with "foo"
d.ItemUnlessZero("Errors", errCount)                 // added unless errCount is 0
```

## Value formatting

Values passed to `Item`, `Values`, `ValuesAny` and `Fields` are formatted
automatically by `columns.Format`, which is also exported for reuse:

| Type                 | Rendered as                                                    |
|----------------------|----------------------------------------------------------------|
| `string`             | as is                                                          |
| `[]string`           | joined with `, `                                               |
| `bool`               | `true` / `false`                                               |
| `int`, `uint`, ...   | thousands separators, `1,234,567`                              |
| `float32`, `float64` | grouped, three decimals, `1,234.500`                           |
| `*big.Int`           | grouped                                                        |
| `time.Duration`      | `HumanizeDuration`, e.g. `1h30m0s`, or `never` for the maximum |
| `time.Time`          | local `2006-01-02 15:04:05`                                    |
| `error`              | its `Error()`                                                  |
| `fmt.Stringer`       | its `String()`                                                 |
| `nil`                | empty                                                          |

Numbers are grouped by default because that is what most reported numbers want.
For identifiers that should not be grouped, such as ports, PIDs and years, wrap
the value with `Plain`.

Because a bare number is ambiguous (is that `int64` a byte count, a duration or
just a number?), the helpers below let the caller state the intent. Each returns
a `columns.Value` that keeps the underlying raw value for the structured
renderers, so a byte count serializes as a number, not the string `"1.5 KiB"`.

| Helper                       | Example call                       | Text                                  |
|------------------------------|------------------------------------|---------------------------------------|
| `IBytes(n)`                  | `IBytes(1610612736)`               | `1.5 GiB`                             |
| `Bytes(n)`                   | `Bytes(1500)`                      | `1.5 kB`                              |
| `Duration(d)`                | `Duration(90*time.Minute)`         | `1h30m0s`                             |
| `Time(t)`                    | `Time(t)`                          | `3 minutes ago`                       |
| `DateTime(t, layout...)`     | `DateTime(t)`                      | `2026-07-17 10:30:00`                 |
| `Percent(part, whole)`       | `Percent(3, 1000)`                 | `0.3%`                                |
| `Plain(v)`                   | `Plain(8080)`                      | `8080`                                |
| `Map(m)`                     | `Map(labels)`                      | a block of aligned `key: value` lines |
| `Annotated(main, note)`      | `Annotated("running", "pid 4823")` | `running (pid 4823)`                  |
| `Join(sep, items...)`        | `Join(", ", "a", "b")`             | `a, b`                                |
| `Bool(b, t, f)`              | `Bool(up, "up", "down")`           | `up`                                  |
| `YesNo(b)`                   | `YesNo(true)`                      | `yes`                                 |
| `Count(n, singular, plural)` | `Count(3, "file", "files")`        | `3 files`                             |

`columns.HumanizeDuration` and `columns.Format` are exported so the same
formatting can be used outside a document.

## Styling

Every value has control characters and ANSI escape sequences stripped, so
untrusted input cannot break the column alignment or inject terminal escapes.
`Style` is the one way to opt a value into terminal styling:

```go
d.Item("State", columns.Style("{green}{bold}running{/bold}{/green}"))
d.Item("Health", columns.Style("{yellow}degraded{/yellow} {dim}(2/3 checks){/dim}"))
```

Supported tags are `bold`, `dim`, `italic`, `underline`, `reverse` and the
colors `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white` and
`gray`. Close a tag with `{/name}`; `{reset}` or `{/}` clears all styling.
Unknown tags and stray braces are passed through as literal text.

The escape sequences are kept in text output only. Markdown, JSON and YAML
receive the plain, unstyled text, so styling never leaks into structured output.

Every `Value` also has a `String` method returning its display form, so `Style`
can build a styled string for use outside a document:

```go
label := columns.Style("{green}{bold}OK{/bold}{/green}").String()
```

`Heading` and `Section` text accepts the same markup, so a heading can be
emphasized without a separate styled-heading API:

```go
d.Heading("{bold}Node{/bold}")
```

The heading renders with ANSI in text output and as its plain visible text in
Markdown, JSON and YAML.

## Rendering

The default rendering is text. The structured formats preserve insertion order
and the raw values.

```go
d := columns.New()
d.Heading("Node")
d.Item("Name", "node-1")
d.Item("Cores", 8)
d.Item("Memory", columns.IBytes(1610612736))
```

Text (`String`, `Bytes`, `WriteTo`):

```
Node:

      Name: node-1
     Cores: 8
    Memory: 1.5 GiB
```

JSON (`JSON`, `RenderJSON`):

```json
{
  "Node": {
    "Name": "node-1",
    "Cores": 8,
    "Memory": 1610612736
  }
}
```

YAML (`YAML`, `RenderYAML`):

```yaml
Node:
  Name: node-1
  Cores: 8
  Memory: 1610612736
```

Markdown (`Markdown`, `RenderMarkdown`):

```markdown
## Node

- **Name:** node-1
- **Cores:** 8
- **Memory:** 1.5 GiB
```

Each heading becomes an ATX heading whose level tracks the indent depth: a
top-level heading is `##` and each further `Indent` or `Section` adds one, clamped
at `######`. Rows become `- **description:** value` bullets; a row with more than
one value line (a `Map`, a `Values` list or a value containing newlines) lists the
lines as nested bullets. Markdown metacharacters in descriptions and values are
escaped so they render literally.

The `String`, `Bytes` and `WriteTo` methods render text; the `JSON`, `YAML` and
`Markdown` methods return bytes and an error, and each has a `Render*` variant
that writes to an `io.Writer`.

## Embedding a table or sub-document

A value passed to `Item` that satisfies the `Embeddable` interface is nested
under its description and rendered in whichever format the document is rendered
to, deferring to the value's own `String`, `Markdown`, `JSON` and `YAML`. A
`*table.Table` and a `*Document` both satisfy it, so a table or a sub-document
can be embedded:

```go
t := table.NewTableWriter("Nodes")
t.AddHeaders("Name", "Cores")
t.AddRow("web1", 8)

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

The text form places the block below the description at the value column; JSON
and YAML nest the value's own structured output at the correct indentation, and
Markdown emits it as a labeled block. The interface is checked structurally, so
an implementer needs no dependency on this package. A nil `Embeddable` renders as
an empty value.

Use `Embed` instead of `Item` to add an embeddable with no description, the way
`Print` adds a free-form line. It renders as a standalone block at the current
indent in text and Markdown, and, having no key to nest under, is ignored by the
JSON and YAML renderers.

```go
d := columns.New()
d.Embed(t)
```

```go
type Embeddable interface {
	String() string
	Markdown() ([]byte, error)
	JSON() ([]byte, error)
	YAML() ([]byte, error)
}
```

## Options

| Option               | Effect                                                                |
|----------------------|-----------------------------------------------------------------------|
| `WithMargin(n)`      | Spaces to the left of the description column. Default 4.              |
| `WithIndentWidth(n)` | Spaces added per indent level. Default 2.                             |
| `WithOmitEmpty()`    | Skip any row whose value renders empty.                               |
| `WithEmptyText(s)`   | Placeholder rendered for an empty or nil value, such as `-` or `n/a`. |

## Structured output notes

- Headings nest by indent: a `Section` or `Indent` places a heading inside the
  enclosing heading, so `JSON` and `YAML` mirror the text and Markdown structure.
  Keys keep the order they were added.
- Rows attach to the heading at their indent; rows added before any heading appear
  at the top level.
- Two entries that would share a key within the same object, whether two rows, two
  sub-headings or a row and a sub-heading, cannot both be object keys, so `JSON`
  and `YAML` return `ErrDuplicateDescription`. Text and Markdown render such a
  document without complaint.
- `Fields` on a non-map value records `ErrInvalidInput`, retrievable via `Err`
  and returned by the structured renderers.
- `JSON` and `YAML` are a faithful serialization of the displayed document, keyed
  by its human labels; for a stable machine contract, serialize your own domain
  model rather than these labels.

## License

Apache-2.0, part of the Choria project.
