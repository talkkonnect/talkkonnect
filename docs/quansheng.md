# Quansheng UV-K5 radio gateway

talkkonnect can control a Quansheng UV-K5 (and clones) and use it as an RF
gateway. The radio must run nicsure's
[quansheng-dock-fw](https://github.com/nicsure/quansheng-dock-fw). talkkonnect
does not talk to the serial port itself. The
[qdock](https://github.com/talkkonnect/quansheng) daemon owns the programming
cable and serves a REST API, and talkkonnect calls that API.

```
Mumble  <->  talkkonnect  --HTTP-->  qdock serve  --USB serial-->  UV-K5
                 audio  <------------- sound card / AIOC ------------>
```

## Setup

1. Flash the dock firmware to the radio and connect the programming cable.
2. Run `qdock serve`. It listens on `127.0.0.1:8099` by default. Add `-token`
   if qdock listens on a network address.
3. Fill in the `<quansheng>` section in `talkkonnect.xml`. It sits under
   `<hardware>`, next to `<radio>`.
4. For voice in either direction you also need an audio path between the
   radio and talkkonnect's sound card, for example an
   [AIOC](https://github.com/skuep/AIOC) cable. The plain programming cable
   carries only data.

## Configuration

```xml
<quansheng enabled="true">
  <url>http://127.0.0.1:8099</url>
  <token></token>
  <timeoutmsecs>3000</timeoutmsecs>
  <startupkeys></startupkeys>
  <mumbletoradio enabled="false">
    <ptttimeoutsecs>30</ptttimeoutsecs>
    <hangtimemsecs>300</hangtimemsecs>
    <maxtxsecs>180</maxtxsecs>
  </mumbletoradio>
  <radiotomumble enabled="false">
    <hangtimemsecs>500</hangtimemsecs>
  </radiotomumble>
</quansheng>
```

| Setting | Meaning |
|---|---|
| `url`, `token` | Where qdock listens, and its `-token`. |
| `timeoutmsecs` | Timeout for API calls. Keying the radio may take up to 10 s: the radio refuses TX for about 6 s after qdock connects. |
| `startupkeys` | Keys pressed once at startup, e.g. `145500` (type a VFO frequency) or `menu up up menu`. |
| `mumbletoradio` | Key the radio while Mumble voice is received (the same moment the `voiceactivity` GPIO turns on). |
| `ptttimeoutsecs` | qdock unkeys by itself if talkkonnect stops refreshing PTT, e.g. after a crash. |
| `hangtimemsecs` (mumbletoradio) | How long the radio stays keyed after Mumble voice stops. |
| `maxtxsecs` | Upper limit for one Mumble transmission. qdock's own `-max-tx` also applies. |
| `radiotomumble` | Transmit into Mumble while the radio's squelch is open. This uses the RX marker on the radio's display, through qdock's event stream. |
| `hangtimemsecs` (radiotomumble) | How long talkkonnect keeps transmitting after the squelch closes. |

**Interlock between the two directions:**

- The radio is not keyed for Mumble voice while it is relaying a received
  signal into Mumble.
- RF is not relayed into Mumble while the radio is keyed, or while someone
  is already transmitting.
- On exit, talkkonnect unkeys the radio.

The section is read at startup. A live config reload does not change it.
Set `<printquansheng>true</printquansheng>` under `<printvariables>` to print
it at startup and to log VFO changes.

## Commands

| HTTP API (`/?command=`) | MQTT | Action |
|---|---|---|
| `quanshengstatus` | `quanshengstatus` | connection, firmware, VFO frequencies, TX/RX, battery |
| `quanshengkeys&keys=145500` | `quanshengkeys 145500` | press a key sequence (`0-9 menu up down exit star f side1 side2`) |
| `quanshengptton` | `quanshengptton` | key the radio (auto off after `ptttimeoutsecs`) |
| `quanshengpttoff` | `quanshengpttoff` | unkey the radio |

As with every command, each one must also be listed in the `<http>`
`<command>` list, or in the MQTT `<commands>` list, in `talkkonnect.xml`.

## Testing against a real radio

`go test -run QuanshengE2E .` with `QUANSHENG_E2E_URL=http://127.0.0.1:8099`
(and `QUANSHENG_E2E_TOKEN` if needed) runs a live test against the radio:
status, UP then DOWN, the event stream, and keying the radio for about a
second for simulated Mumble voice. **Connect a dummy load or antenna, and
make sure you may transmit on the current frequency.**
