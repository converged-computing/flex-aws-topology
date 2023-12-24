package graph

import (
	"fmt"

	"github.com/converged-computing/jsongraph-go/jsongraph/v1/graph"
)

// Create a node for the network node
func (t *TopologyGraph) createNetworkNode(name string, parents []string) graph.Node {
	uid, intuid := t.GetUniqueId(name)
	m := getNetworkNodeMetadata(name, intuid, parents)

	// Each instance is a new node - we have uniqueness here (I think)
	node := graph.Node{
		Label:    &uid,
		Id:       uid,
		Metadata: m,
	}
	return node
}

// generateRoot generates an abstract root for the graph
func generateRoot() graph.Node {

	// Generate metadata for the node
	m := getDefaultMetadata(clusterType)
	m.AddElement("name", clusterType)
	m.AddElement("uniq_id", 0)
	m.AddElement("id", 0)
	m.AddElement("paths", map[string]string{"containment": fmt.Sprintf("/%s", clusterType)})

	return graph.Node{
		Label:    &clusterType,
		Id:       "0",
		Metadata: m,
	}

}
