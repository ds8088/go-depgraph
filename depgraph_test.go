package depgraph

import (
	"errors"
	"slices"
	"testing"
)

// TestResolve tests the dependency resolution algorithm.
func TestResolve(t *testing.T) {
	tbl := []struct {
		in       [][]string // [0]: element; [1:]: element's dependencies
		out      []string
		circular bool
		unknown  bool
	}{
		{
			in:  [][]string{},
			out: []string{},
		},
		{
			in:  [][]string{{"A"}},
			out: []string{"A"},
		},
		{
			in:  [][]string{{"A"}, {"B"}, {"C"}},
			out: []string{"A", "B", "C"},
		},
		{
			in:  [][]string{{"A"}, {"B", "A"}, {"C"}, {"D", "B", "A"}, {"E"}},
			out: []string{"A", "C", "E", "B", "D"},
		},
		{
			in:  [][]string{{"A"}, {"B", "E"}, {"C"}, {"D"}, {"E"}, {"F"}},
			out: []string{"A", "C", "D", "E", "F", "B"},
		},
		{
			in:  [][]string{{"A"}, {"B", "A"}, {"C", "A", "B"}, {"D"}},
			out: []string{"A", "D", "B", "C"},
		},
		{
			in:  [][]string{{"A"}, {"B", "A"}},
			out: []string{"A", "B"},
		},
		{
			in:  [][]string{{"A"}, {"B", "C"}, {"C", "A"}},
			out: []string{"A", "C", "B"},
		},
		{
			in:  [][]string{{"B", "A"}, {"A"}},
			out: []string{"A", "B"},
		},
		{
			in:       [][]string{{"B", "A"}, {"A", "B"}},
			circular: true,
		},
		{
			in:       [][]string{{"A", "B"}, {"B", "C"}, {"C", "D"}, {"D", "A"}},
			circular: true,
		},
		{
			in:      [][]string{{"A", "X"}, {"B"}, {"C"}},
			unknown: true,
		},
		{
			in:  [][]string{{"A", "0"}, {"B", "F", "C"}, {"C", "A"}, {"D"}, {"E"}, {"0"}, {"F"}, {"G", "D", "0"}, {"H"}, {"J", "C"}},
			out: []string{"D", "E", "0", "F", "H", "A", "G", "C", "B", "J"},
		},

		{
			in:       [][]string{{"A", "A"}},
			circular: true,
		},
		{
			in:  [][]string{{"A", "B"}, {"B", "C"}, {"C", "D"}, {"D", "E"}, {"E", "F"}, {"F"}},
			out: []string{"F", "E", "D", "C", "B", "A"},
		},
		{
			in:      [][]string{{"A", "X", "Y"}, {"B", "X", "Y"}, {"C"}},
			unknown: true,
		},
	}

	for _, test := range tbl {
		dg := NewDependencyGraph[string]()

		for _, in := range test.in {
			if len(in) == 0 {
				continue
			}

			dg.Add(in[0], in[1:]...)
		}

		res, err := dg.Resolve()
		if err != nil {
			if test.circular && errors.Is(err, ErrCircularDependency) {
				continue // Circular dependency is expected here
			}

			if test.unknown && errors.Is(err, ErrUnknownDependency) {
				continue
			}

			t.Fatalf("resolving graph: input = %v: %v", test.in, err)
		}

		if test.circular {
			t.Fatalf("resolved graph with circular dependency: input = %v", test.in)
		}

		if test.unknown {
			t.Fatalf("resolved graph with unknown dependency: input = %v", test.in)
		}

		if !slices.Equal(test.out, res) {
			t.Fatalf("graph resolved incorrectly: input = %v; output = %v; expected = %v", test.in, res, test.out)
		}
	}
}

// TestIter checks whether graph's resolution through an iterator (ResolveIter)
// works correctly.
// Additionally, it tests that the iterator allows us to exit early.
func TestIter(t *testing.T) {
	dg := NewDependencyGraph[string]()
	dg.Add("A")
	dg.Add("B", "A")
	dg.Add("C")
	dg.Add("D", "B", "A")
	dg.Add("E")

	res := []string{}
	for el, err := range dg.ResolveIter() {
		if err != nil {
			t.Fatalf("resolving graph iteratively: %v", err)
		}

		res = append(res, el)
		if el == "B" {
			break // Test that the iterator gets cancelled correctly.
		}
	}

	if !slices.Equal(res, []string{"A", "C", "E", "B"}) {
		t.Fatalf("iterative graph resolved incorrectly: %v", res)
	}
}

// TestIterCircularEarlyExit tests that a circular dependency gets unnoticed,
// if an iterator is cancelled early.
func TestIterCircularEarlyExit(t *testing.T) {
	dg := NewDependencyGraph[string]()
	dg.Add("A", "B")
	dg.Add("B")
	dg.Add("C")
	dg.Add("D")
	dg.Add("E", "F")
	dg.Add("F", "E")
	dg.Add("G")

	res := []string{}
	for el, err := range dg.ResolveIter() {
		if err != nil {
			t.Fatalf("resolving graph iteratively: %v", err)
		}

		res = append(res, el)
		if el == "A" {
			break // Intentionally stop at the 5th element (A would be 5th).
		}
	}

	if !slices.Equal(res, []string{"B", "C", "D", "G", "A"}) {
		t.Fatalf("iterative graph resolved incorrectly: %v", res)
	}
}

// TestConsecutiveResolve tests if consecutive graph resolutions work correctly,
// while adding elements in-between the resolutions.
func TestConsecutiveResolve(t *testing.T) {
	dg := NewDependencyGraph[string]()
	dg.Add("A")
	dg.Add("B", "A")
	dg.Add("C")
	dg.Add("D", "B", "A")
	dg.Add("E")

	res, err := dg.Resolve()
	if err != nil {
		t.Fatalf("resolving consecutive-resolve graph, 1st pass: %v", err)
	}

	if !slices.Equal(res, []string{"A", "C", "E", "B", "D"}) {
		t.Fatalf("consecutive-resolve graph resolved incorrectly on 1st pass: %v", res)
	}

	dg.Add("F", "D")
	dg.Add("G", "F")
	dg.Add("H")

	res, err = dg.Resolve()
	if err != nil {
		t.Fatalf("resolving consecutive-resolve graph, 2nd pass: %v", err)
	}

	if !slices.Equal(res, []string{"A", "C", "E", "H", "B", "D", "F", "G"}) {
		t.Fatalf("consecutive-resolve graph resolved incorrectly on 2nd pass: %v", res)
	}
}

// TestMultiAdd tests the graph resolution in the case when dependencies for the same item
// are added sequentially, by calling Add() each time.
func TestMultiAdd(t *testing.T) {
	dg := NewDependencyGraph[string]()
	dg.Add("A")
	dg.Add("B", "A")
	dg.Add("C")
	dg.Add("D")
	dg.Add("C", "A")
	dg.Add("C", "B")

	res, err := dg.Resolve()
	if err != nil {
		t.Fatalf("resolving multi-add graph: %v", err)
	}

	if !slices.Equal(res, []string{"A", "D", "B", "C"}) {
		t.Fatalf("multi-add graph resolved incorrectly: %v", res)
	}
}

// TestPointers tests that a graph gets resolved correctly if a pointer type is provided.
func TestPointers(t *testing.T) {
	type element struct {
		// We have to distinguish the elements somehow,
		// otherwise all of these pointers below will be backed
		// by a single allocation,
		// and this breaks the pointer comparison
		inner int
	}

	a := &element{inner: 1}
	b := &element{inner: 2}
	c := &element{inner: 3}
	d := &element{inner: 4}
	e := &element{inner: 5}

	dg := NewDependencyGraph[*element]()

	dg.Add(a)
	dg.Add(b, a)
	dg.Add(c)
	dg.Add(d, b, a)
	dg.Add(e)

	res, err := dg.Resolve()
	if err != nil {
		t.Fatalf("resolving graph with pointers: %v", err)
	}

	if !slices.Equal(res, []*element{a, c, e, b, d}) {
		t.Fatalf("pointer graph resolved incorrectly: %v", res)
	}
}
