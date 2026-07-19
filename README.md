# ui

A collection of small UI helpers for building Choria command line tools.

## Packages

| Package               | Description                                                                                    |
|-----------------------|------------------------------------------------------------------------------------------------|
| [`columns`](columns/) | Build aligned, columnar CLI output line by line and render it to text, JSON, YAML or Markdown. |
| [`table`](table/)     | Render tabular CLI output on top of go-pretty as text, Markdown, JSON or YAML.                 |

The two compose: a `table.Table` can be embedded directly into a `columns.Document`
via `Item`, rendering in whichever format the document is rendered to.

## License

Apache-2.0, part of the Choria project.
