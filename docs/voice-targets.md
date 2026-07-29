# Voice Targets (Whisper and Shout)

A voice target sends your transmitted audio to specific users or a specific channel instead of the
channel you are joined to — Mumble calls this whispering (to users) and shouting (to a channel). It is
how one device pages a single zone rather than everybody.

Targets are numbered 1 to 31 and are defined per account in `talkkonnect.xml`:

```xml
<voicetargets>
    <id value="1">
        <users>
            <user>zoran-laptop</user>
        </users>
    </id>
    <id value="3">
        <channels>
            <channel>
                <name>Zone-One</name>
                <recursive>true</recursive>
                <links>true</links>
                <group></group>
            </channel>
        </channels>
    </id>
</voicetargets>
```

You do not have to edit the file by hand. The `vt` command manages targets from the CLI or the SSH
console:

| Command | What it does |
| --- | --- |
| `vt list` | Every configured target, which one is active, and any user or channel that is not on the server right now |
| `vt set <id>` | Activate a configured target |
| `vt clear` | Back to normal channel speech (the same as `voicetargetset 0`) |
| `vt next` / `vt prev` | Step through the configured targets |
| `vt whisper <user>` | Whisper to one online user, no config entry needed |
| `vt add <id> user <name> …` | Add users to a target and write it to `talkkonnect.xml` |
| `vt add <id> channel <name> [recursive=true] [links=true] [group=<name>]` | Add a channel to a target |

`vt add` edits only the `<voicetargets>` element, leaves the rest of the file — comments, formatting and
tag order — exactly as it was, keeps a `.bak` copy, and applies the change to the running configuration
immediately, so a new target is usable without a restart.

Targets can also be selected by GPIO preset buttons, a rotary encoder, the `voicetargetset` action over
the HTTP API and MQTT, and per-announcement with the `<voicetarget>` tag of a `<multimedia>` profile.
[Full details and every tag](./running-talkkonnect.md#voice-targets-and-the-vt-command)
