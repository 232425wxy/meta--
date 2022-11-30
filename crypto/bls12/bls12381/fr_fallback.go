//go:build !amd64 || generic
// +build !amd64 generic

package bls12381

import "math/bits"

func addFR(z, x, y *Fr) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], y[0], 0)
	z[1], carry = bits.Add64(x[1], y[1], carry)
	z[2], carry = bits.Add64(x[2], y[2], carry)
	z[3], _ = bits.Add64(x[3], y[3], carry)

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[3] < 8353516859464449352 || (z[3] == 8353516859464449352 && (z[2] < 3691218898639771653 || (z[2] == 3691218898639771653 && (z[1] < 6034159408538082302 || (z[1] == 6034159408538082302 && (z[0] < 18446744069414584321))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744069414584321, 0)
		z[1], b = bits.Sub64(z[1], 6034159408538082302, b)
		z[2], b = bits.Sub64(z[2], 3691218898639771653, b)
		z[3], _ = bits.Sub64(z[3], 8353516859464449352, b)
	}
}

func laddAssignFR(z, y *Fr) {
	var carry uint64

	z[0], carry = bits.Add64(z[0], y[0], 0)
	z[1], carry = bits.Add64(z[1], y[1], carry)
	z[2], carry = bits.Add64(z[2], y[2], carry)
	z[3], _ = bits.Add64(z[3], y[3], carry)
}

func doubleFR(z, x *Fr) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], x[0], 0)
	z[1], carry = bits.Add64(x[1], x[1], carry)
	z[2], carry = bits.Add64(x[2], x[2], carry)
	z[3], _ = bits.Add64(x[3], x[3], carry)

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[3] < 8353516859464449352 || (z[3] == 8353516859464449352 && (z[2] < 3691218898639771653 || (z[2] == 3691218898639771653 && (z[1] < 6034159408538082302 || (z[1] == 6034159408538082302 && (z[0] < 18446744069414584321))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744069414584321, 0)
		z[1], b = bits.Sub64(z[1], 6034159408538082302, b)
		z[2], b = bits.Sub64(z[2], 3691218898639771653, b)
		z[3], _ = bits.Sub64(z[3], 8353516859464449352, b)
	}
}

func subFR(z, x, y *Fr) {
	var b uint64
	z[0], b = bits.Sub64(x[0], y[0], 0)
	z[1], b = bits.Sub64(x[1], y[1], b)
	z[2], b = bits.Sub64(x[2], y[2], b)
	z[3], b = bits.Sub64(x[3], y[3], b)
	if b != 0 {
		var c uint64
		z[0], c = bits.Add64(z[0], 18446744069414584321, 0)
		z[1], c = bits.Add64(z[1], 6034159408538082302, c)
		z[2], c = bits.Add64(z[2], 3691218898639771653, c)
		z[3], _ = bits.Add64(z[3], 8353516859464449352, c)
	}
}

func lsubAssignFR(z, y *Fr) {
	var b uint64
	z[0], b = bits.Sub64(z[0], y[0], 0)
	z[1], b = bits.Sub64(z[1], y[1], b)
	z[2], b = bits.Sub64(z[2], y[2], b)
	z[3], b = bits.Sub64(z[3], y[3], b)
}

func negFR(z, x *Fr) {
	if x.IsZero() {
		z.Zero()
		return
	}
	var borrow uint64
	z[0], borrow = bits.Sub64(18446744069414584321, x[0], 0)
	z[1], borrow = bits.Sub64(6034159408538082302, x[1], borrow)
	z[2], borrow = bits.Sub64(3691218898639771653, x[2], borrow)
	z[3], _ = bits.Sub64(8353516859464449352, x[3], borrow)
}

func mulFR(z, x, y *Fr) {

	var t [4]uint64
	var c [3]uint64
	{
		// round 0
		v := x[0]
		c[1], c[0] = bits.Mul64(v, y[0])
		m := c[0] * 18446744069414584319
		c[2] = madd0(m, 18446744069414584321, c[0])
		c[1], c[0] = madd1(v, y[1], c[1])
		c[2], t[0] = madd2(m, 6034159408538082302, c[2], c[0])
		c[1], c[0] = madd1(v, y[2], c[1])
		c[2], t[1] = madd2(m, 3691218898639771653, c[2], c[0])
		c[1], c[0] = madd1(v, y[3], c[1])
		t[3], t[2] = madd3(m, 8353516859464449352, c[0], c[2], c[1])
	}
	{
		// round 1
		v := x[1]
		c[1], c[0] = madd1(v, y[0], t[0])
		m := c[0] * 18446744069414584319
		c[2] = madd0(m, 18446744069414584321, c[0])
		c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[2], t[0] = madd2(m, 6034159408538082302, c[2], c[0])
		c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[2], t[1] = madd2(m, 3691218898639771653, c[2], c[0])
		c[1], c[0] = madd2(v, y[3], c[1], t[3])
		t[3], t[2] = madd3(m, 8353516859464449352, c[0], c[2], c[1])
	}
	{
		// round 2
		v := x[2]
		c[1], c[0] = madd1(v, y[0], t[0])
		m := c[0] * 18446744069414584319
		c[2] = madd0(m, 18446744069414584321, c[0])
		c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[2], t[0] = madd2(m, 6034159408538082302, c[2], c[0])
		c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[2], t[1] = madd2(m, 3691218898639771653, c[2], c[0])
		c[1], c[0] = madd2(v, y[3], c[1], t[3])
		t[3], t[2] = madd3(m, 8353516859464449352, c[0], c[2], c[1])
	}
	{
		// round 3
		v := x[3]
		c[1], c[0] = madd1(v, y[0], t[0])
		m := c[0] * 18446744069414584319
		c[2] = madd0(m, 18446744069414584321, c[0])
		c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[2], z[0] = madd2(m, 6034159408538082302, c[2], c[0])
		c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[2], z[1] = madd2(m, 3691218898639771653, c[2], c[0])
		c[1], c[0] = madd2(v, y[3], c[1], t[3])
		z[3], z[2] = madd3(m, 8353516859464449352, c[0], c[2], c[1])
	}

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[3] < 8353516859464449352 || (z[3] == 8353516859464449352 && (z[2] < 3691218898639771653 || (z[2] == 3691218898639771653 && (z[1] < 6034159408538082302 || (z[1] == 6034159408538082302 && (z[0] < 18446744069414584321))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744069414584321, 0)
		z[1], b = bits.Sub64(z[1], 6034159408538082302, b)
		z[2], b = bits.Sub64(z[2], 3691218898639771653, b)
		z[3], _ = bits.Sub64(z[3], 8353516859464449352, b)
	}
}

func squareFR(z, x *Fr) {

	var t [4]uint64
	var c [3]uint64
	{
		// round 0
		v := x[0]
		c[1], c[0] = bits.Mul64(v, x[0])
		m := c[0] * 18446744069414584319
		c[2] = madd0(m, 18446744069414584321, c[0])
		c[1], c[0] = madd1(v, x[1], c[1])
		c[2], t[0] = madd2(m, 6034159408538082302, c[2], c[0])
		c[1], c[0] = madd1(v, x[2], c[1])
		c[2], t[1] = madd2(m, 3691218898639771653, c[2], c[0])
		c[1], c[0] = madd1(v, x[3], c[1])
		t[3], t[2] = madd3(m, 8353516859464449352, c[0], c[2], c[1])
	}
	{
		// round 1
		v := x[1]
		c[1], c[0] = madd1(v, x[0], t[0])
		m := c[0] * 18446744069414584319
		c[2] = madd0(m, 18446744069414584321, c[0])
		c[1], c[0] = madd2(v, x[1], c[1], t[1])
		c[2], t[0] = madd2(m, 6034159408538082302, c[2], c[0])
		c[1], c[0] = madd2(v, x[2], c[1], t[2])
		c[2], t[1] = madd2(m, 3691218898639771653, c[2], c[0])
		c[1], c[0] = madd2(v, x[3], c[1], t[3])
		t[3], t[2] = madd3(m, 8353516859464449352, c[0], c[2], c[1])
	}
	{
		// round 2
		v := x[2]
		c[1], c[0] = madd1(v, x[0], t[0])
		m := c[0] * 18446744069414584319
		c[2] = madd0(m, 18446744069414584321, c[0])
		c[1], c[0] = madd2(v, x[1], c[1], t[1])
		c[2], t[0] = madd2(m, 6034159408538082302, c[2], c[0])
		c[1], c[0] = madd2(v, x[2], c[1], t[2])
		c[2], t[1] = madd2(m, 3691218898639771653, c[2], c[0])
		c[1], c[0] = madd2(v, x[3], c[1], t[3])
		t[3], t[2] = madd3(m, 8353516859464449352, c[0], c[2], c[1])
	}
	{
		// round 3
		v := x[3]
		c[1], c[0] = madd1(v, x[0], t[0])
		m := c[0] * 18446744069414584319
		c[2] = madd0(m, 18446744069414584321, c[0])
		c[1], c[0] = madd2(v, x[1], c[1], t[1])
		c[2], z[0] = madd2(m, 6034159408538082302, c[2], c[0])
		c[1], c[0] = madd2(v, x[2], c[1], t[2])
		c[2], z[1] = madd2(m, 3691218898639771653, c[2], c[0])
		c[1], c[0] = madd2(v, x[3], c[1], t[3])
		z[3], z[2] = madd3(m, 8353516859464449352, c[0], c[2], c[1])
	}

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[3] < 8353516859464449352 || (z[3] == 8353516859464449352 && (z[2] < 3691218898639771653 || (z[2] == 3691218898639771653 && (z[1] < 6034159408538082302 || (z[1] == 6034159408538082302 && (z[0] < 18446744069414584321))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 18446744069414584321, 0)
		z[1], b = bits.Sub64(z[1], 6034159408538082302, b)
		z[2], b = bits.Sub64(z[2], 3691218898639771653, b)
		z[3], _ = bits.Sub64(z[3], 8353516859464449352, b)
	}
}

func waddFR(z, y *wideFr) {
	var carry uint64
	z[0], carry = bits.Add64(z[0], y[0], 0)
	z[1], carry = bits.Add64(z[1], y[1], carry)
	z[2], carry = bits.Add64(z[2], y[2], carry)
	z[3], carry = bits.Add64(z[3], y[3], carry)
	z[4], carry = bits.Add64(z[4], y[4], carry)
	z[5], carry = bits.Add64(z[5], y[5], carry)
	z[6], carry = bits.Add64(z[6], y[6], carry)
	z[7], _ = bits.Add64(z[7], y[7], carry)
}

// We applied custom multiplication since goff does generate multiplication code nested with reduction
func wmulFR(w *wideFr, a, b *Fr) {
	// Handbook of Applied Cryptography
	// Hankerson, Menezes, Vanstone
	// 14.12 Algorithm Multiple-precision multiplication

	var w0, w1, w2, w3, w4, w5, w6, w7 uint64
	var a0 = a[0]
	var a1 = a[1]
	var a2 = a[2]
	var a3 = a[3]
	var b0 = b[0]
	var b1 = b[1]
	var b2 = b[2]
	var b3 = b[3]
	var u, v, c, t uint64

	// i = 0, j = 0
	c, w0 = bits.Mul64(a0, b0)

	// i = 0, j = 1
	u, v = bits.Mul64(a1, b0)
	w1 = v + c
	c = u + (v&c|(v|c)&^w1)>>63

	// i = 0, j = 2
	u, v = bits.Mul64(a2, b0)
	w2 = v + c
	c = u + (v&c|(v|c)&^w2)>>63

	// i = 0, j = 3
	u, v = bits.Mul64(a3, b0)
	w3 = v + c
	w4 = u + (v&c|(v|c)&^w3)>>63

	// i = 1, j = 0
	c, v = bits.Mul64(a0, b1)
	t = v + w1
	c += (v&w1 | (v|w1)&^t) >> 63
	w1 = t

	// i = 1, j = 1
	u, v = bits.Mul64(a1, b1)
	t = v + w2
	u += (v&w2 | (v|w2)&^t) >> 63
	w2 = t + c
	c = u + (t&c|(t|c)&^w2)>>63

	// i = 1, j = 2
	u, v = bits.Mul64(a2, b1)
	t = v + w3
	u += (v&w3 | (v|w3)&^t) >> 63
	w3 = t + c
	c = u + (t&c|(t|c)&^w3)>>63

	// i = 1, j = 3
	u, v = bits.Mul64(a3, b1)
	t = v + w4
	u += (v&w4 | (v|w4)&^t) >> 63
	w4 = t + c
	w5 = u + (t&c|(t|c)&^w4)>>63

	// i = 2, j = 0
	c, v = bits.Mul64(a0, b2)
	t = v + w2
	c += (v&w2 | (v|w2)&^t) >> 63
	w2 = t

	// i = 2, j = 1
	u, v = bits.Mul64(a1, b2)
	t = v + w3
	u += (v&w3 | (v|w3)&^t) >> 63
	w3 = t + c
	c = u + (t&c|(t|c)&^w3)>>63

	// i = 2, j = 2
	u, v = bits.Mul64(a2, b2)
	t = v + w4
	u += (v&w4 | (v|w4)&^t) >> 63
	w4 = t + c
	c = u + (t&c|(t|c)&^w4)>>63

	// i = 2, j = 3
	u, v = bits.Mul64(a3, b2)
	t = v + w5
	u += (v&w5 | (v|w5)&^t) >> 63
	w5 = t + c
	w6 = u + (t&c|(t|c)&^w5)>>63

	// i = 3, j = 0
	c, v = bits.Mul64(a0, b3)
	t = v + w3
	c += (v&w3 | (v|w3)&^t) >> 63
	w3 = t

	// i = 3, j = 1
	u, v = bits.Mul64(a1, b3)
	t = v + w4
	u += (v&w4 | (v|w4)&^t) >> 63
	w4 = t + c
	c = u + (t&c|(t|c)&^w4)>>63

	// i = 3, j = 2
	u, v = bits.Mul64(a2, b3)
	t = v + w5
	u += (v&w5 | (v|w5)&^t) >> 63
	w5 = t + c
	c = u + (t&c|(t|c)&^w5)>>63

	// i = 3, j = 3
	u, v = bits.Mul64(a3, b3)
	t = v + w6
	u += (v&w6 | (v|w6)&^t) >> 63
	w6 = t + c
	w7 = u + (t&c|(t|c)&^w6)>>63

	w[0] = w0
	w[1] = w1
	w[2] = w2
	w[3] = w3
	w[4] = w4
	w[5] = w5
	w[6] = w6
	w[7] = w7
}