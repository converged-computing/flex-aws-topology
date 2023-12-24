package archspec

import "fmt"

// I'm not sure what a topology request is yet.
type TopologyRequest struct {
	Number uint64
	Spec   string
}

func (i *TopologyRequest) Satisfied() bool {
	return i.Spec != ""
}

// Show the customer their final request
func (i *TopologyRequest) Show() {
	if i.Satisfied() {
		fmt.Printf("\n😍️ Your Topology Request was satisfied!\n")
		fmt.Printf("Number: %d\n", i.Number)
		fmt.Printf("Spec:\n%s", i.Spec)
	} else {
		fmt.Printf("\n😭️ We could not satisfy your request!\n")
	}
}
