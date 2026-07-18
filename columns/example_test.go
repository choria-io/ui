// Copyright (c) 2026, R.I. Pienaar and the Choria Project contributors
//
// SPDX-License-Identifier: Apache-2.0

package columns_test

import (
	"fmt"
	"os"
	"time"

	"github.com/choria-io/ui/columns"
)

// Example shows the common pattern: create a Document, add entries line by line
// using helpers to format values, then render to text.
func Example() {
	d := columns.New()
	d.Heading("Service")
	d.Item("Name", "web-frontend")
	d.Item("State", columns.Annotated("running", "pid 4823"))
	d.Item("Uptime", 36*time.Hour+12*time.Minute)
	d.Item("Memory", columns.IBytes(1610612736))
	d.Item("Requests", 1048576)
	d.Item("Error Rate", columns.Percent(3, 1000))
	d.Item("Healthy", columns.YesNo(true))

	fmt.Print(d.String())

	// Output:
	// Service:
	//
	//           Name: web-frontend
	//          State: running (pid 4823)
	//         Uptime: 1d12h12m0s
	//         Memory: 1.5 GiB
	//       Requests: 1,048,576
	//     Error Rate: 0.3%
	//        Healthy: yes
}

// ExampleDocument_Section shows an indented sub-section. The heading and its body
// both shift to the right, and the value columns stay aligned with the rest of
// the document.
func ExampleDocument_Section() {
	d := columns.New()
	d.Heading("Cluster")
	d.Item("Nodes", 3)
	d.Section("Leader", func(d *columns.Document) {
		d.Item("Name", "node-1")
		d.Item("Term", 7)
	})

	fmt.Print(d.String())

	// Output:
	// Cluster:
	//
	//     Nodes: 3
	//
	//   Leader:
	//
	//        Name: node-1
	//        Term: 7
}

// ExampleDocument_Fields turns a map into one row per key, while columns.Map
// renders a map as a block under a single description.
func ExampleDocument_Fields() {
	d := columns.New()
	d.Heading("Labels")
	d.Fields(map[string]string{"env": "prod", "region": "eu-west", "tier": "web"})

	fmt.Print(d.String())

	// Output:
	// Labels:
	//
	//        env: prod
	//     region: eu-west
	//       tier: web
}

// ExampleMap renders a map as an aligned block occupying the value column.
func ExampleMap() {
	d := columns.New()
	d.Item("Labels", columns.Map(map[string]string{"env": "prod", "tier": "web"}))

	fmt.Print(d.String())

	// Output:
	//     Labels: env:  prod
	//             tier: web
}

// ExampleDocument_JSON renders the same model as JSON. Raw values are preserved,
// so numbers and byte counts stay numeric.
func ExampleDocument_JSON() {
	d := columns.New()
	d.Heading("Node")
	d.Item("Name", "node-1")
	d.Item("Cores", 8)
	d.Item("Memory", columns.IBytes(1610612736))

	_ = d.RenderJSON(os.Stdout)

	// Output:
	// {
	//   "Node": {
	//     "Name": "node-1",
	//     "Cores": 8,
	//     "Memory": 1610612736
	//   }
	// }
}

// ExampleDocument_YAML renders the model as YAML with key order preserved.
func ExampleDocument_YAML() {
	d := columns.New()
	d.Heading("Node")
	d.Item("Name", "node-1")
	d.Item("Cores", 8)
	d.Item("Memory", columns.IBytes(1610612736))

	_ = d.RenderYAML(os.Stdout)

	// Output:
	// Node:
	//   Name: node-1
	//   Cores: 8
	//   Memory: 1610612736
}

// ExampleDocument_Markdown renders the model as headings with bullet lists, the
// heading level following the indent depth.
func ExampleDocument_Markdown() {
	d := columns.New()
	d.Heading("Node")
	d.Item("Name", "node-1")
	d.Item("Cores", 8)
	d.Item("Memory", columns.IBytes(1610612736))

	_ = d.RenderMarkdown(os.Stdout)

	// Output:
	// ## Node
	//
	// - **Name:** node-1
	// - **Cores:** 8
	// - **Memory:** 1.5 GiB
}

// ExampleDocument_omitEmpty drops rows whose value renders empty, removing the
// usual "if value != \"\"" guards.
func ExampleDocument_omitEmpty() {
	d := columns.New(columns.WithOmitEmpty())
	d.Item("Present", "here")
	d.Item("Missing", "")
	d.Item("Also present", "there")

	fmt.Print(d.String())

	// Output:
	//          Present: here
	//     Also present: there
}

func ExampleIBytes() {
	fmt.Print(columns.New().Item("Memory", columns.IBytes(1610612736)).String())
	// Output:
	//     Memory: 1.5 GiB
}

func ExampleDuration() {
	fmt.Print(columns.New().Item("Uptime", columns.Duration(90*time.Minute)).String())
	// Output:
	//     Uptime: 1h30m0s
}

// ExamplePlain renders a number without thousands separators, for identifiers
// such as ports that should not be grouped.
// ExampleDocument_ItemIf adds rows conditionally, replacing "if cond { ... }"
// guards. ItemUnless is the inverse.
func ExampleDocument_ItemIf() {
	replicas := 3
	note := ""

	d := columns.New()
	d.Item("Name", "web")
	d.ItemIf("Replicas", replicas, replicas > 0)
	d.ItemIf("Note", note, note != "")

	fmt.Print(d.String())

	// Output:
	//         Name: web
	//     Replicas: 3
}

func ExamplePlain() {
	fmt.Print(columns.New().Item("Port", columns.Plain(8080)).String())
	// Output:
	//     Port: 8080
}

// ExampleStyle opts a value into terminal markup. The escape sequences are kept
// in text output but stripped from Markdown and the structured renderers.
func ExampleStyle() {
	d := columns.New()
	d.Item("State", columns.Style("{green}{bold}healthy{/bold}{/green}"))

	// Rendered to a terminal this prints "healthy" in bold green; the plain
	// text is used everywhere ANSI cannot go.
	md, _ := d.Markdown()
	fmt.Print(string(md))
	// Output:
	// - **State:** healthy
}

func ExamplePercent() {
	fmt.Print(columns.New().Item("Ratio", columns.Percent(1, 3)).String())
	// Output:
	//     Ratio: 33.3%
}

func ExampleAnnotated() {
	fmt.Print(columns.New().Item("State", columns.Annotated("running", "pid 1234")).String())
	// Output:
	//     State: running (pid 1234)
}

func ExampleCount() {
	fmt.Print(columns.New().Item("Files", columns.Count(3, "file", "files")).String())
	// Output:
	//     Files: 3 files
}
