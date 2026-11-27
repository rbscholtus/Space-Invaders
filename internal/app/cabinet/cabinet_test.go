package cabinet

import (
	"encoding/json"
	"errors"
	"testing"
)

func testCabinet() Cabinet {
	return Cabinet{
		Name: "Test upright", Width: 1000, Height: 2000, Background: "#383838",
		Image: "cabinet", Backdrop: "backdrop.jpg",
		Screen:  Rect{X: 300, Y: 300, W: 400, H: 460},
		Marquee: Rect{X: 100, Y: 50, W: 800, H: 200},
		Upper:   Rect{W: 1000, H: 1250},
		Controls: []Control{
			{ID: Left, Kind: Button, Pressed: "left-pressed", CX: 200, CY: 1100, R: 30},
			{ID: Right, Kind: Button, Pressed: "right-pressed", CX: 280, CY: 1100, R: 30},
			{ID: Fire, Kind: Button, Pressed: "fire-pressed", CX: 500, CY: 1100, R: 30},
			{ID: Start1, Kind: Button, Pressed: "start1-pressed", CX: 800, CY: 1050, R: 30},
			{ID: Start2, Kind: Button, Pressed: "start2-pressed", CX: 800, CY: 1150, R: 30},
			{ID: Coin1, Kind: Slot, X: 400, Y: 1400, W: 40, H: 60},
			{ID: Coin2, Kind: Slot, X: 560, Y: 1400, W: 40, H: 60},
			{ID: Door, Kind: Lock, CX: 700, CY: 1500, R: 15},
		},
	}
}

func TestParseValid(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(testCabinet())
	if err != nil {
		t.Fatal(err)
	}

	c, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	if ctl, ok := c.Control(Fire); !ok || ctl.R != 30 {
		t.Errorf("Control(fire) = %+v, %v", ctl, ok)
	}

	if got := c.Colour(); got != [3]uint8{0x38, 0x38, 0x38} {
		t.Errorf("Colour = %x", got)
	}
}

func TestValidateRejects(t *testing.T) {
	t.Parallel()

	for name, change := range map[string]func(*Cabinet){
		"no size":            func(c *Cabinet) { c.Width = 0 },
		"no name":            func(c *Cabinet) { c.Name = "" },
		"no backdrop":        func(c *Cabinet) { c.Backdrop = "" },
		"bad colour":         func(c *Cabinet) { c.Background = "grey" },
		"screen outside":     func(c *Cabinet) { c.Screen.W = 2000 },
		"marquee outside":    func(c *Cabinet) { c.Marquee.Y = -1 },
		"control outside":    func(c *Cabinet) { c.Controls[0].CX = 990 },
		"over the screen":    func(c *Cabinet) { c.Controls[2].CY = 500 },
		"overlapping":        func(c *Cabinet) { c.Controls[1].CX = 240 },
		"duplicate id":       func(c *Cabinet) { c.Controls[1].ID = Left },
		"no radius":          func(c *Cabinet) { c.Controls[7].R = 0 },
		"slot without size":  func(c *Cabinet) { c.Controls[5].W = 0 },
		"unknown kind":       func(c *Cabinet) { c.Controls[7].Kind = "joystick" },
		"no pressed sprite":  func(c *Cabinet) { c.Controls[2].Pressed = "" },
		"no fire button":     func(c *Cabinet) { c.Controls[2].ID = "shoot" },
		"no lock":            func(c *Cabinet) { c.Controls = c.Controls[:7] },
		"no coin slot":       func(c *Cabinet) { c.Controls[5].ID = "coin3" },
		"upper outside":      func(c *Cabinet) { c.Upper.H = 2500 },
		"upper cuts screen":  func(c *Cabinet) { c.Upper.H = 700 },
		"button below upper": func(c *Cabinet) { c.Upper.H = 1160 },
	} {
		c := testCabinet()
		change(&c)

		if err := c.Validate(); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: Validate = %v, want ErrInvalid", name, err)
		}
	}

	c := testCabinet()
	c.Controls = append(c.Controls[:6], c.Controls[7]) // no second slot: fine

	if err := c.Validate(); err != nil {
		t.Errorf("one coin slot: %v", err)
	}

	if _, err := Parse([]byte("{")); !errors.Is(err, ErrInvalid) {
		t.Errorf("Parse(bad JSON) = %v", err)
	}
}

func TestScaleFor(t *testing.T) {
	t.Parallel()

	c := testCabinet() // 1000 wide at @3x: @1x is 333, @2x 667

	for _, tc := range []struct {
		width float64
		want  int
	}{{100, 1}, {333, 1}, {334, 2}, {666, 2}, {667, 3}, {5000, 3}} {
		if got := c.ScaleFor(tc.width); got != tc.want {
			t.Errorf("ScaleFor(%g) = %d, want %d", tc.width, got, tc.want)
		}
	}

	if got := FileName("cabinet", 2, "jpg"); got != "cabinet@2x.jpg" {
		t.Errorf("FileName = %q", got)
	}
}

func TestHitTest(t *testing.T) {
	t.Parallel()

	c := testCabinet()

	for _, tc := range []struct {
		name   string
		x, y   float64
		minR   float64
		want   string
		wantOK bool
	}{
		{"fire centre", 500, 1100, 0, Fire, true},
		{"fire edge", 529, 1100, 0, Fire, true},
		{"outside all", 500, 1800, 0, "", false},
		{"just outside the lock", 716, 1500, 0, "", false},
		{"grown lock", 725, 1500, 30, Door, true},
		{"overlap: nearest wins", 238, 1100, 45, Left, true},
		{"overlap: other side", 242, 1100, 45, Right, true},
		{"slot", 420, 1430, 0, Coin1, true},
		{"grown slot", 395, 1430, 30, Coin1, true},
		{"the screen is no control", 500, 500, 0, "", false},
	} {
		ctl, ok := c.HitTest(tc.x, tc.y, tc.minR)
		if ok != tc.wantOK || ctl.ID != tc.want {
			t.Errorf("%s: HitTest = %q, %v; want %q, %v", tc.name, ctl.ID, ok, tc.want, tc.wantOK)
		}
	}
}
