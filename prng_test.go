package prng_test

import (
	"math"
	"math/bits"
	"testing"
	"math/rand"
	"io/ioutil"

	// . "prng"
	. "github.com/pekkizen/prng"
	"github.com/MichaelTJones/pcg"
	newrand "github.com/golang/exp/rand"
	vpxyz "github.com/vpxyz/xorshift/xoroshiro256plus"
	gonumprng "gonum.org/v1/gonum/mathext/prng"
)

func abs(x float64) float64 {
	if x > 0 {
		return x
	}
	return -x
}

func ulpsBetween(x, y float64) int64 {
	if x == y {
		return 0
	}
	if (x < 0 && y > 0) || (x > 0 && y < 0) {
		return ulpsBetween(x, 0) + ulpsBetween(y, 0)
	}
	k := math.Float64bits(y) //&^ (1 << 63)
	n := math.Float64bits(x) //&^ (1 << 63)
	i := int64(n - k)
	if i < 0 {
		return -i
	}
	return i
}

func adjacent(f1, f2 float64) bool {
	if f1 == f2 {
		return false
	}
	mean := (f1 + f2) / 2
	return mean == f1 || mean == f2
}
func TestAdjacenulpsBetween(t *testing.T) {
	x := NewXoro(1)
	const rounds int = 1e8
	for i := 0; i < rounds; i++ {
		k := x.Uint64() >> 11
		f1 := float64(k) / (1 << 53)
		f2 := float64(k+1) / (1 << 53)

		if adjacent(f1, f2) != (ulpsBetween(f1, f2) == 1) {
			t.Fatalf("Failed: f2-f1=%v ulps=%d", f2-f1, ulpsBetween(f1, f2))
		}
	}
}

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

//--------------------------------------------- benchmarks -----------//
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
// ---------------------------------------- New --------------------//
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
	x := gonumprng.NewXoshiro256starstar(1)
	for n := 0; n < b.N; n++ {
		x = gonumprng.NewXoshiro256starstar(uint64(n))
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
func BenchmarkNewNewrand(b *testing.B) {
	x := newrand.New(newrand.NewSource(1))
	for n := 0; n < b.N; n++ {
		x = newrand.New(newrand.NewSource(uint64(n)))
	}
	usink = x.Uint64()
}
func BenchmarkNewRandSlice(b *testing.B) {
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
	const mul = 1.0 / (1 << 53) 
	x := NewXoro(1)
	// z := NewXosh(1)
	for n := 0; n < b.N; n++ {

		// y = Float64_64(uint64(n) )
		// y = Float64_64(uint64(n) + (1 << 52))
		// y = x.Float64_64()
		y = x.Float64Full()
		// y = float64(x.Xoroshiro128plus() >> 11) / (1 << 53) 
		
		// y = math.Float64frombits(0x3FF<<52| (x.Xoroshiro128plus() >> 11)) - 1
	}
	fsink = y
}
func BenchmarkFloat64_64(b *testing.B) {
	var y float64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.Float64_64()
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
func BenchmarkFloat64NewRand(b *testing.B) {
	var y float64
	x := newrand.New(newrand.NewSource(1))
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
func BenchmarkRomuDuo(b *testing.B) {
	var y uint64
	x := NewXoro(1)
	for n := 0; n < b.N; n++ {
		y = x.RomuDuo()
	}
	usink = y
}
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
	x := &newrand.PCGSource{}
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
	x := newrand.New(newrand.NewSource(1))
	for n := 0; n < b.N; n++ {
		y = x.Uint64()
	}
	usink = y
}
func Benchmark128sourceInterface(b *testing.B) {
	var y uint64
	z := NewXoro(1)
	x := newrand.New(&z)
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
	x := gonumprng.NewXoshiro256starstar(1)
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
	x := gonumprng.NewXoshiro256plus(1)
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

//-------------------------------- tests ---------------------------------//
func TestOverlapProbability(t *testing.T) {
	var n float64 = 1e23
	var L float64 = (1<<64)
	var P float64 = (1<<256)
	lower, upper := OverlapProbability(n, L, P)
	t.Logf("lower= %15.14e\n", lower)
	t.Logf("upper= %15.14e", upper)
}

func TestNewOutlet(t *testing.T) {
	s := NewOutlet(1)
	x := s.NextXoro()
	y := s.NextXoro()
	r := s.Next()
	z := x
	z.Jump()
	t.Logf("x.Uint64 =\t%X", x.Uint64())
	t.Logf("y.Uint64 =\t%X", y.Uint64())
	t.Logf("r.Uint64 =\t%X", r.Uint64())
	t.Logf("z.Uint64 =\t%X", z.Uint64())
	if y.Uint64() != z.Uint64() {
		t.Errorf("y.Uint64() != z.Uint64()")
	}
	if x.Uint64() != r.Uint64() {
		t.Errorf("x.Uint64() != r.Uint64()")
	}
}

func TestResetGlobalOutlet(t *testing.T) {
	ResetGlobalOutlet(1)
	x := Next()
	ResetGlobalOutlet(2)
	y := Next()
	z := x
	z.Jump()
	t.Logf("x.Uint64 =\t%X", x.Uint64())
	t.Logf("y.Uint64 =\t%X", y.Uint64())
	t.Logf("z.Uint64 =\t%X", z.Uint64())
	if y.Uint64() != z.Uint64() {
		t.Errorf("y.Uint64() != z.Uint64()")
	}
}
func TestState(t *testing.T) {
	// x := NewXoro(1)
	// z := x
	// const size = 16
	x := NewXosh(1)
	z := x
	const size = 32
	const rounds int = 1e6
	var b []byte
	for i := 0; i < rounds; i++ {
		x.Jump()
		b = append(b, x.State()...)
		if i == 1000 {
			z = x
		}
	}
	ioutil.WriteFile("statebytes", b, 0644)
	c, _ := ioutil.ReadFile("statebytes")
	// c := b
	x.SetState(c[1000*size:])
	if x.Uint64() != z.Uint64() {
		t.Errorf("TestState: x.Uint64() != z.Uint64()")
	}
}

func TestJump(t *testing.T) {
	rr := NewXosh(1)
	rx := NewXosh(1)
	for i := 0; i < 10; i++ {
		rr.Jump()
		rx.Jump()
	}
	t.Logf("rx.Uint64 =\t%X", rx.Uint64())
	t.Logf("rr.Uint64 =\t%X", rr.Uint64())

	for i := 0; i < 10; i++ {
		rr.JumpLong()
		rx.JumpLong()
	}
	t.Logf("rx.Uint64 =\t%X", rx.Uint64())
	t.Logf("rr.Uint64 =\t%X", rr.Uint64())

	if rx.Uint64() != rr.Uint64() {
		t.Errorf("rx.Uint64() != rr.Uint64()")
	}
}
func TestSplitmixJump(t *testing.T) {
	
	var jump int64 = -(1<<32)
	// seed := uint64((1<<64) - 100)
	seed := uint64(0)
	seed2 := seed
	neg := false
	rounds := jump
	if jump < 0 {
		neg = true
		rounds = -rounds
	}
	for i := int64(0); i < rounds; i++ {
		if neg {
			seed -= 0x9e3779b97f4a7c15
			continue
		}
		seed += 0x9e3779b97f4a7c15
	}
	SplitmixJump(&seed2, jump)
	if seed != seed2 {
		t.Errorf("TestSplitmixJump: seed != seed2")
	}
}
func TestSplitmixJump2(t *testing.T) {
	
	var jump int64 = (1<<32) - 1
	// seed := uint64(0)
	seed := uint64((1<<64) - 1)
	u1 := Splitmix(&seed)
	SplitmixJump(&seed, jump)
	_ = Splitmix(&seed)
	SplitmixJump(&seed, -(jump+2))
	u2 := Splitmix(&seed)
	if u1 != u2 {
		t.Errorf("TestSplitmixJump2: u1 != u2")
	}
}

func TestJump32(t *testing.T) {
	// This test makes 2^32 calls of Uint64 and gets the same state as single Jump32
	const rounds = (1 << 32)
	y := NewXoro(1)
	z := y
	z.JumpShort()
	for i := 1; i <= rounds; i++ {
		y.Uint64()
		if z == y {
			if i == rounds {
				t.Logf("jump32 equals to 2^32 x Uint64")
				return
			}
			t.Errorf("Same state found at i =%d", i)
		}
	}
	t.Errorf("Same state not found before %d", rounds)
}
func TestJump64(t *testing.T) {
	// This test makes 2^32 jump32 and gets the same state as single Jump64
	// go test -timeout 2000s prng_test -run ^(TestJump64)$ -v
	const rounds = (1 << 32)
	y := NewXoro(1)
	z := y
	z.Jump()
	for i := 1; i <= rounds; i++ {
		y.JumpShort()
		if z == y {
			if i == rounds {
				t.Logf("jump64 equals to 2^32 x jump32")
				return
			}
			t.Errorf("Same state found at i =%d", i)
		}
	}
	t.Errorf("Same state not found before %d", rounds)
}
func TestJump96(t *testing.T) {
	// This test makes 2^32 jump64 and gets the same state as single Jum96
	// go test -timeout 2000s prng_test -run ^(TestJump96)$ -v
	const rounds = (1 << 32)
	y := NewXoro(1)
	z := y
	z.JumpLong()
	for i := 1; i <= rounds; i++ {
		y.Jump()
		if z == y {
			if i == rounds {
				t.Logf("jump96 equals to 2^32 x jump64")
				return
			}
			t.Errorf("Same state found at i =%d", i)
		}
	}
	t.Errorf("Same state not found before %d", rounds)
}
func TestNewRandSlice(t *testing.T) {
	const size = 100
	y := New(1)
	x := NewRandSlice(size, 1)
	for i := 0; i < size; i++ {
		z := y
		uy := y.Uint64()
		y = z
		ux := x[i].Uint64()
		if i < 3 {
			t.Logf("   y.Uint64=\t%X", uy)
			t.Logf("x[%d].Uint64=\t%X", i, ux)
		}
		y.Jump()
		if ux != uy {
			t.Errorf("ux != uy")
		}
	}
}
func TestNewXoshSlice(t *testing.T) {
	const size = 100
	y := NewXosh(1)
	x := NewXoshSlice(size, 1)
	for i := 0; i < size; i++ {
		z := y
		uy := y.Uint64()
		y = z
		ux := x[i].Uint64()
		if i < 4 {
			t.Logf("   y.Uint64=\t%X", uy)
			t.Logf("x[%d].Uint64=\t%X", i, ux)
		}
		y.Jump()
		if ux != uy {
			t.Errorf("ux != uy")
		}
	}
}
func TestSplitmixBitsChanged(t *testing.T) {
	const rounds int = 1e9 * 5
	var sum int
	seed := uint64(5)
	last := Splitmix(&seed)
	for i := 1; i <= rounds; i++ {
		seed = uint64(i)
		n := Splitmix(&seed)
		sum += bits.OnesCount64(last ^ n)
		last = n
	}
	ratio := float64(sum) / (64 * float64(rounds))
	t.Logf("Ratio of changed bits  %1.9f", ratio)
	if abs(ratio-0.5) > 0.00001 {
		t.Errorf("Ratio failed")
	}
}
func TestBitsChanged(t *testing.T) {
	const rounds int = 1e9 * 5
	var sum int
	// x := NewXoro(1)
	x := NewXosh(1)
	last := x.Uint64()
	for i := 0; i < rounds; i++ {
		n := x.Uint64()
		sum += bits.OnesCount64(last ^ n)
		last = n
	}
	ratio := float64(sum) / (64 * float64(rounds))
	t.Logf("Ratio of changed bits  %1.9f", ratio)
	if abs(ratio-0.5) > 0.00001 {
		t.Errorf("Ratio failed")
	}
}

func TestUint64LowHigh(t *testing.T) {
	const rounds int = 1e9 * 2
	const size uint64 = 1e13
	const failLim = 1e-1
	ResetGlobalOutlet(1)
	r := NextXoro()
	// r := NextXosh()
	low := 0
	high := 0
	var n uint64
	e := float64(size) / (1 << 64) * float64(rounds)
	expected := int(e + 0.5)
	for i := 0; i < rounds; i++ {
		n = r.Uint64()
		if n < size {
			low++
		}
		if n > (1<<64)-1-size {
			high++
		}
	}
	t.Logf("low        %d", low)
	t.Logf("high       %d", high)
	t.Logf("expected   %d", expected)
	r1 := abs(float64(low-expected) / float64(expected))
	r2 := abs(float64(high-expected) / float64(expected))
	if r1 > failLim || r2 > failLim {
		t.Errorf("Fail limit exeeded")
	}
}
func TestFloat64Tab(t *testing.T) {

	const rounds int = 1e9
	const cells = 10000
	const failLim = 1.5e-2
	// x := newrand.New(newrand.NewSource(1))
	x := NewXoro(1)
	// x := NewXosh(1)

	var tab [cells]int
	for i := 0; i < rounds; i++ {
		// f := float64(x.Uint64()>> 11) / (1<<53)
		f := x.Float64_64()
		tab[int(cells*f)]++
	}
	expected := rounds / cells
	failed := 0
	for i := 0; i < cells; i++ {
		list := i%500 == 0
		diff := tab[i] - expected
		reldiff := float64(diff) / float64(expected)
		if list {
			t.Logf("%d %d %4.2e", i, tab[i], reldiff)
		}

		if abs(reldiff) > failLim && failed < 20 {
			failed++
			t.Logf("%d %d %4.2e", i, tab[i], reldiff)
			t.Errorf("Fail limit exeeded")
		}
	}
}
func TestUint64nTab(t *testing.T) {
	Seed(1)
	const rounds int = 1e9
	const cells = 10000
	const failLim = 1.5e-2
	r := NewXosh(1)
	// r := NewXoro(1)

	var tab [cells]int
	for i := 0; i < rounds; i++ {
		tab[r.Uint64()%cells]++
	}
	expected := rounds / cells
	failed := 0
	for i := 0; i < cells; i++ {
		list := i%500 == 0
		diff := tab[i] - expected
		reldiff := float64(diff) / float64(expected)
		if list {
			t.Logf("%d %d %4.2e", i, tab[i], reldiff)
		}
		if abs(reldiff) > failLim && failed < 20 {
			failed++
			t.Logf("%d %d %4.2e", i, tab[i], reldiff)
			t.Errorf("Fail limit exeeded")
		}
	}
}
func TestFloat64FourSlots(t *testing.T) {

	const rounds int = 1e8 
	var slotsize = 1e-5
	const failLim = 6e-2
	var tab [4]int
	x := NewXoro(18)
	// z := NewXosh(18)
	for i := 0; i < rounds; i++ {

		f := x.Float64_64()
		// f := x.Float64_128()

		if f < slotsize {
			tab[0]++
		}
		if f >= 0.5 && f < 0.5+slotsize {
			tab[1]++
		}
		if f >= 1-slotsize && f < 1 {
			tab[2]++
		}
		r := x.Float64()
		for r+slotsize > 1 {
			r = x.Float64()
		}
		if f >= r && f < r+slotsize {
			tab[3]++
		}
	}
	expected := int(slotsize*float64(rounds) + 0.5)
	t.Logf("low      %d", tab[0])
	t.Logf("middle   %d", tab[1])
	t.Logf("high     %d", tab[2])
	t.Logf("random   %d", tab[3])
	t.Logf("expected %d", expected)
	sum := float64(tab[0] + tab[1] + tab[2] + tab[3])
	t.Logf("mean     %d", int(sum/4+0.5))
	for i := 0; i < 4; i++ {
		sum += float64(tab[i])
		diff := tab[i] - expected
		reldiff := float64(diff) / float64(expected)
		t.Logf("%d %d %4.2e", i, tab[i], reldiff)
		if abs(reldiff) > failLim {
			t.Fatalf("Fail limit exeeded %d %d %v", i, tab[i], reldiff)
		}
	}
}
func TestFloat64NearZeroSlot(t *testing.T) {

	const rounds int = 1e8
	var slotsize = 1e-6
	const failLim = 6e-2
	hit := 0
	x := NewXoro(881)
	for i := 0; i < rounds; i++ {
		// f := x.Float64()
		f := x.Float64_64()
		
		if f < slotsize {
			hit++
		}
		
	}
	expected := int(slotsize*float64(rounds) + 0.5)
	t.Logf("hits      %d", hit)
	t.Logf("expected %d", expected)

}
func TestUint64ToFloatDistribution(t *testing.T) {

	x := NewXoro(1)
	const rounds = (1 << 28)
	const wid = 0 // scaling with 2^wid keeps most above properties
	const equidist = 1.0 / (1 << (53 - wid))
	const minAdjacent = (1 << 52)

	for i := 0; i < rounds; i++ {

		k := x.Uint64() >> 11
		f1 := float64(k) / (1 << (53 - wid))
		f2 := float64(k+1) / (1 << (53 - wid))

		j := uint64(f1 * (1 << (53 - wid)))
		if j != k {
			t.Fatalf("Inverse function failed: j - k =%d", j-k)
		}
		if f2-f1 != equidist {
			t.Fatalf("Equidistance failed: f1= %v f2%v", f1, f2)
		}
		if k >= minAdjacent && !adjacent(f1, f2) {
			t.Fatalf("adjacent failed: ulps=%d f1=%v", ulpsBetween(f1, f2), f1)
		}

	}
}
func TestUint64ToFloatDistribution2(t *testing.T) {
	// same as above but sequential

	const equidist = 1.0 / (1 << 53)
	const div = (1 << 53)

	var rounds uint64 = (1 << 30)
	var start uint64 = (1 << 52) - 1000

	stop := start + rounds
	if stop > (1<<53)-1 {
		stop = (1 << 53) - 1
	}
	f1 := float64(start) / div

	for i := uint64(start) + 1; i < stop; i++ {
		f2 := float64(i) / div

		if f2-f1 != equidist {
			t.Fatalf("Equidistance failed: f1= %v f2= %v", f1, f2)
		}
		k := uint64(f2 * div)
		if k != i {
			t.Fatalf("Inverse function failed: k - i =%d", k-i)
		}
		if f1 >= 0.5 && !adjacent(f1, f2) {
			t.Fatalf("adjacent failed: ulps=%d f1=%v", ulpsBetween(f1, f2), f1)
		}
		f1 = f2
	}
}
func TestUint64ToFloatCompareMethods52(t *testing.T) {
	// 52-bit division method and the 52-bit explicit method are same
	x := NewXoro(1)
	const rounds int = 1e9
	for i := 0; i < rounds; i++ {

		k := x.Uint64() >> 12
		// f1 := float64(k) / (1 << 52)
		f1 := float64(k) * (1.0 / (1 << 52))
		f2 := math.Float64frombits(1023<<52|k) - 1
		if f1-f2 != 0 {
			t.Fatalf("Methods not same: diff = %v", f1-f2)
		}
	}
}
func TestUint64ToFloatCompareMethods53(t *testing.T) {
	// 53-bit division method and the 53-bit explicit method are same
	x := NewXoro(1)
	const rounds int = 1e9
	for i := 0; i < rounds; i++ {

		k := x.Uint64() >> 11
		f1 := float64(k) / (1 << 53)
		f2 := math.Float64frombits(1023<<52 ^ k)
		if f2 >= 1 {
			f2 = (f2 - 1) * 0.5
		}
		if f1-f2 != 0 {
			t.Fatalf("Methods not same: diff = %v", f1-f2)
		}
	}
}
func TestFloat64_64Distribution(t *testing.T) {

	var rounds int = 1e8*2
	x := NewXoro(1)
	
	for i := 0; i < rounds; i++ {

		u := x.Uint64() 
		u >>= u % 64
		f1 := Float64_64(u) 
		zeros := uint64(bits.LeadingZeros64(u))
		if zeros > 11 {
			zeros = 11
		}
		// u2 = u + 1 to the least significant float position
		u2 := u << zeros
		u2 += (1 << 11)
		u2 >>= zeros

		f2 := Float64_64(u2) 
		zeros2 := uint64(bits.LeadingZeros64(u2))
			if zeros2 > 11 {
			zeros2 = 11
		}
		if zeros == zeros2 && f2-f1 != 1.0 / (1 << 53) / float64(uint64(1 << zeros)) {
			t.Logf("Equidistance failed: i=%d zeros=%d", i, zeros)
			t.Logf("Distance= %v", f2-f1)
			t.Fatalf("f1= %v f2= %v", f1, f2)
		}
		if f1 >= 1.0 / (1<<12) && !adjacent(f1, f2) {
			t.Logf("adjacent failed: f1=%v f2=%v", f1, f2)
			t.Fatalf("ulps=%d", ulpsBetween(f1, f2))
		}
		z := uint64(f1 * (1 << 53) * float64(uint64(1 << zeros))) >> zeros
		if z != u >> 11 {
			t.Logf("Inverse failed: i=%d zeros=%d", i, zeros)
			t.Fatalf("i=%d z=%d u >> 11 =%d", i, z, u >> 11)
		}
	}
}
func TestFloat64_128Distribution(t *testing.T) {

	var rounds int = 1e8
	x := NewXoro(1)
	for i := 0; i < rounds; i++ {
		hi := x.Uint64() >> 1
		hi = 0
		lo := x.Uint64()
		f1 := Float64_128(hi, lo) 
		zeros := uint64(bits.LeadingZeros64(lo))
		if zeros > 11 {
			zeros = 11
		}
		hi2 := hi
		z := lo << zeros
		z >>= 11
		z += 1
		lo2 := z << 11
		lo2 >>= zeros
		f2 := Float64_128(hi2, lo2) 
		zeros2 := uint64(bits.LeadingZeros64(lo2))
		if zeros2 > 11 {
			zeros2 = 11
		}
		if zeros == zeros2 && f2-f1 != 1.0 / (1 << 53) / float64(uint64(1 << zeros)) / (1 << 64) {
			t.Logf("Equidistance failed: i=%d zeros=%d", i, zeros)
			t.Logf("lo= %v lo2= %v", lo, lo2)
			t.Logf("Distance= %v", f2-f1)
			t.Fatalf("f1= %v f2= %v", f1, f2)
		}
		if f1 >= 1.0 / (1<<76) && !adjacent(f1, f2) {
			t.Logf("adjacent failed: f1=%v f2=%v", f1, f2)
			t.Fatalf("i=%d ulps=%d",i, ulpsBetween(f1, f2))
		}
	}
}
func TestFloat64FullDoubleDistribution(t *testing.T) {

	var rounds int = 1e8
	listed := 0
	x := NewXoro(1)
	for i := 0; i < rounds; i++ {
		
		f1, f2 := x.Float64FullDouble() 
		if  !adjacent(f1, f2) {
			t.Logf("adjacent failed: f1=%v f2=%v", f1, f2)
			t.Fatalf("i=%d ulps=%d",i, ulpsBetween(f1, f2))
		}
		lim := 1.0 / (1 << 256)
		lim *= lim
		if f1 < lim && listed < 50 {
			listed++
			t.Logf("f1=%v f2=%v", f1, f2)
		}
	}
}
func TestFloat64FullSmall(t *testing.T) {

	var hi, lo uint64
	rounds := 15
	const ulps = 200
	hi = 1
	lo = (1<<14)
	// lo = (1<<61)
	f1 := Float64FullSmall(hi, lo, rounds)
	lo += (1<<14) * ulps
	f2 := Float64FullSmall(hi, lo, rounds)
	
	t.Logf("%v", f1) 
	t.Logf("%v", f2)
	t.Logf("%X", math.Float64bits(f1))
	t.Logf("%X", math.Float64bits(f2))
	t.Logf("ulps %v", ulpsBetween(f2, f1))
	t.Logf("adjacent %v", adjacent(f2, f1))
}