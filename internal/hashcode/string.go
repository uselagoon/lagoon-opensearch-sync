package hashcode

import "strconv"

// StringHash implements the Java String hashcode function for the bytes making
// up a string:
//
//	s[0]*31^(n-1) + s[1]*31^(n-2) + ... + s[n-1]
//
// https://docs.oracle.com/javase/6/docs/api/java/lang/String.html#hashCode()
// https://web.archive.org/web/20130703081745/http://www.cogs.susx.ac.uk/courses/dats/notes/html/node114.html
//
// Note that the result of this hash must be cast to a 32-bit int and formatted
// to base-10 to match Java semantics:
//
//	strconv.FormatInt(int64(int32(h.Sum32())), 10)
//
// See the String() convenience function that implements this.
//
// StringHash implements the hash.Hash32, and fmt.Stringer interfaces from the
// Go standard library.
type StringHash uint32

// NewStringHash returns a new zero-valued StringHash.
func NewStringHash() *StringHash {
	var s StringHash
	return &s
}

// Sum32 is part of the hash.Hash32 interface.
func (h StringHash) Sum32() uint32 { return uint32(h) }

// Write is part of the hash.Hash interface.
func (h *StringHash) Write(data []byte) (int, error) {
	s := h.Sum32()
	for _, b := range data {
		s = 31*s + uint32(b)
	}
	*h = StringHash(s)
	return len(data), nil
}

// BlockSize is part of the hash.Hash interface.
func (StringHash) BlockSize() int { return 1 }

// Size is part of the hash.Hash interface.
func (StringHash) Size() int { return 4 }

// Reset is part of the hash.Hash interface.
func (h *StringHash) Reset() { *h = 0 }

// Sum is part of the hash.Hash interface.
func (h StringHash) Sum(in []byte) []byte {
	s := h.Sum32()
	return append(in, byte(s>>24), byte(s>>16), byte(s>>8), byte(s))
}

// String is a convenience function wrapper around StringHash. It calculates
// the Java String hashcode of the given string, and returns it formatted as a
// base-10 integer.
func String(s string) string {
	h := NewStringHash()
	_, _ = h.Write([]byte(s)) // Write never returns an error.
	return strconv.FormatInt(int64(int32(h.Sum32())), 10)
}
