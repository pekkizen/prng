package prng

import "math/bits"

type jumpPolynoms struct {
	p32  []uint64
	p64  []uint64
	p96  []uint64
	p128 []uint64
	p192 []uint64
}

var jumpdist = jumpPolynoms{
	// xoroshiro128+/**
	p32: []uint64{0xfad843622b252c78, 0xd4e95eef9edbdbc6},
	p64: []uint64{0xdf900294d8f554a5, 0x170865df4b3201fc},
	p96: []uint64{0xd2a98b26625eee7b, 0xdddf9b1090aa7ac1},
	// xoshiro256 all
	p128: []uint64{0x180ec6d33cfd0aba, 0xd5a61266f0c9392c, 0xa9582618e03fc9aa, 0x39abdc4529b1661c},
	p192: []uint64{0x76e15d3efefdcbbf, 0xc5004e441c522fb3, 0x77710069854ee241, 0x39109bb02acbe635},
}

// JumpShort sets x to the same state as 2^32 calls to x.Uint64.
func (x *Xoro) JumpShort() {
	x.jump(jumpdist.p32)
}

// Jump sets x to the same state as 2^64 calls to x.Uint64
// or 2^32 calls to x.JumpShort.
func (x *Xoro) Jump() {
	x.jump(jumpdist.p64)
}

// JumpLong sets x to the same state as 2^96 calls to x.Uint64
// or 2^32 calls to x.Jump.
func (x *Xoro) JumpLong() {
	x.jump(jumpdist.p96)
}

// Jump sets x to the same state as 2^128 calls to x.Uint64
func (x *Xosh) Jump() {
	x.jump(jumpdist.p128)
}

// JumpLong sets x to the same state as 2^192 calls to x.Uint64
// or 2^64 calls to x.Jump.
func (x *Xosh) JumpLong() {
	x.jump(jumpdist.p192)
}

func (x *Xoro) jump(dist []uint64) {
	var s0, s1 uint64
	x0, x1 := x.s0, x.s1

	for i := 0; i < 2; i++ {
		xorbits := dist[i]
		for b := 0; b < 64; b++ {

			if (xorbits & 1) != 0 {
				s0 ^= x0
				s1 ^= x1
			}
			xorbits >>= 1
			x1 ^= x0 //one step of linear engine forward
			x0 = bits.RotateLeft64(x0, 24) ^ x1 ^ (x1 << 16)
			x1 = bits.RotateLeft64(x1, 37)
		}
	}
	x.s0, x.s1 = s0, s1
}

func (x *Xosh) jump(dist []uint64) {
	var s0, s1, s2, s3 uint64
	x0, x1, x2, x3 := x.s0, x.s1, x.s2, x.s3

	for i := 0; i < 4; i++ {
		xorbits := dist[i]
		for b := uint(0); b < 64; b++ {

			if xorbits & (1 << b) != 0 {
				s0 ^= x0
				s1 ^= x1
				s2 ^= x2
				s3 ^= x3
			}
			// one step of linear engine forward
			t := x1 << 17
			x2 ^= x0
			x3 ^= x1
			x1 ^= x2
			x0 ^= x3
			x2 ^= t
			x3 = bits.RotateLeft64(x3, 45)
		}
	}
	x.s0, x.s1, x.s2, x.s3 = s0, s1, s2, s3
}
