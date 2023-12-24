package graph

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/converged-computing/flex-aws-topology/src/utils"
	"github.com/converged-computing/jsongraph-go/jsongraph/metadata"
	"github.com/converged-computing/jsongraph-go/jsongraph/v1/graph"
)

var (
	instanceType = "instance"
)

// NewInstanceNode creates a new instance node for the graph
func (t *TopologyGraph) NewInstanceNode(instance *ec2.InstanceTopology) *graph.Node {

	// Generate metadata for the node
	name := *instance.InstanceId
	uid := t.GetUniqueId(name)
	m := t.getInstanceMetadata(instance, uid)

	node, ok := t.nodes[uid.String()]

	// Create the node if we haven't seen it yet
	if !ok {
		fmt.Printf("Creating instance node for %s\n", name)
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
func (t *TopologyGraph) getInstanceMetadata(instance *ec2.InstanceTopology, uid *UniqueId) metadata.Metadata {

	m := getDefaultMetadata(instanceType)

	// The node network path is given to us
	// We need to unwrap (remove) pointers so list of strings
	nodes := utils.UnwrapPointers(instance.NetworkNodes)
	parents := t.assembleParentsPath(nodes)

	path := strings.Join(parents, "/")
	path = fmt.Sprintf("/%s/%s/%s%d", clusterType, path, instanceType, uid.Uid)

	m.AddElement("name", instanceType)
	m.AddElement("uniq_id", uid.Uid)
	m.AddElement("id", uid.Uid)
	m.AddElement("availability_zone", instance.AvailabilityZone)
	m.AddElement("instance_type", instance.InstanceType)
	m.AddElement("paths", map[string]string{"containment": path})
	m.AddElement("zone_id", instance.ZoneId)
	m.AddElement("group", instance.GroupName)
	return m
}
