// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns_test

import (
	"fmt"

	"github.com/choria-io/ui/columns"
	"github.com/choria-io/ui/table"
)

// ExampleDocument_embed embeds a table under an Item. The table renders in
// whichever format the Document is rendered to; here it is the text form, placed
// below the description and aligned to the value column.
func ExampleDocument_embed() {
	t := table.NewTableWriter("Nodes")
	t.AddHeaders("Name", "Cores")
	t.AddRow("web1", 8)

	d := columns.New()
	d.Item("Count", 3)
	d.Item("Detail", t)

	fmt.Print(d.String())

	// Output:
	//      Count: 3
	//     Detail:
	//             ╭──────────────╮
	//             │     Nodes    │
	//             ├──────┬───────┤
	//             │ Name │ Cores │
	//             ├──────┼───────┤
	//             │ web1 │ 8     │
	//             ╰──────┴───────╯
}
