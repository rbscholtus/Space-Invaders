package platform

// Params are the browser's start-up options, from the page's URL query
// (?volume=30). On the desktop the command-line flags do this and
// URLParams returns Params with nothing given.
type Params struct {
	// Volume is 0–100, or -1 when not given.
	Volume int
}
