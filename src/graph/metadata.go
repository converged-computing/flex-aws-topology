package graph

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/converged-computing/flex-aws-topology/src/utils"
	"github.com/converged-computing/jsongraph-go/jsongraph/metadata"
	// TODO update back to flux-sched when merged
)

var (
	nodeType     = "node"
	instanceType = "instance"
	clusterType  = "cluster"
)

// Generate the containment path for a listing of parent nodes
func (t *TopologyGraph) assembleParentsPath(nodes []string) []string {
	parents := []string{}
	for _, parent := range nodes {
		parent_uid, _ := t.GetUniqueId(parent)
		parents = append(parents, nodeType+parent_uid)
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

// getInstanceMetadata starts with default metadata and adds on instance specific attributes
func (t *TopologyGraph) getInstanceMetadata(instance *ec2.InstanceTopology) metadata.Metadata {

	m := getDefaultMetadata(instanceType)
	uid, intuid := t.GetUniqueId(*instance.InstanceId)

	// The node network path is given to us
	// We need to unwrap (remove) pointers so list of strings
	nodes := utils.UnwrapPointers(instance.NetworkNodes)
	parents := t.assembleParentsPath(nodes)

	path := strings.Join(parents, "/")
	path = fmt.Sprintf("/%s/%s/%s%s", clusterType, path, instanceType, uid)

	m.AddElement("name", instanceType)
	m.AddElement("uniq_id", intuid)
	m.AddElement("id", intuid)
	m.AddElement("availability_zone", instance.AvailabilityZone)
	m.AddElement("instance_type", instance.InstanceType)
	m.AddElement("paths", map[string]string{"containment": path})
	m.AddElement("zone_id", instance.ZoneId)
	m.AddElement("group", instance.GroupName)
	return m
}

// getInstanceMetadata starts with default metadata and adds on instance specific attributes
// Parents should be most distant to most recent relative
// IMPORTANT: uid and intuid must be integers
func getNetworkNodeMetadata(uid string, intuid int32, parents []string) metadata.Metadata {

	m := getDefaultMetadata(nodeType)

	// The node network path is given to us - we assume no parents is a root
	// We need to unwrap (remove) pointers so list of strings
	path := strings.Join(parents, "/")
	path = fmt.Sprintf("/%s/%s/%s%s", clusterType, path, nodeType, uid)
	m.AddElement("uniq_id", intuid)
	m.AddElement("id", intuid)
	m.AddElement("paths", map[string]string{"containment": path})
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
