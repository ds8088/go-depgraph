package main

import (
	"fmt"

	"github.com/ds8088/go-depgraph"
)

func main() {
	// Create a graph and add a bunch of string values.
	dg := depgraph.NewDependencyGraph[string]()
	dg.Add("A")
	dg.Add("B", "A")
	dg.Add("C")
	dg.Add("D", "B", "A")
	dg.Add("E")

	// Resolve the graph.
	res, err := dg.Resolve()
	if err != nil {
		panic(fmt.Errorf("resolving dependency graph: %w", err))
	}

	// Should print "[A C E B D]".
	// Note that the ordering between A, C and E is preserved.
	fmt.Printf("Order of resolution: %v\n", res)
}
