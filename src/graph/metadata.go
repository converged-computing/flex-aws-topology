package graph

import (
	"github.com/converged-computing/jsongraph-go/jsongraph/metadata"
	// TODO update back to flux-sched when merged
)

var (
	nodeType    = "node"
	clusterType = "cluster"
)

// Generate the containment path for a listing of parent nodes
func (t *TopologyGraph) assembleParentsPath(nodes []string) []string {
	parents := []string{}
	for _, parent := range nodes {
		uid := t.GetUniqueId(parent)
		parents = append(parents, nodeType+uid.String())
	}
	return parents
}

// getEdgeMetadata returns default edge metadata.
// We assume an "in" relationship of a node being in (a child of) a parent
func getEdgeMetadata(containment string) metadata.Metadata {
	m := metadata.Metadata{}
	nameKey := map[string]string{"containment": containment}
	m.AddElement("name", nameKey)
	return m
}

// getDefaultMetadata ensures required fields are added
func getDefaultMetadata(typ string) metadata.Metadata {

	m := metadata.Metadata{}

	// These are required metadata fields
	// See https://github.com/flux-framework/flux-sched/blob/745e3e097fe1368e53fcf46b0a2c2cd7b95ad369/resource/readers/resource_reader_jgf.cpp#L383-L389
	m.AddElement("type", typ)
	m.AddElement("basename", typ)
	m.AddElement("name", typ)
	m.AddElement("rank", -1)
	m.AddElement("status", -1)
	m.AddElement("exclusive", false)
	m.AddElement("unit", "")
	m.AddElement("size", 1)
	return m
}
