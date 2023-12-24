package main

import (
	"flag"
	"fmt"

	"github.com/converged-computing/flex-aws-topology/src/graph"
)

func main() {
	fmt.Println("This is the flex aws topology prototype")
	matchPolicy := flag.String("policy", "first", "Match policy")
	region := flag.String("region", "us-east-2", "AWS region")
	instance := flag.String("instance", "", "instance ID to get topology for")
	group := flag.String("group", "", "placement group to get topology for")
	saveFile := flag.String("file", "", "save json graph to this file instead of temporary one")

	flag.Parse()

	// Create an ice cream graph, and match the spec to it.
	g := graph.NewTopologyGraph(*matchPolicy, *region)
	err := g.Topology(*group, *instance, *saveFile)
	if err != nil {
		fmt.Printf("Oh no! There was a problem generating the topology: %s", err)
		return
	}
	//match, err := g.Match(specFile)
	//if err != nil {
	//	fmt.Printf("Oh no! There was a problem with your machine matching: %x", err)
	//	return
	//}
	//match.Show()
}
