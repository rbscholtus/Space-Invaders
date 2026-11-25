# Asset origins and licences

Every file in `assets/`, and every file shipped next to the app, is listed
here with where it came from and its licence (CONTRIBUTING, "Assets").

## Cabinet artwork (`assets/cabinet/`)

**Licence: none needed; believed to be in the public domain.** The masters
were AI-generated with Google Gemini. AI-generated images without
substantial human authorship lack copyright under US law (and in many other
jurisdictions), so these images are believed to be in the public domain and
are distributed here without copyright restrictions. The project's BSD
3-Clause licence covers its own code, not these images.

- **Google's terms** (as reviewed for Chippie in November 2026): Google does
  not claim ownership of content generated with Gemini, and allows its use
  for personal and commercial purposes, subject to Google's Prohibited Use
  Policy.
- **Original art only** (DESIGN §12): the prompts are our own words, from
  the style notes in DESIGN §12. No photo or scan of a real cabinet was given
  to the generator, and the prompts ask for no text, logos or creatures.
- **The masters** are kept outside the repository in `assets/source/`
  (git-ignored). Each was generated with Google Gemini image generation (the
  Gemini app) in November 2026, from the prompts in
  [docs/cabinet-artwork.md](../docs/cabinet-artwork.md):

  | Master | Prompt | Size | Used for |
  |---|---|---|---|
  | `cabinet03.png` | §1, the cabinet | 1536×2752 | the cabinet |
  | `backdrop02.png` | §2, the glass and its moon backdrop | 2048×2048 | the glass, and `backdrop.jpg` |
  | `marquee02.png` | §3, the marquee | 3168×1344 | the marquee |

- **What was done to them** (`cmd/cabinet-gen`, from
  `assets/cabinet/recipe.json`):
  - the cabinet cropped, and its edges faded into the background colour;
  - Gemini's visible sparkle mark painted out of the cabinet's background,
    and cropped off the glass and the marquee art (Google's invisible
    SynthID watermark remains in the pixels);
  - the glass art and the marquee art scaled into the cabinet's blank glass
    and marquee panels;
  - our lettering drawn on: "1978 INVASION" on the marquee, the button
    labels, the instruction plate and the price labels (below);
  - the pressed buttons made from the cabinet's own buttons, pushed in and
    darkened;
  - every image at three sizes (`@1x`, `@2x`, `@3x`), and `backdrop.jpg`,
    the part of the glass art behind the picture, at its own resolution.

**The lettering** is set in Go Bold and Go Medium (Bigelow & Holmes, for the
Go project), BSD 3-Clause style licence, from `golang.org/x/image/font/gofont`.
The fonts are used only by `cmd/cabinet-gen`; the app and the site ship no
font file.
