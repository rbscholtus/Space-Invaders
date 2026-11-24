package invaders

// SoundEvent is a change to a sound port (DESIGN §7.3, §14.3): Port is 3 or
// 5, Value the new byte, and Cycle the CPU cycle within the frame at which
// the OUT instruction began (at most 10 cycles, 5 µs, before the write;
// well under one audio sample), for sample-accurate playback.
type SoundEvent struct {
	Cycle int
	Port  byte
	Value byte
}

// maxSoundEvents is the room preallocated per frame. The game writes the
// sound ports a few times per frame; changes beyond this are still applied
// to the ports, and only the last ones' timing is merged into the final
// event.
const maxSoundEvents = 64

// SoundEvents returns the sound-port changes of the last frame, in order. The
// slice is reused by the next RunFrame.
func (b *Board) SoundEvents() []SoundEvent { return b.events }

// recordSound notes a write to a sound port, if it changes the port.
func (b *Board) recordSound(port byte, old *byte, v byte) {
	if *old == v {
		return
	}

	*old = v

	e := SoundEvent{Cycle: int(b.CPU.Cycles - b.frameStart), Port: port, Value: v}
	if len(b.events) == cap(b.events) {
		b.events[len(b.events)-1] = e // full: keep the latest state

		return
	}

	b.events = append(b.events, e)
}
