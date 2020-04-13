package prng_test

import (
	"math"
	"math/bits"
	"testing"
	"io/ioutil"

	. "prng"
	// . "github.com/pekkizen/prng"
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
	k := math.Float64bits(y) //&^ (1<<63)
	n := math.Float64bits(x) //&^ (1<<63)
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
		f1 := float64(k) / (1<<53)
		f2 := float64(k+1) / (1<<53)

		if adjacent(f1, f2) != (ulpsBetween(f1, f2) == 1) {
			t.Fatalf("Failed: f2-f1=%v ulps=%d", f2-f1, ulpsBetween(f1, f2))
		}
	}
}

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
	const rounds = (1<<32)
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
	const rounds = (1<<32)
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
	const rounds = (1<<32)
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
	x := NextXoro()
	// r := NextXosh()
	low := 0
	high := 0
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
func TestFloat64FourSlots(t *testing.T) {

	const rounds int = 1e8
	var slotsize = 1e-5
	const failLim = 6e-2
	var tab [4]int
	x := NewXoro(1)
	// z := NewXosh(18)
	for i := 0; i < rounds; i++ {
		f := x.RandomReal()
		// f := Float64_64R(x.Uint64())
		// f := x.Float64_1024()
		// f := x.Float64Bisect(false)
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

	const rounds int = 1e7 * 2
	var slotsize = 1e-7
	hit := 0
	x := NewXoro(1)
	for i := 0; i < rounds; i++ {
		// f := x.Float64()
		f := x.Float64Bisect(false)
		// f := x.Float64_1024()
		// f := x.RandomReal()
		
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
	const rounds = (1<<28)
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
			t.Fatalf("Equidistance failed: f1= %v f2%v", f1, f2)
		}
		if k >= minAdjacent && !adjacent(f1, f2) {
			t.Fatalf("adjacent failed: ulps=%d f1=%v", ulpsBetween(f1, f2), f1)
		}

	}
}
func TestUint64ToFloatDistribution2(t *testing.T) {
	// same as above but sequential

	const equidist = 1.0 / (1<<53)
	const div = (1<<53)

	var rounds uint64 = (1<<30)
	var start uint64 = (1<<52) - 1000

	stop := start + rounds
	if stop > (1<<53)-1 {
		stop = (1<<53) - 1
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
func TestUint64ToFloat52bit(t *testing.T) {
	// 52-bit division method and the 52-bit explicit method are same
	x := NewXoro(1)
	const rounds int = 1e9
	for i := 0; i < rounds; i++ {

		k := x.Uint64() >> 12
		f1 := float64(k) / (1<<52)
		f2 := math.Float64frombits(1023<<52|k) - 1
		if f1-f2 != 0 {
			t.Fatalf("Methods not same: diff = %v", f1-f2)
		}
	}
}
func Test_64_64Distribution(t *testing.T) {

	var rounds int = 1e8*4
	x := NewXoro(2)
	
	for i := 0; i < rounds; i++ {

		u := x.Uint64() 
		u >>= u % 64 // 0 - 63 leading zeros
		f1 := Float64_64(u) 
		zeros := uint64(bits.LeadingZeros64(u))
		if zeros > 11 {
			zeros = 11
		}
		u2 := u << zeros
		u2 += (1 << 11) // next adjacent
		u2 >>= zeros

		f2 := Float64_64(u2) 
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
func Test_1024Singles(t *testing.T) {
	var hi, lo uint64
	rounds := 15
	const ulps = 1
	hi = 1
	f1 := Float64_1024test(hi, lo, rounds)
	lo += (1<<14) * ulps
	f2 := Float64_1024test(hi, lo, rounds)
	
	t.Logf("%v", f1) 
	t.Logf("%b", f1) 
	t.Logf("%v", f2)
	t.Logf("%b", f2) 
	t.Logf("%X", math.Float64bits(f1))
	t.Logf("%X", math.Float64bits(f2))
	t.Logf("ulps %v", ulpsBetween(f2, f1))
	t.Logf("ulpsX %v", ulpsBetween(0.5, 0.5/2))
}
func Test_BisectSingles(t *testing.T) {

	key := make([]uint64, 17)
	key[16] = (1 << 14)  			// Min. subnormal positive float64 4.940656 × 10−324
	// key[0] = (1 << 64) - 1  		// 0.9999999999999999
	// key[0] = (1 << 63) 			// 0.5
	f1 := Float64Bisect(key)
	t.Logf("%v", f1) 
	t.Logf("%b", f1) 
	t.Logf("%16X", math.Float64bits(f1))
	
}
func Test_1024_Bisect(t *testing.T) {
	var rounds int = 1e6 * 2
	key := make([]uint64, 17) 
	x := NewXoro(1)
	for i := 0; i < rounds; i++ {
		index := i % 15
		for k := 0; k < index; k++ {
			key[k] = 0
		}
		u := x.Uint64() 
		u >>= u % 54
		key[index] = u
		if u == 0 {
			u = 1
		}
		key[index] = u
		key[index + 1] = x.Uint64() >> (u % 54)
		f1 := Float64_1024(key) 
		f2 := Float64Bisect(key) 
		if f1 == f2 {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)

	}
}
func Test_1024_Bisect2(t *testing.T) {
	var rounds int = 1e7
	x1 := NewXoro(1)
	x2 := NewXoro(1)
	for i := 0; i < rounds; i++ {
		f1 := x1.Float64_1024() 
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
func Test_64Range(t *testing.T) {
	var rounds int = 1e8
	x1 := NewXoro(1)
	x2 := NewXoro(1)
	maxdifferent := 0.0
	minequal := 1.0
	for i := 0; i < rounds; i++ {
		f1 := x1.Float64_64() 
		// f1 := x1.Float64_64R() 
		// f1 := x1.Float64_1024() 
		// f1 := x1.Float64_1024R() 
		f2 := x2.Float64Bisect(false) 
		// f2 := x2.Float64Bisect(true) 
		x1.Seed(uint64(i))
		x2.Seed(uint64(i))
		if f1 == f2 {
			if f1 < minequal {
				minequal = f1
			}
			continue
		}
		if f2 > maxdifferent {
			maxdifferent = f2
		}

	}
	t.Logf("Min equal:     %2.8f" , 100*(1 - minequal))
	t.Logf("Max different: %2.8f" , 100*(1 - maxdifferent))
}

func Test_1024R_Bisect(t *testing.T) {
	var rounds int = 1e7
	x1 := NewXoro(1)
	x2 := NewXoro(1)
	for i := 0; i < rounds; i++ {
		f1 := x1.Float64_1024R() 
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

func Test_1024R_RandomReal(t *testing.T) {
	var rounds int = 1e8
	listed := 0

	x1 := NewXoro(1)
	x2 := NewXoro(1)
	for i := 0; i < rounds; i++ {
		f1 := x1.Float64_1024R() 
		f2 := x2.RandomReal() 
		x1.Seed(uint64(i))
		x2.Seed(uint64(i))
		if f1 == f2 {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Logf("F2=  %v", f2)
		listed++
		if listed > 20 {
			t.Fatalf(" ")
		}

	}
}

func Test_RandomReal_Bisect(t *testing.T) {
	var rounds int = 1e6 * 2
	key := make([]uint64, 17) 
	x := NewXoro(1)
	
	for i := 0; i < rounds; i++ {
		index := i % 15
		for k := 0; k < index; k++ {
			key[k] = 0
		}
		u := x.Uint64() 
		u >>= u % 54
		if u == 0 {
			u = 1
		}
		key[index] = u
		key[index + 1] = x.Uint64() >> index
		f1 := RandomReal(key) 
		f2 := Float64BisectR(key) 
		if f1 == f2 {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)

	}
}
func TestKey_1024R_Bisect(t *testing.T) {
	var rounds int = 1e6 * 2
	key := make([]uint64, 17) 
	x := NewXoro(1)
	
	for i := 0; i < rounds; i++ {
		index := i % 15
		for k := 0; k < index; k++ {
			key[k] = 0
		}
		u := x.Uint64() 
		u >>= u % 54
		if u == 0 {
			u = 1
		}
		key[index] = u
		key[index + 1] = x.Uint64() >> index
		f1 := Float64_1024R(key) 
		f2 := Float64BisectR(key) 
		if f1 == f2 {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)

	}
}

func Test_RoundingMethods(t *testing.T) {
	var rounds int = 1e8 * 5
	x := NewXoro(1)
	
	for i := 0; i < rounds; i++ {
		
		u := x.Uint64() 
		u >>= u % 64
		u <<= i % 64
		if u == 0 {
			u = 1 << 64 -1
			if i % 2 == 0 {
				u = 1
			}
		}
	
		zeros := uint64(bits.LeadingZeros64(u))
		f1 := float64(((u << zeros) >> 10 + 1) >> 1) / (1<<53)
		f2 := float64((u << zeros) | 1) / (1<<64) 
		if f1 == f2 {
			continue
		}
		t.Logf("Not same: i=%d" , i)
		t.Logf("Ulps %v", ulpsBetween(f1, f2))
		t.Logf("F1=  %v", f1)
		t.Fatalf("F2=  %v", f2)

	}
}
func Test_BitsChangedFloat64_1024(t *testing.T) {
	const rounds int = 1e9
	var sum int
	x := NewXoro(1)
	last := math.Float64bits(x.Float64_1024()) & ((1<<52) - 1)
	for i := 0; i < rounds; i++ {
		n := math.Float64bits(x.Float64_1024()) & ((1<<52) - 1)
		sum += bits.OnesCount64(last ^ n)
		last = n
	}
	ratio := float64(sum) / (52 * float64(rounds))
	t.Logf("Ratio of changed bits  %1.9f", ratio)
	if abs(ratio-0.5) > 0.00001 {
		t.Errorf("Ratio failed")
	}
}
// --------------------------------------- functions for testing-------------------
// Float64_1024test --
func Float64_1024test(hi, lo uint64, rounds int) float64 {

	pow := 1.0
	for i := 0; i < rounds; i++ { 
		pow *= (1<<64)
	}
	
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	return float64(hi >> 11) / (1<<53) / (pow * float64(uint64(1 << zeros)))
}
// Float64_1024 --
func Float64_1024(bitsequence []uint64) float64 {

	hi := bitsequence[0]
	if hi >= 1<<52 {  //99.95% of cases 
		return Float64_64(hi)
	} 
	pow := 1.0
	i := 1
	for hi == 0 { 
		hi = bitsequence[i]
		i++
		pow *= (1<<64)
	}
	lo := bitsequence[i]
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	return float64(hi >> 11) / (1<<53) / pow / float64(uint64(1 << zeros))
}
func Float64_1024R(bitsequence []uint64) float64 {

	hi := bitsequence[0]
	if hi >= 1<<53 {  //99.95% of cases 
		return Float64_64R(hi)
	} 
	pow := 1.0
	i := 1
	for hi == 0 { 
		hi = bitsequence[i]
		i++
		pow *= (1<<64)
	}
	lo := bitsequence[i]
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	hi = (hi >> 10 + 1) >> 1  
	return float64(hi) / (1<<53) / pow / float64(uint64(1 << zeros))
}

// RandomReal http://prng.di.unimi.it/random_real.c
func RandomReal(bitsequence []uint64) float64 {

	hi := bitsequence[0]
	if hi >= (1<<63) {
		return float64(hi | 1) / (1<<64) 
	}
	pow := 1.0
	i := 1
	for hi == 0 { 
		hi = bitsequence[i]
		i++
		pow *= (1<<64)
	}
	lo := bitsequence[i]
	zeros := uint64(bits.LeadingZeros64(hi))
	hi = (hi << zeros) | (lo >> (64 - zeros))
	return float64(hi | 1) / (1<<64) / pow / float64(uint64(1 << zeros))
	// return float64(hi) / (1<<64) / pow / float64(uint64(1 << zeros))
}

// Float64Bisect --
func Float64Bisect(bitsequence []uint64) float64 {
	
	left, mean, right := 0.0, 0.5, 1.0
	for k :=0; k < len(bitsequence); k++ {
		u := bitsequence[k]
		for b := uint64(0); b < 64; b++ {
	
			if u & (1<<63) != 0 {
				left = mean						// '1' bit -> take the right half, big numbers			
			} else {
				right = mean					// '0' bit -> take the left half, small numbers		
			}
			u <<= 1
			mean = (left + right) / 2
			if mean == left || mean == right {	// right - left = 1 ULP
				return left
			}
		}
	}
	return -1
}
// Float64BisectR  --
func Float64BisectR(bitsequence []uint64) float64 {
	
	left, mean, right := 0.0, 0.5, 1.0
	for k :=0; k < len(bitsequence); k++ {
		u := bitsequence[k]
		for b := uint64(0); b < 64; b++ {

			if u & (1<<(63 - b)) != 0 {
				left = mean								
			} else {
				right = mean					
			}
			mean = (left + right) / 2
			if mean == left || mean == right {	
				b++
				if b > 63 {
					u = bitsequence[k+1]
					b = 0
				}
				if u & (1<<(63 - b)) != 0 {
					return right								
				} 
				return left
			}
		}
	}
	return -1
}