package prng

import (
	"math/bits"
	"math"
	"unsafe"
)

// A Xosh with a xoshiro256 prng implements a 64-bit generator with 256-bit state.
type Xosh struct {
	s0, s1, s2, s3 uint64
}

// NewXosh returns a new xoshiro256 generator seeded by the seed.
func NewXosh(seed uint64) Xosh {
	x := Xosh{}
	x.Seed(seed)
	return x
}

// Seed seeds a xoshiro256 by the seed using splitMix64. Any seed is ok.
func (x *Xosh) Seed(seed uint64) {
	x.s0 = Splitmix(&seed)
	x.s1 = Splitmix(&seed)
	x.s2 = Splitmix(&seed)
	x.s3 = Splitmix(&seed)
}

// NextXosh returns the next xoshiro256 generator from Outlet. Each generator has
// 2^128 long random streams, which is not overlapping with other generators streams.
// NextXosh is safe for concurrent use by multiple goroutines.
func (s *Outlet) NextXosh() Xosh {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.xosh.Jump()
	return s.xosh
}

// NewXoshSlice returns a slice of n xoshiro256 generators with non-overlapping 2^128
// long random streams. First generator is seeded by the seed.
func NewXoshSlice(n int, seed uint64) []Xosh {
	s := make([]Xosh, n)
	s[0].Seed(seed)
	for i := 1; i < n; i++ {
		s[i] = s[i-1]
		s[i].Jump()
	}
	return s
}

// Uint64 returns a pseudo-random uint64. Uint64 is xoshiro256**.
func (x *Xosh) Uint64() (next uint64) {

	next = bits.RotateLeft64(x.s1 * 5, 7) * 9
	*x = x.NextState()
	return
}

// Xoshiro256plus is xoshiro256+
func (x *Xosh) Xoshiro256plus() (next uint64) {

	next = x.s0 + x.s3
	*x = x.NextState()
	return
}

//Xoshiro256plusplus is xoshiro256++
func (x *Xosh) Xoshiro256plusplus() (next uint64) {

	next = bits.RotateLeft64(x.s0 + x.s3, 23) + x.s0
	*x = x.NextState()
	return
}

// NextState returns the next Xosh state of the xoshiro256 linear engine.
func (x Xosh) NextState() Xosh {
	//gc compiler detects similar expressions if given in parentheses

	return Xosh{
		s0: x.s0 ^ (x.s1 ^ x.s3),
		s1: (x.s0 ^ x.s2) ^ x.s1,
		s2: (x.s0 ^ x.s2) ^ (x.s1 << 17),
		s3: bits.RotateLeft64(x.s1 ^ x.s3, 45),
	}
}

// Float64 returns a uniformly distributed pseudo-random float64 from [0, 1).
// The distribution includes  2^53 evenly spaced floats with spacing 2^-53.
func (x *Xosh) Float64() float64 {

	return float64(x.Xoshiro256plus() >> 11) / (1<<53)
}

// Float64_64 returns a uniformly distributed pseudo-random float64 from [0, 1).
// The distribution includes all floats in [2^-12, 1) and 2^52 evenly spaced 
// floats in [0, 2^-12) with spacing 2^-64.
func (x *Xosh) Float64_64() float64 {

	u := x.Uint64()
	if u == 0 { return 0 }  // without this min returned is 2^-65
	z := uint64(bits.LeadingZeros64(u)) + 1
	return math.Float64frombits((1023 - z) << 52 | u << z >> 12)
}

// Float64_117 returns a uniformly distributed pseudo-random float64 from [0, 1). 
// The distribution includes all floats in [2^-65, 1) and 2^52  evenly spaced 
// floats in [0, 2^-65) with spacing 2^-117.
func (x *Xosh) Float64_117() float64 {

	u := x.Uint64()
	z := uint64(bits.LeadingZeros64(u)) + 1
	if z <= 12 {  
		return math.Float64frombits((1023 - z) << 52 | u << z >> 12)
	}
	z--
    u = u << z | x.Uint64() >> (64 - z)
	return float64(u >> 11) * twoToMinus(53 + z)
}

// Float64full returns a uniformly distributed pseudo-random float64 from [0, 1). 
// The distribution includes all floats in [0, 1). 
// Float64full is equivalent to Float64Bisect in truncate mode.
func (x *Xosh) Float64full() float64 {

	u := x.Uint64()
	z := uint64(bits.LeadingZeros64(u)) + 1
	if z <= 12 {                                 //99.975% of cases 
		return math.Float64frombits((1023 - z) << 52 | u << z >> 12)
	}
	z--
	exp := uint64(0)
	for u == 0 { 
		u = x.Uint64() 
		z = uint64(bits.LeadingZeros64(u))
		exp += 64
		if exp + z >= 1074 { return 0 }
	}
	u = u << z | x.Uint64() >> (64 - z)
	exp += z
	if exp < 1022 {
		return math.Float64frombits((1022 - exp) << 52 | u << 1 >> 12)
	}
	return math.Float64frombits(u >> (exp - 1022) >> 12) // 2^52 subnormal floats
}


// WriteState writes the current state of the generator x to b.
// WriteState without allocations is faster than State().
func (x *Xosh) WriteState(b []byte)  {
	if len(b) < XoshStateSize {
		panic("ReadState: byte slice too short")
	}
	// This expects a little endian cpu, eg. all amd64.
	*(*uint64)(unsafe.Pointer(&b[ 0])) = bits.ReverseBytes64(x.s0)
	*(*uint64)(unsafe.Pointer(&b[ 8])) = bits.ReverseBytes64(x.s1)
	*(*uint64)(unsafe.Pointer(&b[16])) = bits.ReverseBytes64(x.s2)
	*(*uint64)(unsafe.Pointer(&b[24])) = bits.ReverseBytes64(x.s3)
}

// State returns the current binary state of the generator x as []byte.
func (x *Xosh) State() []byte {
	var b[XoshStateSize]byte
	
	*(*uint64)(unsafe.Pointer(&b[ 0])) = bits.ReverseBytes64(x.s0)
	*(*uint64)(unsafe.Pointer(&b[ 8])) = bits.ReverseBytes64(x.s1)
	*(*uint64)(unsafe.Pointer(&b[16])) = bits.ReverseBytes64(x.s2)
	*(*uint64)(unsafe.Pointer(&b[24])) = bits.ReverseBytes64(x.s3)
	return b[:]
}

// ReadState reads the state of the generator x from b []byte.
func (x *Xosh) ReadState(b []byte) {
	if len(b) < XoshStateSize {
		panic("ReadState: byte slice too short")
	}
	x.s0 = bits.ReverseBytes64(*(*uint64)(unsafe.Pointer(&b[ 0])))
	x.s1 = bits.ReverseBytes64(*(*uint64)(unsafe.Pointer(&b[ 8])))
	x.s2 = bits.ReverseBytes64(*(*uint64)(unsafe.Pointer(&b[16])))
	x.s3 = bits.ReverseBytes64(*(*uint64)(unsafe.Pointer(&b[24])))
}

	// Alternative ReadState
	// x.s0 = binary.BigEndian.Uint64(b[0:])
	// x.s1 = binary.BigEndian.Uint64(b[8:])
	// x.s2 = binary.BigEndian.Uint64(b[16:])
	// x.s3 = binary.BigEndian.Uint64(b[24:])

	// Alternative State
	// binary.BigEndian.PutUint64(b[0:],  x.s0)
	// binary.BigEndian.PutUint64(b[8:],  x.s1)
	// binary.BigEndian.PutUint64(b[16:], x.s2)
	// binary.BigEndian.PutUint64(b[24:], x.s3)
