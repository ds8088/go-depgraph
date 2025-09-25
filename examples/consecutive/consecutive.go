package main

import (
    "fmt"

    "github.com/ds8088/go-depgraph"
)

func main() {
    dg := depgraph.NewDependencyGraph[string]()
    dg.Add("A")
    dg.Add("B", "A")

    // Resolve the graph.
    res, err := dg.Resolve()
    if err != nil {
        panic(fmt.Errorf("resolving dependency graph, 1st pass: %w", err))
    }

    // Should print "[A B]".
    fmt.Printf("Order of resolution in 1st pass: %v\n", res)

    // Add a bunch of dependencies after the initial resolution.
    dg.Add("A", "E")
    dg.Add("C")
    dg.Add("D", "B")
    dg.Add("E")

    // Resolve the graph once more.
    res, err = dg.Resolve()
    if err != nil {
        panic(fmt.Errorf("resolving dependency graph, 2nd pass: %w", err))
    }

    // Should print "[C E A B D]".
    fmt.Printf("Order of resolution in 2nd pass: %v\n", res)
}
