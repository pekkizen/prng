package prng
// https://en.wikipedia.org/wiki/Lehmer_random_number_generator

import (
	"math"
	"math/bits"
)

// A MCG implements 64-bit multiplicative congruential pseudorandom number 
// generator (MCG) modulo 2^64 with 64-bit state and maximun period of 2^62.
type MCG struct {
	state uint64
}

// Steele and Vigna https://arxiv.org/pdf/2001.05304.pdf:
// For a MCG with modulus of power of two, the state must be odd for 
// maximun period 2^64 / 4 = 2^62.

// Seed --
func (x *MCG) Seed(seed uint64) {
	x.state = Splitmix(&seed) | 1
}

// NewMCG --
func NewMCG(seed uint64) MCG {
	x := MCG{}
	x.Seed(seed)
	return x
}

// Uint64 returns a  pseudo-random uint64 by MCG mod 2^64.
// The multiplier is picked from Table 6 in Steele & Vigna. Without the
// xor-rotate scrambler, the last bits are not uniformly distributed.
// This is a very fast generator, but not properly tested or proved 
// to give anything good.
// 
func (x *MCG) Uint64() (next uint64) {
	next = x.state ^ bits.RotateLeft64(x.state, 27)
	x.state *= 0x83b5b142866da9d5
	return 
}
// Alternative scrambler
// next = x.state ^ (x.state >> 17)

// Uint64 compiles to 7 instructions + in and out.
// 00000 MOVQ	"".x+8(SP), AX
// 00005 MOVQ	(AX), CX
// 00008 MOVQ	$-8956057384675071531, DX
// 00018 IMULQ	CX, DX
// 00022 MOVQ	DX, (AX)
// 00025 MOVQ	CX, AX
// 00028 ROLQ	$27, CX
// 00032 XORQ	CX, AX
// 00035 MOVQ	AX, "".next+16(SP)

// Uint64Mul uses 128-bit multiplication and the high bits of it.
// 
func (x *MCG) Uint64Mul() (next uint64) {
	hi, lo := bits.Mul64(x.state, 0x83b5b142866da9d5)
	next = hi ^ lo
	x.state = lo
	return 
}
// Uint64Mul compiles to 5 instructions + in and out, but is not faster.
// 00000 MOVQ	"".x+8(SP), CX
// 00005 MOVQ	(CX), AX
// 00008 MOVQ	$-8956057384675071531, DX
// 00018 MULQ	DX
// 00021 MOVQ	AX, (CX)
// 00024 XORQ	AX, DX
// 00027 MOVQ	DX, "".next+16(SP)

// Lehmer64 is pure Lehmer generator.
func (x *MCG) Lehmer64() uint64 {
	x.state *= 0x83b5b142866da9d5
	return x.state
}

// Float64 returns a uniformly distributed pseudo-random float64 from [0, 1).
// The distribution is 2^53 evenly spaced floats with spacing 2^-53.
// Float64 uses multiplicative congruential pseudorandom number generator (MCG) 
// mod 2^64. 53 high bits of the MCG are considered good enough for a fast float64, 
// but they don't pass random tests for the last ~3 bits.
// 
func (x *MCG) Float64() (next float64) {
	next = float64(x.state >> 11) * 0x1p-53
	x.state *= 0x83b5b142866da9d5
    return 
}

// Float64_64 returns a uniformly distributed pseudo-random float64 from [0, 1).
// The distribution includes all floats in [2^-12, 1) and 2^52 evenly spaced 
// floats in [0, 2^-12) with spacing 2^-64.
// 
func (x *MCG) Float64_64() float64 {
	u := x.Uint64()
	if u == 0 { return 0 }  // without this the smallest returned is 2^-65
	z := uint64(bits.LeadingZeros64(u)) + 1
	// z := uint64(65 - bits.Len64(u))
	return math.Float64frombits((1023 - z) << 52 | u << z >> 12)
}
