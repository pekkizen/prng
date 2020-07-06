package prng

import (
	"math"
	"math/bits"
	"unsafe"
)
const (
	signbit = 1<<63
	posInf = 0x7FF0000000000000
	maxUint64 = 1<<64 - 1
)

// ulpsBetween returns the distance between x and y in ULPS.
// The distance in ULPS is the same as the difference of the ordinal 
// numbers of x and y in the sorted sequence of all floats.
//
// Special cases:
// x and/or y = +/-Inf or NaN -> ulpsBetween(x, y) = maxUint64 (2^64 - 1).
func ulpsBetween(x, y float64) (u uint64) {

	// uncomment if ulpsBetween(Inf, Inf) = 0 is wanted.
	// if x == y {
	// 	return 0
	// }

	k := *(*uint64)(unsafe.Pointer(&x)) 
	n := *(*uint64)(unsafe.Pointer(&y)) 
	// k := math.Float64bits(x)
	// n := math.Float64bits(y)

	diffsign := (k ^ n) & signbit != 0
	k &^= signbit 
	n &^= signbit 

	switch  {
	case k >= posInf || n >= posInf: // Infs and NaNs
		u = maxUint64	
	case diffsign:
		u = n + k
	case n > k:
		u = n - k
	default:
		u = k - n
	}
	return
}

// adjacent(x, y) returns true, if x and y are adjacent floats.
// It is equivalent to ulpsBetween(x, y) == 1.
// Special cases:
// adjacent(+/-Inf, +/-MaxFloat64) = true
// adjacent(-0, -2^-1074) = true
// adjacent(0, 2^-1074)   = true
// adjacent(-0, 0)        = false
// adjacent(0, -2^-1074)  = false,  these two are the only failures this far
// adjacent(-0, 2^-1074)  = false
func adjacent(x, y float64) bool {
	d := int64(*(*uint64)(unsafe.Pointer(&y)) - *(*uint64)(unsafe.Pointer(&x)))
	return d == 1 || d == -1
}

// adjacentByMean returns true, if x and y are adjacent floats.
// It is slightly slower than adjacent but doesn't fail at zero.
// Special cases: (differ from adjacent)
// adjacent(+/-Inf, +/-MaxFloat64) = false
func adjacentByMean(x, y float64) bool {
	const maxf = math.MaxFloat64 
	if x == y {
		return false
	}
	mean := x/2 + y/2	// this avoids overflowing x + y to Inf
	return (mean == x || mean == y) && -maxf <= mean && mean <= maxf // NaN and Inf -> false
}

// ulp(x) returns the ULP of x as a positive float64. The ULP is calculated as
// the distance to the next float64 away from zero. 
// Special cases:
// ulp(+/-Inf) = NaN
// ulp(NaN)    = NaN
func ulp(x float64) float64 {
	u := math.Float64bits(x) &^ signbit
	exp := u >> 52
	if exp == 0x7ff { return math.NaN() }
	if exp > 52 {
		return math.Float64frombits((exp - 52) << 52)
	}
	return math.Float64frombits(u + 1) - math.Float64frombits(u)
}

// log2Ulp(x) returns log2(ulp(x)) as an int.
// All ULPs are exact powers of two -> 
// normal values have a significand = 0 and 
// subnormal values have only a single significand bit.
// Special cases:
// log2Ulp(+/-Inf) = 1024, 2^1024 = +Inf 
// log2Ulp(NaN)    = 1024
func log2Ulp(x float64) int {
	return log2pow2(ulp(x))
}

// log2pow2 return log2(x) as an int. If x is an integer power log2pow2(2^n) = n.
// Special cases:
// log2pow2(2^n.ddd..) = n    (truncated value for non integer powers)
// log2pow2(+/-Inf)    = 1024
// log2pow2(NaN)       = 1024
func log2pow2(x float64) int {
	u := math.Float64bits(x)
	exp := int(u >> 52)
	if exp == 0 {                    // u is subnormal significand
		return bits.Len64(u) - 1075  // Len64(u=2^n) = n + 1, n = 0 - 51
	}
	return exp - 1023
}

// ulpFloat(x) returns the ULP of x for all normal floats abs(x) > 0x1p-2022.
// See nextToZeroFloat for x * (1 - 0x1p-53)
// Special cases:
// ulpFloat(+/-Inf) = +/-Inf
// ulpFloat(NaN)    = NaN
// For abs(x) <= 0x1p-2022 ulpFloat fails and returns 0.
func ulpFloat(x float64) (y float64) {
	y = x - x * (1 - 0x1p-53)
	// y = x * 0x1p-53  // doesn't work. Needs rounding of x * (1 - 0x1p-53).
	if y < 0 { y = -y } 
	return 
}

// isPowerOfTwo returns true if float64 x is integer power of two, x = 2^n.
func isPowerOfTwo(x float64) (is bool) {
	u := math.Float64bits(x) &^ signbit
	exp := int(u >> 52)
	sig := u &^ (0x7ff << 52)
	switch {
	case exp == 0x7ff || u == 0:
		is = false
	case sig > 0 && exp > 0:	
		is = false
	case sig == 0:	                            // normal floats					
		is = true								// sig = 0 and exp > 0	
	case sig & (sig - 1) == 0:	                // subnormal floats					
		is = true								// exp = 0 and sig is power of two
	}
	return 
}

func isInf(x float64) bool {
	return *(*uint64)(unsafe.Pointer(&x)) &^ signbit == posInf
}

func isInfOrNaN(x float64) bool {
	return *(*uint64)(unsafe.Pointer(&x)) &^ signbit >= posInf
}

func isNaN(x float64) bool {
	return *(*uint64)(unsafe.Pointer(&x)) &^ signbit > posInf
}

// nextToZero returns the next float64 after x towards zero.
// nextToZero(x) is equivalent to math.Nextafter(x, 0)
func nextToZero(x float64) float64 {
	u := math.Float64bits(x)
	if u &^ signbit > posInf || x == 0 { return x }  // NaNs and 0
	return math.Float64frombits(u - 1)
}

// nextToZeroFloat is equivalent to nextToZero for floats x in 
// (2^-2022, Maxfloat64]. In (0, 2^-2022] nextToZeroFloat fails and return x.
// Constant 1 - 0x1p-53 converts to the next float from 1 towards zero. 
// Float64bits(1):           3FF0000000000000.
// Float64bits(1 - 0x1p-53): 3FEFFFFFFFFFFFFF.
func nextToZeroFloat(x float64) (y float64) {
	return x * (1 - 0x1p-53)
}

// nextFromZero returns the next float64 after x away from zero.
// nextFromZero(x) is equivalent to math.Nextafter(+/-x, math.Inf(1/-1))
func nextFromZero(x float64) float64 {
	u := math.Float64bits(x)
	if u &^ signbit >= posInf { return x }  // NaNs and +/-Inf
	return math.Float64frombits(u + 1)
}

// float64Random() returns random float64's from [-MaxFloat64, MaxFloat64].
// Every float has the same probability ~2^-63.999. 64-bit values with eleven 
// '1' bits (7ff) in the exponent part are the only not valid float64 bit 
// representations.
func (x *Xoro) float64Random() float64 {
	again:
	u := x.Uint64()
	if u & (0x7ff << 52) == (0x7ff << 52) {  
		goto again                        // resample 1/2048 of cases
	}
	return math.Float64frombits(u)
}
