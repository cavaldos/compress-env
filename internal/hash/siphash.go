package hash

import (
	"encoding/binary"
	"math/bits"
)

// sipHash13 computes SipHash-1-3 with an all-zero key. This is the algorithm
// behind Rust's std DefaultHasher::new(), which the original implementation
// used to derive the XOR key from the password. It is kept so that files
// written by the Rust version stay readable.
func sipHash13(msg []byte) uint64 {
	const k0, k1 = 0, 0
	v0 := uint64(0x736f6d6570736575) ^ k0
	v1 := uint64(0x646f72616e646f6d) ^ k1
	v2 := uint64(0x6c7967656e657261) ^ k0
	v3 := uint64(0x7465646279746573) ^ k1

	round := func() {
		v0 += v1
		v1 = bits.RotateLeft64(v1, 13)
		v1 ^= v0
		v0 = bits.RotateLeft64(v0, 32)
		v2 += v3
		v3 = bits.RotateLeft64(v3, 16)
		v3 ^= v2
		v0 += v3
		v3 = bits.RotateLeft64(v3, 21)
		v3 ^= v0
		v2 += v1
		v1 = bits.RotateLeft64(v1, 17)
		v1 ^= v2
		v2 = bits.RotateLeft64(v2, 32)
	}

	n := len(msg)
	for len(msg) >= 8 {
		m := binary.LittleEndian.Uint64(msg)
		v3 ^= m
		round()
		v0 ^= m
		msg = msg[8:]
	}

	last := uint64(n) << 56
	for i, b := range msg {
		last |= uint64(b) << (8 * uint(i))
	}
	v3 ^= last
	round()
	v0 ^= last

	v2 ^= 0xff
	round()
	round()
	round()

	return v0 ^ v1 ^ v2 ^ v3
}

// hashStr mirrors `str::hash` fed into DefaultHasher: the bytes of the string
// followed by a 0xff terminator.
func hashStr(s string) uint64 {
	buf := make([]byte, 0, len(s)+1)
	buf = append(buf, s...)
	buf = append(buf, 0xff)
	return sipHash13(buf)
}
