package graph

import (
	"fmt"

	"github.com/converged-computing/jsongraph-go/jsongraph/metadata"
	"github.com/converged-computing/jsongraph-go/jsongraph/v1/graph"
)

var (
	nodeType = "node"
)

// NewNetworkNode creates a new network node
// We return a graph node, and a boolean to indicate created or not
// nodes should be the complete list of nodes, and idx the index of the network node in it
func (t *TopologyGraph) NewNetworkNode(name string, nodes []string, idx int) *graph.Node {

	uid := t.GetUniqueId(name)
	node, ok := t.nodes[uid.String()]

	// But if we haven't seen it yet, add to graph
	if !ok {
		fmt.Printf("Creating network node for %s\n", name)

		// Assemble parents - the unique ids we've already seen
		path := t.assembleParentsPath(nodes[:idx])
		m := t.getNetworkNodeMetadata(name, path)

		node = &graph.Node{
			Label:    &uid.Name,
			Id:       uid.String(),
			Metadata: m,
		}
		t.AddNode(node)
	}
	return node
}

// getInstanceMetadata starts with default metadata and adds on instance specific attributes
// Parents should be most distant to most recent relative
// IMPORTANT: uid and intuid must be integers
func (t *TopologyGraph) getNetworkNodeMetadata(name string, path string) metadata.Metadata {
	uid := t.GetUniqueId(name)
	m := getDefaultMetadata(nodeType, uid)
	if path == "" {
		path = fmt.Sprintf("/%s/%s%s", clusterPath, nodeType, uid)
	} else {
		path = fmt.Sprintf("/%s/%s/%s%s", clusterPath, path, nodeType, uid)
	}
	m.AddElement("paths", map[string]string{"containment": path})
	return m
}
