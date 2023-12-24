package graph

import (
	"fmt"

	"github.com/converged-computing/jsongraph-go/jsongraph/v1/graph"
)

// I don't know why the flux examples have the unique id / id in metadata as int
// and the ones on the outside string. I suspect because the node->id doesn't need
// to be a number (but haven't tested yet)
var (
	rootUid     = 0
	rootId      = "0"
	clusterType = "cluster"
	clusterPath = "cluster0"
)

// generateRoot generates an abstract root for the graph
func generateRoot() *graph.Node {

	// Generate metadata for the node
	m := getDefaultMetadata(clusterType, &UniqueId{Uid: int32(rootUid), Name: clusterType})

	// The cluster path
	m.AddElement("paths", map[string]string{"containment": fmt.Sprintf("/%s", clusterPath)})

	return &graph.Node{
		Label:    &clusterType,
		Id:       rootId,
		Metadata: m,
	}

}
