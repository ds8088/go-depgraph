// Package depgraph implements a stable dependency graph
// and an algorithm for its resolution.
package depgraph

import (
	"errors"
	"fmt"
	"iter"
)

var (
	// ErrCircularDependency is used, as per its name, in cases when a graph cannot be resolved
	// due to at least two of its nodes being dependent on each other,
	// either directly or transitively.
	// This error may be wrapped; to account for this, use either "errors.Is" or "errors.As"
	// instead of a simple comparison.
	ErrCircularDependency = errors.New("circular dependency")

	// ErrUnknownDependency is used when resolving the graph and encountering a node,
	// which depends on an other unknown node (meaning that the graph doesn't know anything about it).
	// This error may be wrapped; to account for this, use either "errors.Is" or "errors.As"
	// instead of a simple comparison.
	ErrUnknownDependency = errors.New("unknown dependency")
)

type (
	depList[T comparable] = map[T]struct{}
	depEdge[T comparable] = struct {
		name T
		deps depList[T]
	}
)

// DependencyGraph represents a stable dependency graph.
// By stable we mean that it keeps the initial ordering of elements,
// according to the insertion order, which may only be broken to solve a dependency.
type DependencyGraph[T comparable] struct {
	edges   []*depEdge[T]
	edgeMap map[T]*depEdge[T]
}

// NewDependencyGraph creates a new stable dependency graph.
func NewDependencyGraph[T comparable]() *DependencyGraph[T] {
	return &DependencyGraph[T]{
		edgeMap: map[T]*depEdge[T]{},
	}
}

// validate iterates over all graph edges and checks if their dependencies exist.
func (dg *DependencyGraph[T]) validate() error {
	for _, edge := range dg.edgeMap {
		for dep := range edge.deps {
			if _, ok := dg.edgeMap[dep]; !ok {
				return fmt.Errorf("looking up dependency \"%v\": %w", dep, ErrUnknownDependency)
			}
		}
	}

	return nil
}

// Add adds an element to the end of dependency graph's edge list.
// It may be called multiple times during the graph construction,
// in which case, its dependencies get concatenated together.
func (dg *DependencyGraph[T]) Add(name T, deps ...T) {
	// Determine whether we already have this edge.
	edge, ok := dg.edgeMap[name]
	if !ok {
		// If not, create and add it.
		edge = &depEdge[T]{
			name: name,
			deps: depList[T]{},
		}

		dg.edgeMap[name] = edge
		dg.edges = append(dg.edges, edge)
	}

	// Irregardless of whether this edge is new or existing,
	// add all deps to its dep list.
	for _, dep := range deps {
		edge.deps[dep] = struct{}{}
	}
}

// ResolveIter returns an iterator that yields the graph's elements in dependency order.
// If a circular dependency is detected, or if the graph is invalid,
// the iterator yields a pair of (zero element, error) and stops.
func (dg *DependencyGraph[T]) ResolveIter() iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T

		err := dg.validate()
		if err != nil {
			yield(zero, fmt.Errorf("validating dependency graph: %w", err))
			return
		}

		fmax := 0
		edges := dg.edges
		refcounts := make(map[T]int, len(edges))

		// Save the current number of dependencies for each edge.
		for _, edge := range edges {
			refcounts[edge.name] = len(edge.deps)
		}

		// Promote all free edges to the start of the edge list,
		// whilst keeping the stable ordering.
		for i, edge := range edges {
			if len(edge.deps) == 0 {
				edges[fmax], edges[i] = edges[i], edges[fmax]
				fmax++
			}
		}

		// Keep iterating while we still have at least one remaining free edge.
		for fcur := 0; fcur < fmax; fcur++ {
			this := edges[fcur]

			// Since this edge has no dependencies - yield it to our caller.
			if !yield(this.name, nil) {
				return
			}

			// If a later edge depends on this edge - clear the (already resolved) dependency.
			// If, after clearing, an edge becomes free - promote it to the free list.
			for i := fcur + 1; i < len(edges); i++ {
				if _, ok := edges[i].deps[this.name]; ok {
					// We can't really clear a dependency because that would require tracking
					// a lot of state; however, we can simply decrease the reference counter.
					// This is enough to track when an edge becomes free.
					refcounts[edges[i].name]--

					if refcounts[edges[i].name] == 0 {
						// Promote the edge.
						edges[fmax], edges[i] = edges[i], edges[fmax]
						fmax++
					}
				}
			}
		}

		// If we stopped before reaching fmax,
		// not all edges have been processed, thus there is a circular dependency.
		if fmax != len(edges) {
			yield(zero, ErrCircularDependency)
		}
	}
}

func (dg *DependencyGraph[T]) Resolve() ([]T, error) {
	// The resulting slice will be the same length as the graph's edge count,
	// therefore allocate all the memory beforehand.
	res := make([]T, 0, len(dg.edges))

	for el, err := range dg.ResolveIter() {
		if err != nil {
			return nil, err
		}

		res = append(res, el)
	}

	return res, nil
}
