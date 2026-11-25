# Cabinet artwork: image brief

The prompts behind the cabinet, the glass with its moon backdrop and the
marquee, and how to generate them. How the art is cut up and used is in
[DESIGN.md §11–§12](DESIGN.md#12-cabinet-artwork); the style notes there
are the source of these prompts.

Written for Gemini but generator-neutral: plain-language prompts, with the
aspect ratio set in the tool and things to avoid listed in the prompt.

**The rules:**

- **Never give the generator a reference photo,** not even "for the style".
  The prompts are in our own words; the result must not be derived from
  Taito's or Midway's art (DESIGN §12).
- **No text in any image.** Generated letters come out garbled, and the
  title must be ours. `cmd/cabinet-gen` draws all the text: "INVASION 1978"
  on the marquee, LEFT, RIGHT, FIRE, 1 PLAYER and 2 PLAYERS on the control
  panel, and the price labels on the coin door.
- **No creatures, aliens, spaceships or figures.** They would be the game's
  or the real cabinet's; the mood comes from the moon and the stars.
- **The same light in every image:** "soft even lighting" and the same
  background, so the parts match.

## 1. The cabinet

Front-on, the whole cabinet, with the glass and the marquee left black:
`cmd/cabinet-gen` puts the glass art (§2) and the marquee art (§3) into
them. 9:16, at the highest resolution the tool gives (4K if it can): the
picture is only about a quarter of the cabinet's height.

```
front-facing straight-on product photograph of a late-1970s upright arcade
cabinet standing on the floor, seen from the eye height of a standing
player so the top of the control panel is clearly visible, the cabinet
centred, the whole cabinet in frame from the top of the marquee to the
floor. The upper front holds a large rectangular glass panel that is
completely black, with no reflection; above it a wide rectangular backlit
marquee panel that is completely black and blank. The cabinet sides are
deep royal blue with thin white horizontal pinstripes, seen as narrow edges
beside the front; the lower front of the cabinet is glossy white. Below
the glass, a wide, nearly flat brushed-silver control panel with a black
trim, divided by thin dark lines into three framed sections, with five round
white plastic push buttons. The left section, the widest: two buttons side
by side close together in its middle, set between painted
yellow-orange-red chevron arrows pointing outwards to the left and to the
right, like one long double-headed arrow, and below them a blank plain silver
plate for instructions. The middle section: one button under painted
yellow-orange-red chevrons pointing upwards. The right section: two buttons
stacked one above the other at its right side. Below the panel, in the
white lower front, a dark-grey metal coin door with two coin slots side by
side near its top, a small blank red label at the outer side of each slot,
two small coin-return levers below them, and a rectangular black speaker
grille to the left of the coin door. Plain dark grey studio background, soft even lighting, no logo, no brand name,
orthographic, high detail, photorealistic, 9:16 aspect ratio. Avoid: text,
letters, numbers, logos, brand names, creatures, aliens, spaceships,
figures, people, joystick, reflections on the glass, side view, tilt
```

**If the control panel is hidden,** add "the control panel is shown from
slightly above, like a shelf". **If the buttons are bunched together,**
generate the panel alone first (the same text from "a wide, nearly flat
brushed-silver control panel" to "at its right side", as "top-down
photograph of …"), to see whether the generator understands the layout.
**If the cabinet is too narrow in the frame,** add "the cabinet fills most
of the frame's width".

## 2. The glass and its moon backdrop

The art for the whole glass panel. The middle of it is behind the picture:
the game's lit pixels shine over it, and it shows where the screen is black
(DESIGN §11.3). The rest of the glass shows around the picture. 1:1.

```
flat front-on painting for the glass panel of a 1970s arcade game cabinet,
a deep midnight-blue starfield with many small white and pale-blue stars,
a few larger twinkling stars, and along the bottom third a yellow and ochre
cratered moon landscape with low rolling hills, the horizon slightly curved,
icy blue-white rocky peaks rising at the far left and right edges, the upper
two thirds mostly dark open sky so a game picture can sit over it, painted
illustration in a bold late-1970s airbrush style, saturated colours, soft
even lighting, no border, flat, no perspective, square 1:1 aspect ratio.
Avoid: text, letters, numbers, logos, creatures, aliens, spaceships,
planets with rings, figures, people, frame, border, glass reflections
```

**What matters most:** the dark sky in the middle, so the game stays easy to
read over it. A bright or busy middle is unusable, however good it looks.

## 3. The marquee

The backlit strip above the glass. `cmd/cabinet-gen` letters "INVASION 1978"
over it, so it needs a calm area for the words. 21:9 (the widest the tool
offers; the recipe crops it to the marquee's shape).

```
wide backlit translucent arcade marquee artwork from the late 1970s, a
deep-blue starfield glowing from behind, the curved edge of a large yellow
cratered moon rising at the right end, a few streaking white stars at the
left end, a broad calm dark band across the middle left empty for a title,
painted illustration in a bold late-1970s airbrush style, saturated colours,
soft even backlight, flat, front-on, no perspective, 21:9 aspect ratio.
Avoid: text, letters, numbers, logos, creatures, aliens, spaceships,
figures, people, frame, border
```

**The lettering is ours, drawn by `cmd/cabinet-gen`:** bold capitals in
white with a red outline (a proposal; not the original's 3D yellow and
orange title), in Go's fonts (BSD).

## Tips

- **Generate several, pick the most front-on.** Generators rarely give a
  perfectly straight view. Small skew can be fixed with a perspective
  correction in any image editor.
- **Check the buttons:** five, round, in three sections: left and right
  together inside the outward arrow, fire alone under the upward chevrons,
  1P and 2P stacked at the right. The pressed-button sprites are cut from
  the same image, so no separate button prompt is needed. The instruction
  plate stays blank; `cmd/cabinet-gen` writes our own short instructions on
  it.
- **Save the masters in `assets/source/`** (git-ignored, outside the
  repository's history) with the generator's own file names, and tell me
  which is which. The recipe (`assets/cabinet/recipe.json`) records the
  crops, the glass and marquee rectangles, the picture's place on the
  glass, and the controls.
- **Record the generator, model, date and prompt** for every image in
  `assets/LICENSES.md`. For Gemini this was checked for Chippie
  (Chippie's `assets/LICENSES.md`, "TV artwork"): Google claims no ownership
  of the images and allows their use, and AI-generated images are believed
  to be in the public domain.
