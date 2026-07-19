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
