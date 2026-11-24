package platform

// Picked is the result of the file picker: the chosen files by name, or an
// error.
type Picked struct {
	Files map[string][]byte
	Err   error
}

// PickFiles opens the browser's file picker (several files allowed: a zip
// or the four ROMs) and delivers the result on the channel, which has room
// for it. Call it straight after a click or key press: browsers only open
// the picker during a user gesture. If the player cancels, nothing is sent.
// On the desktop it sends ErrUnsupported: drop the files or use -roms.
//
// func PickFiles() <-chan Picked and func CanPickFiles() bool are in
// pick_js.go and pick_other.go.
