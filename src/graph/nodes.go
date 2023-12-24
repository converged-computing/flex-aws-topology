package graph

import (
	"fmt"
	"strings"

	"github.com/converged-computing/jsongraph-go/jsongraph/metadata"
	"github.com/converged-computing/jsongraph-go/jsongraph/v1/graph"
)

// NewNetworkNode creates a new network node
// We return a graph node, and a boolean to indicate created or not
func (t *TopologyGraph) NewNetworkNode(name string, nodes []string) *graph.Node {

	uid := t.GetUniqueId(name)
	node, ok := t.nodes[uid.String()]

	// But if we haven't seen it yet, add to graph
	if !ok {
		fmt.Printf("Creating network node for %s\n", name)

		// Assemble parents - the unique ids we've already seen
		parents := t.assembleParentsPath(nodes)
		m := t.getNetworkNodeMetadata(name, parents)

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
func (t *TopologyGraph) getNetworkNodeMetadata(name string, parents []string) metadata.Metadata {

	uid := t.GetUniqueId(name)
	m := getDefaultMetadata(nodeType)

	// The node network path is given to us - we assume no parents is a root
	// We need to unwrap (remove) pointers so list of strings
	path := strings.Join(parents, "/")
	path = fmt.Sprintf("/%s/%s/%s%s", clusterType, path, nodeType, uid)
	m.AddElement("uniq_id", uid.Uid)
	m.AddElement("id", uid.Uid)
	m.AddElement("paths", map[string]string{"containment": path})
	return m
}
