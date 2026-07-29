package talkkonnect

import (
	"encoding/xml"
	"strings"
	"testing"
)

// vtTestDoc mirrors the account shape in the shipped configs: two accounts, the
// first with a populated <voicetargets>, comments in the middle, and the mixed
// indentation the hand written file actually uses.
const vtTestDoc = `<?xml version="1.0" encoding="UTF-8"?>
<document>
  <accounts>
    <account name="primary_account" default="true">
      <username>mycall</username>
      <!-- a commented out example must not be mistaken for markup
      <voicetargets>
        <id value="99"><users><user>ghost</user></users></id>
      </voicetargets>
      -->
      <listentochannels/>
      <voicetargets>
           <id value="1">
             <users>
               <user>zoran-laptop</user>
             </users>
           </id>
           <id value="3">
             <channels>
               <channel>
                 <name>TEST1</name>
                 <recursive>true</recursive>
                 <links>true</links>
                 <group>"all"</group>
               </channel>
             </channels>
           </id>
      </voicetargets>
    </account>
    <account name="secondary_account" default="false">
      <username>user2</username>
      <voicetargets/>
    </account>
  </accounts>
  <global>
    <multimedia>
      <id value="main_announcement" enabled="true"/>
    </multimedia>
  </global>
</document>
`

func vtParse(t *testing.T, doc string) ConfigStruct {
	t.Helper()
	var cfg ConfigStruct
	if err := xml.Unmarshal([]byte(doc), &cfg); err != nil {
		t.Fatalf("edited document does not parse: %v\n%s", err, doc)
	}
	return cfg
}

func vtTarget(t *testing.T, cfg ConfigStruct, accountIndex int, id uint32) (users []string, channels []voiceTargetChannelEntry) {
	t.Helper()
	if accountIndex >= len(cfg.Accounts.Account) {
		t.Fatalf("no account %v in parsed config", accountIndex)
	}
	for _, target := range cfg.Accounts.Account[accountIndex].Voicetargets.ID {
		if target.Value != id {
			continue
		}
		users = append(users, target.Users.User...)
		for _, c := range target.Channels.Channel {
			channels = append(channels, voiceTargetChannelEntry{Name: c.Name, Recursive: c.Recursive, Links: c.Links, Group: c.Group})
		}
		return users, channels
	}
	t.Fatalf("voice target %v not found in account %v", id, accountIndex)
	return nil, nil
}

func TestVoiceTargetXMLAddUserToExistingTarget(t *testing.T) {
	out, err := voiceTargetXMLAdd(vtTestDoc, 0, 1, []string{"newuser"}, nil)
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	users, _ := vtTarget(t, vtParse(t, out), 0, 1)
	if len(users) != 2 || users[0] != "zoran-laptop" || users[1] != "newuser" {
		t.Errorf("users = %v, want [zoran-laptop newuser]", users)
	}

	// The commented out block and the other account must be untouched.
	if !strings.Contains(out, "<user>ghost</user>") {
		t.Error("the comment block was modified")
	}
	if strings.Count(out, "<voicetargets/>") != 1 {
		t.Error("the second account's <voicetargets/> should be left alone")
	}
	if strings.Contains(out, "<username>mycall</username>\n      <voicetargets>") {
		t.Error("markup was inserted in the wrong place")
	}
}

func TestVoiceTargetXMLAddChannelToTargetWithOnlyUsers(t *testing.T) {
	out, err := voiceTargetXMLAdd(vtTestDoc, 0, 1, nil,
		[]voiceTargetChannelEntry{{Name: "HAM-CB", Recursive: true, Links: false, Group: "all"}})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	users, channels := vtTarget(t, vtParse(t, out), 0, 1)
	if len(users) != 1 {
		t.Errorf("users = %v, want the existing one untouched", users)
	}
	if len(channels) != 1 {
		t.Fatalf("channels = %v, want 1", channels)
	}
	want := voiceTargetChannelEntry{Name: "HAM-CB", Recursive: true, Links: false, Group: "all"}
	if channels[0] != want {
		t.Errorf("channel = %+v, want %+v", channels[0], want)
	}
}

func TestVoiceTargetXMLAddNewTarget(t *testing.T) {
	out, err := voiceTargetXMLAdd(vtTestDoc, 0, 7, []string{"someone"}, nil)
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	cfg := vtParse(t, out)
	users, _ := vtTarget(t, cfg, 0, 7)
	if len(users) != 1 || users[0] != "someone" {
		t.Errorf("users = %v, want [someone]", users)
	}
	if got := len(cfg.Accounts.Account[0].Voicetargets.ID); got != 3 {
		t.Errorf("account has %v targets, want 3", got)
	}
	// A new <id> should pick up the indentation the existing siblings use.
	if !strings.Contains(out, "\n           <id value=\"7\">") {
		t.Errorf("new <id> not indented like its siblings:\n%s", out)
	}
}

func TestVoiceTargetXMLExpandsSelfClosingVoicetargets(t *testing.T) {
	out, err := voiceTargetXMLAdd(vtTestDoc, 1, 2, []string{"bob"}, nil)
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	users, _ := vtTarget(t, vtParse(t, out), 1, 2)
	if len(users) != 1 || users[0] != "bob" {
		t.Errorf("users = %v, want [bob]", users)
	}
	if strings.Contains(out, "<voicetargets/>") {
		t.Error("<voicetargets/> should have been expanded")
	}
	// The first account keeps its two targets.
	if got := len(vtParse(t, out).Accounts.Account[0].Voicetargets.ID); got != 2 {
		t.Errorf("first account has %v targets, want 2 untouched", got)
	}
}

func TestVoiceTargetXMLCreatesMissingVoicetargets(t *testing.T) {
	doc := `<document>
  <accounts>
    <account name="only" default="true">
      <username>x</username>
    </account>
  </accounts>
</document>
`
	out, err := voiceTargetXMLAdd(doc, 0, 5, nil,
		[]voiceTargetChannelEntry{{Name: "Zone A", Recursive: false, Links: false}})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	_, channels := vtTarget(t, vtParse(t, out), 0, 5)
	if len(channels) != 1 || channels[0].Name != "Zone A" {
		t.Errorf("channels = %+v, want one named %q", channels, "Zone A")
	}
}

func TestVoiceTargetXMLHandlesSingleLineContainer(t *testing.T) {
	doc := `<document>
  <accounts>
    <account name="only" default="true">
      <voicetargets>
        <id value="4"><users></users></id>
      </voicetargets>
    </account>
  </accounts>
</document>
`
	out, err := voiceTargetXMLAdd(doc, 0, 4, []string{"solo"}, nil)
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	users, _ := vtTarget(t, vtParse(t, out), 0, 4)
	if len(users) != 1 || users[0] != "solo" {
		t.Errorf("users = %v, want [solo]", users)
	}
}

func TestVoiceTargetXMLEscapesNames(t *testing.T) {
	out, err := voiceTargetXMLAdd(vtTestDoc, 0, 9, []string{`a<b&c"d`}, nil)
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	users, _ := vtTarget(t, vtParse(t, out), 0, 9)
	if len(users) != 1 || users[0] != `a<b&c"d` {
		t.Errorf("users = %q, want the name to survive a round trip", users)
	}
}

func TestVoiceTargetXMLRejectsMissingAccount(t *testing.T) {
	if _, err := voiceTargetXMLAdd(vtTestDoc, 9, 1, []string{"x"}, nil); err == nil {
		t.Error("adding to a non existent account should fail")
	}
	if _, err := voiceTargetXMLAdd(vtTestDoc, 0, 1, nil, nil); err == nil {
		t.Error("adding nothing should fail")
	}
}

func TestVoiceTargetCLITokens(t *testing.T) {
	got := voiceTargetCLITokens(`vt add 3 channel "Zone One" recursive=true group="all hands"`)
	want := []string{"vt", "add", "3", "channel", "Zone One", "recursive=true", "group=all hands"}
	if len(got) != len(want) {
		t.Fatalf("tokens = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestVoiceTargetVerifyDocumentCatchesMisplacedEdit(t *testing.T) {
	// A document where the target was not actually added must be rejected.
	if err := voiceTargetVerifyDocument(vtTestDoc, 0, 1, []string{"absent"}, nil); err == nil {
		t.Error("verify should reject a document missing the added user")
	}
	if err := voiceTargetVerifyDocument("<document", 0, 1, []string{"x"}, nil); err == nil {
		t.Error("verify should reject a document that does not parse")
	}
	if err := voiceTargetVerifyDocument(vtTestDoc, 0, 1, []string{"zoran-laptop"}, nil); err != nil {
		t.Errorf("verify rejected a valid document: %v", err)
	}
}
