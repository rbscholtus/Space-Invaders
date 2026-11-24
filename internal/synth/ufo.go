package synth

// ufo is the UFO sound: an SN76477 (step 5 of M3 ports it).
type ufo struct{ on bool }

func newUFO(float64) *ufo { return &ufo{} }

func (u *ufo) enable(on bool) { u.on = on }

func (u *ufo) step() float64 { return 0 }
