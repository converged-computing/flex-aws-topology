package graph

import (
	"fmt"
	"strings"

	"github.com/converged-computing/jsongraph-go/jsongraph/metadata"
	// TODO update back to flux-sched when merged
)

// Generate the containment path for a listing of parent nodes
func (t *TopologyGraph) assembleParentsPath(nodes []string) string {
	parents := []string{}
	for _, parent := range nodes {
		uid := t.GetUniqueId(parent)
		parents = append(parents, nodeType+uid.String())
	}
	// No parents, no path!
	return strings.Join(parents, "/")
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
func getDefaultMetadata(typ string, uid *UniqueId) metadata.Metadata {

	m := metadata.Metadata{}

	// These are required metadata fields
	// See https://github.com/flux-framework/flux-sched/blob/745e3e097fe1368e53fcf46b0a2c2cd7b95ad369/resource/readers/resource_reader_jgf.cpp#L383-L389

	// Unique id and id need to be integers
	m.AddElement("uniq_id", uid.Uid)
	m.AddElement("id", uid.Uid)

	// The basename should be the name minus the uid number (e.g., "tiny" is basename, "tiny0" is name)
	name := fmt.Sprintf("%s%d", typ, uid.Uid)

	// Metadata fields are mostly required (you'll get an error missing these)
	m.AddElement("type", typ)
	m.AddElement("basename", typ)
	m.AddElement("name", name)
	m.AddElement("rank", -1)
	m.AddElement("status", -1)
	m.AddElement("exclusive", false)
	m.AddElement("unit", "")
	m.AddElement("size", 1)
	return m
}
