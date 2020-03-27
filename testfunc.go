package prng

import (
	"math/bits"

)

// Float64_128 returns a uniformly distributed pseudo-random float64 value in [0, 1).
// The distribution is much denser near 0.
func (x *Xoro) Float64_128() float64 {

	hi := x.Uint64()
	if hi >= 1<<52 {
		return  Float64_64(hi)
	} 
	lo := x.Uint64()
	if hi == 0 {
		return  Float64_64(lo) / (1 << 64)
	} 
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	return float64(hi >> 11) / (1 <<53 ) / float64(uint64(1 << zeros))
}

// Float64_128 is for testing.
func Float64_128(hi, lo uint64) float64 {

	if hi >= 1<<52 {
		return  Float64_64(hi)
	} 
	if hi == 0 {
		return  Float64_64(lo) / (1 << 64)
	} 
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	return float64(hi >> 11) / (1 << 53) / float64(uint64(1 << zeros))
}

// Float64FullDouble --
func (x *Xoro) Float64FullDouble() (float64, float64) {

	hi := x.UintFake()
	 
	lo := x.UintFake()
	powz := 1.0
	i := 1
	for hi == 0 { 
		hi = x.UintFake() 
		
		if i++; i > 15 { 
			hi++
			break
		}
		powz *= (1 << 64)
	}
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	hi >>= 11

	f1 := float64(hi) / (1 << 53) / powz / float64(uint64(1 << zeros))
	f2 := float64(hi+1) / (1 << 53) / powz / float64(uint64(1 << zeros))
	if f1 == 0 {
		return f1, powz
	}
	return f1, f2
}


// UintFake --
func (x *Xoro) UintFake() (next uint64) {

	next = bits.RotateLeft64(x.s0 * 5, 7) * 9
	if next % 2 == 0 {
		next = 0
	}
	*x = x.NextState()
	return
}

// Float64FullSmall --
func Float64FullSmall(hi, lo uint64, rounds int) float64 {

	powz := 1.0
	for i := 0; i < rounds; i++ { 
		powz *= (1 << 64)
	}
	
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	// return float64(hi >> 11) / (1 << 53) / powz / float64(uint64(1 << zeros))
	return float64(hi >> 11) / (1 << 53) / (powz * float64(uint64(1 << zeros)))
}

// RomuDuo https://arxiv.org/pdf/2002.11331.pdf
func (x *Xoro) RomuDuo() uint64 {
	s0, s1 := x.s0, x.s1
	
	x.s0 = 15241094284759029579 * s1
	x.s1 = bits.RotateLeft64(s1, 36) + bits.RotateLeft64(s1, 15) - s0
	return s0
}
