package sound

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// playerBuffer is Ebitengine's own buffer: short, for low latency, as in
// Chippie.
const playerBuffer = 50 * time.Millisecond

// Output plays a Stream through Ebitengine's audio. Browsers allow audio
// only after a user gesture, so Start is separate from creating the stream
// (DESIGN §14.4; the browser side comes in M4).
type Output struct {
	player *audio.Player
}

// Start creates the audio context and plays stream. Calling it again does
// nothing.
func (o *Output) Start(stream *Stream) error {
	if o.player != nil {
		return nil
	}

	p, err := audio.NewContext(SampleRate).NewPlayerF32(stream)
	if err != nil {
		return fmt.Errorf("starting the audio: %w", err)
	}

	p.SetBufferSize(playerBuffer)
	p.Play()
	o.player = p

	return nil
}
