package chessengine

type ranctx struct {
	a uint64
	b uint64
	c uint64
	d uint64
}

// Rotate left function
func rot(x uint64, k uint) uint64 {
	return (x << k) | (x >> (64 - k))
}

// PRNG function
func ranval(x *ranctx) uint64 {
	e := x.a - rot(x.b, 7)
	x.a = x.b ^ rot(x.c, 13)
	x.b = x.c + rot(x.d, 37)
	x.c = x.d + e
	x.d = e + x.a
	return x.d
}

// Initialization function
func raninit(x *ranctx, seed uint64) {
	x.a = 0xf1ea5eed
	x.b = seed
	x.c = seed
	x.d = seed
	for i := 0; i < 20; i++ {
		ranval(x)
	}
}
