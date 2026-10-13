// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package table_test

import (
	"fmt"

	"github.com/choria-io/ui/table"
)

// Example shows the common pattern: create a table, add headers and rows, then
// render it to text.
func Example() {
	t := table.NewTableWriter("Nodes")
	t.AddHeaders("Name", "Cores")
	t.AddRow("web1", 8)
	t.AddRow("web2", 4)

	fmt.Print(t.Render())

	// Output:
	// ╭──────────────╮
	// │     Nodes    │
	// ├──────┬───────┤
	// │ Name │ Cores │
	// ├──────┼───────┤
	// │ web1 │ 8     │
	// │ web2 │ 4     │
	// ╰──────┴───────╯
}

// ExampleTable_JSON renders the same table as JSON. Plain scalar cells keep their
// native type, so numbers stay numeric.
func ExampleTable_JSON() {
	t := table.NewTableWriter("Nodes")
	t.AddHeaders("Name", "Cores")
	t.AddRow("web1", 8)
	t.AddRow("web2", 4)

	b, _ := t.JSON()
	fmt.Print(string(b))

	// Output:
	// {
	//   "title": "Nodes",
	//   "headers": [
	//     "Name",
	//     "Cores"
	//   ],
	//   "rows": [
	//     [
	//       "web1",
	//       8
	//     ],
	//     [
	//       "web2",
	//       4
	//     ]
	//   ]
	// }
}

// ExampleTable_YAML renders the table as YAML with the same shape as JSON.
func ExampleTable_YAML() {
	t := table.NewTableWriter("Nodes")
	t.AddHeaders("Name", "Cores")
	t.AddRow("web1", 8)
	t.AddRow("web2", 4)

	b, _ := t.YAML()
	fmt.Print(string(b))

	// Output:
	// title: Nodes
	// headers:
	// - Name
	// - Cores
	// rows:
	// - - web1
	//   - 8
	// - - web2
	//   - 4
}

// ExampleTable_Markdown renders the table as a Markdown table with an H1 title.
func ExampleTable_Markdown() {
	t := table.NewTableWriter("Nodes")
	t.AddHeaders("Name", "Cores")
	t.AddRow("web1", 8)
	t.AddRow("web2", 4)

	b, _ := t.Markdown()
	fmt.Print(string(b))

	// Output:
	// # Nodes
	// | Name | Cores |
	// | --- | --- |
	// | web1 | 8 |
	// | web2 | 4 |
}

// ExampleTable_FitWidth keeps the text output within 60 cells by wrapping the
// prose columns, leaving the path whole. Use FitTerminal to fit the terminal.
func ExampleTable_FitWidth() {
	t := table.NewTableWriter("Refused")
	t.AddHeaders("Target", "Reason", "To get past it")
	t.AddRow("items/auth-rework.md", "the item has landed and cannot be modified", "reopen the item, then retry the change")
	t.FitWidth(60)

	fmt.Print(t.Render())

	// Output:
	// ╭──────────────────────────────────────────────────────────╮
	// │                          Refused                         │
	// ├──────────────────────┬─────────────────┬─────────────────┤
	// │ Target               │ Reason          │ To get past it  │
	// ├──────────────────────┼─────────────────┼─────────────────┤
	// │ items/auth-rework.md │ the item has    │ reopen the      │
	// │                      │ landed and      │ item, then      │
	// │                      │ cannot be       │ retry the       │
	// │                      │ modified        │ change          │
	// ╰──────────────────────┴─────────────────┴─────────────────╯
}

// ExampleTable_WrapColumns wraps the second column at 20 cells.
func ExampleTable_WrapColumns() {
	t := table.NewTableWriter("")
	t.AddHeaders("Name", "Description")
	t.AddRow("web1", "serves the public website and the api")
	t.WrapColumns(20, 2)

	fmt.Print(t.Render())

	// Output:
	// ╭──────┬──────────────────────╮
	// │ Name │ Description          │
	// ├──────┼──────────────────────┤
	// │ web1 │ serves the public    │
	// │      │ website and the api  │
	// ╰──────┴──────────────────────╯
}
