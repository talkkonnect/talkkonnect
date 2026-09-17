package gumble

type audioEventItem struct {
	parent     *AudioListeners
	prev, next *audioEventItem
	listener   AudioListener
	streams    map[*User]chan *AudioPacket
}

func (e *audioEventItem) Detach() {
	if e.prev == nil {
		e.parent.head = e.next
	} else {
		e.prev.next = e.next
	}
	if e.next == nil {
		e.parent.tail = e.prev
	} else {
		e.next.prev = e.prev
	}
}

// AudioListeners is a list of audio listeners. Each attached listener is
// called in sequence when a new user audio stream begins.
type AudioListeners struct {
	head, tail *audioEventItem
}

// Attach adds a new audio listener to the end of the current list of listeners.
func (e *AudioListeners) Attach(listener AudioListener) Detacher {
	item := &audioEventItem{
		parent:   e,
		prev:     e.tail,
		listener: listener,
		streams:  make(map[*User]chan *AudioPacket),
	}
	if e.tail == nil {
		e.head = item
	} else {
		e.tail.next = item
	}
	// tail must advance to the newly appended item on every Attach, not just
	// the first. Left stale it keeps pointing at whatever was attached first,
	// so the third Attach overwrites the second's next pointer (that listener
	// disappears from the list), and an Attach after the head item detaches
	// links onto an item that is no longer in the list. Either way the
	// listener is unreachable from head and never receives audio.
	//
	// A Detach+Attach cycle with a single listener happens to recover, because
	// Detach resets tail to nil - which is what makes this look intermittent.
	e.tail = item
	return item
}
