package prng

import (
	"math"
	"unsafe"
)
const (
	signbit = 1<<63
	signmask = 1<<63 - 1
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

// adjacent(x, y) is equivalent to ulpsBetween(x, y) == 1.
// Special cases:
// adjacent(+/-Inf, +/-MaxFloat64) = true
// adjacent(-2^-1074, -0) = true
// adjacent(-2^-1074, 0)  = false
// adjacent(2^-1074, 0)   = true
// adjacent(0, -0)        = false
func adjacent(x, y float64) bool {
	d := *(*int64)(unsafe.Pointer(&y)) - *(*int64)(unsafe.Pointer(&x))
	return d == 1 || d == -1
}

func adjacentByMean(x, y float64) bool {
	if x == y {
		return false
	}
	mean := x/2 + y/2	// this avoids overflowing x + y to Inf
	if isInfOrNaN(mean) { 
		return false
	}
	return mean == x || mean == y
}

func ulp(x float64) float64 {
	u := math.Float64bits(x) &^ signbit
	return math.Float64frombits(u + 1) - math.Float64frombits(u)
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

// nextToZero(x) is equivalent to math.Nextafter(x, 0)
// Nextafter returns the next representable float64 value after x towards y.
func nextToZero(x float64) (y float64) {
	u := math.Float64bits(x)
	if u &^ signbit > posInf || x == 0 { return x }  // NaNs and 0
	return math.Float64frombits(u - 1)
}

// nextToZeroFast(x) is equivalent to math.Nextafter(x, 0) for floats x in 
// (2^-2022, Maxfloat64]. In (0, 2^-2022] nextToZeroFast(x) fails and return x.
func nextToZeroFast(x float64) (y float64) {
	return x * (1 - 0x1p-53)
}

// nextFromZero(x) is equivalent to math.Nextafter(x, math.Inf(1/-1))
func nextFromZero(x float64) float64 {
	u := math.Float64bits(x)
	if u &^ signbit >= posInf { return x }  // NaNs and +/-Inf
	return math.Float64frombits(u + 1)
}

// float64Random() returns random float64's from [-MaxFloat64, MaxFloat64].
// Every float has the same probability ~2^-64. 64-bit values with eleven 
// '1' bits (7ff) in the exponent part are the only not valid float64 numbers.
func (x *Xoro) float64Random() float64 {
	again:
	u := x.Uint64()
	if u & (0x7ff << 52) == 0x7ff << 52 {  // 1/2048 of cases
		goto again
	}
	return math.Float64frombits(u)
}
