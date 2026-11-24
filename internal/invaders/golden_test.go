package invaders_test

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/rbscholtus/Space-Invaders/internal/invaders"
	"github.com/rbscholtus/Space-Invaders/internal/testrig"
)

var update = flag.Bool("update", false, "rewrite the golden hashes (and write PNGs to look at)")

// Golden frames (DESIGN §15.3): a scripted session with the real ROM, with
// a hash of the picture and of the sound events every goldenEvery frames.
// Only the hashes are committed; the pictures are Taito's. Local only:
// skipped without your own ROMs.
const (
	goldenFrames = testrig.ScriptFrames // 6 min 40 s of game time
	goldenEvery  = 30                   // a checkpoint every half second
	goldenPath   = "testdata/golden/session.txt"
	pngDir       = "testdata/golden/png" // written by -update; git-ignored
	failureDir   = "testdata/failures"   // written on a mismatch; git-ignored
	maxReported  = 5                     // mismatching checkpoints shown in detail
)

// Work RAM the session checks, so the script keeps covering every phase
// (addresses checked against the game's behaviour).
const (
	gameModeAddr = 0x20EF // 1 while a game runs
	ships1Addr   = 0x21FF // player 1's ships in reserve
	score1Addr   = 0x20F8 // player 1's score, BCD, low byte first
	score2Addr   = 0x20FC // player 2's score
)

// checkpoint is one line of the golden file.
type checkpoint struct {
	frame        int
	video, sound string // SHA-256, hex
}

func (c checkpoint) String() string { return fmt.Sprintf("%6d %s %s", c.frame, c.video, c.sound) }

// phases records what the session reached, to check the script's coverage.
type phases struct {
	games, gameOvers int
	shipsLost        bool
	score1, score2   int
}

func bcd(b *invaders.Board, addr uint16) int {
	lo, hi := b.Read(addr), b.Read(addr+1)

	return int(hi>>4)*1000 + int(hi&15)*100 + int(lo>>4)*10 + int(lo&15)
}

// runSession plays the script and returns the checkpoints and the pictures
// at them (one byte per pixel).
func runSession(t *testing.T) ([]checkpoint, [][]byte, phases) {
	t.Helper()

	b, err := invaders.New(testrig.ROMs(t), invaders.DIPs{})
	if err != nil {
		t.Fatal(err)
	}

	var (
		points   []checkpoint
		pictures [][]byte
		seen     phases
		inGame   bool
	)

	sound := sha256.New()
	ev := make([]byte, 6)

	for f := range goldenFrames {
		b.SetInputs(testrig.Script(f))
		b.RunFrame()

		for _, e := range b.SoundEvents() {
			binary.LittleEndian.PutUint32(ev, uint32(e.Cycle))
			ev[4], ev[5] = e.Port, e.Value
			sound.Write(ev)
		}

		running := b.Read(gameModeAddr) == 1
		if running && !inGame {
			seen.games++
		}

		if !running && inGame {
			seen.gameOvers++
		}

		inGame = running

		if running && b.Read(ships1Addr) == 0 {
			seen.shipsLost = true
		}

		seen.score1 = max(seen.score1, bcd(b, score1Addr))
		seen.score2 = max(seen.score2, bcd(b, score2Addr))

		if (f+1)%goldenEvery == 0 {
			video := sha256.Sum256(b.Video())
			points = append(points, checkpoint{f + 1, hex.EncodeToString(video[:]), hex.EncodeToString(sound.Sum(nil))})

			picture := make([]byte, invaders.ScreenWidth*invaders.ScreenHeight)
			invaders.Screen(b.Video(), picture)
			pictures = append(pictures, picture)

			sound.Reset()
		}
	}

	return points, pictures, seen
}

// TestGoldenSession plays the scripted session and compares every
// checkpoint with testdata/golden/session.txt. With -update, it rewrites
// the hashes and writes the pictures to testdata/golden/png/ to check by
// eye. Never weaken a golden to make this pass (CONTRIBUTING).
func TestGoldenSession(t *testing.T) {
	t.Parallel()

	points, pictures, seen := runSession(t)

	// The script must keep reaching every phase, or the hashes cover less
	// than they seem to.
	if seen.games != 2 || seen.gameOvers != 2 || !seen.shipsLost || seen.score1 == 0 || seen.score2 == 0 {
		t.Errorf("the session no longer covers every phase: %+v; want 2 games, 2 game overs, "+
			"player 1's ships lost, and both players scoring", seen)
	}

	if *update {
		writeGolden(t, points)
		writePNGs(t, pngDir, points, pictures, nil)

		return
	}

	want := readGolden(t)
	if len(want) != len(points) {
		t.Fatalf("%s has %d checkpoints, the session %d; run with -update after checking why",
			goldenPath, len(want), len(points))
	}

	var bad []int

	for i := range points {
		if points[i] != want[i] {
			bad = append(bad, i)
		}
	}

	if len(bad) == 0 {
		return
	}

	for _, i := range bad[:min(len(bad), maxReported)] {
		got, exp := points[i], want[i]
		t.Errorf("frame %d: picture %v, sound %v", got.frame, same(got.video, exp.video), same(got.sound, exp.sound))
	}

	writePNGs(t, failureDir, points, pictures, bad)
	t.Errorf("%d of %d checkpoints differ, the first at frame %d; the pictures are in %s/",
		len(bad), len(points), points[bad[0]].frame, failureDir)
}

func same(got, want string) string {
	if got == want {
		return "same"
	}

	return "DIFFERENT"
}

func readGolden(t *testing.T) []checkpoint {
	t.Helper()

	f, err := os.Open(goldenPath)
	if err != nil {
		t.Fatalf("%v; run with -update to create it", err)
	}
	defer f.Close()

	var points []checkpoint

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 3 {
			t.Fatalf("%s: bad line %q", goldenPath, line)
		}

		frame, err := strconv.Atoi(fields[0])
		if err != nil {
			t.Fatalf("%s: bad line %q", goldenPath, line)
		}

		points = append(points, checkpoint{frame, fields[1], fields[2]})
	}

	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	return points
}

func writeGolden(t *testing.T, points []checkpoint) {
	t.Helper()

	var sb strings.Builder

	sb.WriteString("# Golden hashes of the scripted session in golden_test.go (DESIGN §15.3).\n")
	sb.WriteString("# frame, SHA-256 of video RAM, SHA-256 of the sound events since the\n")
	sb.WriteString("# previous checkpoint. Only hashes: the pictures are Taito's.\n")
	sb.WriteString("# Regenerate with: go test ./internal/invaders -run TestGoldenSession -update\n")

	for _, p := range points {
		sb.WriteString(p.String())
		sb.WriteByte('\n')
	}

	if err := os.MkdirAll(filepath.Dir(goldenPath), 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(goldenPath, []byte(sb.String()), 0o600); err != nil {
		t.Fatal(err)
	}
}

// writePNGs writes the pictures at the checkpoints in only (all when only
// is nil) to dir, for a person to look at.
func writePNGs(t *testing.T, dir string, points []checkpoint, pictures [][]byte, only []int) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	write := func(i int) {
		img := image.NewGray(image.Rect(0, 0, invaders.ScreenWidth, invaders.ScreenHeight))
		for j, lit := range pictures[i] {
			img.Pix[j] = lit * 255
		}

		path := filepath.Join(dir, fmt.Sprintf("frame-%05d.png", points[i].frame))

		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}

		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}

		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}

	if only == nil {
		for i := range points {
			write(i)
		}

		return
	}

	for _, i := range only {
		write(i)
	}
}
