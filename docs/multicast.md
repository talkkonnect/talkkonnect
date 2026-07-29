# Multicast to IP Speakers and SIP Phones

talKKonnect can re-transmit the audio it receives from Mumble to a multicast group as RTP, so hardware IP
PA speakers (CyberData, Algo, Barix, Advanced Network Devices) and SIP desk phones such as Yealink can
hear a Mumble channel without being Mumble clients. This turns a talKKonnect node into a gateway between
a Mumble channel and an existing paging system.

Audio goes out as **8 kHz mono G.711, 20 ms per packet** — the format those devices decode — which is the
same wire format [gochimesd](https://github.com/talkkonnect/gochimesd) sends, so one receiver
configuration serves both. Several people talking at once are mixed into a single RTP stream, and nothing
is sent while the channel is quiet.

Add a `<multicast>` section inside `<global><software>` of `talkkonnect.xml`:

```xml
<multicast enabled="true">
    <group>239.0.1.10</group>
    <port>5004</port>
    <codec>pcmu</codec>
    <ttl>1</ttl>
    <interface>eth0</interface>
    <packetms>20</packetms>
    <allchannels>false</allchannels>
    <include>
        <!-- empty means every talker in the channel -->
    </include>
    <exclude>
        <!-- <user>noisybox</user> -->
    </exclude>
</multicast>
```

* `include` and `exclude` decide **whose** audio is carried. An empty `include` carries every talker in
  the channel; a name in `exclude` is never carried, and exclude wins
* `allchannels` left false carries only the channel talKKonnect is joined to, so a monitored channel is
  not re-broadcast over the PA by accident
* `codec` should stay `pcmu` (G.711 u-law) unless every listener on the group is known to support
  something else. `l16` is uncompressed PCM on a dynamic payload type and G.711-only receivers drop it
  silently — the stream looks correct on the network while the speaker stays mute
* a `<multimedia>` profile with `<multicast>true</multicast>` pages the same group with a pre-recorded
  announcement, independently of local playback and playing into the Mumble channel

Control it at run time with `mc status`, `mc on`, `mc off` and `mc toggle` in the CLI or SSH console, with
the `multicaston`, `multicastoff` and `multicasttoggle` actions over the HTTP API and MQTT, and watch it
in the `multicast` object of the `/uistatus` telemetry snapshot. To check the stream from another machine
on the same subnet: `ffplay rtp://@239.0.1.10:5004`.

[Every tag, and the run time and network caveats](./running-talkkonnect.md#the-multicast-section-rtp-to-ip-speakers-and-sip-phones)
