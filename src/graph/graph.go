package graph

import (
	"encoding/json"
	"os"

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
	MatchPolicy string
	Region      string

	// These get reset between topology generations
	counter int32
	seen    map[string]int32

	cli *fluxcli.ReapiClient
	ec2 *ec2.EC2
}

// Start counting at 1, the root is 0
func (t *TopologyGraph) Reset() {
	t.counter = 1
	t.seen = map[string]int32{}
}

// A NewTopologyGraph is associated with a region and match policy
func NewTopologyGraph(matchPolicy string, region string) *TopologyGraph {

	// Set default match policy
	if matchPolicy == "" {
		matchPolicy = "first"
	}

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

// generateTopologyInput generates the parameters for the topology request
// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DescribeInstanceTopologyInput
func generateTopologyInput(group string, instance string) *ec2.DescribeInstanceTopologyInput {

	groups := []*string{}
	instances := []*string{}
	dryRun := false

	// For larger sets we might want NextToken (string) or Filters []*ec2Filter{}
	input := ec2.DescribeInstanceTopologyInput{
		DryRun: &dryRun,
	}

	// Don't add these empty if not provided, likely weird errors
	if group != "" {
		groups = append(groups, &group)
		input.GroupNames = groups
	}
	if instance != "" {
		instances = append(instances, &instance)
		input.InstanceIds = instances
	}
	return &input
}

// Get a unique id for a node (instance or network node)
// We need both int and string, so generate both here
func (t *TopologyGraph) GetUniqueId(name string) (string, int32) {

	// Have we seen it before?
	intuid, ok := t.seen[name]

	// Nope, create a node for it!
	if !ok {
		intuid := t.counter
		t.counter += 1
		t.seen[name] = intuid
	}
	uid := fmt.Sprintf("%d", intuid)
	return uid, intuid
}

// Init a new FlexGraph from a graphml filename
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

	// prepare a graph to load targets into
	g := graph.NewGraph()
	created := map[string]bool{}

	// Create the root node (cluster)
	root := generateRoot()
	g.Graph.Nodes = append(g.Graph.Nodes, root)

	// Create a node for each instance and network nodes
	for _, instance := range topology.Instances {

		// Generate metadata for the node
		m := t.getInstanceMetadata(instance)
		instance_uid, _ := t.GetUniqueId(*instance.InstanceId)

		// Each instance is a new node - we have uniqueness here (I think)
		node := graph.Node{
			Label:    instance.InstanceId,
			Id:       instance_uid,
			Metadata: m,
		}
		g.Graph.Nodes = append(g.Graph.Nodes, node)

		// Unwrap the network nodes once
		nodes := utils.UnwrapPointers(instance.NetworkNodes)

		// There are edges between each of the network nodes
		for i, nnode := range nodes {

			// Again, get the unique id for the network node
			// If it's newly created, also create the node
			uid, _ := t.GetUniqueId(nnode)
			_, ok := created[uid]
			if !ok {
				fmt.Printf("Creating node for %s\n", nnode)
				// Assemble parents - the unique ids we've already sen
				parents := t.assembleParentsPath(nodes[i:])
				node = t.createNetworkNode(nnode, parents)
				g.Graph.Nodes = append(g.Graph.Nodes, node)
				created[uid] = true
			}

			// If we are at the node 0, make edge to root
			if i == 0 {
				g.Graph.Edges = append(g.Graph.Edges, getBidirectionalEdges("0", uid)...)
				continue
			}

			// If we are > 0, we can add an edge from parent to child and back (bidirectional)
			parent_uid, _ := t.GetUniqueId(nodes[i-1])
			g.Graph.Edges = append(g.Graph.Edges, getBidirectionalEdges(parent_uid, uid)...)

			// If we are at the last entry in the list, make edges between the last one and our instance
			if i == len(nodes)-1 {
				g.Graph.Edges = append(g.Graph.Edges, getBidirectionalEdges(uid, instance_uid)...)
			}
		}
	}

	// Serialize the struct to string
	conf, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}

	if saveFile == "" {
		jsonFile, err := os.CreateTemp("", "aws-topology-*.json") // in Go version older than 1.17 you can use ioutil.TempFile
		if err != nil {
			fmt.Printf("Error creating temporary json file: %x", err)
			return err
		}
		defer jsonFile.Close()
		defer os.Remove(jsonFile.Name())
		saveFile = jsonFile.Name()
	}

	// Write to file!
	err = os.WriteFile(saveFile, conf, os.ModePerm)
	if err != nil {
		fmt.Printf("Error writing json to file: %x", err)
		return err
	}

	// Alert the user to all the chosen parameters
	// Note that "grug" == "graphml" but probably nobody knows what grug means
	fmt.Printf(" Match policy: %s\n", t.MatchPolicy)
	fmt.Println(" Load format: JSON Graph Format (JGF)")

	// 2. Create the context, the default format is JGF
	// 3. Remainder of defaults should work out of the box
	// Note that the options get passed as a json string to here:
	// https://github.com/flux-framework/flux-sched/blob/master/resource/reapi/bindings/c%2B%2B/reapi_cli_impl.hpp#L412
	opts := `{"matcher_policy": "%s", "load_file": "%s", "load_format": "jgf", "match_format": "jgf"}`
	p := fmt.Sprintf(opts, t.MatchPolicy, saveFile)

	// 4. Then pass in a jobspec... err, ice cream request :)
	err = t.cli.InitContext(string(conf), p)
	if err != nil {
		fmt.Printf("Error creating context: %s", err)
		return err
	}
	fmt.Printf("\n✨️ Init context complete!\n")
	return nil

}

/*
// Order is akin to doing a Satisfies, but right now it's a MatchAllocate
// The result of an order is a traversal of the graph that could satisfy the ice cream request
func (f *TopologyGraph) Match(specFile string) (instance.IceCreamRequest, error) {
	fmt.Printf("   🍦️ Request: %s\n", specFile)

	// Prepare the ice cream request!
	request := instance.IceCreamRequest{}

	spec, err := os.ReadFile(specFile)
	if err != nil {
		return request, errors.New("Error reading jobspec")
	}

	// TODO this could be f.cli.Satisfies
	// Note that number originally was a jobid (it's now a number for the ice cream in the shop)
	// Note that recipe was originally "allocated"
	_, recipe, _, _, number, err := f.cli.MatchAllocate(false, string(spec))
	if err != nil {
		return request, err
	}

	// Populate the ice cream request for the customer
	request.Recipe = recipe
	request.Number = number
	return request, nil
}
*/
