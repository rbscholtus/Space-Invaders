package platform

// maxExerciserSize bounds the download: 8080EXM.COM is 4.5 KB.
const maxExerciserSize = 64 << 10

// ReadExerciser returns an 8080 exerciser for the CPU-test demo (DESIGN
// §9.4), such as "8080EXM.COM": in the browser fetched from exercisers/
// next to the page, on the desktop read from third_party/8080-exercisers/
// in a repository checkout. It blocks: call it from a goroutine.
//
// func ReadExerciser(ctx context.Context, name string) ([]byte, error) is
// in exerciser_js.go and exerciser_other.go.
