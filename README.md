
## PRNG

Package prng has methods of delivering pseudo-random number generators concurrent
safely for multiple goroutines for large scale parallel computations.
Additionally, prng implements a set of 64-bit pseudo-random number functions with the same API as standard library math/rand.
For these functions you can import rand "github.com/pekkizen/prng" instead of "math/rand". Prng functions are faster 
but not safe for concurrent use.  Package prng is still experimental and there is
no guarantee of backward compatibility.

Package prng uses Xoroshiro128 and xoshiro256 random generators and jump functions from 
Dipartimento di Informatica Università degli Studi di Milano.
Written by David Blackman and Sebastiano Vigna and licensed under 
http://creativecommons.org/publicdomain/zero/1.0/. Background: [*Scrambled Linear Pseudorandom Number Generators* by David Blackman and Sebastiano Vigna.](http://vigna.di.unimi.it/ftp/papers/ScrambledLinear.pdf) Package prng functions are adapted from the C-source code from http://prng.di.unimi.it.  

#### The authors recommendations for generator use

- _Xoroshiro128+ 1.0 is our best and fastest small-state generator
   for floating-point numbers. We suggest to use its upper bits for
   floating-point generation, as it is slightly faster than
   xoroshiro128**._

 - _Xoroshiro128** 1.0 is one of our all-purpose, rock-solid,
   small-state generators. ... it passes all tests we are aware of,   
   but its state space is large enough only for mild parallelism._

- _Xoshiro256+ 1.0 is our best and fastest generator for floating-point
   numbers. We suggest to use its upper bits for floating-point
   generation, as it is slightly faster than xoshiro256++/xoshiro256**._

- _Xoshiro256** 1.0 is one of our all-purpose, rock-solid generators. 
   It has excellent (sub-ns) speed, a state (256 bits) that is
   large enough for any parallel application, and it passes all tests we
   are aware of._

Package prng functions Float64 are implemented by the suggested + generators and Uint64 by ** generators. 
xoroshiro128+ and **  have the same linear engine. Also xoshiro256+ and **. 
So, the same random state receiver variable can used for floats and uints without
disturbing random stream properties. The generators can be used as a random source 
for github.com/golang/exp/rand. Functions for random floating-point numbers are
documented in https://github.com/pekkizen/prng/wiki/floats.


### Benchmarking generator speeds
The original C functions have been modified to more efficient Go code, 
especially for xoshiro256. The baseline 128-bit state minimal "random" generator below was used
to get baseline time reference for returning a result and updating the state.
An analogous function was used for 256-bit baseline. The functions NexState returns the
next state of the linear engine.
The tests were run on a standard Windows 10 pro tablet PC with Intel i7-1065G7 CPU running 
benchmarks @ ~3.5 GHz.  Standard Go compiler version 1.13.8 was used. For reference two
PCG, two xoshiro256 implementations and math/rand functions were included.    

```Go
func (x *Xoro) Baseline128() uint64 {
	result := x.s0 
	*x = Xoro {
		s0: x.s1,
		s1: x.s0,
	}
	return result
}
```

#### Time (ns) for baseline reference functions
|  Function                     | Time    |  
|-------------------------------|---------|
| Empty loop                    | 0.27    |        
| (1) Baseline128               | 0.45    |    
| (1) Baseline256               | 0.71    |        
| (1) NextState128              | 0.87    |  
| (1) NextState256              | 1.08    |       

#### Time (ns) to generate an uint64 
|     Generator                 | Time    | 
|-------------------------------|---------|
| (1) xoroshiro128+             | 0.93    |
| (1) xoroshiro128**            | 1.07    |
| (1) SplitMix64                | 1.11    |
| (1) xoshiro256+               | 1.19    | 
| (5) xoshiro256+               | 1.34    | 
| (1) xoshiro256**              | 1.34    | 
| (1) xoshiro256++              | 1.35    |  
| (3) PCG                       | 1.90    | 
| (2) xoshiro256+               | 2.43    | 
| (2) xoshiro256**              | 2.44    | 
|  math/rand rng.Int63()        | 2.62    |      
| (1/3) 128**/Source interface  | 1.86    |     
| (3) PCG/Source interface      | 2.70    |     
| (4) PCG                       | 3.40    | 

#### Time (ns) to generate a float64 in [0, 1)
|       Generator               | Time    |
|-------------------------------|---------|
| (1) Float64/xoroshiro128+     | 1.34    | 
| (1) Float64/xoshiro256+       | 1.63    |
| (1) Float64_64/xoroshiro128** | 2.70    | 
| (1) Float64_1024/xoroshiro128** | 2.80    | 
| math/rand rng.Float64()       | 2.88    |
| (3) rng.Float64()             | 4.92    |

(1) github.com/pekkizen/prng   
(2) gonum.org/v1/gonum/mathext/prng      
(3) github.com/golang/exp/rand     
(4) github.com/MichaelTJones/pcg    
(5) github.com/vpxyz/xorshift/xoroshiro256plus     

The tables were calculated by the benchmark function below. The benchmark loop was run
10 - 25 x 10^9 times, so that each benchmark lasted ~30 s. Between the individual benchmarks
a 4 minutes cooling timeout was kept. In 30 seconds, the CPU did not seem to cumulate heat enough to
set any thermal control slow down in effect. If the result u is not carried out of the benchmark for loop, 
the Go compiler optimizes its calculation away from the inlined function code. 

```Go
var usink uint64
func BenchmarkMethod(b *testing.B) {
    var u uint64
    x := <initialized receiver>
    for n := 0; n < b.N; n++ {
        u = x.<Method>
    }
    usink = u
}
```
The results somewhat differ from the times given in prng.di.unimi.it. Most remarkably
xoroshiro128+/** are now clearly faster than xoshiro256+/**. The differences may be related to
the random state updating: C/C++ has static state variables and in Go you must update by a pointer 
referencing the state variable. If the state variable is declared outside of the benchmark function,
the times increases over 1 ns. The state variable in stack vs heap. These benchmarks measure
Go functions implementing prng algorithms, not C-functions or prng algorithms.

### Jump functions 

Xoro/Xoshiro generator is a combination of a scrambler and a linear engine. The linear engine is a 
linear generator, which
> *have several advantages: they are
fast, it is easy to create full-period generators with large state spaces, and thanks to their connection
with linear-feedback shift registers (LFSRs) [18] many of their properties, such as full period, are
mathematically provable. Moreover, if suitably designed, they are rather easy to implement using
simple xor and shift operations. In particular, Marsaglia [31] introduced the family of xorshift
generators, which have a very simple structure* (Blackman and Vigna).

A scrambler is a nonlinear function that reduces or deletes the linear artifacts of 
the state array of the linear engine.
From the linear engine properties follows, that it is possible to create jump functions
to roll the linear engine forward for a desired number of steps in constant time. 
Package prng has for xoroshiro128+/** the jump methods:    
- x.JumpShort sets x to the same state as 2^32 calls to x.Uint64.    
- x.Jump sets x to the same state as 2^64 calls to x.Uint64 or 2^32 calls to x.JumpShort   
- x.JumpLong sets x to the same state as 2^96 calls to x.Uint64 or 2^32 calls to x.Jump   

prng_test.go has test functions, which prove that the jump functions above actually work, exactly. 
By jump functions it is easy to generate  non-overlapping subsequences for parallel computations.

### Implementing concurrent safe delivery of generators with non-overlapping random streams    

Below is a stripped version of the full code. The main concept is type Outlet, which is a mutex
protected source of random generators. Outlet has Next() method, which returns a generator after
a jump from the previous generator.
Type Rand is just a light wrapper around the actual generator. Xoroshiro128 and 
xoshiro256 can also be used directly, if Rands extra methods are not needed. 

```Go
type Rand struct {
    // rng Xosh //xoshiro256
    rng Xoro //xoroshiro128 
}
type Outlet struct {
    mu    sync.Mutex
    rand  Rand
}
func NewOutlet(seed uint64) *Outlet {
    s := &Outlet{}
    s.rand.Seed(seed)
    return s
}
func (s *Outlet) Next() Rand {
    s.mu.Lock()
    defer s.mu.Unlock()
	
    s.rand.Jump()
    return s.rand
}
```
The globalOutlet global implements the delivery of generators without creating an own Outlet. It is initialized by
UnixNano time but can be reset once by a seed. 
```Go
type globalOutlet struct {
    once    sync.Once
    outlet  *Outlet
}
var global = globalOutlet {outlet: NewOutlet(uint64(time.Now().UnixNano()))    

func ResetGlobalOutlet(seed uint64) {
    global.once.Do ( func() {
        global.outlet = NewOutlet(seed)
    })
}
func Next() Rand {
    return global.outlet.Next()
}
```
Function NewRandSlice returns a slice of n generators. It can be used to create the generators faster in batch.
```Go
func NewRandSlice(n int, seed uint64) []Rand {
    s := make([]Rand, n)
    s[0].Seed(seed)
    for i := 1; i < n; i++ {
        s[i] = s[i-1]
        s[i].Jump()
    }
    return s
}
```
### Providing generators to goroutines
Each worker function retrieves a generator from the globalOutlet.
```Go
func worker() { 
    myPrivateNonOverlappingGenerator := prng.Next()  
    ... 
}
```
As a parameter. A Rand is a value type and each worker gets a local copy of the Rand.
A Rand is only 2 or 4 x uint64 of data. As a value type a Rand is as concurrent safe as
any other value variable, so far you don't use global Rands and don't pass pointers to a same Rand
to concurrent functions. You can pass a single Rand as a value parameter to multiple concurrent 
functions, but all the passed copy Rands have the same random stream.

```Go
func worker(r Rand) { ... }

go worker(prng.Next())
```
Putting a lot of workers go fast to work.
```Go
workers := 1000000
rng := prng.NewRandSlice(workers, 1)
for i := 0; i < workers; i++ {
    go worker(rng[i])
}
```
Creating a batch of generators as a binary file of generator states.
```Go
x := prng.New(1)
rngs := 1000000
var statebytes []byte
for i := 0; i < rngs; i++ {
    statebytes = append(statebytes, x.State()...)
    x.Jump()
}
WriteFile("statefile", statebytes)
```
Setting up generators from a saved generator state file.
```Go
func worker(me int, statebytes []byte) {
    myRng := prng.Rand{}
    myRng.SetState(statebytes[me * prng.RandStateSize:])
    ...
}
statebytes := ReadFile("statefile")
workers := 1000000
for i := 0; i < workers; i++ {
    go worker(i, statebytes)
}
```

Providing random seeded generators, if possible overlapping random streams are no issue. 
Creating a seeded xoro/xoshiro generator is faster than using a jump. A jump is not remarkably slow but
still takes 170 ns for xoroshiro128 and 270 ns for xoshiro256. Creating a seeded xoroshiro128 takes 3.2 ns and a xoshiro256 5 ns.
Creating a seeded math/rand generator with 607 x 64-bit state takes ~10000 ns.
Seeding with index i effectively is 64-bit pseudo random seeding, because all seeding goes thru SplitMix64 prng.

```Go
workers := 100
for i := 0; i < workers; i++ {
    go worker(prng.New(uint64(i)))
}
```
Function OverlapProbability calculates lower and upper bound of the probability for an event 
that at least two random streams overlap when splitting a single prng by **random** seeding.
Formulas from [*On the probability of overlap of random subsequences of pseudorandom 
number generators*](http://vigna.di.unimi.it/ftp/papers/overlap.pdf).

```Go
func OverlapProbability(n, L, P float64) (lower, upper float64)
    n = processes/number of splitted parallel prngs
    L = length of the random stream for each prng
    P = full period of the prng.

```
### Package prng API functions and methods


#### Random number functions
Functions are not safe for concurrent use.
Functions Int63n, Intn and Uint64n don't make any bias correction. The bias with 
64-bit numbers is very small and probably not detectable from the random stream.
All functions are also implemented as methods of type Rand. A single Rand should not be shared concurrently.
```Go
func Uint64() uint64
    Uint64 returns a pseudo-random uint64.
```
```Go
func Uint64n(n uint64) uint64
    Uint64n returns a pseudo-random number in [0,n) as an uint64.
```
```Go
func Int() int
    Int returns a non-negative pseudo-random int.
```
```Go
func Int63() int64
    Int63 returns a non-negative pseudo-random int64.
```
```Go
func Int63n(n int64) int64
    Int63n return a pseudo-random number in [0,n) as an int64
```
```Go
func Intn(n int) int
    Intn returns a pseudo-random number in [0,n) as an int.
```
```Go
func Float64() float64
    Float64 returns a uniformly distributed pseudo-random float64 from [0, 1).
    The distribution is 2^53 evenly spaced floats with spacing 2^-53.
    
```
```Go
func Float64_64() float64
    Float64_64 returns a uniformly distributed pseudo-random float64 from [0, 1).
    The distribution includes all floats in [2^-12, 1) and 2^52 evenly spaced 
    floats in [0, 2^-12) with spacing 2^-64.
    
```
```Go
func Float64_117() float64
    Float64_117 returns a uniformly distributed pseudo-random float64 from [0, 1).
    The distribution includes all floats in [2^-65, 1) and 2^52 evenly spaced 
    floats in [0, 2^-65) with spacing 2^-117.  
```
```Go
func Float64_1024() float64
    Float64_1024 returns a uniformly distributed pseudo-random float64 from [0, 1).
    The distribution includes all floats in [2^-1024, 1) and 0.
```

#### Random number generator functions and methods
All seeding goes thru Splitmix prng/shuffler and the seeds do not need to be complicated, eg. 0, 1, etc. are ok.
```Go
func New(seed uint64) Rand
    New returns a new Rand seeded with the seed.
```
```Go
func Seed(seed uint64)
    Seed seeds system global Rand globalRand by the seed. globalRand is
    used by non-method functions above.
```
```Go
func (r *Rand) Seed(seed uint64)
    Seed seeds a Rand by the seed. Any seed is ok. Do not seed Rands created by
    Next or NewRandSlice.
```
```Go
func NewRandSlice(n int, seed uint64) []Rand
    NewRandSlice returns a slice of n Rands with non-overlapping random streams. The
    first Rand is seeded by seed.
```
```Go
func NewOutlet(seed uint64) *Outlet
    NewOutlet returns a new generator delivery Outlet seeded by the seed.
```
```Go
func ResetGlobalOutlet(seed uint64)
    ResetGlobalOutlet recreates the globalOutlet seeded by the seed. This can be
    made only once.
```
```Go
func (s *Outlet) Next() Rand
    Next returns the next Rand from Outlet. Each Rand has 2^64 long random
    stream, which is not overlapping with other Rands streams. Next is safe for
    concurrent use by multiple goroutines.
```
```Go
func Next() Rand
    Next returns the next non-overlapping stream Rand from globalOutlet. Next is
    safe for concurrent use by multiple goroutines.
```
```Go
func (r *Rand) Jump()
    r.Jump sets r to the same state as 2^64 calls to r.Uint64. Jump can be used to
    generate 2^64 non-overlapping subsequences for parallel computations.

```
```Go
func (r *Rand) State() []byte
    State returns the current binary state of the generator r as []byte.
```
```Go
func (r *Rand) SetState(b []byte) 
    SetState sets the state of the generator r from the state in b []byte.
```

#### Direct non-wrapped methods of type Xosh (Xoshiro256)   
Type Xoro (Xorohiro128) has the same methods. Only shorter jumps and random streams. 
Just replace "Xosh" by "Xoro".
```Go
func (x *Xosh) Float64() (next float64)
    Float64 returns a uniformly distributed pseudo-random float64 value in [0, 1). 
    Float64 uses 53 high bits of xoshiro256+.
```
```Go
func (x *Xosh) Uint64() (next uint64)
    Uint64 returns a pseudo-random 64-bit value as a uint64. Uint64 is
    xoshiro256**.
```
```Go
func (x *Xosh) Seed(seed uint64)
    Seed seeds a xoshiro256 by the seed using SplitMix64. Any seed is ok.
```
```Go
func NewXosh(seed uint64) Xosh
    NewXosh returns a new xoshiro256 generator seeded by the seed.
```
```Go
func NewXoshSlice(n int, seed uint64) []Xosh
    NewXoshSlice returns a slice of n xoshiro256 generators with non-overlapping
    2^128 long random streams. First generator is seeded by seed.
```
```Go
func NextXosh() Xosh
    NextXosh returns the next non-overlapping stream xoshiro256 from globalOutlet. 
```
```Go
func (x *Xosh) Jump()
    x.Jump sets x to the same state as 2^128 calls to x.Uint64
```
```Go
func (x *Xosh) JumpLong()
    x.JumpLong sets x to the same state as 2^192 calls to x.Uint64 or
    2^64 calls to x.Jump.

```
