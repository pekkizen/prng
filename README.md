
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
for github.com/golang/exp/rand.


### Generating a uniform distributed float64 in [0, 1)

The uniform distribution of floats in [0, 1) is not a generally defined concept.
If every float is given a same probability, the distribution is not uniform
by any normal definition, except for the ordinal numbers of the floats, and has rare use in anything.   

The continuous uniform distribution of real random numbers r in [0, 1] can be defined
by its _cumulative distribution function (CDF)_: F(x) = P(r <= x) = x (P = probability).
This distribution is called _standard uniform distribution_.
Generally, library random functions use a discrete uniform distribution of evenly spaced and 
equally likely float values in [0, 1), the 1 not included.
The minimum possible even spacing for float64's is 2^-53.
An other approach is to use all or less floats in [0, 1) and assign individual
float probabilities so, that the resulting distribution implements
the CDF of the standard uniform distribution, within the limits of discretization.
This means that P(f <= fx) ~ P(r <= x), where fx = x rounded to float64
and f is a random float64 from the distribution.



Package prng functions Float64 use the formula 
```Go
    n = Uint64() >> 11
    f = float64(n) / (1 << 53)  
```
The discrete uniform distribution of the floats f is 2^53 equally likely and evenly spaced values with
spacing 1/2^53: [0, 1/2^53, 2/2^53, ... , (2^53-1)/2^53]. In [1/2, 1) they are
adjacent floats. These dyadic rationals have a finite and unique float64 representation in [0, 1).
The mapping from n to f is a bijective function with an inverse function
```Go   
    n = uint64(f * (1 << 53))
```
This method gives the greatest possible number (2^53) of evenly spaced and equally likely floats in [0, 1).     
Scaling f by power of two
```Go   
    f = f * 2^k, where k <= 53 and k+53 >= -1022
```
gives 2^53 evenly spaced floats in [0, 2^k). Also 2^53 is the maximum number of
evenly spaced floats in any [0, 2^k).

An alternative method is to explicitly construct a 64-bit representation of a float64 
in [1, 2), take it as a float64 and subtract 1. The number 1023 goes to the exponent bits and
represents an actual exponent 0. The IEEE 754 floating-point format interprets n as a binary 
decimal with an implicit '1' bit as the whole part: 1.n is a number in [1.0, 1.1111...] ~ [1.0, 2.0) in decimal.

```Go  
    n = Uint64() >> 12
    f = math.Float64frombits(1023 << 52 | n ) - 1 
 ```
with an inverse function
```Go   
    n = uint64(f * (1 << 52))
```
The explicit method uses 52 bits and gives only every second float64 compared to the previous one. 
On the test hardware the explicit method is not faster. 
Mainly because the compiler and CPU handle the division by power of two (1 << 53) as a single
CPU cycle operation, much faster than a general division. The division by (1 << 53) takes the same time 
as the multiplication by the constant 1.0 / (1 << 53), which gives the same result.
In http://prng.di.unimi.it the multiply/divide method is preferred.

The previous methods unnecessarily lose accuracy, if the random uint64 has leading zeros. 
We can have up to 11 leading zeros and still have 53 bits for the full float64 accuracy.
Function Float64_64 counts leading zeros and shifts random number urand to start with 
the first '1' bit or the 11th bit.
If the first bit is the 11th bit, 0 or 1, urand is mapped to 2^53 evenly spaced float64 values in [0, 1).
Otherwise urand is mapped to 2^52 evenly spaced float64 values in [1/2, 1). 
In both cases the resulting float64 is then scaled to a right subinterval by dividing by 2^leading_zeros.

```Go   
func Float64_64(urand uint64) float64 {

    zeros := uint64(bits.LeadingZeros64(urand))
    if zeros > 11 { 
        zeros = 11 
    }                                                // zeros = 11, f goes to [0, 1)
    f := float64((urand << zeros) >> 11) / (1 << 53) // zeros < 11, f goes to [1/2, 1)
    return f / float64(uint64(1 << zeros))           
}
```
In [2^-11, 1) every float64 is included in the distribution and adding 1 to the 53-bit 
value ((urand << zeros) >> 11) gives the next adjacent float. 
In [0, 2^-11) Float64_64 gives 2^53 evenly spaced floats with spacing 2^-64. 
Each leading zero doubles the float density and halves the individual float probability.
Near 1 the individual float probability is 2048 x the probability near 0, but near 0 there are 2048 x more 
floats to select in a same length interval.
The float distribution has a discrete CDF       
- F(f<sub>j</sub>) = P(f < f<sub>j</sub>) = f<sub>j</sub> or    
- F(f<sub>j</sub>) = P(f <= f<sub>j</sub>) = f<sub>j+1</sub>     

where f, f<sub>j</sub> and f<sub>j+1</sub> are floats in the distribution and f<sub>j+1</sub> is 
the next float to f<sub>j</sub>.
The single float point probability is 
- P(f = f<sub>j</sub>) = f<sub>j+1</sub> - f<sub>j</sub> = Δ<sub>j</sub> where
- Δ<sub>j</sub> is one ULP in [2^-11, 1) and
- Δ<sub>j</sub> is 2^-64 in [0, 2^-11) 

An _ULP_ means a _unit of least precision_ or a _unit in the last place_ and one ULP is
the gap between two adjacent floats. The maximum ULP in [0, 1) is 2^-53 in [0.5, 1).

In P(f <= f<sub>j</sub>) = f<sub>j+1</sub> the difference to standard uniform CDF is within
the normal floating point aritmetic rounding errors, eg.
- P(r <= 0.5) = 0.5 for standard uniform distribution and
- P(f <= 0.5) = 0.5 + 2^-53 = 0.5000000000000001 for Float64_64

The difference is one ULP, which is quite impossible to detect from the random stream of floats.
In any case, for all f<sub>j</sub> in the distribution
- P(f <= f<sub>j</sub>)) + P(f > f<sub>j</sub>)) = 1 and
- P(f < f<sub>j</sub>)) + P(f >= f<sub>j</sub>)) = 1

Besides, each f<sub>j</sub> may contain rounding inaccuracy in [0, ULP/2).

In the Float64_64 distribution every float has a probability equal to the 
space (Δ<sub>j</sub>, mostly 1 ULP) it occupies in [0, 1) and the distribution is 
a kind of unevenly spaced discrete uniform distribution, although this concept hardly exists.

This simple algorithm can produce every float64 in 99.95% of the length of the unit interval 
and the rest 0.05% has an even spacing of 2^-64. 
The spacing 2^-53 near zero and especially the gap [0, 2^-53) has been criticized.
The spacing 2^-64 puts 2048 floats in [0, 2^-53).

bits.LeadingZeros64 is a fast intrinsic function (amd64) and it makes Float64_64 a real choice with 
3.2 ns execution time, random number generation included.

Function Float64Full expands the leading zeros concept to an arbitrary number of zeros and
implements an uniform distribution of all float64's in [0, 1).

```Go 
func (x *Xoro) Float64Full() float64 {
    hi := x.Uint64()
    if hi >= 1<<52 {  //99.95% of cases 
        return Float64_64(hi)
    } 
    powz := 1.0
    for hi == 0 { 
        hi = x.Uint64() 
        powz *= (1 << 64)
    }
    lo := x.Uint64()
    zeros := uint64(bits.LeadingZeros64(hi))
    hi = (hi << zeros) | (lo >> (64 - zeros))
    f := float64(hi >> 11) / (1 << 53)               // f goes to [1/2, 1)
    return  f / (powz * float64(uint64(1 << zeros))) // divide by 2^all_zeros
}
```
Float64Full converts any random sequence of bits to a float64 in [0, 1).
The outcome float distribution covers 99.98% of all floats in [0, 1). 
The smallest nonzero number returned is 2^-1024 (5.6e-309). This is an IEEE 754 
subnormal float64 and in (0, 2^-1024) there are 2^50 subnormals more,
which Float64Full misses from the full range of floats in [0, 1). 
After 2^-1024 the divider variable powz just overflows to +inf and the rest are zeros.

All single floats (>= 2^-1024) have a probability 1 ULP, a gap to the next float. 
This quite likely implements the most accurate, computationally light
floating point approximation of the standard uniform distribution.
For all float64 numbers f<sub>j</sub> in [0, 1) it has
a cumulative distribution function F which satisfies
- F(f<sub>j</sub>) = P(f <= f<sub>j</sub>)) = 1 - P(f > f<sub>j</sub>)).

Float64Full is a random number generator and those longer random zero sequences
simply don't happen even they are not impossible.
It will take 405 years to get only 64 leading random zeros with probability 0.5, 
if one try takes one nanosecond. After 405 more years we have these zeros with 
probability 0.75.

Float64Full is a fast 3.3 ns function, because 99.95% of cases are handled by Float64_64.
A specific function for 128 bits seems to be not simpler or faster.
An other implementation of a full accuracy random float64 function is presented in 
http://prng.di.unimi.it/random_real.c.


### Benchmarking generator speeds
The original C functions have been modified to more efficient Go code, 
especially for xoshiro256. The baseline 128-bit state minimal "random" generator below was used
to get a baseline time reference for returning a result and updating the state.
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
| (1) Baseline128               | 0.41    |    
| (1) Baseline256               | 0.71    |        
| (1) NextState128              | 0.87    |  
| (1) NextState256              | 1.06    | 

#### Time (ns) to generate an uint64 
|     Generator                 | Time    | 
|-------------------------------|---------|
| (1) xoroshiro128+             | 0.93    |
| (1) xoroshiro128**            | 1.08    |
| (1) SplitMix64                | 1.11    |
| (1) xoshiro256+               | 1.20    | 
| (5) xoshiro256+               | 1.36    | 
| (1) xoshiro256**              | 1.34    | 
| (1) xoshiro256++              | 1.36    |  
| (3) PCG                       | 1.88    | 
| (2) xoshiro256+               | 2.43    | 
| (2) xoshiro256**              | 2.43    | 
|  math/rand rng.Int63()        | 2.66    |      
| (1/3) 128**/Source interface  | 1.86    |     
| (3) PCG/Source interface      | 2.70    |     
| (4) PCG                       | 3.40    | 

#### Time (ns) to generate a float64 in [0, 1)
|       Generator               | Time    |
|-------------------------------|---------|
| (1) Float64/xoroshiro128+     | 1.32    | 
| (1) Float64/xoshiro256+       | 1.61    |
| math/rand rng.Float64()       | 2.92    |
| (3) rng.Float64()             | 5.01    |

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
Functions Int63n, Intn and Uint64n doesn't make any bias correction. The bias with 64-bit numbers is very small
and probably not detectable from the random stream.
All functions are also implemented as methods of type Rand. A single Rand should not be shared concurrently.
The methods of Rand are over 2 x faster, if a local Rand is used.
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
    Float64 returns a uniformly distributed pseudo-random float64 value in [0, 1).
    
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
