package prng_test

import (
	"math/bits"
	"testing"
	"math/rand"


	// . "prng"
	. "github.com/pekkizen/prng"
	"github.com/MichaelTJones/pcg"
	exprand "github.com/golang/exp/rand"
	vpxyz "github.com/vpxyz/xorshift/xoroshiro256plus"
	gonum "gonum.org/v1/gonum/mathext/prng"
)

var usink uint64
var fsink float64
var isink int
var bsink []byte

func workerBaseLine(i int, c chan float64) {
	c <- float64(i)
}
func workerSeeded(i int, c chan float64) {
	r := NewXoro(uint64(i))
	c <- r.Float64()
}
func workerNonOverLap(c chan float64) {
	r := NextXoro()
	c <- r.Float64()
}
func work() float64 {
	const workers = 1000000
	c := make(chan float64, workers)
	for i := 0; i < workers; i++ {
		// go workerBaseLine(i, c)
		// go workerSeeded(i, c)
		go workerNonOverLap(c)
	}
	return <-c
}

func BenchmarkGoWorkers(b *testing.B) {
	var y float64
	for n := 0; n < b.N; n++ {
		y = work()
	}
	fsink = y
}
func BenchmarkEmpty(b *testing.B) {
	var i int
	for n := 0; n < b.N; { // 1 CPU cycle, 0.28 ns @ 3.6 MHz
		n++
		i = n 
	}
	isink = i
}

func BenchmarkBaseline128(b *testing.B) {
	var y uint64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.Baseline128()
	}
	usink = y
}
func BenchmarkBaseline256(b *testing.B) {
	var y uint64
	x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		y = x.Baseline256()
	}
	usink = y
}
func BenchmarkState128(b *testing.B) {
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		x = x.NextState()
	}
	usink = x.Uint64()
}
func BenchmarkState256(b *testing.B) {
	x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		x = x.NextState()
	}
	usink = x.Uint64()
}
func BenchmarkSeed(b *testing.B) {
	var x Xoro
	// var x Xosh
	for n := 0; n < b.N; n++ {
		x.Seed(1)
	}
	usink = x.Uint64()
}

func BenchmarkJump(b *testing.B) {
	x := NewXoro(1)
	// x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		x.Jump()
	}
	usink = x.Uint64()
}
func BenchmarkState(b *testing.B) {
	// x := NewXosh(1)
	x := NewXoro(1)
	s := make([]byte, 32)
	for n := 0; n < b.N; n++ {
		s = x.State()
	}
	x.SetState(s)
	usink = x.Uint64()
}
func BenchmarkSetState(b *testing.B) {
	x := NewXosh(1)
	// x := NewXoro(1)
	s := make([]byte, 32)
	for n := 0; n < b.N; n++ {
		x.SetState(s)
	}
	usink = x.Uint64()
}
func BenchmarkBitsLeadingZeros(b *testing.B) {
	var zeros uint64
	for n := 0; n < b.N; n++ {
		zeros = uint64(bits.LeadingZeros64(uint64(n) + (1 << 52)))
	}
	usink = zeros
}
// ---------------------------------------- New generator--------------------//
func BenchmarkNew(b *testing.B) {
	var x Xoro
	// var x Xosh
	for n := 0; n < b.N; n++ {
		x = NewXoro(uint64(n))
		// x = NewXosh(uint64(n))
	}
	usink = x.Uint64()
}
func BenchmarkNewGonum(b *testing.B) {
	x := gonum.NewXoshiro256starstar(1)
	for n := 0; n < b.N; n++ {
		x = gonum.NewXoshiro256starstar(uint64(n))
	}
	usink = x.Uint64()
}
func BenchmarkExprand(b *testing.B) {
	x := rand.New(rand.NewSource(1))
	for n := 0; n < b.N; n++ {
		x = rand.New(rand.NewSource(int64(n)))
	}
	usink = uint64(x.Int63())
}
func BenchmarkNewExprand(b *testing.B) {
	x := exprand.New(exprand.NewSource(1))
	for n := 0; n < b.N; n++ {
		x = exprand.New(exprand.NewSource(uint64(n)))
	}
	usink = x.Uint64()
}
func BenchmarkExprandSlice(b *testing.B) {
	const size = 1000000
	var x []Rand
	for n := 0; n < b.N; n++ {
		x = NewRandSlice(size, 1)
	}
	usink = x[size/2].Uint64()
}
func BenchmarkNext(b *testing.B) {
	var x Xoro
	// var x Xosh
	for n := 0; n < b.N; n++ {
		x = NextXoro()
		// x = NextXosh()
	}
	usink = x.Uint64()
}

//-----------------------------------------------------Float64--------//
func BenchmarkFloat64Methods(b *testing.B) {
	var y float64
	var u uint64
	const c = 1.0 /  ((1<<53) - 1)
	x := NewXoro(1)
	// x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		u = x.Uint64() | (1 << 63)
		// y = float64(u)
		y = float64((u >> 10 + 1) >> 1) / (1<<53) 
		// y = float64(u | 1) / (1<<64)

		// y = float64(u >> 11) / (1<<53) 
		// y = float64(u ) / (1<<64)

		// y = float64((1 << 64)-1-uint64(n))  // 50% slower
		// y = float64(((1 << 64)-1-uint64(n)) >> 11)
		// y = x.RandomReal()
		// y = x.Float64_64R()
		// y = x.Float64_1024()
		// y = x.Float64()
		// y = float64(x.Xoroshiro128plus() >> 11) / (1 << 53) 
		// y = math.Float64frombits(0x3FF<<52| (x.Xoroshiro128plus() >> 12)) - 1
	}
	fsink = y
}
func BenchmarkFloat64_64(b *testing.B) {
	var y float64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.Float64_64()
		// y = x.Float64_64R()
	}
	fsink = y
}
func BenchmarkFloat64_1024(b *testing.B) {
	var y float64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.Float64_1024R()
	}
	fsink = y
}
func BenchmarkRandomReal(b *testing.B) {
	var y float64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.RandomReal()
	}
	fsink = y
}
func BenchmarkFloat64Bisect(b *testing.B) {
	var y float64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.Float64Bisect(false)
	}
	fsink = y
}

func BenchmarkFloat64Xoro(b *testing.B) {
	var y float64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.Float64()
	}
	fsink = y
}
func BenchmarkFloat64Xosh(b *testing.B) {
	var y float64
	x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		y = x.Float64()
	}
	fsink = y
}
func BenchmarkFloat64Exprand(b *testing.B) {
	var y float64
	x := exprand.New(exprand.NewSource(1))
	for n := 0; n < b.N; n++ {
		y = x.Float64()
	}
	fsink = y
}
func BenchmarkFloat64Rand(b *testing.B) {
	var y float64
	x := rand.New(rand.NewSource(1))
	for n := 0; n < b.N; n++ {
		y = x.Float64()
	}
	fsink = y
}

//----------------------------------Uint64----------------Uint64------//
func BenchmarkSplitmix(b *testing.B) {
	var y uint64
	var seed uint64
	for n := 0; n < b.N; n++ {
		y = Splitmix(&seed)
	}
	usink = y
}
func Benchmark128plus(b *testing.B) {
	var y uint64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.Xoroshiro128plus()
	}
	usink = y
}
func BenchmarkInt63Rand(b *testing.B) {
	var y int64
	x := rand.New(rand.NewSource(1))
	for n := 0; n < b.N; n++ {
		y = x.Int63()
		// y = x.Uint64()
	}
	usink = uint64(y)
}
func Benchmark128starstar(b *testing.B) {
	var y uint64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.Uint64()
	}
	usink = y
}
func Benchmark128starstarGlobalRand(b *testing.B) {
	var y uint64
	for n := 0; n < b.N; n++ {
		y = Uint64()
	}
	usink = y
}
func BenchmarkPCG(b *testing.B) {
	var y uint64
	x := &exprand.PCGSource{}
	x.Seed(1)
	for n := 0; n < b.N; n++ {
		y = x.Uint64()
	}
	usink = y
}
func BenchmarkPCGMTJ(b *testing.B) {
	var y uint64
	x := pcg.NewPCG64()
	x = x.Seed(1, 2, 3, 4)
	for n := 0; n < b.N; n++ {
		y = x.Random()
	}
	usink = y
}
func BenchmarkPCGsourceInterface(b *testing.B) {
	var y uint64
	x := exprand.New(exprand.NewSource(1))
	for n := 0; n < b.N; n++ {
		y = x.Uint64()
	}
	usink = y
}
func Benchmark128sourceInterface(b *testing.B) {
	var y uint64
	z := NewXoro(1)
	x := exprand.New(&z)
	for n := 0; n < b.N; n++ {
		y = x.Uint64()
	}
	usink = y
}
func Benchmark256starstar(b *testing.B) {
	var y uint64
	x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		y = x.Uint64()
	}
	usink = y
}
func Benchmark256starstarGonum(b *testing.B) {
	var y uint64
	x := gonum.NewXoshiro256starstar(1)
	for n := 0; n < b.N; n++ {
		y = x.Uint64()
	}
	usink = y
}
func Benchmark256plus(b *testing.B) {
	var y uint64
	x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		y = x.Xoshiro256plus()
	}
	usink = y
}
func Benchmark256plusGonum(b *testing.B) {
	var y uint64
	x := gonum.NewXoshiro256plus(1)
	for n := 0; n < b.N; n++ {
		y = x.Uint64()
	}
	usink = y
}
func Benchmark256plusVpxyz(b *testing.B) {
	var y uint64
	x := vpxyz.XoroShiro256Plus{}
	x.Seed(1)
	for n := 0; n < b.N; n++ {
		y = x.Uint64()
	}
	usink = y
}
func Benchmark256plusplus(b *testing.B) {
	var y uint64
	x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		y = x.Xoshiro256plusplus()
	}
	usink = y
}
