# talKKonnect HTTP API Specification

This document is the authoritative specification of the HTTP interfaces talKKonnect exposes, and a build
guide for the two kinds of external user interface people most often want:

1. A **framebuffer screen** — a full-screen status display drawn directly to a Linux framebuffer (or any
   local display surface) on the talkkonnect device itself, replacing or supplementing a small physical
   LCD/OLED.
2. A **web UI** — a browser dashboard that shows live state and sends control commands.

Both are built entirely on top of the endpoints described here. talKKonnect itself needs no code changes,
which means a UI can be written in any language, and can be generated almost end-to-end by an AI
programming agent. See [Building a UI with an AI programming agent](#building-a-ui-with-an-ai-programming-agent).

* [Overview of endpoints](#overview-of-endpoints)
* [Quick start](#quick-start)
* [Configuration reference](#configuration-reference)
* [Security model](#security-model)
* [Telemetry API — `GET /uistatus`](#telemetry-api--get-uistatus)
* [Control API — `GET /?command=…`](#control-api--get-command)
* [Config editor — `/config`](#config-editor--config)
* [Building a framebuffer screen client](#building-a-framebuffer-screen-client)
* [Building a web UI](#building-a-web-ui)
* [Building a UI with an AI programming agent](#building-a-ui-with-an-ai-programming-agent)

Specification version: talkkonnect **4.16.02**.

---

## Overview of endpoints

| Method | Path | Purpose | Content type | Served by |
| --- | --- | --- | --- | --- |
| `GET` | `/uistatus` | Read the full live state of the client as JSON | `application/json` | `<uistatus>` listener |
| `GET` | `/?command=<name>&…` | Execute one control command | `text/plain` | `<http>` listener |
| `GET` | `/config` | XML configuration editor (HTML page) | `text/html` | `<http>` listener |
| `POST` | `/config` | Save `talkkonnect.xml` and live-reload it | `text/html` | `<http>` listener |

The design is deliberately asymmetric and this shapes every client:

* **Telemetry is pull-only.** There is no WebSocket, no server-sent events, no long poll, and no webhook.
  A UI stays current by polling `/uistatus`.
* **Control is one command per request.** There is no batching and no request body — everything is a
  query string.
* **State is a full snapshot.** `/uistatus` always returns every field. There are no deltas, no ETags and
  no `If-Modified-Since` support, so a client can render each response standalone and never has to
  reconstruct state from a history of events.

Relevant source files: `uistatus.go` (telemetry), `httpapi.go` (control), `httpconfig.go` (config editor),
`remotecontrol_acl.go` (network ACL), `xmlparser.go` (configuration structs).

---

## Quick start

Enable both interfaces in `talkkonnect.xml` under `<global><software><remotecontrol>`:

```xml
<remotecontrol>
    <http enabled="true" listenport="8080">
        <!-- command list, see the Control API section -->
    </http>
    <uistatus enabled="true" listenip="0.0.0.0" listenport="8080" url="/uistatus"/>
</remotecontrol>
```

Restart talkkonnect, then:

```bash
# Read live state
curl -s http://127.0.0.1:8080/uistatus | jq .

# List the control commands this device will accept
curl -s "http://127.0.0.1:8080/?command=listapi"

# Transmit for two seconds
curl -s "http://127.0.0.1:8080/?command=starttransmitting"
sleep 2
curl -s "http://127.0.0.1:8080/?command=stoptransmitting"
```

> `listenip="0.0.0.0"` is shown above so the endpoint is reachable from another machine while you develop.
> The shipped default is `127.0.0.1`, which only serves clients on the talkkonnect device itself — the
> right choice for a framebuffer client. See [Security model](#security-model) before exposing either
> port on a network you do not control.

---

## Configuration reference

### `<http>` — the control API listener

```xml
<http enabled="true" listenport="8080">
    <command action="channelup" funcparamname="" message="Channel Up" enabled="true"/>
    <command action="joinchannel" funcparamname="value" message="Join Channel" enabled="true"/>
</http>
```

| Attribute | Meaning |
| --- | --- |
| `enabled` | Master switch for the control API **and** `/config`. |
| `listenport` | TCP port. Binds all interfaces. Defaults to `8080` when blank. |

Each `<command>` element declares one command as callable:

| Attribute | Meaning |
| --- | --- |
| `action` | The command name, matched case-insensitively against `?command=`. Must be one of the built-in names in the [command reference](#command-reference). |
| `funcparamname` | Empty for commands that take no argument. The literal string `value` for commands whose argument comes from the query string. Any other value is passed to the handler as a fixed argument (this is how `mute`, `unmute` and `mute-toggle` share one handler). |
| `message` | Human-readable label, echoed by `listapi` and in some acknowledgements. Use it as the button caption in a UI. |
| `enabled` | Currently honoured only when building the `listapi` listing. **Presence of the element is what makes a command callable** — remove the element, not just the flag, to deny a command. |

### `<uistatus>` — the telemetry listener

```xml
<uistatus enabled="true" listenip="127.0.0.1" listenport="8080" url="/uistatus"/>
```

| Attribute | Default | Meaning |
| --- | --- | --- |
| `enabled` | mirrors `<http enabled>` when the whole `<uistatus>` element is absent | Master switch for `/uistatus`. |
| `listenip` | `127.0.0.1` | Bind address. `0.0.0.0` or empty binds all interfaces. |
| `listenport` | `<http listenport>`, else `8080` | TCP port. |
| `url` | `/uistatus` | Path. A leading `/` is added if you omit it. |

Listener behaviour:

* When `listenport` equals the `<http>` port **and** `<http>` is enabled, `/uistatus` is registered on the
  same listener — one port serves everything.
* When the ports differ, or `<http>` is disabled, `/uistatus` gets its own listener on
  `listenip:listenport`.
* Set `enabled="false"` when the device uses a built-in OLED/LCD and you have no external screen client.

### `<networkacl>` — who may call in

```xml
<networkacl enabled="true">
    <network cidr="127.0.0.1/32"/>
    <network cidr="192.168.1.0/24"/>
</networkacl>
```

Applies to the control API, `/config`, `/uistatus`, and the SSH remote console. Invalid CIDRs are logged
and skipped. If `enabled="true"` but no CIDR parses, the ACL is **inactive and every address is allowed** —
a warning is logged, so check your log after editing.

---

## Security model

Be explicit with yourself about this before you build anything that faces a network:

* **There is no authentication.** No API key, no bearer token, no HTTP Basic, no session. Any client that
  can open a TCP connection and pass the network ACL has full control of the device, including PTT,
  channel changes, TTS into the channel, e-mail sending, and — through `/config` — arbitrary rewriting of
  `talkkonnect.xml`.
* **There is no TLS.** Everything is plaintext HTTP.
* **There is no CSRF protection**, and every control command is a plain `GET`. An HTML `<img>` tag on any
  page a browser on your LAN visits can trigger a command.
* **The only access control is the CIDR ACL** in `<networkacl>`.

Therefore:

* Keep `listenip="127.0.0.1"` for framebuffer clients that run on the device.
* For a remote web UI, do not expose port 8080 directly. Put it behind a reverse proxy (nginx, Caddy,
  Traefik) that terminates TLS and enforces authentication, and restrict `<networkacl>` to the proxy's
  address. This also solves CORS — see [Building a web UI](#building-a-web-ui).
* Treat `/config` as root-equivalent. If your deployment does not need remote configuration editing,
  block `/config` at the proxy.
* Reachable over a VPN or WireGuard tunnel is fine. Reachable from the public internet is not.

---

## Telemetry API — `GET /uistatus`

### Request

```
GET /uistatus HTTP/1.1
```

No parameters. No request body. Methods other than `GET` return `405`.

### Response

`200` with `Content-Type: application/json` and `Cache-Control: no-store`, body indented with two spaces.

| Status | Cause |
| --- | --- |
| `200` | Success. Always a complete snapshot. |
| `403` | Client address rejected by `<networkacl>`. |
| `405` | Method was not `GET`. |

A connection refused / timeout means talkkonnect is not running or `<uistatus>` is disabled. Treat it as
"device offline" and keep retrying — it is the normal state during a talkkonnect restart.

### Example

```json
{
  "connected": true,
  "transmitting": false,
  "serverName": "TalkkonnectCommunity",
  "server": "mumble.talkkonnect.com:64738",
  "channel": "HAM-CB",
  "usersOnline": 3,
  "channelUsers": [
    { "name": "suvir",   "status": "Speaking", "self": false },
    { "name": "tk-base", "status": "idle",     "self": true  },
    { "name": "visitor", "status": "Muted",    "self": false }
  ],
  "channelTree": [
    { "name": "Root",    "depth": 0, "userCount": 0, "active": false, "accessible": true },
    { "name": "HAM-CB",  "depth": 1, "userCount": 3, "active": true,  "accessible": true },
    { "name": "Private", "depth": 1, "userCount": 1, "active": false, "accessible": false }
  ],
  "receiving": true,
  "lastSpeaker": "suvir",
  "lastMessage": { "sender": "suvir", "text": "moving to channel 2" },
  "rxVolume": 74,
  "audioLevel": 62,
  "rxAudioLevel": 62,
  "txAudioLevel": 0,
  "muted": false,
  "voiceTarget": { "id": 0 },
  "internetRadio": {
    "enabled": true,
    "playing": false,
    "status": "stopped",
    "stationName": "",
    "stationIndex": 0,
    "stationCount": 4,
    "volume": 50
  },
  "ipAddress": "192.168.1.42",
  "bitrate": "",
  "uptimeSec": 18342,
  "scanning": false,
  "scanHold": false,
  "activity": "rx",
  "mumbleUsername": "tk-base",
  "version": "4.16.02"
}
```

### Field reference

#### Connection and identity

| Field | Type | Notes |
| --- | --- | --- |
| `connected` | bool | Connected to a Mumble server. When `false`, most other fields are empty or stale — drive your "offline" screen off this. |
| `serverName` | string | Friendly name of the current account/server entry from the config. |
| `server` | string | `host:port` actually in use. |
| `mumbleUsername` | string | Username talkkonnect is registered as. |
| `ipAddress` | string | First non-loopback IPv4 of the device. Empty if none. Useful on a headless screen so you know where to point a browser. |
| `version` | string | talkkonnect version, e.g. `4.16.02`. |
| `uptimeSec` | int | Seconds since the talkkonnect process started. Not since connection. |

#### Channel and users

| Field | Type | Notes |
| --- | --- | --- |
| `channel` | string | Name of the currently joined channel. |
| `usersOnline` | int | **Users in the current channel**, not on the whole server. Equals `len(channelUsers)`. |
| `channelUsers` | array | Users in the current channel. Pre-sorted: speakers first, then case-insensitive by name — render in the given order and the active speaker is always at the top. |
| `channelUsers[].name` | string | Mumble username. |
| `channelUsers[].status` | string | `"Speaking"`, `"Muted"` or `"idle"`. `Speaking` only appears while audio is actually being received. `Muted` covers server-muted, self-muted and suppressed. |
| `channelUsers[].self` | bool | True for this talkkonnect instance's own user. Highlight this row. |
| `channelTree` | array | Flattened, pre-ordered server channel tree. Render sequentially and indent by `depth`. |
| `channelTree[].depth` | int | 0 for root. |
| `channelTree[].userCount` | int | Users in that channel. |
| `channelTree[].active` | bool | True for the joined channel. |
| `channelTree[].accessible` | bool | False when this user lacks Enter permission — grey it out and do not offer `joinchannel` for it. |

When the channel tree has not been walked yet, `channelTree` degrades to a single entry describing the
current channel, or is absent. Do not assume a root node exists.

#### Audio activity

| Field | Type | Notes |
| --- | --- | --- |
| `transmitting` | bool | PTT is currently keyed. |
| `receiving` | bool | Audio is currently arriving. |
| `lastSpeaker` | string | Most recent speaker's name. Persists after they stop, so gate it on `receiving` if you only want the *current* speaker. |
| `activity` | string | Precomputed state for a single status badge: `"tx"`, `"rx"`, `"radio"`, `"idle"`, `"offline"`. Precedence is TX → RX → radio → idle → offline. Prefer this over deriving your own. |
| `muted` | bool | Output device muted at the OS mixer. |
| `rxVolume` | int | OS output volume, 0–100. |
| `rxAudioLevel` | int | Peak received PCM level, 0–100. |
| `txAudioLevel` | int | Peak transmitted PCM level, 0–100. |
| `audioLevel` | int | Convenience VU value: `txAudioLevel` while transmitting, `rxAudioLevel` while receiving, otherwise 0. |
| `bitrate` | string | `"stream"` while internet radio plays, otherwise empty. |

> **The three audio level fields are consumed by reading them.** They report the *peak since the previous
> `/uistatus` request* and are reset to zero on every read. Two consequences:
>
> * **Exactly one client should poll for VU purposes.** If a framebuffer screen and a web UI both poll,
>   each sees only the peaks that landed between its own request and the other's, so both VU meters read
>   low and flicker. Where two consumers are unavoidable, have one poll and fan the result out to the
>   other, or accept that only one gets a usable meter.
> * **Poll interval sets the meter's response.** 200–500 ms gives a meter that looks alive. Polling once
>   per second yields a one-second peak-hold, which looks sluggish.
>
> Every other field is a true instantaneous snapshot and is safe to read from any number of clients.

#### Voice targets, scanning, radio

| Field | Type | Notes |
| --- | --- | --- |
| `voiceTarget.id` | int | Mumble voice target ID. **`0` means no target** — audio goes to the joined channel normally. |
| `voiceTarget.kind` | string | `"user"` or `"channel"`. Omitted when `id` is 0. |
| `voiceTarget.names` | array | Who or what the target points at, so you can display *whom* you are whispering to. Omitted when empty. |
| `scanning` | bool | Channel scan is running. |
| `scanHold` | bool | Scan has paused on an active channel. Show a distinct indicator — this is the state where the operator needs to know why scanning stopped. |
| `internetRadio.enabled` | bool | Streaming radio is configured. |
| `internetRadio.playing` | bool | Currently playing. |
| `internetRadio.status` | string | Player state text. |
| `internetRadio.stationName` | string | Current station. |
| `internetRadio.stationIndex` | int | Index of the current station. |
| `internetRadio.stationCount` | int | Number of configured stations. |
| `internetRadio.volume` | int | Player volume. |

#### Text messages

| Field | Type | Notes |
| --- | --- | --- |
| `lastMessage.sender` | string | Sender of the most recent Mumble text message. Omitted when empty. |
| `lastMessage.text` | string | Message body. Omitted when empty. |

`lastMessage` is always present as an object; when there is no message it serialises as `{}`. Check for an
empty `text` rather than for the key's absence. Sender and text are sanitised for display but retain full
Unicode (Thai, CJK and similar render correctly) — pick a font with the coverage your users need.

Only the single most recent message is retained. talKKonnect keeps no message history, so a chat log in
your UI must be accumulated client-side by noticing when `lastMessage` changes. That is inherently lossy:
two messages arriving inside one poll interval means you only ever see the second.

### Polling guidance

| Consumer | Interval | Rationale |
| --- | --- | --- |
| Framebuffer screen with VU meter | 200–300 ms | Smooth meter; the device is local so cost is negligible. |
| Framebuffer screen without VU | 1 s | Text fields do not need faster. |
| Web dashboard, tab focused | 500 ms–1 s | Responsive without hammering an SBC. |
| Web dashboard, tab hidden | 5–10 s, or pause | Use the Page Visibility API. |

The handler builds each snapshot on demand, including an OS mixer query for volume and mute. On a
Raspberry Pi Zero, sub-100 ms polling is wasteful. Never poll from multiple browser tabs at full rate.

---

## Control API — `GET /?command=…`

### Request

```
GET /?command=<name>[&<param>=<value>…] HTTP/1.1
```

* The command name is matched case-insensitively and trimmed.
* Parameter **names** are case-insensitive; values are used as given.
* Any path works — the handler is registered at `/`, so `/api?command=channelup` behaves identically to
  `/?command=channelup`. A request to any unrecognised path on this port without a `command` parameter
  returns `400`, not `404`.
* The method is not checked, but use `GET`; nothing reads a request body.

A command executes only when **both** conditions hold:

1. It is one of the built-in handler names below, and
2. a `<command action="…">` element for it exists in `<http>` in `talkkonnect.xml`.

Condition 2 is how you scope what a device will accept. Several handlers listed below are **not** present
in the shipped sample configs and must be added before use: `setrxvolume`, `joinchannel`, `whisperuser`,
`whisperclear`, `radiotoggle`, `radionext`, `radioprev`, `radiovolup`, `radiovoldown`.

### Query parameters

| Parameter | Type | Used by | Notes |
| --- | --- | --- | --- |
| `command` | string | all | Required. |
| `channel` | string | `joinchannel` | Channel name. Required, non-empty. |
| `user` | string | `whisperuser` | Username. Required, non-empty. |
| `volume` | int | `setrxvolume` | Required, 0–100 inclusive. |
| `id` | int | `voicetargetset` | Voice target ID. Non-numeric input is a `400`. |
| `mediaid` | string | `announcement` | Media ID from the config's multimedia section. Required, non-empty. |
| `ttsmessage` | string | `ttsannouncement` | Text to speak. URL-encode it. |
| `ttslocalplay` | bool | `ttsannouncement` | `true`/`false`. Play on the local speaker. |
| `ttsplayintostream` | bool | `ttsannouncement` | `true`/`false`. Play into the Mumble channel. |
| `gpioenabled` | bool | `ttsannouncement` | `true`/`false`. Drive a GPIO while speaking. |
| `gpioname` | string | `ttsannouncement` | GPIO name from the config. |
| `predelay` | int | `ttsannouncement` | **Seconds** before speaking. |
| `postdelay` | int | `ttsannouncement` | **Seconds** after speaking. |
| `language` | string | `ttsannouncement` | TTS language code, e.g. `en`, `th`. |

Booleans accept only the exact lowercase-insensitive strings `true` and `false`; anything else leaves the
value at its zero (`false`). Numeric parameters that fail to parse produce `400` with a message naming the
offending parameter. Unknown parameters are ignored.

### Command reference

Commands with an empty **Params** column take no query parameters.

#### Transmit and mute

| Command | Params | Effect |
| --- | --- | --- |
| `starttransmitting` | | Key PTT. Stays keyed until `stoptransmitting`. |
| `stoptransmitting` | | Unkey PTT. |
| `repeattxloop` | | Run the configured repeat-TX loop. |
| `mute` | | Mute output. |
| `unmute` | | Unmute output. |
| `mute-toggle` | | Toggle mute. |

> `starttransmitting` latches. A UI that keys on mouse-down **must** send `stoptransmitting` on mouse-up,
> on `blur`, on `visibilitychange`, and on page unload, or the device transmits indefinitely. Also consider
> a client-side maximum key time as a safety net.

#### Volume

| Command | Params | Effect |
| --- | --- | --- |
| `volumerxup` / `volumerxdown` | | Step receive volume. |
| `volumetxup` / `volumetxdown` | | Step transmit volume. |
| `currentrxvolume` / `currenttxvolume` | | Report current volume in the response body. |
| `setrxvolume` | `volume` | Set receive volume to an absolute 0–100. |

#### Channels

| Command | Params | Effect |
| --- | --- | --- |
| `channelup` / `channeldown` | | Move to the next/previous channel. |
| `joinchannel` | `channel` | Join a channel by name. |
| `listserverchannels` | | List channels in the response body. |
| `scanchannels` | | Toggle channel scanning. |
| `listeningstart` / `listeningstop` | | Start/stop listening to configured extra channels. |

#### Whisper / voice targets

| Command | Params | Effect |
| --- | --- | --- |
| `whisperuser` | `user` | Set a whisper target to one user. |
| `whisperclear` | | Clear the whisper target. |
| `voicetargetset` | `id` | Set the Mumble voice target ID. `0` restores normal channel talk. |

#### Audio playback and announcements

| Command | Params | Effect |
| --- | --- | --- |
| `playback` | | Toggle playback of the configured audio file. |
| `announcement` | `mediaid` | Play a configured multimedia announcement. |
| `ttsannouncement` | `ttsmessage`, `ttslocalplay`, `ttsplayintostream`, `gpioenabled`, `gpioname`, `predelay`, `postdelay`, `language` | Speak text via TTS. |

#### Internet radio

| Command | Params | Effect |
| --- | --- | --- |
| `radiotoggle` | | Start/stop the stream. |
| `radionext` / `radioprev` | | Change station. |
| `radiovolup` / `radiovoldown` | | Change player volume. |

#### Servers and connection

| Command | Params | Effect |
| --- | --- | --- |
| `connnextserver` | | Connect to the next configured server. |
| `previousserver` | | Connect to the previous configured server. |
| `pingservers` | | Ping configured servers; results in the response body. |

#### Status, diagnostics, misc

| Command | Params | Effect |
| --- | --- | --- |
| `listapi` | | List every command this device accepts, with its configured `message`. |
| `listonlineusers` | | List online users in the response body. |
| `showuptime` | | Report uptime in the response body. |
| `showversion` | | Report the talkkonnect version. |
| `displaymenu` | | Print the console menu. |
| `clearscreen` | | Clear the local screen. |
| `dumpxmlconfig` | | Dump the running configuration. |
| `gpsposition` | | Report the GPS position. |
| `sendemail` | | Send the configured e-mail. |
| `panicsimulation` | | Trigger the panic function. **Sends alerts and e-mail** — never wire this to an unconfirmed button. |
| `thanks` | | Print the credits. |

### Responses

Plain text. On success the first line is an acknowledgement:

```
200 OK: http command channelup OK
```

Acknowledgement wording varies by command — `200 OK: http command joinchannel for channel HAM-CB`,
`200 OK: http command setrxvolume to 70%`, and so on. **Do not parse these strings.** Treat any `2xx` as
success and re-read `/uistatus` for the resulting state.

Commands that report information (`showuptime`, `listonlineusers`, `pingservers`, `currentrxvolume`,
`listserverchannels`, radio commands, …) append their human-readable output to the response body after
the acknowledgement line. This text is captured from the same reply channel the SSH console uses. It is
free-form, may be multi-line, is truncated beyond an internal cap, and its format is not stable across
releases. Display it verbatim in a log pane; never scrape it for values that `/uistatus` already exposes
as typed fields.

Because some commands emit their output from background goroutines, text may occasionally arrive too late
to be captured. A successful command with an empty informational body is normal, not an error.

| Status | Body | Cause |
| --- | --- | --- |
| `200` | acknowledgement, plus optional captured text | Command ran. |
| `400` | `400 bad request: …` | Missing `command`; a name that is not a built-in; a missing or malformed required parameter; `volume` outside 0–100. |
| `403` | `403 forbidden: …` | Rejected by `<networkacl>`. |
| `404` | `404 not found: API command "x" is not defined in configuration` | Built-in name, but no matching `<command>` element. **Fix the config, not the client.** |
| `500` | `500 internal server error: …` | Handler invoked with the wrong parameters — normally a `funcparamname` mismatch in the config. |

Note that `200` means *dispatched*, not *achieved*. `joinchannel` for a channel you lack permission to
enter still returns `200`. Confirm every state-changing command by polling `/uistatus` and checking the
field you expected to change; surface a timeout in the UI if it does not.

---

## Config editor — `/config`

`GET /config` renders the whole of `talkkonnect.xml` in a textarea. `POST /config` with form field
`xmlcontent` writes the file (mode `0644`), reloads it, and re-runs the configuration sanity check. The
page reports success, or a save-succeeded-but-reload-failed state — in which case the file on disk has
already been overwritten while the running configuration may be partly the old one.

It is an editor, not an API: HTML in, HTML out, no JSON, no schema validation beyond XML parsing, and no
authentication beyond the network ACL. Do not build a UI on top of it. If your UI needs to change
configuration, generate the XML, `POST` it, and then verify by re-reading `GET /config` — and gate that
capability behind your reverse proxy's authentication.

---

## Building a framebuffer screen client

A framebuffer client turns any HDMI monitor, SPI TFT or small LCD into a talkkonnect status display, with
far more room than the 4×20 character LCD the built-in driver targets.

### Architecture

```
┌──────────────────── talkkonnect device ────────────────────┐
│                                                            │
│  talkkonnect ──── HTTP :8080 /uistatus ───► screen client   │
│   (Go)          127.0.0.1, JSON snapshot     (any language) │
│                                                    │        │
│                                              /dev/fb0 or    │
│                                              SDL / DRM      │
└────────────────────────────────────────────────────────────┘
```

The client is a separate process on the same device. Keep `listenip="127.0.0.1"` — nothing needs to reach
this endpoint from off-box.

Set `<uistatus enabled="true">` and disable the built-in LCD/OLED driver in the hardware section of
`talkkonnect.xml` if you are replacing a physical display rather than supplementing it.

### Requirements

1. **Poll** `GET http://127.0.0.1:8080/uistatus` every 200–300 ms if drawing a VU meter, else every
   second.
2. **Render a full snapshot each frame.** Every response is complete; keep no state except what you need
   for animation and for detecting `lastMessage` changes.
3. **Never block the draw loop on the network.** Use a short connect and read timeout (1–2 s). On error,
   keep drawing the last good frame with a visible "stale"/"disconnected" marker plus the age in seconds.
   talkkonnect restarts are routine and the screen must not go black or freeze.
4. **Redraw only on change** where the display makes that cheap. On SPI panels, pushing an unchanged
   framebuffer wastes most of the bus bandwidth; compare against the previous snapshot and skip.
5. **Degrade by resolution.** Drive the layout off the actual framebuffer size rather than assuming one
   panel.

### Suggested layout

Priority order, for an operator glancing at the screen from a metre away:

| Priority | Element | Source fields |
| --- | --- | --- |
| 1 | Big activity indicator: colour-filled bar or badge — red TX, green RX, blue radio, grey idle, dark offline | `activity` |
| 2 | Current channel, large | `channel` |
| 3 | Who is talking | `lastSpeaker` gated on `receiving` |
| 4 | VU meter, 0–100 | `audioLevel` |
| 5 | User list with speaking/muted marks, own row highlighted | `channelUsers` |
| 6 | Whisper/voice-target banner when `voiceTarget.id != 0` | `voiceTarget` |
| 7 | Scan indicator, distinct when holding | `scanning`, `scanHold` |
| 8 | Radio station and volume when playing | `internetRadio` |
| 9 | Last text message | `lastMessage` |
| 10 | Status bar: server name, username, IP, RX volume, mute icon, uptime, version | `serverName`, `mumbleUsername`, `ipAddress`, `rxVolume`, `muted`, `uptimeSec`, `version` |

A 480×320 SPI panel comfortably fits items 1–5 plus the status bar. On 1920×1080 add the `channelTree`
in a side column, indenting by `depth` and greying entries where `accessible` is false.

Design notes that matter in the field:

* **Colour alone is not enough.** Use text or shape as well as colour for TX/RX so the state reads under
  bright sunlight and for colour-blind operators.
* **Reserve fixed space** for the speaker name and the message line. Text that appears and disappears
  reflows the layout and makes the screen feel unstable.
* **Choose a font with the coverage your users need** — `lastMessage` and usernames carry full Unicode.
* **Dim rather than blank** on idle if the panel has a backlight you control; a black screen is
  indistinguishable from a crashed client.

### Implementation options

| Approach | Good for | Notes |
| --- | --- | --- |
| Python + `pygame` on the `fbcon`/KMSDRM driver | Fastest to get working | Set `SDL_VIDEODRIVER=kmsdrm` (or `fbcon` on older stacks) and run without X. Very little code. |
| Python + `Pillow` writing `/dev/fb0` directly | No SDL dependency | Render to an RGB image, convert to the framebuffer's pixel format (query with `fbset`), write bytes. Must handle 16-bit RGB565 vs 32-bit yourself. |
| Go + `gioui` / `ebiten` | Single static binary, no runtime deps | Matches talkkonnect's own toolchain; easy to ship alongside it. |
| Chromium in kiosk mode against your own web UI | One codebase for screen and browser | Heavy for a Pi Zero; fine on a Pi 4. Lets you reuse the web UI below verbatim. |
| Rust + `embedded-graphics` | Tiny SPI panels driven over SPI directly | Best where there is no framebuffer device at all. |

For SPI panels wired to a controller such as ILI9341 or ST7789, either use `fbtft` to expose `/dev/fb1`
and treat it as a framebuffer, or drive the controller directly and skip the framebuffer entirely.

### Reference polling loop

```python
import time, requests

URL = "http://127.0.0.1:8080/uistatus"
POLL = 0.25
session = requests.Session()          # reuse the TCP connection
last, last_ok = None, 0.0

while True:
    try:
        st = session.get(URL, timeout=1.5).json()
        last, last_ok = st, time.monotonic()
    except Exception:
        pass                          # keep the previous frame, mark it stale

    if last is None:
        draw_waiting_screen()
    else:
        draw(last, stale_secs=time.monotonic() - last_ok)

    time.sleep(POLL)
```

Run it under systemd with `Restart=always` and no ordering dependency on talkkonnect — the client must
tolerate talkkonnect being down at any moment, including at boot.

---

## Building a web UI

A web UI adds control to the picture: poll `/uistatus` to render, `GET /?command=…` to act.

### CORS: read this first

**Neither endpoint sends any `Access-Control-Allow-Origin` header.** Browser JavaScript served from a
different origin than talkkonnect **cannot read** `/uistatus` and cannot read command responses. This is
the single most common thing that blocks a first attempt, and no amount of client-side code fixes it.

Three workable options:

1. **Reverse proxy, same origin (recommended).** Serve your static UI and proxy the talkkonnect paths
   under one origin. This also gets you TLS and authentication:

   ```nginx
   server {
       listen 443 ssl;
       server_name tk.example.com;

       location / {                     # your UI's static files
           root /var/www/tk-ui;
       }
       location /uistatus {
           proxy_pass http://127.0.0.1:8080/uistatus;
       }
       location /api {                  # control API
           auth_basic "talkkonnect";
           auth_basic_user_file /etc/nginx/.htpasswd;
           proxy_pass http://127.0.0.1:8080/;
       }
       location /config { deny all; }   # root-equivalent; block unless needed
   }
   ```

   With `<networkacl>` limited to `127.0.0.1/32`, the proxy becomes the only way in.

2. **A small backend of your own** (Node, Python, Go) that talks to talkkonnect server-side and exposes
   whatever API and auth you want. Best if you need to aggregate several talkkonnect devices, keep a
   message history, or add per-user permissions.

3. **Serve the UI from the talkkonnect device itself** on the same port — simplest for a single device on
   a trusted LAN, but you still have no authentication.

While developing, a dev-server proxy (`vite.config.js` `server.proxy`, `webpack-dev-server` `proxy`) gives
you option 1 without any infrastructure.

### Recommended structure

```
Browser
  ├── poll   GET /uistatus         every 500 ms–1 s (paused when the tab is hidden)
  ├── act    GET /api?command=…    on user interaction
  └── verify re-poll after each command, confirm the field changed
```

Concrete guidance:

* **Single source of truth.** Hold one state object from the latest `/uistatus` and render from it. Do not
  optimistically mutate local state on a click — show a pending indicator on the control instead, and let
  the next poll confirm. Anything else drifts from the device, which is also being driven by GPIO buttons,
  a USB keyboard, MQTT and the SSH console.
* **Build the control surface from `listapi`.** Call `?command=listapi` at startup and enable only the
  commands it reports, using the configured `message` as each label. A device whose config omits
  `joinchannel` should not show a channel-join control.
* **PTT safety.** Bind `starttransmitting` to pointer-down and `stoptransmitting` to pointer-up,
  `pointercancel`, `blur`, `visibilitychange` and `pagehide`. Add a hard client-side maximum key time
  (30–60 s). A browser tab that closes mid-transmission must not leave the device keyed.
* **Confirm destructive commands.** `panicsimulation` sends real alerts and e-mail; `sendemail` sends
  mail; `previousserver`/`connnextserver` drop the current connection. Require an explicit confirmation
  for each.
* **Handle the states, not just the happy path.** Fetch failure (device down or restarting), `403`
  (ACL), `404` (command not in config — say "not enabled on this device", not "error"), `500`
  (config mismatch). Show a clear "disconnected, retrying" banner and keep polling with backoff.
* **Back off when hidden.** Use the Page Visibility API to slow or pause polling. Several open tabs each
  polling an SBC at full rate is a real load, and remember that each poll consumes the audio-level peaks
  (see the note in the [telemetry section](#audio-activity)) — expect VU meters to misbehave with more
  than one poller.
* **Log pane.** Render the informational text from command responses verbatim in a scrolling pane. It is
  the most useful debugging aid you can give an operator, and it costs nothing to add.

### Minimal working example

```html
<!doctype html>
<meta charset="utf-8">
<title>talkkonnect</title>
<div id="activity">…</div>
<div id="channel"></div>
<div id="speaker"></div>
<progress id="vu" max="100" value="0"></progress>
<ul id="users"></ul>
<button id="ptt">PTT</button>
<pre id="log"></pre>

<script>
const BASE = "";                       // same origin, via the reverse proxy above

async function poll() {
  try {
    const st = await (await fetch(`${BASE}/uistatus`, {cache: "no-store"})).json();
    activity.textContent = st.activity.toUpperCase();
    channel.textContent  = st.channel || "—";
    speaker.textContent  = st.receiving ? (st.lastSpeaker || "") : "";
    vu.value             = st.audioLevel;
    users.replaceChildren(...(st.channelUsers || []).map(u => {
      const li = document.createElement("li");
      li.textContent = `${u.name} — ${u.status}${u.self ? " (me)" : ""}`;
      return li;
    }));
  } catch (e) {
    activity.textContent = "OFFLINE";
  }
}

async function cmd(name, params = {}) {
  const qs = new URLSearchParams({command: name, ...params});
  const res  = await fetch(`${BASE}/api?${qs}`);
  const text = await res.text();
  log.textContent = text + log.textContent;   // newest first
  poll();                                     // confirm from the device
  return res.ok;
}

// Latching PTT: every path out of "keyed" must unkey.
let keyed = false;
const key   = () => { if (!keyed) { keyed = true;  cmd("starttransmitting"); } };
const unkey = () => { if (keyed)  { keyed = false; cmd("stoptransmitting");  } };
ptt.addEventListener("pointerdown", key);
for (const ev of ["pointerup", "pointercancel", "pointerleave"]) ptt.addEventListener(ev, unkey);
for (const ev of ["blur", "pagehide"]) addEventListener(ev, unkey);
addEventListener("visibilitychange", () => { if (document.hidden) unkey(); });

poll();
setInterval(() => { if (!document.hidden) poll(); }, 700);
</script>
```

That is a complete, working dashboard in under 60 lines. Everything beyond it — channel tree navigation,
radio controls, voice targets, a chat pane, charts — is more of the same fields and commands.

---

## Building a UI with an AI programming agent

Both clients above are well suited to being written by an AI coding agent (Claude Code, or any agent that
can read a repository and run commands). The endpoints are small and fully specified, the data model is
one flat JSON object, and the result is easy to check by eye. What makes the difference between a working
result and a plausible-looking one is how much of the specification you hand over, and how you verify.

### Give the agent the specification, not a description of it

The single most effective step: point the agent at **this file** rather than paraphrasing it. Everything
it needs — field names, types, status codes, the latching-PTT hazard, the CORS constraint, the
peak-reset behaviour of the audio levels — is here. An agent working from a summary will invent field
names that look right and are not.

If the agent can read the repository, `uistatus.go` and `httpapi.go` are the ground truth and are short
enough to read in full.

**Capture a real snapshot first** and give the agent that too. It removes all guesswork about shape:

```bash
curl -s http://127.0.0.1:8080/uistatus > uistatus-sample.json
curl -s "http://127.0.0.1:8080/?command=listapi" > listapi-sample.txt
```

Take the snapshot while the device is in an interesting state — connected, someone talking, a couple of
users in the channel. A snapshot captured while disconnected has most fields empty and leads the agent to
build for the empty case.

### Starting prompt — framebuffer screen

> Read `docs/api.md` in this repository, in particular the "Telemetry API" and "Building a framebuffer
> screen client" sections, and `uistatus.go` for the exact JSON structure. I have also attached
> `uistatus-sample.json`, a real response from my device.
>
> Build a full-screen status display for a talkkonnect device in Python using pygame on the KMSDRM
> driver, so it runs without X on a Raspberry Pi. Target panel: 480×320.
>
> Requirements:
> - Poll `http://127.0.0.1:8080/uistatus` every 250 ms in a background thread with a 1.5 s timeout; the
>   draw loop must never block on the network.
> - If polling fails, keep drawing the last good snapshot with a visible "STALE 12s" marker. Never blank
>   or freeze the screen — talkkonnect restarts are normal.
> - Layout, in priority order: large activity badge driven by the `activity` field (red tx, green rx,
>   blue radio, grey idle, dark offline, with the state also written as text so it does not rely on
>   colour); channel name large; current speaker (`lastSpeaker`, only while `receiving`); a 0–100 VU
>   meter from `audioLevel`; the `channelUsers` list in the order given, marking speaking and muted and
>   highlighting the row where `self` is true; a status bar with `serverName`, `mumbleUsername`,
>   `ipAddress`, `rxVolume`, a mute icon, `uptimeSec` formatted as h:mm, and `version`.
> - Show a whisper banner whenever `voiceTarget.id` is non-zero, naming `voiceTarget.names`.
> - Show a scan indicator, visually distinct when `scanHold` is true.
> - Reserve fixed space for the speaker line and the status bar so the layout never reflows.
> - Read the framebuffer size at startup and scale the layout; do not hard-code 480×320 in the drawing
>   code.
> - Use a font with Thai and CJK coverage — usernames and messages carry full Unicode.
> - Single file, standard library plus pygame and requests only. Include a systemd unit with
>   `Restart=always` and no ordering dependency on talkkonnect.
>
> Then write a small mock server that serves canned `/uistatus` responses cycling through offline,
> idle, receiving, transmitting, whispering and scan-hold states, so I can check every visual state
> without a live device.

### Starting prompt — web UI

> Read `docs/api.md` in this repository — the "Telemetry API", "Control API" and "Building a web UI"
> sections — plus `uistatus.go` and `httpapi.go` for exact behaviour. I have attached
> `uistatus-sample.json` and `listapi-sample.txt` from my device.
>
> Build a single-page dashboard for one talkkonnect device.
>
> Constraints from the specification you must respect:
> - talkkonnect sends no CORS headers, so the UI and the API must share an origin. Set up a dev-server
>   proxy for development and give me the nginx config for production.
> - `starttransmitting` latches. Send `stoptransmitting` on pointer-up, pointercancel, pointerleave,
>   window blur, visibilitychange-to-hidden and pagehide, and enforce a 60 s maximum key time
>   client-side.
> - There is no push channel. Poll `/uistatus` every 700 ms while the tab is visible, every 10 s when
>   hidden.
> - `/uistatus` is the only source of truth. Never optimistically update local state after a command;
>   show the control as pending and let the next poll confirm.
> - A `200` from a command means it was dispatched, not that it worked. After each state-changing
>   command, poll and check the field you expected to change; if it has not changed within 3 seconds,
>   surface that in the UI.
> - Call `?command=listapi` at startup and only render controls for the commands it reports, using each
>   configured `message` as the label.
> - Require an explicit confirmation dialog for `panicsimulation`, `sendemail`, `previousserver` and
>   `connnextserver`.
> - Distinguish the failure modes for the user: fetch failure = "device offline, retrying"; `403` =
>   "blocked by network ACL"; `404` = "not enabled on this device"; `500` = "configuration mismatch".
>
> Panels: connection and identity; activity with a VU meter; channel users; channel tree from
> `channelTree` indented by `depth` with inaccessible channels greyed and non-clickable; PTT and mute;
> volume; internet radio; voice target; last message; and a log pane showing command response text
> verbatim, newest first.
>
> Plain HTML, CSS and JavaScript with no build step, or React with Vite — your call, but say which and
> why before you start. Mobile-friendly: the PTT button must be reachable with a thumb.

### Iterating well

* **Have the agent build a mock server early.** Canned `/uistatus` responses covering offline, idle, RX,
  TX, whisper, scan-hold and radio-playing let both agent and you exercise every visual state without a
  live device, and without transmitting on a real channel while testing.
* **Test control commands against a mock or a private Mumble channel first.** An agent iterating on a PTT
  button against your production channel will key the transmitter repeatedly.
* **For a framebuffer client, ask for screenshots.** Have the agent render to an offscreen surface and
  save PNGs; if it can view images, it can then critique and fix its own layout. This closes the loop far
  faster than describing what looks wrong.
* **Ask it to verify against the source, not against its own memory.** "Check every field name you used
  against the `UIStatus` struct in `uistatus.go`" reliably catches invented fields such as `isMuted`,
  `channel_name` or `users`.
* **Feed real errors back verbatim.** A pasted `404 not found: API command "joinchannel" is not defined
  in configuration` is enough for the agent to tell you to add the missing `<command>` element — which is
  the correct fix, in the config, not in the client.

### Review checklist before you trust it

Whatever generated the code, check these — they are the mistakes that actually show up:

- [ ] Every JSON field name matches the `UIStatus` struct in `uistatus.go` exactly, including case.
- [ ] `voiceTarget.id == 0` is treated as "no target", not as target zero.
- [ ] `usersOnline` is labelled as users in the *channel*, not on the server.
- [ ] `lastSpeaker` is gated on `receiving` wherever it is presented as who is talking *now*.
- [ ] `lastMessage` handles the `{}` case without throwing.
- [ ] `channelTree` handles being absent, and a single-entry tree with no root node.
- [ ] `accessible: false` channels are not offered as join targets.
- [ ] Poll failures leave the last frame visible with a staleness indicator; nothing blanks or freezes.
- [ ] Network timeouts are set. An unbounded fetch on a hung connection hangs the UI.
- [ ] `stoptransmitting` fires on *every* path out of the keyed state, including tab close and page
      navigation. Test by keying PTT and closing the tab, then check `transmitting` in `/uistatus`.
- [ ] A maximum key time exists as a safety net.
- [ ] Destructive commands (`panicsimulation`, `sendemail`, server switching) are confirmed.
- [ ] `404` responses are presented as "not enabled on this device", with the config fix, not as a bug.
- [ ] Nothing parses acknowledgement strings for state; state comes from `/uistatus`.
- [ ] Polling backs off when the page is hidden.
- [ ] If two clients poll the same device, the audio-level behaviour is understood and accepted.
- [ ] No credentials, tokens or `Authorization` headers were invented — talkkonnect has no
      authentication, and code that pretends otherwise is misleading about the deployment's security.

---

## Related documentation

* [Configuring and Running talKKonnect](./running-talkkonnect.md)
* [Functionality and Configurability](./functionality.md) — including the other remote control interfaces
  (MQTT, SSH console, GPIO, USB keyboard)
* [Getting Started](./getting-started.md)

Questions, corrections and contributed UIs are welcome — <suvir@talkkonnect.com>, or open an issue on
[GitHub](https://github.com/talkkonnect).
