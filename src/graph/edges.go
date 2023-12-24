package graph

import (
	"github.com/converged-computing/jsongraph-go/jsongraph/v1/graph"
)

// Helper function to get a set of edges (bidirectional)
func getBidirectionalEdges(source string, dest string) []graph.Edge {
	return []graph.Edge{
		getEdge(source, dest, "contains"),
		getEdge(dest, source, "in"),
	}
}

// Get an edge with a specific containment (typically "contains" or "in")
func getEdge(source string, dest string, containment string) graph.Edge {
	m := getEdgeMetadata(containment)
	return graph.Edge{Source: source, Target: dest, Metadata: m}

}
