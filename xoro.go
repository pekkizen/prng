package prng

import (
	"math/bits"
	"math"
	"unsafe"
)

// const fakeUint = true
const fakeUint = false

// A Xoro with a xoroshiro prng implements a 64-bit generator with 128-bit state.
// A Xoro is the den of the xoroshiros holding their two 64-bit eggs.
type Xoro struct {
	s0, s1 uint64
}

// NewXoro returns a new xoroshiro128 generator seeded by the seed.
// Float64 uses xoroshiro128+ and Uint64 xoroshiro128**. Both xoroshiros update
// a Xoro in the same way (same linear engine) and we can use a single Xoro for both
// functions without interfering random stream properties.
func NewXoro(seed uint64) Xoro {
	x := Xoro{}
	x.Seed(seed)
	return x
}

// Seed seeds a xorohiro128 generator by seed using splitMix64. Any seed is ok.
func (x *Xoro) Seed(seed uint64) {
	x.s0 = Splitmix(&seed)
	x.s1 = Splitmix(&seed)
}

// NextXoro returns the next xoroshiro128 from Outlet. Each generator has
// 2^64 long random streams, which is not overlapping with other generators streams.
// NextXoro is safe for concurrent use by multiple goroutines.
func (s *Outlet) NextXoro() Xoro {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.xoro.Jump()
	return s.xoro
}

// NewXoroSlice returns a slice of n xoroshiro128 generators with non-overlapping 2^64
// long random streams. First generator is seeded by seed.
func NewXoroSlice(n int, seed uint64) []Xoro {
	s := make([]Xoro, n)
	s[0].Seed(seed)
	for i := 1; i < n; i++ {
		s[i] = s[i-1]
		s[i].Jump()
	}
	return s
}

// Uint64 returns a  pseudo-random uint64. Uint64 is xoroshiro128**.
func (x *Xoro) Uint64() (next uint64) {

	next = bits.RotateLeft64(x.s0 * 5, 7) * 9
	*x = x.NextState()
	if fakeUint {
		if next % 3 == 0 {
			next >>= next % 66
		}
		if next % 5 == 0 {
			next <<= next % 66
		}
	}
	return
}

// Xoroshiro128plus is xoroshiro128+
func (x *Xoro) Xoroshiro128plus() (next uint64) {

	next = x.s0 + x.s1 
	*x = x.NextState()
	return
}

// NextState returns the next Xoro state of the xoroshiro128+/** linear engine.
func (x Xoro) NextState() Xoro {
	//gc compiler detects similar expressions if given in parentheses

	return Xoro{
		s0: bits.RotateLeft64(x.s0, 24) ^ (x.s0 ^ x.s1) ^ ((x.s0 ^ x.s1) << 16),
		s1: bits.RotateLeft64(x.s0 ^ x.s1, 37),
	}
}

// State returns the current state of the generator x as []byte.
func (x *Xoro) State() []byte {
	var b[16]byte
	
	// This expects a little endian cpu, eg. all amd64.
	*(*uint64)(unsafe.Pointer(&b[0])) = bits.ReverseBytes64(x.s0)
	*(*uint64)(unsafe.Pointer(&b[8])) = bits.ReverseBytes64(x.s1)
	return b[:]
}

// SetState sets the state of the generator x from the state in b []byte.
func (x *Xoro) SetState(b []byte) {
	if len(b) < 16 {
		panic("Xoro SetState bytes < 16")
	}
	x.s0 = bits.ReverseBytes64(*(*uint64)(unsafe.Pointer(&b[0])))
	x.s1 = bits.ReverseBytes64(*(*uint64)(unsafe.Pointer(&b[8])))
}

// Float64 returns a uniformly distributed pseudo-random float64 value in [0, 1). 
// The distribution is  2^53 evenly spaced floats with spacing 2^-53.
func (x *Xoro) Float64() float64 {

	return float64(x.Xoroshiro128plus() >> 11) / (1<<53)
}

// Float64_64 returns a uniformly distributed pseudo-random float64 value 
// in [0, 1). The distribution includes all floats in [2^-12, 1) and 2^52 
// evenly spaced floats in [0, 2^-12) with spacing 2^-64.
func (x *Xoro) Float64_64() float64 {

	return float64_64(x.Uint64()) 
}

// Float64_64R returns a uniformly distributed pseudo-random float64 from [0, 1] using rounding.
// The distribution includes all floats in [2^-12, 1) and 2^52 evenly spaced floats in [0, 2^-12).
func (x *Xoro) Float64_64R() float64 {

	return float64_64Round(x.Uint64())
}
// float64_64 transforms the given 64-bit value to a float64 in [0, 1).
func float64_64(u uint64) float64 {

	z := uint64(bits.LeadingZeros64(u)) + 1
	if z > 11 {  
		return float64(u) / (1<<64) 
	}
	return math.Float64frombits((1023 - z) << 52 | u << z >> 12)
}
// float64_64Round transforms the given 64-bit value to a float64 in [0, 1] using rounding.
// Rounding: add 1 to a 53-bit value and truncate to a 52-bit value
func float64_64Round(u uint64) float64 {
	
	z := uint64(bits.LeadingZeros64(u)) + 1
	if z > 11 { 
		return float64(u) / (1<<64) 
	}
	u <<= z
	return math.Float64frombits((1023 - z) << 52 | (u >> 11 + 1) >> 1)
}

// Float64_1024 returns a uniformly distributed pseudo-random float64 value 
// in [0, 1). The distribution includes all floats in [2^-1024, 1) and  0.
func (x *Xoro) Float64_1024() float64 {

	u := x.Uint64()
	if u >= 1<<52 {  //99.95% of cases 
		return float64_64(u)
	} 
	pow := 1.0
	for u == 0 { 
		u = x.Uint64() 
		pow *= 1<<64
	}
	z := uint64(bits.LeadingZeros64(u))
	u = u << z | x.Uint64() >> (64 - z)
	return float64(u >> 11) / (1<<53) / pow / float64(uint64(1 << z))
}

// Float64_1024R returns a uniformly distributed pseudo-random float64 value 
// in [0, 1] using rounding. The distribution includes all floats in [2^-1024, 1] and  0.
// Rounding: add 1 to a 54-bit value and truncate to a 53-bit value
func (x *Xoro) Float64_1024R() float64 {

	u := x.Uint64()
	if u >= 1<<53 {  //max 10 zeros rounded by float64_64Round
		return float64_64Round(u)
	} 
	pow := 1.0
	for u == 0 { 
		u = x.Uint64() 
		pow *= 1<<64
	}
	z := uint64(bits.LeadingZeros64(u))
	u = u << z | x.Uint64() >> (64 - z)
	return float64((u >> 10 + 1) >> 1) / (1<<53) / pow / float64(uint64(1 << z))
}

// Float64_117 returns a uniformly distributed pseudo-random float64 value 
// in [0, 1). The distribution includes all floats in [2^-65, 1) and  
// 2^52 evenly spaced floats in [0, 2^-65) with spacing 2^-117.
func (x *Xoro) Float64_117() float64 {

	u := x.Uint64()
	z := uint64(bits.LeadingZeros64(u)) + 1
	if z <= 12 {  //need 52 bits
		return math.Float64frombits((1023 - z) << 52 | u << z >> 12)
	}
	z--
	u = u << z | x.Uint64() >> (64 - z)
	return math.Float64frombits((1022 - z) << 52 | u << 1 >> 12)
}

// Float64_117R returns a uniformly distributed pseudo-random float64 value 
// in [0, 1] using rounding. The distribution includes all floats in [2^-65, 1) 
// and 2^52 evenly spaced floats in [0, 2^-65) with spacing 2^-117.
func (x *Xoro) Float64_117R() float64 {

	u := x.Uint64()
	z := uint64(bits.LeadingZeros64(u)) + 1
	if z <= 11 {  //need 52 bits + 1 rounding bit
		return math.Float64frombits((1023 - z) << 52 | (u << z >> 11 + 1) >> 1)
	}
	z--
	u = u << z | x.Uint64() >> (64 - z)
	return math.Float64frombits((1022 - z) << 52 | (u << 1 >> 11 + 1) >> 1)
}

// RandomReal returns a uniformly distributed pseudo-random float64 value in [0, 1].
// http://prng.di.unimi.it/random_real.c
// This version returns 0 after 1024 leading zeros.
func (x *Xoro) RandomReal() float64 {

	u := x.Uint64()
	pow := 1.0
	for u == 0 { 
		u = x.Uint64() 
		pow *= 1<<64
	}
	z := uint64(bits.LeadingZeros64(u))
	u = u << z | x.Uint64() >> (64 - z)
	return float64(u | 1) / (1<<64) / (pow * float64(uint64(1 << z)))
}

// Float64Bisect returns a uniformly distributed pseudo-random float64 value in [0, 1).
// If round == true, rounding is applied and the range is [0, 1].
// All floats, normal and subnormal, are included.
func (x *Xoro) Float64Bisect(round bool) float64 {

	left, mean, right := 0.0, 0.5, 1.0
	for {
		u := x.Uint64()
		for b := 0; b < 64; b++ {

			if u & (1<<63) != 0 {
				left = mean						// '1' bit -> take the right half, big numbers			
			} else {
				right = mean					// '0' bit -> take the left half, small numbers		
			}
			u <<= 1
			mean = (left + right) / 2
			if mean == left || mean == right {	// right - left = 1 ULP
				if math.Float64bits(mean) & 1 == 1 {
					panic("  ")
				}
				if !round {
					return left					// no rounding
				}
				if b == 63 {
					u = x.Uint64()
				}
				if u & (1<<63) != 0 {			// '1' bit -> round up
					return right								
				} 
				return left
			}
		}
	}
}

// functions without math.Float64frombits.
var scale = [11]float64 {
	1<<53, 1<<54, 1<<55, 1<<56, 1<<57, 1<<58, 
	1<<59, 1<<60, 1<<61, 1<<62, 1<<63, 
}
func float64_64B(u uint64) float64 {

	z := uint64(bits.LeadingZeros64(u))
	if z <= 10 {  
		return float64((u << z) >> 11) / scale[z]  
	}
	return float64(u) / (1<<64) 
}
func float64_64RoundB(u uint64) float64 {
	
	z := uint64(bits.LeadingZeros64(u))
	if z <= 10 { 
		return float64(((u << z) >> 10 + 1) >> 1) / scale[z]	
	}
	return float64(u) / (1<<64) 
}

