package varint

import "testing"

func TestRange(t *testing.T) {

	fn := func(i int64) {
		var b [MaxVarintLen]byte
		size := Encode(b[:], i)
		if size == 0 {
			t.Error("Encode returned size 0\n")
		}
		s := b[:size]

		val, size := Decode(s)
		if size == 0 {
			t.Error("Decode return size 0\n")
		}

		if i != val {
			t.Errorf("Encoded %d (%v) equals decoded %d\n", i, s, val)
		}
	}

	for i := int64(-10000); i <= 10000; i++ {
		fn(i)
	}

	fn(134342525)
	fn(10282934828342)
	fn(1028293482834200000)
}
