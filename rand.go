package prng

import (
	"sync"
	"time"
)

// Random state sizes in bytes.
const (
	RandStateSize = 16
	XoroStateSize = 16
	XoshStateSize = 32	
)

// A Rand is a wrapper around the actual pseudo-random number generator.
// It is now fixed to xoroshiro128 generator instead of having more flexible rng
// interface. This way we get faster inlineable functions, but cannot change
// the rng in an application.
type Rand struct {
	// rng Xosh	// xoshiro256+/** generator
	rng Xoro // xoroshiro128+/** generator
}

// New returns a new Rand seeded with the seed.
func New(seed uint64) Rand {
	r := Rand{}
	r.rng.Seed(seed)
	return r
}

// NewSource is only for math/rand compability.
func NewSource(seed int64) uint64 {
	return uint64(seed)
}

// Outlet is a delivery type of pseudo-random number generators with
// non-overlapping random streams. Methods of Outlet use sync.Mutex
// to protect the Outlet for simultaneous access.
type Outlet struct {
	mu   sync.Mutex
	xoro Xoro
	xosh Xosh
	rand Rand
}

// NewOutlet returns a new generator delivery Outlet seeded by the seed.
func NewOutlet(seed uint64) *Outlet {
	s := &Outlet{}
	s.xoro.Seed(seed)
	s.xosh.Seed(seed)
	s.rand.Seed(seed)
	return s
}

// Next returns the next Rand from Outlet. Each Rand has 2^64 long
// random stream, which is not overlapping with other Rands streams.
// Next is safe for concurrent use by multiple goroutines.
func (s *Outlet) Next() Rand {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rand.Jump()
	return s.rand
}

// globalOutlet is a delivery type of generators without creating an own Outlet.
// globalOutlet may mostly be the preferred way to deliver.
type globalOutlet struct {
	once   sync.Once
	outlet *Outlet
}

// globalOutlet global is initially seeded by UnixNano time but
// it can be reset once by ResetGlobalOutlet(seed).
var global = globalOutlet{outlet: NewOutlet(uint64(time.Now().UnixNano()))}

// ResetGlobalOutlet resets the globalOutlet by the seed.
// This can be made only once.
func ResetGlobalOutlet(seed uint64) {
	global.once.Do(func() {
		global.outlet = NewOutlet(seed)
	})
}

// Next returns the next non-overlapping stream Rand from globalOutlet.
// Next is safe for concurrent use by multiple goroutines.
func Next() Rand {
	return global.outlet.Next()
}

// NextXosh returns the next non-overlapping stream xoshiro256 from
// globalOutlet. This generator has only Float64 and Uint64 prn methods.
func NextXosh() Xosh {
	return global.outlet.NextXosh()
}

// NextXoro returns the next non-overlapping stream xoroshiro128 from
// globalOutlet. This generator has only Float64 and Uint64 prn methods.
func NextXoro() Xoro {
	return global.outlet.NextXoro()
}

// NewRandSlice returns a slice of n Rands with non-overlapping
// random streams. The first Rand is seeded by seed.
func NewRandSlice(n int, seed uint64) []Rand {
	s := make([]Rand, n)
	s[0].Seed(seed)
	for i := 1; i < n; i++ {
		s[i] = s[i-1]
		s[i].Jump()
	}
	return s
}

// A Rand's rng has methods Float64, Uint64, Jump and Seed.
// Rand & math/rand functions are defined below.

// Seed seeds a Rand by the seed. Any seed is ok.
// Do not seed Rands created by Next or NewRandSlice.
func (r *Rand) Seed(seed uint64) {
	r.rng.Seed(seed)
}

// Jump sets r to the same state as 2^64 calls to r.Uint64.
// Jump can be used to generate 2^64 non-overlapping subsequences for
// parallel computation.
func (r *Rand) Jump() {
	r.rng.Jump()
}

// State returns the current state of the generator r as []byte.
func (r *Rand) State() []byte {
	return r.rng.State()
}

// SetState sets the state of the generator r from the state in b []byte.
func (r *Rand) SetState(b []byte)  {
	r.rng.SetState(b)
}

// Float64 returns a uniformly distributed pseudo-random float64 value in [0, 1).
func (r *Rand) Float64() float64 {
	return r.rng.Float64()
}

// Uint64 returns a pseudo-random uint64.
func (r *Rand) Uint64() uint64 {
	return r.rng.Uint64()
}

// Int63 returns a non-negative pseudo-random int64.
func (r *Rand) Int63() int64 {
	return int64(r.rng.Uint64() >> 1) //take high bits
}

// Int returns a non-negative pseudo-random int.
func (r *Rand) Int() int {
	return int(r.rng.Uint64() >> 1)
}

// Uint64n returns a pseudo-random number in [0,n) as an uint64.
// Uint64n doesn't make any bias correction. The bias with 64-bit numbers is very small
// and propably not detectable from the random stream.
func (r *Rand) Uint64n(n uint64) uint64 {
	if n == 0 {
		panic("invalid argument to Uint64n")
	}
	return r.rng.Uint64() % n
}

// Int63n return a pseudo-random number in [0,n) as an int64.
func (r *Rand) Int63n(n int64) int64 {
	if n <= 0 {
		panic("invalid argument to Int64n")
	}
	return int64((r.rng.Uint64() % uint64(n)) &^ (1 << 63))
}

// Intn returns a pseudo-random number in [0,n) as an int.
func (r *Rand) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	return int((r.rng.Uint64() % uint64(n)) &^ (1 << 63))
}

// The top level non method functions using system generator globalRand -----------//
// globalRand is initially seeded by UnixNano time. It can be reseeded
// by Seed function. These functions are not safe for concurrent use by
// multiple goroutines.

var globalRand = New(uint64(time.Now().UnixNano()))

// Seed seeds system global generator globalRand by seed.
func Seed(seed uint64) {
	globalRand.rng.Seed(seed)
}

// Float64 returns a uniformly distributed pseudo-random float64 value in [0, 1).
func Float64() float64 {
	return globalRand.rng.Float64()
}

// Uint64 returns a pseudo-random 64-bit value as an uint64
func Uint64() uint64 {
	return globalRand.rng.Uint64()
}

// Int63 returns a non-negative pseudo-random int64.
func Int63() int64 {
	return int64(globalRand.rng.Uint64() >> 1)
}

// Int returns a non-negative pseudo-random int.
func Int() int {
	return int(globalRand.rng.Uint64() >> 1)
}

// Uint64n returns a pseudo-random number in [0,n) as an uint64.
func Uint64n(n uint64) uint64 {
	if int64(n) <= 0 {
		panic("invalid argument to Uint64n")
	}
	return globalRand.rng.Uint64() % n
}

// Int63n return a pseudo-random number in [0,n) as an int64
func Int63n(n int64) int64 {
	if n <= 0 {
		panic("invalid argument to Int63n")
	}
	return int64((globalRand.rng.Uint64() % uint64(n)) &^ (1 << 63))
}

// Intn returns a pseudo-random number in [0,n)  as an int.
func Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	return int((globalRand.rng.Uint64() % uint64(n)) &^ (1 << 63))
}

// Splitmix is a 64-bit state SplitMix64 pseudo-random number generator
// from http://prng.di.unimi.it/splitmix64.c .
// The pointer parameter seed is used as the random state.
// Splitmix is used here to blend seeds for the other generators.
// It is a good and quite fast 64-bit state generator for other uses too.
func Splitmix(seed *uint64) uint64 {

	*seed += 0x9e3779b97f4a7c15 //any seed is ok
	z := *seed
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// SplitmixJump (&seed, jump) sets the seed  to the same value
// as jump calls to Splitmix(&seed). Jump can be negative.
func SplitmixJump(seed *uint64, jump int64) {
	if jump >= 0 {
		*seed += uint64(jump) * 0x9e3779b97f4a7c15
		return
	} 
	*seed -= uint64(-jump) * 0x9e3779b97f4a7c15
}

// OverlapProbability function calculates lower and upper bound of the 
// probability for an event that at least two random streams overlap when 
// splitting a single prng by random seeding. 
// Formulas from http://vigna.di.unimi.it/ftp/papers/overlap.pdf.
// n = number of splitted parallel prng's
// L = lenght of the random stream for each prng
// P = full period of the prng.
func OverlapProbability(n, L, P float64) (lower, upper float64){

	upper = n*(n-1) * L / P
	lower = n*(n-1) * (L-1) / P * (1 - n*n*L / (2*P))
	return
}

// Baseline128 is for benchmarking minimal 128-bit state generator.
func (x *Xoro) Baseline128() uint64 {
	next := x.s0
	*x = Xoro{x.s1, x.s0}
	return next
}

// Baseline256 is for benchmarking minimal 256-bit state generator.
func (x *Xosh) Baseline256() uint64 {
	next := x.s0
	*x = Xosh{x.s3, x.s0, x.s1, x.s2}
	return next
}

