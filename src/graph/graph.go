package graph

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/converged-computing/flex-aws-topology/src/utils"
	"github.com/converged-computing/jsongraph-go/jsongraph/v1/graph"

	// TODO update back to flux-sched when merged
	"github.com/researchapps/flux-sched/resource/reapi/bindings/go/src/fluxcli"

	"fmt"
)

type TopologyGraph struct {

	// Clients needed for aws and fluxion
	cli *fluxcli.ReapiClient
	ec2 *ec2.EC2

	// User preferences
	MatchPolicy string
	Region      string

	// These get reset between topology generations
	graph   *graph.JsonGraph
	counter int32

	// Lookup for unique id objects
	seen map[string]UniqueId

	// Lookup of nodes and edges for graph (to create at once)
	nodes map[string]*graph.Node
	edges map[string]*graph.Edge
}

// Reset the topology graph to a "zero" count and no nodes seen or created
func (t *TopologyGraph) Reset() {
	// Start counting at 1, the root is 0
	t.counter = 1
	t.seen = map[string]UniqueId{}
	t.nodes = map[string]*graph.Node{}
	t.edges = map[string]*graph.Edge{}

	// prepare a graph to load targets into
	t.graph = graph.NewGraph()

}

// AddNode adds a node to the graph
func (t *TopologyGraph) AddNode(node *graph.Node) {
	t.nodes[node.Id] = node
}

// CreateNodes creates all nodes at once
func (t *TopologyGraph) CreateNodes() {
	for _, node := range t.nodes {
		fmt.Printf("Creating node %s %s\n", node.Id, *node.Label)
		t.graph.Graph.Nodes = append(t.graph.Graph.Nodes, *node)
	}
}

// CreateEdges creates all edges at once
func (t *TopologyGraph) CreateEdges() {
	for _, edge := range t.edges {
		fmt.Printf("Creating edge (%s contains->%s) (%s in-> %s) \n", edge.Source, edge.Target, edge.Target, edge.Source)
		t.graph.Graph.Edges = append(t.graph.Graph.Edges, getBidirectionalEdges(edge.Source, edge.Target)...)
	}
}

// AddEdge adds a bidirectional edge to the graph
func (t *TopologyGraph) AddEdge(source string, dest string) {

	// Since we create bi-directional, always sort
	ids := []string{source, dest}
	sort.Strings(ids)
	key := strings.Join(ids, "-")

	_, ok := t.edges[key]
	if !ok {
		t.edges[key] = &graph.Edge{Source: source, Target: dest}
	}

}

// A NewTopologyGraph is associated with a region and match policy
func NewTopologyGraph(matchPolicy string, region string) *TopologyGraph {

	// Set default match policy
	if matchPolicy == "" {
		matchPolicy = "first"
	}

	// Alert the user to all the chosen parameters
	// Note that "grug" == "graphml" but probably nobody knows what grug means
	// We are using JGF for now because XML is slightly evil
	fmt.Printf(" Match policy: %s\n", matchPolicy)
	fmt.Println(" Load format: JSON Graph Format (JGF)")

	t := TopologyGraph{MatchPolicy: matchPolicy, Region: region}
	t.Reset()

	// instantiate fluxion
	t.cli = fluxcli.NewReapiClient()
	fmt.Printf("Created flex resource graph %s\n", t.cli)

	// Create a EC2 client from a session
	s := session.Must(session.NewSession())
	t.ec2 = ec2.New(s, aws.NewConfig().WithRegion(region))
	return &t
}

// A unique id can hold the id and return string and other derivates of it
type UniqueId struct {
	Uid  int32
	Name string
}

// String converts the int uid to a string
func (u *UniqueId) String() string {
	return fmt.Sprintf("%d", u.Uid)
}

// Get a unique id for a node (instance or network node)
// We need both int and string, so we return a struct
func (t *TopologyGraph) GetUniqueId(name string) *UniqueId {

	// Have we seen it before?
	uid, ok := t.seen[name]

	// Nope, create a node for it!
	if !ok {
		fmt.Printf("%s is not yet seen, adding with uid %d\n", name, t.counter)
		uid = UniqueId{Uid: t.counter, Name: name}
		t.seen[name] = uid
		t.counter += 1
	}
	return &uid
}

// Init a new TopologyGraph from a graphml filename
// Each instance in the topology result has a listing of network nodes like this:

// NetworkNodes: ["nn-ec17a935b39a06f41","nn-dd9ec3119ca6ea9dc","nn-a59759166e67e7c02"]

// This is to say that nn-ec17* is at the top, and the instance is connected directly
// to nn-a59. This means that two instances connected to that node are close together.
// The closer two instances are in the graph, overall, the closer. That is all of
// the information that we have!
func (t *TopologyGraph) Topology(group string, instance string, saveFile string) error {

	// Reset counter and ids
	t.Reset()

	// Generate empty params for topology input
	params := generateTopologyInput(group, instance)
	fmt.Printf("Topology Query Parameters:\n%s\n", params)

	// Get topology for instances
	// TODO see paginatined example here
	// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeInstanceTopology
	topology, err := t.ec2.DescribeInstanceTopology(params)
	if err != nil {
		return fmt.Errorf("describe instance topology: %s", err)
	}

	// Show the user the found topology
	fmt.Println(topology)

	// No instances found
	if len(topology.Instances) == 0 {
		return fmt.Errorf("No instances were found for this query.")
	}

	// Create the root node (cluster)
	t.AddNode(generateRoot())

	// Create a node for each instance and network nodes
	for _, instance := range topology.Instances {

		instance_node := t.NewInstanceNode(instance)

		// Unwrap the network nodes once
		nodes := utils.UnwrapPointers(instance.NetworkNodes)

		// There are edges between each of the network nodes
		for i, networkNode := range nodes {

			// Create the network node, providing parents up to it
			node := t.NewNetworkNode(networkNode, nodes, i)

			// If we are at the node 0, make edge to root
			if i == 0 {
				t.AddEdge(rootId, node.Id)
				continue
			}

			// If we are > 0, we can add an edge from parent to child and back (bidirectional)
			parent_uid := t.GetUniqueId(nodes[i-1])
			t.AddEdge(parent_uid.String(), node.Id)

			// If we are at the last entry in the list, make edges between the last one and our instance
			if i == len(nodes)-1 {
				t.AddEdge(node.Id, instance_node.Id)
			}
		}
	}

	// Create nodes and edges once
	t.CreateNodes()
	t.CreateEdges()

	// Init the context for fluxion
	return t.initFluxionContext(saveFile)

}

// initFluxionContext, and also save the graph to file if desired.
// If a saveFile is not provided, we save to temporary file (and clean up)
// I'm not sure why fluxion requires both the bytes and the filename path, it seems redundant.
func (t *TopologyGraph) initFluxionContext(saveFile string) error {

	// Serialize the struct to string
	conf, err := json.MarshalIndent(t.graph, "", "  ")
	if err != nil {
		return err
	}

	if saveFile == "" {
		jsonFile, err := os.CreateTemp("", "aws-topology-*.json")
		if err != nil {
			return fmt.Errorf("Error creating temporary json file: %s", err)
		}
		defer jsonFile.Close()
		defer os.Remove(jsonFile.Name())
		saveFile = jsonFile.Name()
	}

	// 1. Write to file!
	err = os.WriteFile(saveFile, conf, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error writing json to file: %s", err)
	}

	// 2. Create the context, the default format is JGF
	// 3. Remainder of defaults should work out of the box
	// Note that the options get passed as a json string to here:
	// https://github.com/flux-framework/flux-sched/blob/master/resource/reapi/bindings/c%2B%2B/reapi_cli_impl.hpp#L412
	opts := `{"matcher_policy": "%s", "load_file": "%s", "load_format": "jgf", "match_format": "jgf"}`
	p := fmt.Sprintf(opts, t.MatchPolicy, saveFile)

	// 4. Then pass in the JGF as a string of bytes
	err = t.cli.InitContext(string(conf), p)
	if err != nil {
		return fmt.Errorf("Error creating context: %s", err)
	}
	fmt.Printf("\n✨️ Init context complete!\n")
	return nil
}

/*
// This function isn't written yet because I don't know what a topology request is
// I think likely we want something more like closeness (distance) in the graph
// as opposed to resources with some attributes.
func (f *TopologyGraph) Match(specFile string) (instance.TopologymRequest, error) {
	fmt.Printf("  💻️  Request: %s\n", specFile)

	// Prepare the ice cream request!
	request := instance.TopologyRequest{}

	spec, err := os.ReadFile(specFile)
	if err != nil {
		return request, errors.New("Error reading jobspec")
	}

	// TODO this could be f.cli.Satisfies
	// Note that number originally was a jobid (it's now some number for the request)
	// Note that recipe was originally "allocated"
	_, match, _, _, number, err := f.cli.MatchAllocate(false, string(spec))
	if err != nil {
		return request, err
	}

	// Populate the ice cream request for the customer
	request.Spec = match
	request.Number = number
	return request, nil
}
*/
