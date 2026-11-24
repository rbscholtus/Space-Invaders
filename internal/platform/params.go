package platform

// Params are the browser's start-up options, from the page's URL query
// (?volume=30&demo=fast). On the desktop the command-line flags do this and
// URLParams returns the zero Params.
type Params struct {
	// Volume is 0–100, or -1 when not given.
	Volume int
	// Demo is "fast" to run the CPU-test demo quickly (for the browser
	// tests), else "".
	Demo string
}

// DemoFast is Params.Demo's value for a fast demo.
const DemoFast = "fast"
