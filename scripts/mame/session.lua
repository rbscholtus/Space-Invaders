-- Plays the golden session's script (internal/testrig/script.go) in MAME's
-- Space Invaders, for the session sound check (DESIGN §14.5). Run by
-- scripts/mame-reference.sh session, with -wavwrite.
--
-- MAME latches inputs once per frame, at the start of vertical blank (line
-- 224), and runs this callback there too: values set in callback N take
-- effect at line 224 of frame N + 1. So callback N sets the script's inputs
-- for frame N + 2, and the Go side reproduces it with
-- Board.SetInputsAt(224, Script(f + 1)) before frame f (checked: work and
-- video RAM then match MAME's at every frame).

local SCRIPT_FRAMES = 24000
local FIRST_COIN, COIN_GAP, COINS = 5400, 100, 3
local START1, PLAY_FROM, START2 = 5700, 6000, 11400
local PRESS, SWEEP, FIRE_EVERY, FIRE_FRAMES = 5, 45, 16, 2

local ports = manager.machine.ioport.ports
local coin = ports[":IN1"].fields["Coin 1"]
local start1 = ports[":IN1"].fields["1 Player Start"]
local start2 = ports[":IN1"].fields["2 Players Start"]
local fire = ports[":CONTP1"].fields["P1 Button 1"]
local left = ports[":CONTP1"].fields["P1 Left"]
local right = ports[":CONTP1"].fields["P1 Right"]
local screen = manager.machine.screens[":screen"]

local function held(from, f) return f >= from and f < from + PRESS end

-- script returns the inputs held at frame f, as 0 or 1.
local function script(f)
  local c = 0
  for i = 0, COINS - 1 do
    if held(FIRST_COIN + i * COIN_GAP, f) then c = 1 end
  end
  local l, r, b = 0, 0, 0
  if f >= PLAY_FROM then
    if (f // SWEEP) % 2 == 0 then r = 1 else l = 1 end
    if f % FIRE_EVERY < FIRE_FRAMES then b = 1 end
  end
  return c, held(START1, f) and 1 or 0, held(START2, f) and 1 or 0, b, l, r
end

session_subscription = emu.add_machine_frame_notifier(function()
  local f = screen:frame_number() + 2
  local c, s1, s2, b, l, r = script(f)
  coin:set_value(c)
  start1:set_value(s1)
  start2:set_value(s2)
  fire:set_value(b)
  left:set_value(l)
  right:set_value(r)
  if f >= SCRIPT_FRAMES then manager.machine:exit() end
end)
