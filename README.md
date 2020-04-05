
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


### The uniform distribution in [0, 1)

The term _float_ is used for both general floating point numbers and 
the 64-bit IEEE 754 floating point numbers, which are also referenced by _float64_.

The uniform distribution of floats in [0, 1) is not a generally well defined concept.
If every float is given a same probability, the distribution is not uniform
by any normal definition.  The distribution normally wanted is a floating point
approximation of the  _continuous uniform distribution_ of real random numbers r in [0, 1).
This can be defined
by its _cumulative distribution function (CDF)_

- F(x) = P(r <= x) = x

This distribution is called the _standard uniform distribution_. In a continuos 
uniform distribution a random variable has the same probability 
within all subintervals of the same size. _Standard uniform distribution_
may also be used for discrete uniform distributions in [0, 1). 

Random numbers k = 1, 2, ... n, each with probability 1/n, can be mapped
to a _discrete uniform distribution_ of random variables f<sub>k</sub> in [0, 1) by 

- f<sub>k</sub> = (k-1) / n  &nbsp; &nbsp; &nbsp; f<sub>k</sub> = 0, 1, ... (n-1)/n.

The random numbers f<sub>k</sub> have a CDF
- F(f) = P(f <= f<sub>k</sub>) = f<sub>k+1</sub> &nbsp; &nbsp; for k <= n and f<sub>n+1</sub> = 1

The probability of a single f<sub>k</sub> is
- P(f = f<sub>k</sub>) = P(f <= f<sub>k</sub>) - P(f <= f<sub>k-1</sub>) = f<sub>k+1</sub> - f<sub>k</sub> = 1/n 
 &nbsp; &nbsp; &nbsp; where P(f <= f<sub>0</sub>) = f<sub>1</sub> = 0

The probability of f<sub>k</sub> is the same as its distance to the next number in the distribution.
This is trivial for an evenly spaced distribution.
For an unevenly spaced distribution, eg. floating point numbers, this is
a necessary condition for a distribution to be uniform. 
**Every random number in a discrete uniform distribution in [0, 1) must have 
a probability equal to the space it occupies in [0, 1)**.
This condition doesn't depend on the spacing and [0, 1) can
be divided subintervals of different, even or uneven, spacings. 

In measuring floating point number distances, the concept of _ULP_ is generally used.
An ULP means a _unit of least precision_ or a _unit in the last place_. 
The gap between two adjacent floats is 1 ULP. 
Two adjacent floats differ only in the least significant bit.
As a property of a float ULP is (here) defined
- The ULP of a float f<sub>k</sub> = f<sub>k+1</sub> - f<sub>k</sub>, &nbsp; the distance to next bigger adjacent float.

The maximum ULP in [0, 1) is 1/2<sup>53</sup> in [1/2, 1).
The floats in [0, 1) are exponentially distributed so that 
- all the intervals 
[1/2<sup>k+1</sup>, 1/2<sup>k</sup>), &nbsp; 0 <= k <= 1021, 
have 2<sup>52</sup> evenly spaced adjacent floats with an ULP 1/2<sup>53+k</sup>. 

Most of the floats are near 0 and in [1/2, 1) there are only ~0.1% of all the floats in [0, 1).

In  _**the complete uniform distribution of floats in [0, 1)**_ every float
is included and has  probability equal to its ULP. 
It is simple to draw random numbers from this distribution,
even with fast functions. See below.

The concept of uniform distribution of floats in [0, 1) and drawing random numbers from it,
seems to be hard to find . The "producing random floats" is mostly described as to somehow 
compute a uniform "real number" in [0, 1) and then transform it to a floating point number in [0, 1).

[Goualard (2020)](https://hal.archives-ouvertes.fr/hal-02427338/document)
_computes floating-point numbers with a standard uniform distribution_
with the focus on _generating random floating-point numbers by dividing integers_, 
which includes all generally used methods. Of the overall situation:
_Many studies are devoted to the analysis of RNGs producing integers; they
are much fewer to consider RNGs producing 
floats, and we are not aware of
other works considering the structure of the 
floats produced at the bit level._

[Downey (2007)](http://allendowney.com/research/rand/downey07randfloat.pdf)
_generates uniformly-distributed pseudo-random floating-point values in the range [0; 1]_
and _chooses floating-point values in the range such that the
probability that a given value is chosen is proportional to the distance between
it and its two neighbors,_ but don't explain why to choose this way. 

Stackoverflow has a discussion of the subject [Generate uniformly random float which can return all possible values](https://stackoverflow.com/questions/53277105/generate-uniformly-random-float-which-can-return-all-possible-values),
which brings up  the problems.


[Goualard (2020)](https://hal.archives-ouvertes.fr/hal-02427338/document)
presents a case study of implementations in software 
and there seems to be a lot of quite unnecessary problems in the current uniform float functions. 
One example is Go math/rand Float64, which is documented:
_Float64 returns, as a float64, a pseudo-random number in [0.0,1.0)_.
Go Float64 doesn't return a strictly uniformly distributed number. 


### Functions for a uniformly distributed float64 in [0, 1)

Generally, library random functions use a discrete uniform distribution of evenly spaced 
float values in [0, 1). The maximum number of floats is 2<sup>53</sup> with
spacing 2<sup>-53</sup>. An other approach is to use all or less floats in [0, 1) and assign individual
float probabilities so, that the resulting distribution implements the
uniform distribution condition.

Package prng functions Float64 use the formula 
```Go
    n = Uint64() >> 11
    f = float64(n) / (1 << 53)  
```
The discrete uniform distribution of the floats f is 2<sup>53</sup> equally likely and evenly 
spaced values with spacing 1/2<sup>53</sup>: 
[0, 1/2<sup>53</sup>, 2/2<sup>53</sup>, ... , (2<sup>53</sup>-1)/2<sup>53</sup>]. 
In [1/2, 1) they are adjacent floats, every float is included. 
These _dyadic rationals_ have a finite and unique float64 representation in [0, 1).
This is a discrete uniform distribution also by a definition, not only
a rounded floating point approximation.

The mapping from n to f is a bijective function with an inverse function
```Go   
    n = uint64(f * (1 << 53))
```
This 53-bit divide method gives the greatest possible number (2<sup>53</sup>) of evenly 
spaced and equally likely floats in [0, 1). For a more accurate distribution, an uneven
spacing must be used.
Scaling f by power of two
```Go   
    f = f * 2^k, where k <= 53 and k+53 >= -1022
```
gives 2<sup>53</sup> evenly spaced floats in [0, 2<sup>k</sup>). Also 2<sup>53</sup> is the maximum number of
evenly spaced floats in any [0, 2<sup>k</sup>).

An alternative method is to explicitly construct a 64-bit representation of a float64 
in [1, 2), take it as a float64 and subtract 1. The number 1023 goes to the exponent bits and
represents an actual exponent 0. The IEEE 754 floating-point format interprets n as a binary 
decimal with an implicit '1' bit as the whole part: 1.n is a number in [1.0, 1.1111...] ~ [1.0, 2.0) in decimal.

```Go  
    n = Uint64() >> 12
    f = math.Float64frombits(1023 << 52 | n ) - 1 
 ```
The explicit method is equivalent to float64(Uint64() >> 12) / (1 << 52) and
gives only every second float compared to the 53-bit divide. 
On the test setup the explicit method is not faster. 
The compiler and CPU handle the division by power of two (1 << 53) as a single
CPU cycle operation, faster than a general division. The division by (1 << 53) takes the same time 
as the multiplication by the constant 1.0 / (1 << 53), which gives the same result.
Is the same operation?

The previous methods unnecessarily lose accuracy, if the random uint64 has leading zeros. 
The uint64 can have up to 11 leading zeros and still have 53 bits for the full float64 accuracy.
Float64_64 uses leading zeros (< 11) and the following '1' bit to select 
an evenly spaced float64 interval and
the rest 52 bits as the float64 significand bits. If leading zeros >= 11, a float from an evenly
spaced distribution in [0, 2<sup>-11</sup>) is returned.

```Go   
func Float64_64(urand uint64) float64 {

    zeros := uint64(bits.LeadingZeros64(urand))
    if zeros >= 11 { 
        return float64(urand) / (1 << 64)             // the result is in [0, 2^-11)
    }                                                
    f := float64((urand << zeros) >> 11) / (1 << 53)  // f goes to [1/2, 1)
    return f / float64(uint64(1 << zeros))          
}                                                   
```
In [2<sup>-12</sup>, 1) the Float64_64 distribution is complete, every float is included.
In [0, 2<sup>-12</sup>) the distribution has 2<sup>52</sup> evenly spaced floats with spacing 2<sup>-64</sup>. 
Each leading zero doubles the float density and halves the individual float probability.
Near 1 the individual float probability is 2048 x the probability near 0, but near 0 there are 2048 x more 
floats to select in a same length interval.
The float distribution has a discrete CDF       
- F(f<sub>k</sub>) = P(f <= f<sub>k</sub>) = f<sub>k+1</sub>     

where f, f<sub>k</sub> and f<sub>k+1</sub> are floats in the distribution and f<sub>k+1</sub> is 
the next float to f<sub>k</sub> (or 1).
The single float point probability is 
- P(f = f<sub>k</sub>) = f<sub>k+1</sub> - f<sub>k</sub> = Δ<sub>k</sub> where
- Δ<sub>k</sub> is one ULP in [2<sup>-12</sup>, 1) and
- Δ<sub>k</sub> is 2<sup>-64</sup> in [0, 2<sup>-12</sup>) 


There are inevitably differences between a continous and a discrete distribution, eg.

- P(r <= 0.5) = 0.5 for standard uniform distribution 

 - P(r < 0.5) = 0.5 for standard uniform distribution 

 - P(f < 0.5) = 0.5 for Float64_64     

- P(f <= 0.5) = 0.5 + 2<sup>-53</sup> ~ 0.5000000000000001 for Float64_64

The P(f <= 0.5) could be fixed to 0.5, but then we will get other problems,
like losing 0 and 1 coming to the distribution.
In any case, for all f<sub>k</sub> in the distribution
- P(f <= f<sub>k</sub>) + P(f > f<sub>k</sub>) = 1 and
- P(f < f<sub>k</sub>) + P(f >= f<sub>k</sub>) = 1

In the Float64_64 distribution every float has a probability equal to the 
space (mostly 1 ULP) it occupies in [0, 1). So, the distribution is 
an unevenly spaced discrete uniform distribution.

This simple algorithm can produce every float (is complete) in 99.975% of the length of the unit interval 
and the rest 0.025% has an even spacing of 2<sup>-64</sup>. 
The spacing 2<sup>-53</sup> near zero and especially the gap [0, 2<sup>-53</sup>) has been criticized.
The spacing 2<sup>-64</sup> puts 2048 floats in the [0, 2<sup>-53</sup>) gap.

bits.LeadingZeros64 is a fast intrinsic function (amd64) and it makes Float64_64 a real choice with 
2.7 ns execution time, random number generation (xoroshiro128**) included. This 2.7 ns
version is slightly modified and can be found in the xoro.go file. Float64_64 is two
times slower than Float64, but it has 6.5 times more floats to give.

Function Float64_1024 expands the leading zeros concept to an arbitrary number of zeros and
implements a uniform distribution of all the normal floats in [0, 1).

```Go 
func (x *Xoro) Float64_1024() float64 {
    hi := x.Uint64()
    if hi >= 1<<52 {  //99.95% of cases 
        return Float64_64(hi)
    } 
    pow := 1.0
    for hi == 0 { 
        hi = x.Uint64() 
        pow *= (1 << 64)
    }
    lo := x.Uint64()
    zeros := uint64(bits.LeadingZeros64(hi))
    hi = (hi << zeros) | (lo >> (64 - zeros))
    f := float64(hi >> 11) / (1 << 53)              // f goes to [1/2, 1)
    return  f / (pow * float64(uint64(1 << zeros))) // divide by 2^all_zeros
}
```
Float64_1024 converts a random sequence of bits to a float64 in [0, 1).
The outcome float distribution covers 99.98% of the floats in [0, 1). 
The smallest nonzero number returned is 2<sup>-1024</sup> (5.6e-309). This is an IEEE 754 
subnormal float64 and in (0, 2<sup>-1024</sup>) there are 2<sup>50</sup> subnormals more,
which Float64_1024 misses from the full range of floats in [0, 1).
Every float (>= 2<sup>-1024</sup>) has a probability equal to its ULP. 

After 2<sup>1024</sup> leading zeros the divider variable pow just overflows to +inf and 
Float64_1024 returns zeros.
These longer random zero sequences simply don't happen even they are not impossible.
It will take 405 years to get only 64 leading random zeros with probability 0.5, 
if one try takes one nanosecond. However, in widely used applications or large parallel 
simulations these 64 zeros are likely to happen. But not 128 or more zeros.

Float64_1024 is a fast under 3 ns function, because 99.95% of cases are handled by Float64_64.
A specific function for 128 bits seems to be not simpler or faster.    

An other implementation of a full accuracy random float function _random_real.c_ is 
presented by [Campbell (2014)](http://prng.di.unimi.it/random_real.c). 
The main difference to Float64_1024 is how
random_real.c rounds a random uint64 value to a float64 in the final phase. 
Simplified and translated to Go: random_real.c uses float64(u | 1) / (1 << 11) 
instead of float64(u >> 11), where u is an uint64 starting with a '1' bit. 
At the end of the next section random_real.c is compared to the random bisecting.
A Go implementation, function RandomReal, is in the xoro.go file.

#### Random bisecting
We can draw an any length finite approximation of a real number from the standard 
uniform distribution by random bisecting:
- Split the interval [0, 1) into two equal length halves. Select either one by random
and repeat with the selected half. Stop when the interval length is less 
than desired accuracy. Return the interval middle value.

Float64Bisect implements the random bisecting for an arbitrary length random bit sequence.
For an enough long sequence, it stops after the interval is one ULP and cannot 
be splitted anymore. Every float in [0, 1) has a possibility to be selected. 
The probability of a returned float is 1/2^bisectings. This is 
also the ULP of the float, which ensures that the floats returned are from 
the complete uniform distribution of floats in [0, 1). 

The bit sequence is interpreted as a binary desimal random number starting with an
implicit '0' bit as the whole part: 0.bitsequence eg. 0.0110... in [0, 1).
Float64Bisect encodes this to a float64 by selecting the right half of the interval
with a '1' bit and the left half with a '0' bit. 
The random binary selection is essential, but
interpreting the bit sequence to be anything is not.
From two adjacent floats, the smaller is selected, which
ensures, that we never get a 1. 

```Go
func Float64Bisect(bitsequence []uint64) float64 {
	
    left, mean, right := 0.0, 0.5, 1.0
    for k := 0; k < len(bitsequence); k++ {
        u := bitsequence[k]
        for bit := uint64(0); bit < 64; bit++ {

            if u & (1 << (63 - bit)) != 0 {
                left = mean                 // '1' bit -> take the right half, big numbers					
            } else {
                right = mean                // '0' bit -> take the left half, small numbers					
            }
            mean = (left + right) / 2
            if mean == left || mean == right { // right - left = 1 ULP
                return left
            }
       }   
    }   
    return -1
}
```
In the interval [2<sup>-12</sup>, 1) Float64Bisect returns the same float as Float64_64
and in the interval [2<sup>-1024</sup>, 1)  it returns the same float as Float64_1024.
Expressed as another way: 
**Functions Float64_64 and Float64_1024 are implementations of the random bisecting algoritm**. 
They are only much more efficient software implementations than Float64Bisect, but
also, only partial implementations within the intervals above.
Float64Bisect can return every float in [0, 1),
eg. the minimum subnormal positive value 2<sup>-1074</sup> and the zero after that.

Float64Bisect can be modified to apply "rounding". Instead of finally always selecting 
the smaller float, it takes the next bit and if it is a '1', selects the bigger one. 
The modified Float64Bisect returns the same float as random_real.c. 
The forcing the last bit to '1' by (u | 1) in random_real.c is essential. 
Also random_real.c implements the (modified) random bisecting algoritm.
This is a somewhat unexpected result considering the very different
point of view in the introduction to random_real.c by
[Campbell (2014)](http://prng.di.unimi.it/random_real.c).
The other way around, Float64Bisect with rounding implements the real number to float
process described by Campbell.
The random bisecting with rounding gives a 1 with probability 2<sup>-54</sup>.

Another view to the random bisecting algorithm. 
The algorithm isn't generating or calculating the floats.
The about 1024 x 2<sup>52</sup> floats are already "there" and ready, just
waiting for to be picked up.
The bit sequence is a key, that provides a unique binary search path to the float.
As an algoritm this binary search by bisecting is quite dumb. From the first
'1' bit it does 52 bisects and finally finds its own search key as the significand
bits of the float found. 
Nevertheless, with its few code lines Float64Bisect implements the complete uniform distribution 
of floats in [0, 1). And it is also a binary desimal to float64 encoder in [0, 1).


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
