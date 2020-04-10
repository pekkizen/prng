package prng

import (
	"math/bits"
	"unsafe"
)

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

// Float64 returns a uniformly distributed pseudo-random float64 value in [0, 1).
func (x *Xoro) Float64() float64 {

	return float64(x.Xoroshiro128plus() >> 11) / (1<<53)
}

// Float64_64 returns a uniformly distributed pseudo-random float64 value in [0, 1).
// The distribution of the returned float64's is 6.5 x denser than in Float64.
func (x *Xoro) Float64_64() float64 {

	return Float64_64(x.Uint64()) 
}
// Float64_64R returns a uniformly distributed pseudo-random float64 value in [0, 1).
func (x *Xoro) Float64_64R() float64 {

	f := Float64_64R(x.Uint64())
	for f == 1 {
		f = Float64_64R(x.Uint64())
	}
	return f
}

var scale = [11]float64 {
	1<<53, 1<<54, 1<<55, 1<<56, 1<<57, 1<<58,
	1<<59, 1<<60, 1<<61, 1<<62, 1<<63, 	
}
// Float64_64 transforms the given 64-bit value to a float64 in [0, 1).
func Float64_64(u uint64) float64 {
	
	zeros := uint64(bits.LeadingZeros64(u))
	if zeros >= 11 {
		return float64(u) / (1<<64) 
	}
	return float64((u << zeros) >> 11) / scale[zeros]
}

// Float64_64R --
func Float64_64R(u uint64) float64 {
	
	zeros := uint64(bits.LeadingZeros64(u))
	if zeros >= 11 {
		return float64(u) / (1<<64) 
	}
	//add 1 to a 54-bit value and truncate to a 53-bit value
	return float64(((u << zeros) >> 10 + 1) >> 1 ) /  scale[zeros]
}

// Float64_1024 returns a pseudo-random float64 value from the uniform distribution of
// of all normal floats in [0, 1).
func (x *Xoro) Float64_1024() float64 {

	hi := x.Uint64()
	if hi >= 1<<52 {  //99.95% of cases 
		return Float64_64(hi)
	} 
	pow := 1.0
	for hi == 0 { 
		hi = x.Uint64() 
		pow *= 1<<64
	}
	lo := x.Uint64()
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	return float64(hi >> 11) / (1<<53) / pow / float64(uint64(1 << zeros))
}

// Float64_1024R --
func (x *Xoro) Float64_1024R() float64 {

resample:
	hi := x.Uint64()
	if hi >= 1<<53 {  //99.9% of cases. Float64_1024 hi >= 1<<52
		f := Float64_64R(hi)
		if f == 1 {
			goto resample
		}
		return f
	} 
	pow := 1.0
	for hi == 0 { 
		hi = x.Uint64() 
		pow *= 1<<64
	}
	lo := x.Uint64()
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	hi = (hi >> 10 + 1) >> 1  
	return float64(hi) / (1<<53) / pow / float64(uint64(1 << zeros))
}

// Float64Bisect returns a pseudo-random float64 from the complete 
// uniform distribution of floats in [0, 1).
func (x *Xoro) Float64Bisect() float64 {

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
				return left
			}
		}
	}
}

// Float64RandomReal returns a uniformly distributed pseudo-random float64 value in [0, 1].
// http://prng.di.unimi.it/random_real.c
// This version returns 0 after 1024 leading zeros.
func (x *Xoro) Float64RandomReal() float64 {
	again:
	hi := x.Uint64()
	if hi >= 1<<63 { //50% of the cases
		f := float64(hi | 1) / (1<<64) 
		if f == 1 {
			goto again
		}
		return f
	}
	pow := 1.0
	for hi == 0 { 
		hi = x.Uint64() 
		pow *= 1<<64
	}
	zeros := uint64(bits.LeadingZeros64(hi))
	lo := x.Uint64()
	hi = (hi << zeros) | (lo >> (64 - zeros))
	return float64(hi | 1) / (1<<64) / pow / float64(uint64(1 << zeros))
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

// Baseline128 is for benchmarking minimal 128-bit state generator.
func (x *Xoro) Baseline128() uint64 {
	next := x.s0
	*x = Xoro{x.s1, x.s0}
	return next
}


