package talkkonnect

import (
	"encoding/xml"
	"testing"
	"time"
)

// multimediaTestConfig mirrors the <multimedia> shape shipped in sample-configs,
// where blocking, offset and loop are written as attributes.
const multimediaTestConfig = `<document>
  <global>
    <multimedia>
      <id value="main_announcement" enabled="true">
        <params>
          <announcementtone file="/tone.wav" volume="10" blocking="true" enabled="true"/>
          <localplay>true</localplay>
          <gpio name="multimedia_active_led" enabled="true"/>
          <predelay value="2" enabled="true"/>
          <postdelay value="1" enabled="true"/>
          <playintostream>true</playintostream>
          <streamvolume>40</streamvolume>
          <voicetarget id="3">true</voicetarget>
        </params>
        <schedule intervalsecs="300" enabled="true"/>
        <media>
          <source name="1st-song" file="/song.mp3" volume="25" duration="12" offset="4" loop="2" blocking="true" enabled="true"/>
        </media>
      </id>
      <id value="legacy_profile" enabled="true">
        <params>
          <localplay>true</localplay>
          <playintostream>false</playintostream>
          <voicetarget/>
        </params>
        <media>
          <source name="only" file="/other.mp3" enabled="true"/>
        </media>
      </id>
    </multimedia>
  </global>
</document>`

func TestMultimediaConfigParsesAttributes(t *testing.T) {
	var cfg ConfigStruct
	if err := xml.Unmarshal([]byte(multimediaTestConfig), &cfg); err != nil {
		t.Fatalf("unmarshal multimedia config: %v", err)
	}

	if len(cfg.Global.Multimedia.ID) != 2 {
		t.Fatalf("parsed %v multimedia profiles, want 2", len(cfg.Global.Multimedia.ID))
	}

	profile := cfg.Global.Multimedia.ID[0]

	if !profile.Params.Announcementtone.Blocking {
		t.Error("announcementtone blocking = false, want true (blocking is an attribute)")
	}
	if got, want := profile.Params.Announcementtone.Volume, 10; got != want {
		t.Errorf("announcementtone volume = %v, want %v", got, want)
	}

	if len(profile.Media.Source) != 1 {
		t.Fatalf("parsed %v sources, want 1", len(profile.Media.Source))
	}
	source := profile.Media.Source[0]
	if !source.Blocking {
		t.Error("source blocking = false, want true (blocking is an attribute)")
	}
	if got, want := source.Offset, float32(4); got != want {
		t.Errorf("source offset = %v, want %v", got, want)
	}
	if got, want := source.Duration, float32(12); got != want {
		t.Errorf("source duration = %v, want %v", got, want)
	}
	if got, want := source.Loop, 2; got != want {
		t.Errorf("source loop = %v, want %v", got, want)
	}

	if !multimediaVoicetargetEnabled(profile.Params.Voicetarget.Value) {
		t.Errorf("voicetarget enabled = false for %q, want true", profile.Params.Voicetarget.Value)
	}
	if got, want := profile.Params.Voicetarget.ID, uint32(3); got != want {
		t.Errorf("voicetarget id = %v, want %v", got, want)
	}

	if got, want := multimediaDelaySeconds(profile.Params.Predelay.Value), 2*time.Second; got != want {
		t.Errorf("predelay = %v, want %v", got, want)
	}

	if !profile.Schedule.Enabled || profile.Schedule.IntervalSecs != 300 {
		t.Errorf("schedule = %+v, want enabled every 300 seconds", profile.Schedule)
	}

	// The historic empty <voicetarget/> form must stay parseable and read as off.
	legacy := cfg.Global.Multimedia.ID[1]
	if multimediaVoicetargetEnabled(legacy.Params.Voicetarget.Value) {
		t.Error("empty voicetarget element read as enabled, want disabled")
	}
}

func TestMediaSourceLoopClamps(t *testing.T) {
	for _, tc := range []struct {
		in   int
		want int
	}{
		{in: -1, want: 1},
		{in: 0, want: 1},
		{in: 1, want: 1},
		{in: maxMediaSourceLoops, want: maxMediaSourceLoops},
		{in: maxMediaSourceLoops + 5, want: maxMediaSourceLoops},
	} {
		if got := mediaSourceLoop(tc.in); got != tc.want {
			t.Errorf("mediaSourceLoop(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestMultimediaStreamVolumeFallsBackToProfile(t *testing.T) {
	if got, want := multimediaStreamVolume(25, 40), float32(25); got != want {
		t.Errorf("per source volume = %v, want %v", got, want)
	}
	if got, want := multimediaStreamVolume(0, 40), float32(40); got != want {
		t.Errorf("unset source volume = %v, want profile volume %v", got, want)
	}
}

func TestMultimediaVoicetargetEnabled(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{in: "true", want: true},
		{in: " true ", want: true},
		{in: "1", want: true},
		{in: "false", want: false},
		{in: "", want: false},
		{in: "nonsense", want: false},
	} {
		if got := multimediaVoicetargetEnabled(tc.in); got != tc.want {
			t.Errorf("multimediaVoicetargetEnabled(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
