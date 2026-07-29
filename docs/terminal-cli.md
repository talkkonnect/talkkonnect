# Using the Terminal CLI

When talKKonnect runs in a terminal it keeps a command prompt pinned to the bottom row while the log
scrolls above it. Everything the device can do is reachable from there: single keys for the common
actions, and typed commands for the rest. The same prompt is available remotely over SSH when
`<remotesshconsole enabled="true">` is set in `talkkonnect.xml` — the sample configs listen on port 9999,
so `ssh -p 9999 user@device`.

Type `?`, `help` or `menu` at the prompt for the full banner. In short:

| Input | What it does |
| --- | --- |
| `2` `3` | Channel up / down |
| `4` `5` `6` | Mute-unmute speaker, digital volume up / down |
| `7` `8` | Start / stop transmitting |
| `9` `0` | List online users, show uptime |
| `r` `v` `b` | Channel scan on/off, online radio on/off, playback/stop stream |
| `d` `h` `u` | Dump the XML config, run the config sanity checker, show the version |
| `a` `o` `g` | List API commands, ping the configured servers, show the GPS position |
| `c` / `clear` | Clear the screen and repaint the prompt |
| `q` / `quit` | Close the CLI, leaving talKKonnect running |
| `...` | Shut talKKonnect down |

Typed commands, with **Tab completion** throughout:

* `cfg keys` / `cfg list` / `cfg set <path> <value>` / `cfg save` / `cfg restart` — read and change any
  setting in the running configuration by its dotted path, e.g.
  `cfg set global.software.settings.loglevel debug`. `cfg save` writes it back to `talkkonnect.xml`
* `vt …` — voice targets, see [Voice Targets (Whisper and Shout)](./voice-targets.md)
* `mc …` — multicast output, see [Multicast to IP Speakers and SIP Phones](./multicast.md)
* any HTTP API action name, typed directly — for example `announcement main_announcement`,
  `voicetargetset 3` or `showuptime`. These go through the same allow-list as the HTTP API, so the action
  must be enabled under `<http>` in the XML

Set `TALKKONNECT_NO_BOTTOM_CLI=1` to turn the bottom prompt off when running talKKonnect under a
harness, in a pipeline, or anywhere you want plain unadorned log output. It is also skipped
automatically when there is no terminal, such as under `--daemon` or systemd.
