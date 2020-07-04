package prng

import (
	"math"
	"math/bits"
	"testing"
	"io/ioutil"
)

func abs(x float64) float64 {
	if x > 0 {
		return x
	}
	return -x
}

func TestIsInf(t *testing.T) {
	max := math.MaxFloat64
	t.Logf("%v", isInf(math.Inf(1)))
	t.Logf("%v", isInf(math.Inf(-1)))
	t.Logf("%v", isInf(math.NaN()))
	t.Logf("%v", isInf(max))
	t.Logf("%v", isInf(max + max))
}

func TestInfNaN(t *testing.T) {
	f1 := math.Inf(1)
	f2 := f1 + 1
	f3 := math.Float64frombits(math.Float64bits(f1) + 1)
	f4 := f3 + 1
	t.Logf("f1  %X %v", math.Float64bits(f1), f1)
	t.Logf("f2  %X %v", math.Float64bits(f2), f2)
	t.Logf("f3  %X %v !!!!!", math.Float64bits(f3), f3)
	t.Logf("-f3 %X %v\n", math.Float64bits(-f3), -f3)
	t.Logf("f4  %X %v", math.Float64bits(f4), f4)
	t.Logf("f3 != f3 %v", f3 != f3)
}

func TestFloat64Random(t *testing.T) {
	const rounds int = 1e8*5
	min, max, i01 := 999.0, 0.0, 0.0
	x := NewXoro(11)
	for i := 0; i < rounds; i++ {
		f := x.float64Random() 
		if f < min {
			min = f
		}
		if f > max {
			max = f
		}
		if f >= 0 && f <= 1 {
			i01++
		}

		
	}
	t.Logf("Min   %v", min)
	t.Logf("Max   %v", max)
	t.Logf("0 - 1 %v", 100*float64(i01) / float64(rounds))
}

func TestNextToZero(t *testing.T) {
    const rounds int = 1e8*3
	x := NewXoro(1)
	for i := 0; i < rounds; i++ {
		f1 := x.float64Random() 
		switch i {
		case 1:
			f1 = 0
		case 2:
			f1 = math.Inf(1)
		case 3:
			f1 = math.Inf(-1)
		case 4:
			f1 = math.MaxFloat64
		}
		f2 := nextToZero(f1)
		f3 := math.Nextafter(f1, 0)
		if f2 != f3 {
			t.Logf("Nextafter i=   %d", i)
			t.Logf("F1  %v", f1)
			t.Logf("F2  %v", f2)
			t.Logf("F3  %v", f3)
		}
		ulps := ulpsBetween(f1, f2)
		if ulps != 1 && f1 > 0 && !isInf(f1) {	
			t.Logf("Ulps %v", ulps)
			t.Logf("i=   %d", i)
			t.Logf("F1   %v", f1)
			t.Fatalf("F2   %v", f2)
		}
	}
}

func TestNextToZeroFast(t *testing.T) {
    const rounds int = 1e8
	x := NewXoro(1)
	for i := 0; i < rounds; i++ {
		f1 := x.float64Random() 
		switch i {
		case 1:
			f1 = 0
		case 2:
			f1 = math.Nextafter(0x1p-1022, 1)
		case 3:
			f1 = math.MaxFloat64
		}
		f2 := f1 * (1 - 0x1p-53)
		f3 := math.Nextafter(f1, 0)
		if f2 != f3 && f2 > 0x1p-1072 {
			t.Logf("Nextafter i=   %d", i)
			t.Logf("F1  %v", f1)
			t.Logf("F2  %v", f2)
			t.Fatalf("F3  %v", f3)
		}
	}
}

func TestNextFromZero(t *testing.T) {
    const rounds int = 1e8*3
	x := NewXoro(1)
	for i := 0; i < rounds; i++ {
		f1 := x.float64Random() 
		switch i {
		case 1:
			f1 = 0
		case 2:
			f1 = math.Inf(1)
		case 3:
			f1 = math.Inf(-1)
		case 4:
			f1 = math.MaxFloat64
		}
		f2 := nextFromZero(f1)
		d := math.Inf(1)
		if f1 < 0 { d = -d }
		f3 := math.Nextafter(f1, d)
		if f2 != f3 {
			t.Logf("Nextafter i=   %d", i)
			t.Logf("F1  %v", f1)
			t.Logf("F2  %v", f2)
			t.Logf("F3  %v", f3)
			
			t.Logf("F1  %X" , math.Float64bits(f1))
			t.Logf("F2  %X" , math.Float64bits(f2))
			t.Fatalf("F3  %X" , math.Float64bits(f3))
		
		}
		ulps := ulpsBetween(f1, f2)
		// if ulps != 1 && abs(f2) <= math.MaxFloat64 {
		if ulps != 1 && !isInf(f2) {
			t.Logf("Ulps %v", ulps)
			t.Logf("i=   %d", i)
			t.Logf("F1=  %v", f1)
			t.Fatalf("F2=  %v", f2)
		}
	}
}


func TestAdjacent(t *testing.T) {
    const rounds int = 1e9
	x := NewXoro(1)
	for i := 0; i < rounds; i++ {
		// f1 := x.Float64full() 
		f1 := x.float64Random()
		switch i {
		case 1:
			f1 = 0
		case 2:
			f1 = math.Inf(1)
		case 3:
			f1 = math.Inf(-1)
		case 4:
			f1 = math.MaxFloat64
		}
		
		f2 := nextFromZero(f1)
			if !adjacent(f1, f2) && !isInf(f2) {
			t.Logf("Not adjacent")
			t.Logf("Ulps %v", ulpsBetween(f1, f2))
			t.Logf("i=   %d", i)
			t.Logf("F1=  %v", f1)
			t.Fatalf("F2=  %v", f2)
		}
	}
}

func TestUlpsBetween(t *testing.T) {
	const upper = 1 - 0x1p-53
	const maxFloat = 0x1p1023 * (2 - 0x1p-52)
	const minFloat = 0x1p-1074

	t.Logf("+Inf       %v", ulpsBetween(math.Inf(1), 1.0))
	t.Logf("-Inf       %v", ulpsBetween(math.Inf(-1), 1.0))
	t.Logf("NaN        %v", ulpsBetween(math.NaN(), 1.0))
	t.Logf("+Inf +Inf  %v", ulpsBetween(math.Inf(1), math.Inf(1)))
	t.Logf("0.5 - 1    %v", math.Log2(float64(ulpsBetween(0.5, 1.0))))
	t.Logf("0 - 1      %v", math.Log2(float64(ulpsBetween(0, 1.0))))
 	t.Logf("subNorm    %v", math.Log2(float64(ulpsBetween(0, 0x1p-1022))))
	t.Logf("maxFloats  %v", math.Log2(float64(ulpsBetween(-maxFloat, maxFloat))))
	t.Logf("minFloats  %v", ulpsBetween(-minFloat, minFloat))
	t.Logf("log2(ulps) %v", math.Log2(float64(ulpsBetween(0x1p-1023, 0x1p-1022))))
    t.Logf("log2(ulps) %v", math.Log2(float64(ulpsBetween(0, 0x1p-1023))))
 
    for i := uint64(1); i < 17;  i++ {
        t.Logf("\t%v \t%d \t%d", 
        math.Log2(float64(ulpsBetween(1.0/float64(uint64(1<<i)), upper))), i, 1<<i)
    }
}

func TestUlp(t *testing.T) {
    const rounds int = 1e8*3
	x := NewXoro(1)
	for i := 0; i < rounds; i++ {
		f1 := x.float64Random()
		f2 := f1 + ulp(f1)
		if !adjacent(f1, f2) {
			t.Logf("Not adjacent")
			t.Logf("Ulps %v", ulpsBetween(f1, f2))
			t.Logf("i    %d", i)
			t.Logf("F1   %v", f1)
			t.Logf("F2   %v", f2)
			t.Logf("F1   %X" , math.Float64bits(f1))
			t.Fatalf("F2   %X" , math.Float64bits(f2))
		}
	}
}

func TestOverlapProbability(t *testing.T) {
	var n float64 = 1e6
	var L float64 = 1<<54
	var P float64 = 1<<128
	lower, upper := OverlapProbability(n, L, P)
	t.Logf("lower= %15.16e\n", lower)
	t.Logf("upper= %15.16e", upper)
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
	x := NewXosh(1)
	z := x
	// size := XoroStateSize
	size := XoshStateSize
	const rounds int = 1e4
	var b []byte
	s := make([]byte, size )
	for i := 0; i < rounds; i++ {
		x.Jump()
		// b = append(b, x.State()...)
		x.WriteState(s)
		b = append(b, s...)
		if i == 1000 {
			z = x
		}
	}
	ioutil.WriteFile("statebytes", b, 0644)
	c, _ := ioutil.ReadFile("statebytes")
	x.ReadState(c[1000*size:])
	if x.Uint64() != z.Uint64() {
		t.Errorf("TestState: x.Uint64() != z.Uint64()")
	}
}

func TestJump(t *testing.T) {
	rr := NewXosh(1)
	rx := rr
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
	const rounds int = (1<<32)
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
	// go test -timeout 2000s prng -run ^(TestJump64)$ -v
	const rounds int = (1<<32)
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
	// go test -timeout 2000s prng -run ^(TestJump96)$ -v
	const rounds int = (1<<32)
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
func TestNewPrngSlice(t *testing.T) {
	const size = 100
	y := New(1)
	x := NewPrngSlice(size, 1)
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

// ---------------------------------- testing generator output ------------
func TestSplitmixBitsChanged(t *testing.T) {
	const rounds int = 1e9 * 3
	var sum int
	seed := uint64(0)
	last := Splitmix(&seed)
	for i := 1; i <= rounds; i++ {
		seed = uint64(i) // seeding by index
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

func TestUint64LowHigh(t *testing.T) {
	const rounds int = 1e9 * 2
	const size uint64 = 1e13
	const failLim = 1e-1
	x := NewXoro(1)
	low, high := 0, 0
	var n uint64
	e := float64(size) / (1<<64) * float64(rounds)
	expected := int(e + 0.5)
	for i := 0; i < rounds; i++ {
		n = x.Uint64()
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
func TestUint64nTab(t *testing.T) {

	const rounds int = 1e9
	const cells = 10000
	const failLim = 1.5e-2
	x := NewXoro(1)
	var tab [cells]int
	for i := 0; i < rounds; i++ {
		tab[x.Uint64()%cells]++
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
func TestFloat64Tab(t *testing.T) {

	const rounds int = 1e9
	const cells = 100000
	const failLim = 5e-2
	x := NewXoro(1)
	var tab [cells]int
	for i := 0; i < rounds; i++ {
		f := x.Float64_64()
		tab[int(cells*f)]++
	}
	expected := rounds / cells
	failed := 0
	for i := 0; i < cells; i++ {
		list := i%5000 == 0
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
	const rounds int = 1e8 * 4
	var slotsize =1.0/(1<<53) * (1<<32)
	const failLim = 1e-1
	var tab [4]int
	failed := 0
	x := NewXoro(12)

	for i := 0; i < rounds; i++ {
		// f := x.RandomReal()
		// f := x.Float64_64()
		f := x.float64div63()
		// f := x.Float64_117()
		// f := x.Float64full()

		if f < slotsize {
			tab[0]++
		}
		if f >= 0.5 && f < 0.5+slotsize {
			tab[1]++
		}
		if f >= 1-slotsize && f < 1 {
			tab[2]++
		}
		r := x.Float64full()
		for r+slotsize > 1 {
			r = x.Float64full()
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
	t.Logf("Relative deviation")
	
	for i := 0; i < 4; i++ {
		// sum += float64(tab[i])
		diff := tab[i] - expected
		reldiff := float64(diff) / float64(expected)
		t.Logf("         %d  %4.2e", tab[i], reldiff)
		if abs(reldiff) > failLim {
			failed++
			t.Logf("Fail limit exeeded %d %d %v", i, tab[i], reldiff)
		}
	}
	if failed > 0 {
		t.Fatalf("Failed: %d", failed)
	}
	
}

func TestFloat64NearZeroSlot(t *testing.T) {
	const rounds int = 1e9 * 2
	var slotsize = 1.0/(1<<53) * (1<<30)
	hit := 0
	x := NewXoro(1)
	// x := NewXosh(1)
	for i := 0; i < rounds; i++ {
		f := x.Float64_64()
		// f := x.Float64_117()
		// f := x.Float64Bisect(false)
		// f := x.Float64full()
		// f := x.RandomReal()
		if f < slotsize {
			hit++
		}
	}
	expected := int(slotsize*float64(rounds) + 0.5)
	t.Logf("hits      %d", hit)
	t.Logf("expected  %d", expected)
}

// -------------------------------------------------------52/53/63-bit divide
func Test53BitDivideDistribution(t *testing.T) {

	x := NewXoro(1)
	const rounds int = 1e8*3
	const wid = 0 // scaling with 2^wid keeps most above properties
	const equidist = 1.0 / (1 << (53 - wid))
	const minAdjacent = (1<<52)

	for i := 0; i < rounds; i++ {

		k := x.Uint64() >> 11 
	    f1 := float64(k) / (1 << (53 - wid))
		f2 := float64(k+1) / (1 << (53 - wid))

		j := uint64(f1 * (1 << (53 - wid)))
		if j != k {
			t.Fatalf("Inverse function failed: j - k =%d", j-k)
		}
		if f2-f1 != equidist {
			t.Fatalf("Equidistance failed: f1= %v f2= %v", f1, f2)
		}
		if k >= minAdjacent && !adjacent(f1, f2) {
			t.Fatalf("adjacent failed: ulps=%d f1=%v", ulpsBetween(f1, f2), f1)
		}

	}
}

func Test52BitExplicitVsDivide(t *testing.T) {
	// 52-bit division method and the 52-bit explicit method are same
	x := NewXoro(1)
	const rounds int = 1e8
	for i := 0; i < rounds; i++ {
		k := x.Uint64() >> 12
		f1 := float64(k) / (1<<52)
		f2 := math.Float64frombits(1023<<52|k) - 1
		if f1 == f2 {
			continue
		}
		t.Logf("Methods not same: diff = %v", f1-f2)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
}

func Test63BitDivideUlps(t *testing.T) {
	x := NewXoro(1)

	const rounds int = 1e9
	const scale = 0x1p-63 * (1 - 0x1p-53) 
	for i := 0; i < rounds; i++ {
		k := x.Uint64() >> 1
		f1 := float64(k) / (1<<63)
		f2 := float64(k) * scale
		ulps := ulpsBetween(f1, f2)
		if ulps == 1 {
			continue
		}
		t.Logf("Upls not 1: f1-f2  %v", f1-f2)
		t.Logf("Ulps %v", ulps)
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
}
// ----------------------------------------------------------Float64_64
func Test_64_64Distribution(t *testing.T) {

	const rounds int = 1e8
	x := NewXoro(2)
	
	for i := 0; i < rounds; i++ {

		u := x.Uint64() 
		u >>= u % 64 // 0 - 63 leading zeros
		f1 := float64_64(u) 
		zeros := uint64(bits.LeadingZeros64(u))
		if zeros > 11 {
			zeros = 11
		}
		u2 := u << zeros
		u2 += (1 << 11) // next adjacent
		u2 >>= zeros

		f2 := float64_64(u2) 
		zeros2 := uint64(bits.LeadingZeros64(u2))
		if zeros2 > 11 {
			zeros2 = 11
		}
		// if zeros == zeros2 && f2-f1 != 1.0 / (1 << 53) / float64(uint64(1 << zeros)) {
		if f2-f1 != 1.0 / (1<<53) / float64(uint64(1 << zeros)) {
			t.Logf("Distance failed: i=%d zeros=%d", i, zeros)
			t.Logf("Distance= %v", f2-f1)
			t.Fatalf("f1= %v f2= %v", f1, f2)
		}
		if f1 >= 1.0 / (1<<12) && !adjacent(f1, f2) {
			t.Logf("Adjacent failed: f1=%v f2=%v", f1, f2)
			t.Fatalf("ulps=%d", ulpsBetween(f1, f2))
		}
		z := uint64(f1 * (1<<53) * float64(uint64(1 << zeros))) >> zeros
		if z != u >> 11 {
			t.Logf("Inverse failed: i=%d zeros=%d", i, zeros)
			t.Fatalf("z=%d u >> 11 =%d", z, u >> 11)
		}
	}
}
func Test_64_64Spacing(t *testing.T) {
	var rounds int = 1e8
	x := NewXoro(2)
	for i := 0; i < rounds; i++ {
		u := x.Uint64() 
        f1 := float64_64(u)
		zeros := uint64(bits.LeadingZeros64(u))
		if zeros > 11 {
			zeros = 11
		}
		u2 := u << zeros >> 11
		u2++
		u2 <<= 11
		u2 >>= zeros
        f2 := float64_64(u2)
  		if (adjacent(f1, f2)) && f1 >= 1.0/(1<<12) {
			continue
		}
		if f2 - f1 == 1.0/(1<<64) {
			continue
        }
        t.Logf("i           %d", i)
		t.Logf("Ulps        %v", ulpsBetween(f1, f2))
		t.Logf("Log2(f2-f1) %v", math.Log2(f2-f1))
		t.Logf("F1=         %v", f1)
		t.Fatalf("F2=         %v", f2)
	}
}
func Test_64_64RSpacing(t *testing.T) {
	var rounds int = 1e8
	x := NewXoro(2)
	for i := 0; i < rounds; i++ {
		u := x.Uint64() 
        f1 := float64_64R(u)
		zeros := uint64(bits.LeadingZeros64(u))
		if zeros > 11 {
			zeros = 11
		}
		u2 := u << zeros >> 11
		u2++
		u2 <<= 11
		u2 >>= zeros
        f2 := float64_64R(u2)
		if (f1 == f2 || adjacent(f1, f2)) && f1 >= 1.0/(1<<12) {
			continue
		}
		if f2 - f1 == 1.0/(1<<64) {
			continue
        }
        t.Logf("i           %d", i)
		t.Logf("Ulps        %v", ulpsBetween(f1, f2))
		t.Logf("Log2(f2-f1) %v", math.Log2(f2-f1))
		t.Logf("F1=         %v", f1)
		t.Fatalf("F2=         %v", f2)
	}
}

// set const twistedUint64 = true for following 
func Test_64_64Div(t *testing.T) {
	var rounds int = 1e8 
	x1 := NewXoro(1)
	x2 := x1
	for i := 0; i < rounds; i++ {
		f1 := x1.float64_64Div() 
		f2 := x2.Float64_64()
		if f1 == f2  {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
}
func Test_64_64Tab(t *testing.T) {
	var rounds int = 1e8 
	x1 := NewXoro(1)
	x2 := x1
	for i := 0; i < rounds; i++ {
		f1 := x1.float64_64Tab() 
		f2 := x2.Float64_64()
		if f1 == f2  {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
}
func Test_64R_64DivR(t *testing.T) {
	var rounds int = 1e8 
	x1 := NewXoro(1)
	x2 := x1
	for i := 0; i < rounds; i++ {
		f1 := x1.float64_64DivR() 
		f2 := x2.Float64_64R()
		if f1 == f2  {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
}
func Test_64R_64TabR(t *testing.T) {
	var rounds int = 1e8 
	x1 := NewXoro(1)
	x2 := x1
	for i := 0; i < rounds; i++ {
		f1 := x1.float64_64TabR() 
		f2 := x2.Float64_64R()
		if f1 == f2  {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
}
func Test_64_Bisect(t *testing.T) {
	var rounds int = 1e7
	x1 := NewXoro(2)
	for i := 0; i < rounds; i++ {
		x2 := x1
		f1 := x1.Float64_64() 
		f2 := x2.Float64Bisect(false)
		if f1 == f2 || f1 < 1.0 / (1 << 12) {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
}
func Test_64R_Bisect(t *testing.T) {
	var rounds int = 1e7
	x1 := NewXoro(2)
	for i := 0; i < rounds; i++ {
		x2 := x1
		f1 := x1.Float64_64R() 
		f2 := x2.Float64Bisect(true)
		if f1 == f2 || f1 < 1.0 / (1 << 11) {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
}

// ---------------------------------------------------------Float64_117
// set const twistedUint64 = true for following 
func Test_117_Bisect(t *testing.T) {
	var rounds int = 1e8
	x1 := NewXoro(2)
	for i := 0; i < rounds; i++ {
		x2 := x1
		f1 := x1.Float64_117() 
		// f2 := x2.Float64Bisect(false)
		f2 := x2.Float64full()
		if f1 == f2 {
			continue
        }
        if f1 < 1.0/(1<<65)  {
			continue
        }
 		t.Logf("Not same: i=%d" , i)
        t.Logf("Ulps %v", ulpsBetween(f1, f2))
        t.Logf("Log2(f2-f1) %v", math.Log2(abs(f2-f1)))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
}
func Test_117R_Bisect(t *testing.T) {
	var rounds int = 1e8
	x1 := NewXoro(2)
	x2 := x1
	for i := 0; i < rounds; i++ {
		f1 := x1.Float64_117R() 
		// f2 := x2.Float64Bisect(true)
		f2 := x2.Float64fullR()
		x2 = x1
		if f1 == f2  {
			continue
		}
		if f1 < 1.0/(1<<65)  {
			continue
        }
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
}
// ----------------------------------------------------Float64full
func Test_64fullSingles(t *testing.T) {
	var hi, lo uint64
	rounds := 15   // !!!!!!!!!
	const ulps = 100
	hi = 1
	f1 := float64fulltest(hi, lo, rounds)
	lo += (1<<14) * ulps
	f2 := float64fulltest(hi, lo, rounds)
	
	t.Logf("%v", f1) 
	t.Logf("%b", f1) 
	t.Logf("%v", f2)
	t.Logf("%b", f2) 
	t.Logf("%X", math.Float64bits(f1))
	t.Logf("%X", math.Float64bits(f2))
	t.Logf("ulps:  %v", ulpsBetween(f2, f1))
	t.Logf("exp 2: %v", math.Log2(float64(ulpsBetween(f2, f1))))
	
}

// set const twistedUint64 = true for following 2
func Test_64full_Bisect(t *testing.T) {
	var rounds int = 1e7 * 2
	x1 := NewXoro(1)
	for i := 0; i < rounds; i++ {
		x2 := x1
		f1 := x1.Float64full() 
		f2 := x2.Float64Bisect(false)
		if f1 == f2 {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)

	}
}

func Test_64fullR_Bisect(t *testing.T) {
	var rounds int = 1e7 * 2
	x1 := NewXoro(1)
	for i := 0; i < rounds; i++ {
		x2 := x1
		f1 := x1.Float64fullR() 
		f2 := x2.Float64Bisect(true) 
		if f1 == f2 {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)

	}
}

// ------------------------------------------------------RandomReal
// set const twistedUint64 = true for following 
func TestLdexp(t *testing.T) {
	var rounds int = 1e8
	x := NewXoro(1)
	all, notsame := 0, 0
	for i := 1; i < rounds; i++ {
		f := x.Float64full()
		k := i % 1080 
		f1 := math.Ldexp(f, -k)
		f2 := ldexp(f, uint64(k))
		if f1 == f2  && f1 >= 0x1p-1022 {
			continue
		}
		all++
		if f1 != f2 {
			u := ulpsBetween(f1, f2)
			if notsame++; notsame < 20 {
				t.Logf("F1=  %v", f1)
				t.Logf("F2=  %v", f2)
				t.Logf("Ulps %v", u)
			}
			if u > 1 {
				t.Fatalf("ULPs > 1     %d" , u)
			}
			
		}
	}
	t.Logf("Not same subnormals     %d" , notsame)
	t.Logf("All subnormals          %d" , all)
	t.Logf("Not same subnormal pros %1.8f" , 100*float64(notsame) / float64(all))
}


func Test_RandomReal_64fullR(t *testing.T) {
	var rounds int = 1e8
	x1 := NewXoro(1)
	var even, odd, zeros, same, diff, diff2 int
   
	for i := 0; i < rounds; i++ {
		x2 := x1
		f1 := x1.Float64fullR() 
		f2 := x2.RandomReal() 
		if f1 == f2 && f1 >= 0x1p-1022 {
			continue
		}
		if f1 == f2 {
			same++
		} else {
			diff++
			if diff < 20 {
				t.Logf("F2=  %b", f2)
				t.Logf("F1:  %X" ,  math.Float64bits(f1))
				t.Logf("F2:  %X" ,  math.Float64bits(f2))
			}
			if ulpsBetween(f1, f2) > 1 {
				diff2++
			}
		}
		if math.Float64bits(f2) & 1 == 1 {
			odd++
		} else if f2 != 0 {
			even++
		}
		if f2 == 0 && f1 != 0{
			zeros++
		}
		if f1 < 0x1p-1022 {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)
	}
	t.Logf("Even:  %d" , even)
	t.Logf("Odd:   %d" , odd)
	t.Logf("Zeros: %d" , zeros)
	t.Logf("Same:  %d" , same)
	t.Logf("Diff:  %d" , diff)
	t.Logf("Diff2: %d" , diff2)
}

// -----------------------------------------------------
func Test_RoundingMethods(t *testing.T) {
    var rounds int = 1e7
	x1 := New(1)
	for i := 0; i < rounds; i++ {
        x2 := x1
        u := x1.Uint64() 
		if u == 0 {
            continue
        }
        z := uint64(bits.LeadingZeros64(u))
        
        u = u << z | x1.Uint64() >> (64 - z)
		f1 := float64((u >> 10 + 1) >> 1) / (1<<53) / float64(uint64(1 << z))
        f2 := float64(u | 1) / (1<<64) / float64(uint64(1 << z))
        f3 := math.Float64frombits((((1022 - z) << 53 | u << 1 >> 11) + 1) >> 1)
        f4 := x2.Float64Bisect(true)

		if f1 == f2 && f1 == f3 && f1 == f4 {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f4))
		t.Logf("F1=  %v", f1)
        t.Logf("F2=  %v", f2)
        t.Logf("F3=  %v", f3)
		t.Fatalf("F4=  %v", f4)
	}
}

func TestBitsChanged(t *testing.T) {
    const rounds int = 1e9 
    const shift = 11
	var sum int
	x := NewXoro(1)

	last := x.Uint64() >> shift
	for i := 0; i < rounds; i++ {
		n := x.Uint64() >> shift
		sum += bits.OnesCount64(last ^ n)
		last = n
	}
	ratio := float64(sum) / ((64- shift) * float64(rounds))
	t.Logf("Ratio of changed bits  %1.9f", ratio)
	if abs(ratio-0.5) > 0.00001 {
		t.Errorf("Ratio failed")
	}
}
func Test_1BitRatio(t *testing.T) {
    const rounds int = 1e7*4
    x := NewXoro(1)
	// x := NewXosh(1)
	const left float64 = 0x1p-64
	const right float64 = 0x1p-12
	const errlim = 0.005
    failed := 0
    for bit := 51; bit >= 0; bit-- {
		sum := 0
		n := 0
        for i := 0; i < rounds; i++ {
			// f := x.Float64()
			// f := x.Float64_64()
			f := x.float64div64()
			// f := x.Float64_117()
			// f := x.Float64full()

			if f < left || f >= right {
				continue
			}
			n++
			u := math.Float64bits(f) 
 			u >>= bit 
			sum += int(u & 1)
        }
        ratio := float64(sum) / float64(n)      
        if abs(ratio-0.5) > errlim {
			failed++
			t.Logf("Ratio failed: bit %d ", bit)
            t.Logf("Ratio of 1 bits   %1.9f", ratio)
        }
    }
    if failed > 5 {
        t.Errorf(" ")
    }
}
func Test_1BitRatioSingle(t *testing.T) {
    const rounds int = 1e9
    x := NewXoro(11)
	// x := NewXosh(1)
	sum := uint64(0)
	n := 0
	const bit = 0
	const left float64 = 0x1p-10
	const right float64 = 0x1p-9
  	for i := 0; i < rounds; i++ {
		f := x.float64div63()
		// f := x.Float64_117()
		// f := x.Float64_64()
		// f := x.Float64full()

		if f < left || f >= right {
			continue
		}
		n++
		u := math.Float64bits(f)
		u >>= bit 
		sum += u & 1
	}
	ratio := float64(sum) / float64(n)      
	t.Logf("Ratio of 1 bits  %1.9f", ratio)
	t.Logf("n                %d", n)
}

func Test_OnesCount(t *testing.T) {
    const rounds int = 1e9
    x := NewXoro(1)
	sum := 0
	n := 0
	const left float64 = 0x1p-64
	const right float64 = 0x1p-12

  	for i := 0; i < rounds; i++ {
		// f := x.Float64()
		// f := x.Float64_64()
		// f := x.Float64_64R()
		f := x.float64div63()
		// f := x.Float64_117()
		// f := x.Float64full()
		// f := x.RandomReal()
		
		if f < left || f >= right {
			continue
		}
		n++
		u := math.Float64bits(f) 
		sum += bits.OnesCount64(u << 12)
	}
	ratio := float64(sum) / (52 * float64(n) )    
	t.Logf("All over ratio of 1 bits  %1.9f", ratio)
	t.Logf("Cases                     %d", n)
	t.Logf("Cases pros                %2.4f", 100*float64(n)/ float64(rounds))
}

func Test_Minfloat(t *testing.T) {
    const rounds int = 1e9
    x := NewXoro(1)
	min := 1.0
   
	for i := 0; i < rounds; i++ {
		f := x.Float64full()
		if f < min {
			min = f
			t.Logf("Min         %v", min)
			t.Logf("Log2(Min)   %v", math.Log2(min))
			t.Logf("Log2(round) %v", math.Log2(float64(i)))
		}
		if i % 1e10 == 0 {
			t.Logf("  round/1e9 %d", i / 1e9)
		}
	}
}

// set const twistedUint64 = true 

func Test_Range(t *testing.T) {
	const rounds int = 1e8
	const ulpsLim = 0x1p-1074
	// const exact = false
	const exact = true
	x1 := NewXoro(11)
	max1, max2, minNonzero, minsame, maxdiff := 0.0, 0.0, 1.0, 1.0, 0.0
	samecnt, ulpcnt := 0, 0
	zero, one, maxulps :=  false, false, uint64(0)

	for i := 0; i < rounds; i++ {
		x2 := x1
        // f1 := x1.float64_64Div() 
		// f1 := x1.float64div63()
		// f1 := x1.float64_64Div()
		// f1 := x1.float64div64()
		// f1 := x1.Float64_64() 
		// f1 := x1.Float64_64R()
		f1 := x1.Float64_117() 
		// f1 := x1.Float64_128() 
		// f1 := x1.Float64_117R() 
		// f1 := x1.RandomReal() 
	
		f2 := x2.Float64full() 
		// f2 := x2.Float64fullR() 
		// f2 := x2.Float64Bisect(false) 
        // f2 := x2.Float64Bisect(true) 
        if f1 == 0 {
			zero = true
		}
		if f1 == 1 {
			one = true
        }
        if f1 < minNonzero  && f1 != 0 {
			minNonzero = f1
		}
		diff := abs(f1 - f2)
		if diff > maxdiff {
			maxdiff = diff
		}
		ulps := ulpsBetween(f1, f2)
		if ulps > maxulps && f2 > ulpsLim { 
			maxulps = ulps
		}
		same := f1 == f2
		if !exact {
			same = ulps < 2 //&& f1 <= f2
		}
		if same {
			samecnt++
			if f1 < minsame  && f1 != 0 {
				minsame = f1
			}
			if ulps == 1 { ulpcnt++	} 
 			continue
		}
		if f2 > max2 {
			max2 = f2
		}
		if f1 > max1 {
			max1 = f1
        }
	}
	t.Logf("Range pros       %2.4f (of random bisection)", 100*(1 - max1))
	t.Logf("Max1 not same    %v" , max1)
	t.Logf("                 %X" , math.Float64bits(max1))
	t.Logf("Log2(max1)       %v" , math.Log2(max1))
	t.Logf("Max2 not same    %v" , max2)
	t.Logf("                 %X" ,  math.Float64bits(max2))
	t.Logf("Log2(max2)       %v" , math.Log2(max2))
	t.Logf("1 ulp pros       %2.4f", 100*(float64(ulpcnt)/ float64(samecnt)))
	t.Logf("Log2(maxdiff)    %v" , math.Log2(maxdiff))
	// t.Logf("Ulps(max1, max2) %d" , ulpsBetween(max1, max2))
	t.Logf("Max ulps         %d" , maxulps)
	t.Logf("Min same         %v" , minsame)
    t.Logf("Log2(min same)   %v" , math.Log2(minsame))
    t.Logf("Min non zero     %v" , minNonzero)
    t.Logf("Log2(min non z)  %v" , math.Log2(minNonzero))
	t.Logf("Zero             %v" , zero)
	t.Logf("One              %v" , one)
}

// --------------------------------------- functions for testing-------------------

func float64fulltest(hi, lo uint64, rounds int) float64 {

	pow := 1.0
	for i := 0; i < rounds; i++ { 
		pow *= (1<<64)
	}
	
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	return float64(hi >> 11) / (1<<53) / (pow * float64(uint64(1 << zeros)))
}

func float64_64(u uint64) float64 {
	if u == 0 { return 0 }  
	z := uint64(bits.LeadingZeros64(u)) + 1
	return math.Float64frombits((1023 - z) << 52 | u << z >> 12)
}

func float64_64R(u uint64) float64 {
	if u == 0 { return 0 }
    z := uint64(bits.LeadingZeros64(u)) + 1
    return math.Float64frombits((((1023 - z) << 53 | u << z >> 11) + 1) >> 1)
}

func (x *Xoro) float64_64Div() float64 {
    u := x.Uint64()
    // if u == 0 { return 0 }
	z := uint64(bits.LeadingZeros64(u))
	return float64(u << z >> 11) * twoToMinus(53 + z)
	// return float64(u << z >> 11) / (1 << 53) / float64(uint64(1 << z))
}

func (x *Xoro) float64_64DivR() float64 {
    u := x.Uint64()
    if u == 0 { return 0 }
    z := uint64(bits.LeadingZeros64(u))
    return float64((u << z >> 10 + 1) >> 1) / (1<<53) / float64(uint64(1 << z))
}
var scale = [11]float64 {
	0x1p-53, 0x1p-54, 0x1p-55, 0x1p-56, 0x1p-57, 0x1p-58, 
	0x1p-59, 0x1p-60, 0x1p-61, 0x1p-62, 0x1p-63,  
}
func (x *Xoro) float64_64Tab() float64 {
    u := x.Uint64()
	z := uint64(bits.LeadingZeros64(u))
	if z <= 10 {  
		return float64(u << z >> 11) * scale[z]  
	}
	return float64(u) * 0x1p-64
}

func (x *Xoro) float64_64TabR() float64 {
	
	u := x.Uint64()
	z := uint64(bits.LeadingZeros64(u))
	if z <= 10 { 
		return float64((u << z >> 10 + 1) >> 1) * scale[z]	
	}
	return float64(u) * 0x1p-64
}
func (x *Xoro) float64fullDiv() float64 {

	u := x.Uint64()
	z := uint64(bits.LeadingZeros64(u)) + 1
	if z <= 12 {  
		return math.Float64frombits((1023 - z) << 52 | u << z >> 12)
	}
	z--
	pow := 1.0
	for u == 0 { 
		u = x.Uint64() 
		z = uint64(bits.LeadingZeros64(u))
		pow *= 1<<64
	}
	u = u << z | x.Uint64() >> (64 - z)
	return float64(u >> 11) / (1<<53) / pow / float64(uint64(1 << z))
}

func (x *Xoro) float64fullRDiv() float64 {
	var exp uint64

	u := x.Uint64()
	z := uint64(bits.LeadingZeros64(u)) + 1
    if z <= 11 {  //99.9% of cases 
		return math.Float64frombits((((1023 - z) << 53 | u << z >> 11) + 1) >> 1)
	}
	z--
	for u == 0 { 
		u = x.Uint64() 
		z = uint64(bits.LeadingZeros64(u))
		exp += 64
		if exp == 1024 { return 0 }
	}
	u = u << z | x.Uint64() >> (64 - z)
	return float64((u >> 10 + 1) >> 1) / (1<<53) * twoToMinus(exp + z)
}

// float64div64 produces practically the same distribution as Float64_64R.
func (x *Xoro) float64div64() float64 {

	u := x.Uint64() //&^ 1
	return float64(u) * 0x1p-64 
	// f := float64(u)
	// return (f - f * 0x1p-53)  * 0x1p-64 
	 
}

// float64divE --
func (x *Xoro) float64div63() float64 {
	f := float64(x.Uint64() >> 1) * 0x1p-63
	// f := float64((2 * (x.Uint64() >> 1) +1) >> 1) * 0x1p-63
	// f := float64((x.Uint64()+1) >> 1) * 0x1p-63
	if f == 1 {
		return 1 - 0x1p-53
	}
	return f
}