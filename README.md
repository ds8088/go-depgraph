# go-depgraph

Simple, straightforward implementation of a stable dependency graph in Go
without any external dependencies.

Resolution works in an `O(n)` time and preserves the original insertion order
as much as possible.

It has a reasonably comprehensive test suite and a 100% test coverage.

## Prerequisites

At least Go 1.23 is required.

## Installation

```sh
go get github.com/ds8088/go-depgraph
```

```go
import "github.com/ds8088/go-depgraph"
```

## Examples

Simple graph resolution:

```go
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
```

Iterator-based approach:

```go
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
```

Adding or updating elements in-between graph resolutions is also supported:

```go
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
```
