package main

import (
    "fmt"

    "github.com/ds8088/go-depgraph"
)

func main() {
    dg := depgraph.NewDependencyGraph[string]()
    dg.Add("A")
    dg.Add("B", "A")
    dg.Add("C")

    // ResolveIter returns an Option-like pair of elements.
    // Either `err` is nil, in which case `el` is valid,
    // or `err` is non-nil, and if so, we should handle the error.
    for el, err := range dg.ResolveIter() {
        if err != nil {
            panic(fmt.Errorf("resolving dependency graph: %w", err))
        }

        // Should consecutively print "A", "C" and "B".
        fmt.Println(el)
    }
}
