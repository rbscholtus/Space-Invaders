-- Plays one sound sequence from internal/synth/testdata/reference/sequences.txt
-- in MAME's Space Invaders, for reference recordings (DESIGN §14.5). Run by
-- scripts/mame-reference.sh with MAME's -autoboot_script and -wavwrite.
--
-- INVADERS_SEQUENCES: path of sequences.txt; INVADERS_SOUND: the sequence.

local spec = os.getenv("INVADERS_SEQUENCES")
local name = os.getenv("INVADERS_SOUND")

local writes, total, current = {}, nil, nil
for line in io.lines(spec) do
  line = line:gsub("#.*", "")
  local seq, frames = line:match("^%s*sound%s+(%S+)%s+(%d+)")
  if seq then
    current = seq
    if seq == name then total = tonumber(frames) end
  else
    local f, p, v = line:match("^%s*(%d+)%s+(%d+)%s+(%S+)")
    if f and current == name then
      writes[#writes + 1] = { frame = tonumber(f), port = tonumber(p), value = tonumber(v) }
    end
  end
end
assert(total, "no sequence named " .. tostring(name))

local io_space = manager.machine.devices[":maincpu"].spaces["io"]
local screen = manager.machine.screens[":screen"]

-- Kept in a global: the notifier stops when its subscription is collected.
sequence_subscription = emu.add_machine_frame_notifier(function()
  local f = screen:frame_number()
  for _, w in ipairs(writes) do
    if w.frame == f then io_space:write_u8(w.port, w.value) end
  end
  if f >= total then manager.machine:exit() end
end)
