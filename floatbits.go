package prng

import (
	"math"
	"math/bits"
)

const (
	signbit   = 1<<63
	posInf    = 0x7ff0000000000000
	maxUint64 = 1<<64 - 1
)

// ulpsBetween returns the distance between x and y in ULPS.
// 
// The distance in ULPS is the number of float64's between x and y - 1.
// Special cases:
// ulpsBetween(+/-Inf, +/-MaxFloat64) = 1
// ulpsBetween(-0, 2^-1074)           = 1
// ulpsBetween(+/-Inf, +/-Inf)        = 0
// ulpsBetween(-Inf, +Inf)            = maxUint64 - 2^53 + 1
// ulpsBetween(x, NaN)                = maxUint64 
// 
func ulpsBetween(x, y float64) (u uint64) {
	k := math.Float64bits(x)
	n := math.Float64bits(y)
	signdiff := k ^ n >= signbit
	k &^= signbit 
	n &^= signbit 
	switch {
	case k > posInf || n > posInf:  // NaNs 
		u = maxUint64	
	case signdiff:
		u = n + k
	case n > k:
		u = n - k
	default:
		u = k - n
	}
	return
}

// adjacent(x, y) returns true, if x and y are adjacent floats.
// 
// adjacent(x, y) is a faster equivalent to ulpsBetween(x, y) == 1.
// For x > 0 and y > x  adjacent(x, y) == (math.Nextafter(x, y) == y)
// Special and other cases:
// adjacent(+Inf, +MaxFloat64) = true
// adjacent(-Inf, -MaxFloat64) = true
// adjacent(x, NaN)            = false
// adjacent(-0, -2^-1074)      = true
// adjacent(0, 2^-1074)        = true
// adjacent(-0, 0)             = false
// adjacent(0, -2^-1074)       = false,  A special case failure and
// adjacent(-0, 2^-1074)       = false   also this. Only failures.
// 2^-1074 is the smallest nonzero float64.
// 
func adjacent(x, y float64) bool {

	d := int64(math.Float64bits(x) - math.Float64bits(y))
	return d == 1 || d == -1 
	// or
	// d := math.Float64bits(x) - math.Float64bits(y)
	// return d == 1 || d == maxUint64               
}

// adjacentFP returns true, if x and y are adjacent floats.
// 
// Only floating-point operations are used.
// This is ~35% (0.3 ns) slower than adjacent but doesn't fail at zero.
// Special cases different from func adjacent:
// adjacentFP(+Inf, +MaxFloat64) = false
// adjacentFP(-Inf, -MaxFloat64) = false
// adjacentFP(0, -2^-1074)       = true
// adjacentFP(-0, 2^-1074)       = true
// 
func adjacentFP(x, y float64) bool {
	if x == y {
		return false
	}
	mean := x/2 + y/2                // this avoids overflowing x + y to Inf
	if mean != x && mean != y {
		return false
	}
	return -math.MaxFloat64 <= mean && mean <= math.MaxFloat64 
}

// ulp returns the ULP of x as a positive float64. 
// 
// A ULP returned is the distance to the next float64 away from zero,
// which also means that two floats with a same exponent have a same ULP.
// If x is a power on two, ULP(x) towards zero is ULP(x)/2 away from zero. 
// All ULPs are exact powers of two -> 
//   normal values have a significand = 0 and 
//   subnormal values have only a single significand bit.
// Special cases:
// ulp(+/-Inf) = +Inf
// ulp(NaN)    = NaN
// 
func ulp(x float64) float64 {
	u := math.Float64bits(x) &^ signbit
	exp := u >> 52
	switch {
	case exp == 0x7ff:
	case exp > 52:
		u = (exp - 52) << 52
	case exp > 1:
		u = 1 << (exp - 1)
	default:
		u = 1                // x < 2^-2021, ULP = 2^-1074
	}
	return math.Float64frombits(u)  
}
// logUlp returns log2(ulp(x)) as an int, ulp(x) = 2^logUlp(x).
// Special cases:
// logUlp(+/-Inf) = 1024    (2^1024 = +Inf)
// logUlp(NaN)    = 1024
// 
func logUlp(x float64) (exp int) {
	exp = int(math.Float64bits(x) &^ signbit >> 52)
	switch {
	case exp == 0x7ff:
		exp = 1024
	case exp >= 1:
		exp -= (1023 + 52)
	default:
		exp = -1074        
	}
	return  
}

// A slightly faster version of ulp with FP subtraction.
// Depends on adding 1's to Inf and NaN bits.
// ulp(+/-Inf)        = NaN
// ulp(+/-MaxFloat64) = +Inf !
// 
func ulpSub(x float64) float64 {
	u := math.Float64bits(x) &^ signbit
	return math.Float64frombits(u + 1) - math.Float64frombits(u)
}

// ulpFP returns the ULP of x for abs(x) > 0x1p-1022.
// 
// A ULP is calculated as a difference towards zero.
// Special cases:
// ulpFP(+/-Inf) = NaN
// ulpFP(NaN)    = NaN
// For abs(x) <= 0x1p-1022 ulpFP fails and returns 0.
// 
func ulpFP(x float64) (y float64) {
	y = x - x * (1 - 0x1p-53)             // y = x - nextToZeroFP(x)
	if y < 0 { y = -y } 
	return 
}



// log2pow2 return log2(abs(x)) as an int. 
// 
// If x is an integer power 2^n log2pow2 returns n.
// Special cases:
// log2pow2(2^n.ddd..) = n    (truncated value for non integer powers)
// log2pow2(+/-Inf)    = 1024
// log2pow2(NaN)       = 1024
// log2pow2(-x)        = log2pow2(x)
// 
func log2pow2(x float64) int {
	u := math.Float64bits(x) &^ signbit
	exp := int(u >> 52)
	if exp == 0 {                        // x is subnormal 
		return bits.Len64(u) - 1075      // Len64(u=2^n) = n + 1, n = 0 - 51
	}
	return exp - 1023                    // x is normal
}

// isPowerOfTwo returns true if float64 x is an integer power of two.
// 
// Cases of interest:
// isPowerOfTwo(1)      = true
// isPowerOfTwo(-1)     = false
// isPowerOfTwo(0)      = false
// isPowerOfTwo(+/-Inf) = false
// isPowerOfTwo(NaN)    = false
// 
func isPowerOfTwo(x float64) bool {
	s := math.Float64bits(x) << 12            // 52 significand bits + zeros                
	if s & (s - 1) > 0 {                      // there are only 2046 + 52
		return false                          // power of 2 float64's
	}
	e := math.Float64bits(x) >> 52            // sign bit + 11 exponent bits                   
	return ((s > 0) != (e > 0)) && e < 0x7ff

	// A float64 value x is a power of two if and only if the following 
	// conditions are met:
	//     s & (s - 1) == 0     -> significand is zero or power of two
	//     (s > 0) != (e > 0)   -> significand or exponent is zero, but not both
	//     e < 0x7ff            -> x is not +/-Inf, NaN or negative
	// Above e > 0 is true for a negative x, but the last condition drops this out.

	// An "prettier" but ~25% (0.16 ns) slower equivalent function.
	// s := math.Float64bits(x) 
	// e := s >> 52                                               
	// s <<= 12                                    
	// return s & (s - 1) == 0 && ((s > 0) != (e > 0)) && e < 0x7ff

	// s <<= 12 is faster than masking s &= (1<<52)-1 !? 
	// The position of the bits is not relevant here.
}

// https://stackoverflow.com/questions/27566187/code-for-check-if-double-is-a-power-of-2-without-bit-manipulation-in-c
// This is without bit operations and seems to work, but is over 50% slower than isPowerOfTwo
func isPowerOfTwoFP(x float64) bool { 
	return x > 0 && math.FMA(0x1.0p-51/x, x, -0x1.0p-51) == 0 
	// return x > 0 && 0x1.0p-51/x * x - 0x1.0p-51 == 0 // doesn't work
}

// Java DoubleUtils.isPowerOfTwo(double x) from com.google.common.math.
// https://www.codota.com/code/java/classes/com.romainpiel.guava.math.DoubleUtils
// This checks twice both x > 0 and isFinite(x).
// 
// public static boolean isPowerOfTwo(double x) {
//  return x > 0.0 && isFinite(x) && LongMath.isPowerOfTwo(getSignificand(x));
// }
// public static boolean isPowerOfTwo(long x) {
//     return x > 0 & (x & (x - 1)) == 0;
// }
// DoubleUtils.getSignificand(...)
// static long getSignificand(double d) {
//  checkArgument(isFinite(d), "not a normal value");
//  int exponent = getExponent(d);
//  long bits = doubleToRawLongBits(d);
//  bits &= SIGNIFICAND_MASK;
//  return (exponent == MIN_EXPONENT - 1)
//    ? bits << 1
//    : bits | IMPLICIT_BIT;
// }
// Java implementation's call/dependancy tree:
// DoubleUtils.isPowerOfTwo
//     isFinite
//     LongMath.isPowerOfTwo
//     getSignificand
//         checkArgument
//             isFinite
//         getExponent
//         doubleToRawLongBits
//         SIGNIFICAND_MASK
//         MIN_EXPONENT
//         IMPLICIT_BIT
	
// isPowerOfTwoJava implements DoubleUtils.isPowerOfTwo. 
// The bare algorithm with the same functionality.
// This small and simple function is still over 50% slower than isPowerOfTwo above.
// 
func isPowerOfTwoJava(x float64) bool {
	bits := math.Float64bits(x)     // bits = doubleToRawLongBits(x)
	exp := bits >> 52               // exponent = getExponent(x) 
	bits &= 1<<53 - 1               // bits &= SIGNIFICAND_MASK
	if exp > 0  {                   // not (exponent == MIN_EXPONENT - 1)
		bits |= 1<<52               // bits | IMPLICIT_BIT (this is the point in the algorithm)
	}
	return bits & (bits - 1) == 0 && bits > 0 && exp < 0x7ff // isPowerOfTwo(bits) & isFinite(x) & x > 0.
}

func isInf(x float64) bool {
	// return x < -math.MaxFloat64 || x > math.MaxFloat64
	return math.Float64bits(x) &^ signbit == posInf
}

func isFinite(x float64) bool {
	return math.Float64bits(x) &^ signbit < posInf
}

func isNaN(x float64) bool {
	// return math.Float64bits(x) &^ signbit > posInf
	return x != x
}

// nextToZero returns the next float64 after x towards zero.
// 
// nextToZero(x) is equivalent to math.Nextafter(x, 0)
// Special cases:
// nextToZero(+/-Inf) = +/-MaxFloat64 
// nextToZero(NaN)    = NaN
// nextToZero(0)      = 0
// nextToZero(-0)     = -0
// 
func nextToZero(x float64) float64 {
	u := math.Float64bits(x)
	if u << 1 == 0 { return 0 }  // this is faster than if x == 0
	// if u &^ signbit > posInf || x == 0 { return x }  // NaNs and 0
	return math.Float64frombits(u - 1)
}

// nextToZeroFP is equivalent to nextToZero for floats abs(x) > 2^-1022. 
// In (0, 2^-1022] nextToZeroFP(x) fails and returns x.
// Constant 1 - 0x1p-53 converts exactly to the next float64 from 1 towards zero. 
// Float64bits(1):           3FF0000000000000.
// Float64bits(1 - 0x1p-53): 3FEFFFFFFFFFFFFF.
// Float64bits(1 + 0x1p-53): 3FF0000000000000.
// Float64bits(1 + 0x1p-52): 3FF0000000000001.
// Going away from zero needs more complicated formula than x * (1 + 0x1p-52) and
// it can't win func nextFromZero,
// 
func nextToZeroFP(x float64) float64 {
	return x * (1 - 0x1p-53)
}

// nextFromZero returns the next float64 after x away from zero.
// 
// nextFromZero(+/-abs(x) is equivalent to math.Nextafter(+/-abs(x), math.Inf(1/-1)).
// nextFromZero is as fast as Go Math.Abs.
// Special cases:
// nextFromZero(+/-Inf) = Nan  
// nextFromZero(NaN)    = NaN
// 
func nextFromZero(x float64) float64 {
	// u := math.Float64bits(x)
	// if u &^ signbit >= posInf { return x }  // NaNs and +/-Inf
	// return math.Float64frombits(u + 1)
	return math.Float64frombits(math.Float64bits(x) + 1)

	// For a very fast function use the code line below if nextFromZero(Inf) = NaN
	// is ok and mixing NaNs (quiet and signaling) is no problem.
	// 
	// return math.Float64frombits(math.Float64bits(x) + 1)
	// 
	// This returns a NaN with a bit representation increased by 1.
	// amd64 CPU/Go seems to take as NaN any bit representation greater than 
	// +Inf = 0x7FF0000000000000, e.g 0x7FF0000000000001 is handled as a NaN
	// in computations. If the CPU produces a NaN, it is in the 
	// standard (quiet NaN) format     0x7FF8000000000001.
	// nextFromZero(NaN) returns bits  0x7FF8000000000002.
	// nextFromZero(+Inf) returns bits 0x7FF0000000000001.
}




