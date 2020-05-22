package prng

import (
	"math/bits"
	"math/rand"
	"testing"

	"github.com/MichaelTJones/pcg"
	exprand "github.com/golang/exp/rand"
	vpxyz "github.com/vpxyz/xorshift/xoroshiro256plus"
	gonum "gonum.org/v1/gonum/mathext/prng"
)

var usink uint64
var fsink float64
var isink int
var bsink []byte

// baseline128 is for benchmarking minimal 128-bit state generator.
func (x *Xoro) baseline128() uint64 {
	next := x.s0
	*x = Xoro{x.s1, x.s0}
	return next
}

// baseline256 is for benchmarking minimal 256-bit state generator.
func (x *Xosh) baseline256() uint64 {
	next := x.s0
	*x = Xosh{x.s3, x.s0, x.s1, x.s2}
	return next
}

func workerbaseline(i int, c chan float64) {
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
		// go workerbaseline(i, c)
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
	for n := 0; n < b.N; n++ { // 1 CPU cycle, 0.28 ns @ 3.6 MHz
		i = n
	}
	isink = i
}

func BenchmarkBaseline128(b *testing.B) {
	var y uint64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.baseline128()
	}
	usink = y
}
func BenchmarkBaseline256(b *testing.B) {
	var y uint64
	x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		y = x.baseline256()
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
	x := NewXoro(1)
	sb := make([]byte, 0, 16*1000)
	for n := 0; n < b.N; n++ {
		for i := 0; i < 1000; i++ {
			sb = append(sb, x.State()...)
		}
	}
	x.SetState(sb[500*16:])
	usink = x.Uint64()
}
func BenchmarkGetState(b *testing.B) {
	x := NewXoro(1)

	sb := make([]byte, 16*1000)

	for n := 0; n < b.N; n++ {
		for i := 0; i < 1000; i++ {
			x.GetState(sb[i*16:])
		}
	}
	x.SetState(sb[500*16:])
	usink = x.Uint64()
}
func BenchmarkSetState(b *testing.B) {
	// x := NewXosh(1)
	x := NewXoro(1)
	s := x.State()
	for n := 0; n < b.N; n++ {
		x.SetState(s)
		if n == 1000 {
			x.Uint64()
			s = x.State()
		}
	}
	x.SetState(s)
	usink = x.Uint64()
}

func BenchmarkBitsLeadingZeros(b *testing.B) {
	var zeros uint64
	for n := 0; n < b.N; n++ {
		zeros = uint64(bits.LeadingZeros64(uint64(n) * 0x9e3779b97f4a7c15))
	}
	usink = zeros
}

// ---------------------------------------- New generator--------------------//
func BenchmarkNewPrng(b *testing.B) {
	// var x Xoro
	var x Xosh
	// var x Prng
	for n := 0; n < b.N; n++ {
		// x = New(uint64(n))
		// x = NewXoro(uint64(n))
		x = NewXosh(uint64(n))
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

func BenchmarkNewRand(b *testing.B) {
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
func BenchmarkNewPrngSlice(b *testing.B) {
	const size = 1000000
	var x []Prng
	for n := 0; n < b.N; n++ {
		x = NewPrngSlice(size, 1)
	}
	usink = x[size/2].Uint64()
}
func BenchmarkNextPrng(b *testing.B) {
	// var x Xoro
	// var x Xosh
	var x Prng
	for n := 0; n < b.N; n++ {
		x = Next()
		// x = NextXoro()
		// x = NextXosh()
	}
	usink = x.Uint64()
}

func BenchmarkLdexp(b *testing.B) {
	var y float64
	for n := 0; n < b.N; n++ {
		// y = 1.2 * twoToMinus(uint64(n) & 255)
		y = ldexp(1.2, uint64(n)&255)
		// y = 1.2 * math.Float64frombits((1023 - uint64(n) & 255) << 52)
	}
	fsink = y
}

//-----------------------------------------------------Float64--------//
func BenchmarkMiscFormulas(b *testing.B) {
	var y float64
	var u uint64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {

		// y = float64(x.Uint64() &^ (1<<63))
		// y = float64(x.Uint64() | (1<<63))
		// y = float64(x.Uint64())

		u = x.Uint64() | (1 << 63)

		// y = float64(u )												// 1.7
		// y = float64(u & 1)											// 1.43
		y = float64((u>>10+1)>>1) / (1 << 53) // 1.7
		// y = float64(u | 1) / (1<<64)									// 2.0
		// y = math.Float64frombits(1022 << 52 | (u >> 11 + 1) >> 1)	// 1.55
		// y = math.Float64frombits(1022 << 52 | u >> 12)				// 1.40

		// y = float64(x.Xoroshiro128plus() >> 11) / (1 << 53)
		// y = math.Float64frombits(1023 << 52 | (x.Xoroshiro128plus() >> 12)) - 1
	}
	fsink = y
}

func BenchmarkFloat64_64(b *testing.B) {
	var y float64
	x := NewXoro(1)
	// x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		y = x.Float64_64()
		// y = x.float64_64Div()
		// y = x.float64_64Tab()
		// y = x.Float64_64R()
	}
	fsink = y
}

func BenchmarkFloat64full(b *testing.B) {
	var y float64
	x := NewXoro(1)
	// x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		y = x.Float64full()
		// y = x.Float64fullR()
		// y = x.float64fullDiv()
	}
	fsink = y
}

func BenchmarkFloat64_117(b *testing.B) {
	var y float64
	x := NewXoro(1)
	// x := NewXosh(1)
	for n := 0; n < b.N; n++ {
		y = x.Float64_117()
		// y = x.Float64_128()
		// y = x.Float64_117R()
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
