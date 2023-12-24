package utils

// Unwrap a list of string pointers to just a list of string
// I'm looking at you, network nodes!
func UnwrapPointers(listing []*string) []string {
	unwrapped := []string{}
	for _, item := range listing {
		unwrapped = append(unwrapped, *item)
	}
	return unwrapped
}
