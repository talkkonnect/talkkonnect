package gumble

import "testing"

type nopAudioListener struct{ id int }

func (nopAudioListener) OnAudioStream(*AudioStreamEvent) {}

// attachedIDs walks the list the way dispatch does - from head, following
// next - and reports the listeners dispatch would actually reach.
func attachedIDs(e *AudioListeners) []int {
	var out []int
	for item := e.head; item != nil; item = item.next {
		out = append(out, item.listener.(nopAudioListener).id)
	}
	return out
}

// Attach only ever set tail on the first call, so the third Attach overwrote
// the second's next pointer and that listener disappeared from the list.
func TestAudioListenersAttachAdvancesTail(t *testing.T) {
	var ls AudioListeners
	ls.Attach(nopAudioListener{id: 1})
	ls.Attach(nopAudioListener{id: 2})
	ls.Attach(nopAudioListener{id: 3})

	got := attachedIDs(&ls)
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("dispatch reaches %v, want [1 2 3]", got)
	}
}

// Detaching the head item correctly moves head on, but left tail pointing at
// the item just unlinked, so the next Attach linked onto an item no longer in
// the list and was unreachable from head.
func TestAudioListenersAttachAfterHeadDetach(t *testing.T) {
	var ls AudioListeners
	first := ls.Attach(nopAudioListener{id: 1})
	ls.Attach(nopAudioListener{id: 2})
	first.Detach()
	ls.Attach(nopAudioListener{id: 3})

	got := attachedIDs(&ls)
	if len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("dispatch reaches %v, want [2 3]", got)
	}
}
