
## Generating uniform random floating-point numbers

The term _float_ is used for both general floating point number and 
the 64-bit IEEE 754 floating point number, which is also referenced by _float64_.

### The uniform distribution in [0, 1)

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
may is also used for discrete uniform distributions in [0, 1). 

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
ULP is here used as a **property of a floating point number**:
- ULP(f<sub>k</sub>) = f<sub>k+1</sub> - f<sub>k</sub>, &nbsp; the distance to next bigger adjacent float.

For those interested in ULPs: 
[Jean-Michel Muller (2005): On the definition of ulp(x)](http://ljk.imag.fr/membres/Carine.Lucas/TPScilab/JMMuller/ulp-toms.pdf), where ULP is defined as a property of real number.


The floats in [0, 1) (not including subnormals) are exponentially distributed so that 
- all the intervals 
[1/2<sup>k+1</sup>, 1/2<sup>k</sup>), &nbsp; 0 <= k <= 1021, 
have 2<sup>52</sup> evenly spaced adjacent floats with an ULP 1/2<sup>53+k</sup>. 

The maximum ULP in [0, 1) is 1/2<sup>53</sup> in [1/2, 1).
Most of the floats are near 0 and in [1/2, 1) there are only ~0.1% of all the floats in [0, 1).

 _**The complete uniform distribution of floats in [0, 1)**_
can be defined as distribution, where every float f<sub>k</sub> in [0, 1)
is included and has the probability  ULP(f<sub>k</sub>). 
The connection to the continuous standard uniform distribution is (simplified):
A random number from the continuous standard uniform
distribution has the probability ULP(f<sub>k</sub>) to be 
within distance ULP(f<sub>k</sub>)/2 of f<sub>k</sub>
from where it gets rounded to f<sub>k</sub>. 

It is simple to draw random numbers from the complete uniform distribution of floats, 
even with fast functions. 
Here is a kind of point of view shift: drawing numbers from
a uniform distribution of floats vs. computing floats 
with a standard uniform distribution. 
The concept of uniform distribution of floats in [0, 1) and drawing random numbers from it,
seems to be hard to find anywhere. 
The "producing random floats" is mostly described as to somehow 
compute a uniform "real number" in [0, 1) and then round it to a floating point number in [0, 1).

[Goualard (2020)](https://hal.archives-ouvertes.fr/hal-02427338/document)
uses expression _to compute floating-point numbers with a standard uniform distribution_.
Producing floats: _software offering the most methods
to compute random integers will almost always provide only one means to obtain
random floats, namely through the division of some random integer by another integer_.
The overall situation:
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

Stackoverflow has a subject [Generate uniformly random float which can return all possible values](https://stackoverflow.com/questions/53277105/generate-uniformly-random-float-which-can-return-all-possible-values),
which brings up  the problems. In
[Uniformly distributed secure floating point numbers in \[0,1)](https://crypto.stackexchange.com/questions/31657/uniformly-distributed-secure-floating-point-numbers-in-0-1)
is a definition of the uniform distribution of floats in terms of real numbers and rounding.


[Goualard (2020)](https://hal.archives-ouvertes.fr/hal-02427338/document)
has a case study of implementations in software, which
presents quite unnecessary problems in the current uniform float functions. 
One example is Go math/rand Float64, which is documented:
_Float64 returns, as a float64, a pseudo-random number in [0.0,1.0)_.
Go Float64 doesn't return a strictly uniformly distributed number. 


### Functions for a uniformly distributed float64 in [0, 1)

Generally, library random functions use a discrete uniform distribution of evenly spaced 
float values in [0, 1). The maximum number of floats is 2<sup>53</sup> with
spacing 2<sup>-53</sup>. An other approach is to use all or less floats in [0, 1) and assign individual
float probabilities so, that the resulting distribution implements the
uniform distribution condition.

Mostly uniform floats are made by the formula 
```Go
    n = Uint64() >> 11
    f = float64(n) / (1 << 53)  
```
The discrete uniform distribution of the floats f is 2<sup>53</sup> equally likely and evenly 
spaced values with spacing 1/2<sup>53</sup>: 
[0, 1/2<sup>53</sup>, 2/2<sup>53</sup>, ... , (2<sup>53</sup>-1)/2<sup>53</sup>]. 
In [1/2, 1) they are adjacent floats, every float is included. 
These _dyadic rationals_ have a finite and unique float64 representation in [0, 1).
This is a discrete uniform distribution also by the definition, not only
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
Function Float64_64 uses leading zeros (< 11) and the following '1' bit to select 
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
 - P(r < 0.5) = 0.5 for standard uniform distribution          
 - P(r <= 0.5) = 0.5 for standard uniform distribution         
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
version is slightly modified and can be found in the package prng xoro.go file. Float64_64 is two
times slower than 53-bit divide, but it has 6.5 times more floats to give.

Function Float64_1024 expands the leading zeros concept to an arbitrary number of zeros and
implements a uniform distribution of all the normal floats in [0, 1).

```Go 
func Float64_1024() float64 {
    hi := Uint64()
    if hi >= 1<<52 {  //99.95% of cases 
        return Float64_64(hi)
    } 
    pow := 1.0
    for hi == 0 { 
        hi = Uint64() 
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

An other implementation of a full accuracy random float function _Random_real.c_ is 
presented by [Campbell (2014)](http://prng.di.unimi.it/random_real.c). 
The main difference to Float64_1024 is how
Random_real.c rounds a random uint64 value to a float64 in the final phase. 
Simplified and translated to Go: Random_real.c uses float64(u | 1) / (1 << 11) 
instead of float64(u >> 11), where u is an uint64 starting with a '1' bit. 
At the end of the next section Random_real.c is compared to the random bisecting.
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
Float64Bisect with rounding returns the same float as Random_real.c. 
The forcing/rounding the last bit to '1' by (u | 1) in Random_real.c is essential. 
Random_real.c implements the random bisecting with rounding algoritm.
This is a somewhat unexpected result considering the very different
point of view in the introduction to Random_real.c by
[Campbell (2014)](http://prng.di.unimi.it/random_real.c).
The other way around, Float64Bisect with rounding implements the real number to float
process described by Campbell.  

The rounding in the random bisecting randomly gives every second time one ULP bigger float
and shifts the whole distribution 0.5 ULPs forward. It is not quite clear, that
this gives a better approximation of real numbers in the standard uniform distribution.
The rounding gives a 1 with probability 2<sup>-54</sup>.

The random bisecting can also be regarded as a search algorithm,
which for a given bit sequence key finds the corresponding float.
The bit sequence provides a unique binary search path to the float
As an algoritm this binary search by bisecting is quite dumb. From the first
'1' bit it does 52 bisects and finally finds its own search key as the significand
bits of the float found. The random bisecting is too slow for use as a general 
random number generator, but useful in testing other generators distributions.
