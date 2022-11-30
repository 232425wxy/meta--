//go:build !amd64 || generic
// +build !amd64 generic

package bls12381

import "math/bits"

func add(z, x, y *fe) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], y[0], 0)
	z[1], carry = bits.Add64(x[1], y[1], carry)
	z[2], carry = bits.Add64(x[2], y[2], carry)
	z[3], carry = bits.Add64(x[3], y[3], carry)
	z[4], carry = bits.Add64(x[4], y[4], carry)
	z[5], _ = bits.Add64(x[5], y[5], carry)

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[5] < 1873798617647539866 || (z[5] == 1873798617647539866 && (z[4] < 5412103778470702295 || (z[4] == 5412103778470702295 && (z[3] < 7239337960414712511 || (z[3] == 7239337960414712511 && (z[2] < 7435674573564081700 || (z[2] == 7435674573564081700 && (z[1] < 2210141511517208575 || (z[1] == 2210141511517208575 && (z[0] < 13402431016077863595))))))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 13402431016077863595, 0)
		z[1], b = bits.Sub64(z[1], 2210141511517208575, b)
		z[2], b = bits.Sub64(z[2], 7435674573564081700, b)
		z[3], b = bits.Sub64(z[3], 7239337960414712511, b)
		z[4], b = bits.Sub64(z[4], 5412103778470702295, b)
		z[5], _ = bits.Sub64(z[5], 1873798617647539866, b)
	}
}

func addAssign(z, y *fe) {
	var carry uint64

	z[0], carry = bits.Add64(z[0], y[0], 0)
	z[1], carry = bits.Add64(z[1], y[1], carry)
	z[2], carry = bits.Add64(z[2], y[2], carry)
	z[3], carry = bits.Add64(z[3], y[3], carry)
	z[4], carry = bits.Add64(z[4], y[4], carry)
	z[5], _ = bits.Add64(z[5], y[5], carry)

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[5] < 1873798617647539866 || (z[5] == 1873798617647539866 && (z[4] < 5412103778470702295 || (z[4] == 5412103778470702295 && (z[3] < 7239337960414712511 || (z[3] == 7239337960414712511 && (z[2] < 7435674573564081700 || (z[2] == 7435674573564081700 && (z[1] < 2210141511517208575 || (z[1] == 2210141511517208575 && (z[0] < 13402431016077863595))))))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 13402431016077863595, 0)
		z[1], b = bits.Sub64(z[1], 2210141511517208575, b)
		z[2], b = bits.Sub64(z[2], 7435674573564081700, b)
		z[3], b = bits.Sub64(z[3], 7239337960414712511, b)
		z[4], b = bits.Sub64(z[4], 5412103778470702295, b)
		z[5], _ = bits.Sub64(z[5], 1873798617647539866, b)
	}
}

func ladd(z, x, y *fe) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], y[0], 0)
	z[1], carry = bits.Add64(x[1], y[1], carry)
	z[2], carry = bits.Add64(x[2], y[2], carry)
	z[3], carry = bits.Add64(x[3], y[3], carry)
	z[4], carry = bits.Add64(x[4], y[4], carry)
	z[5], _ = bits.Add64(x[5], y[5], carry)
}

func laddAssign(z, y *fe) {
	var carry uint64

	z[0], carry = bits.Add64(z[0], y[0], 0)
	z[1], carry = bits.Add64(z[1], y[1], carry)
	z[2], carry = bits.Add64(z[2], y[2], carry)
	z[3], carry = bits.Add64(z[3], y[3], carry)
	z[4], carry = bits.Add64(z[4], y[4], carry)
	z[5], _ = bits.Add64(z[5], y[5], carry)
}

func double(z, x *fe) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], x[0], 0)
	z[1], carry = bits.Add64(x[1], x[1], carry)
	z[2], carry = bits.Add64(x[2], x[2], carry)
	z[3], carry = bits.Add64(x[3], x[3], carry)
	z[4], carry = bits.Add64(x[4], x[4], carry)
	z[5], _ = bits.Add64(x[5], x[5], carry)

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[5] < 1873798617647539866 || (z[5] == 1873798617647539866 && (z[4] < 5412103778470702295 || (z[4] == 5412103778470702295 && (z[3] < 7239337960414712511 || (z[3] == 7239337960414712511 && (z[2] < 7435674573564081700 || (z[2] == 7435674573564081700 && (z[1] < 2210141511517208575 || (z[1] == 2210141511517208575 && (z[0] < 13402431016077863595))))))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 13402431016077863595, 0)
		z[1], b = bits.Sub64(z[1], 2210141511517208575, b)
		z[2], b = bits.Sub64(z[2], 7435674573564081700, b)
		z[3], b = bits.Sub64(z[3], 7239337960414712511, b)
		z[4], b = bits.Sub64(z[4], 5412103778470702295, b)
		z[5], _ = bits.Sub64(z[5], 1873798617647539866, b)
	}
}

func doubleAssign(z *fe) {
	var carry uint64

	z[0], carry = bits.Add64(z[0], z[0], 0)
	z[1], carry = bits.Add64(z[1], z[1], carry)
	z[2], carry = bits.Add64(z[2], z[2], carry)
	z[3], carry = bits.Add64(z[3], z[3], carry)
	z[4], carry = bits.Add64(z[4], z[4], carry)
	z[5], _ = bits.Add64(z[5], z[5], carry)

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[5] < 1873798617647539866 || (z[5] == 1873798617647539866 && (z[4] < 5412103778470702295 || (z[4] == 5412103778470702295 && (z[3] < 7239337960414712511 || (z[3] == 7239337960414712511 && (z[2] < 7435674573564081700 || (z[2] == 7435674573564081700 && (z[1] < 2210141511517208575 || (z[1] == 2210141511517208575 && (z[0] < 13402431016077863595))))))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 13402431016077863595, 0)
		z[1], b = bits.Sub64(z[1], 2210141511517208575, b)
		z[2], b = bits.Sub64(z[2], 7435674573564081700, b)
		z[3], b = bits.Sub64(z[3], 7239337960414712511, b)
		z[4], b = bits.Sub64(z[4], 5412103778470702295, b)
		z[5], _ = bits.Sub64(z[5], 1873798617647539866, b)
	}
}

func ldouble(z, x *fe) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], x[0], 0)
	z[1], carry = bits.Add64(x[1], x[1], carry)
	z[2], carry = bits.Add64(x[2], x[2], carry)
	z[3], carry = bits.Add64(x[3], x[3], carry)
	z[4], carry = bits.Add64(x[4], x[4], carry)
	z[5], _ = bits.Add64(x[5], x[5], carry)
}

func sub(z, x, y *fe) {
	var b uint64
	z[0], b = bits.Sub64(x[0], y[0], 0)
	z[1], b = bits.Sub64(x[1], y[1], b)
	z[2], b = bits.Sub64(x[2], y[2], b)
	z[3], b = bits.Sub64(x[3], y[3], b)
	z[4], b = bits.Sub64(x[4], y[4], b)
	z[5], b = bits.Sub64(x[5], y[5], b)
	if b != 0 {
		var c uint64
		z[0], c = bits.Add64(z[0], 13402431016077863595, 0)
		z[1], c = bits.Add64(z[1], 2210141511517208575, c)
		z[2], c = bits.Add64(z[2], 7435674573564081700, c)
		z[3], c = bits.Add64(z[3], 7239337960414712511, c)
		z[4], c = bits.Add64(z[4], 5412103778470702295, c)
		z[5], _ = bits.Add64(z[5], 1873798617647539866, c)
	}
}

func subAssign(z, y *fe) {
	var b uint64
	z[0], b = bits.Sub64(z[0], y[0], 0)
	z[1], b = bits.Sub64(z[1], y[1], b)
	z[2], b = bits.Sub64(z[2], y[2], b)
	z[3], b = bits.Sub64(z[3], y[3], b)
	z[4], b = bits.Sub64(z[4], y[4], b)
	z[5], b = bits.Sub64(z[5], y[5], b)
	if b != 0 {
		var c uint64
		z[0], c = bits.Add64(z[0], 13402431016077863595, 0)
		z[1], c = bits.Add64(z[1], 2210141511517208575, c)
		z[2], c = bits.Add64(z[2], 7435674573564081700, c)
		z[3], c = bits.Add64(z[3], 7239337960414712511, c)
		z[4], c = bits.Add64(z[4], 5412103778470702295, c)
		z[5], _ = bits.Add64(z[5], 1873798617647539866, c)
	}
}

func lsubAssign(z, y *fe) {
	var b uint64
	z[0], b = bits.Sub64(z[0], y[0], 0)
	z[1], b = bits.Sub64(z[1], y[1], b)
	z[2], b = bits.Sub64(z[2], y[2], b)
	z[3], b = bits.Sub64(z[3], y[3], b)
	z[4], b = bits.Sub64(z[4], y[4], b)
	z[5], b = bits.Sub64(z[5], y[5], b)
}

func neg(z, x *fe) {
	if x.isZero() {
		z.zero()
		return
	}
	var borrow uint64
	z[0], borrow = bits.Sub64(13402431016077863595, x[0], 0)
	z[1], borrow = bits.Sub64(2210141511517208575, x[1], borrow)
	z[2], borrow = bits.Sub64(7435674573564081700, x[2], borrow)
	z[3], borrow = bits.Sub64(7239337960414712511, x[3], borrow)
	z[4], borrow = bits.Sub64(5412103778470702295, x[4], borrow)
	z[5], _ = bits.Sub64(1873798617647539866, x[5], borrow)
}

func mul(z, x, y *fe) {

	var t [6]uint64
	var c [3]uint64
	{
		// round 0
		v := x[0]
		c[1], c[0] = bits.Mul64(v, y[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd1(v, y[1], c[1])
		c[2], t[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd1(v, y[2], c[1])
		c[2], t[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd1(v, y[3], c[1])
		c[2], t[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd1(v, y[4], c[1])
		c[2], t[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd1(v, y[5], c[1])
		t[5], t[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}
	{
		// round 1
		v := x[1]
		c[1], c[0] = madd1(v, y[0], t[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[2], t[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[2], t[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd2(v, y[3], c[1], t[3])
		c[2], t[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd2(v, y[4], c[1], t[4])
		c[2], t[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd2(v, y[5], c[1], t[5])
		t[5], t[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}
	{
		// round 2
		v := x[2]
		c[1], c[0] = madd1(v, y[0], t[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[2], t[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[2], t[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd2(v, y[3], c[1], t[3])
		c[2], t[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd2(v, y[4], c[1], t[4])
		c[2], t[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd2(v, y[5], c[1], t[5])
		t[5], t[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}
	{
		// round 3
		v := x[3]
		c[1], c[0] = madd1(v, y[0], t[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[2], t[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[2], t[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd2(v, y[3], c[1], t[3])
		c[2], t[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd2(v, y[4], c[1], t[4])
		c[2], t[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd2(v, y[5], c[1], t[5])
		t[5], t[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}
	{
		// round 4
		v := x[4]
		c[1], c[0] = madd1(v, y[0], t[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[2], t[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[2], t[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd2(v, y[3], c[1], t[3])
		c[2], t[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd2(v, y[4], c[1], t[4])
		c[2], t[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd2(v, y[5], c[1], t[5])
		t[5], t[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}
	{
		// round 5
		v := x[5]
		c[1], c[0] = madd1(v, y[0], t[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd2(v, y[1], c[1], t[1])
		c[2], z[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd2(v, y[2], c[1], t[2])
		c[2], z[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd2(v, y[3], c[1], t[3])
		c[2], z[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd2(v, y[4], c[1], t[4])
		c[2], z[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd2(v, y[5], c[1], t[5])
		z[5], z[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[5] < 1873798617647539866 || (z[5] == 1873798617647539866 && (z[4] < 5412103778470702295 || (z[4] == 5412103778470702295 && (z[3] < 7239337960414712511 || (z[3] == 7239337960414712511 && (z[2] < 7435674573564081700 || (z[2] == 7435674573564081700 && (z[1] < 2210141511517208575 || (z[1] == 2210141511517208575 && (z[0] < 13402431016077863595))))))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 13402431016077863595, 0)
		z[1], b = bits.Sub64(z[1], 2210141511517208575, b)
		z[2], b = bits.Sub64(z[2], 7435674573564081700, b)
		z[3], b = bits.Sub64(z[3], 7239337960414712511, b)
		z[4], b = bits.Sub64(z[4], 5412103778470702295, b)
		z[5], _ = bits.Sub64(z[5], 1873798617647539866, b)
	}
}

func square(z, x *fe) {

	var t [6]uint64
	var c [3]uint64
	{
		// round 0
		v := x[0]
		c[1], c[0] = bits.Mul64(v, x[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd1(v, x[1], c[1])
		c[2], t[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd1(v, x[2], c[1])
		c[2], t[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd1(v, x[3], c[1])
		c[2], t[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd1(v, x[4], c[1])
		c[2], t[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd1(v, x[5], c[1])
		t[5], t[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}
	{
		// round 1
		v := x[1]
		c[1], c[0] = madd1(v, x[0], t[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd2(v, x[1], c[1], t[1])
		c[2], t[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd2(v, x[2], c[1], t[2])
		c[2], t[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd2(v, x[3], c[1], t[3])
		c[2], t[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd2(v, x[4], c[1], t[4])
		c[2], t[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd2(v, x[5], c[1], t[5])
		t[5], t[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}
	{
		// round 2
		v := x[2]
		c[1], c[0] = madd1(v, x[0], t[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd2(v, x[1], c[1], t[1])
		c[2], t[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd2(v, x[2], c[1], t[2])
		c[2], t[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd2(v, x[3], c[1], t[3])
		c[2], t[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd2(v, x[4], c[1], t[4])
		c[2], t[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd2(v, x[5], c[1], t[5])
		t[5], t[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}
	{
		// round 3
		v := x[3]
		c[1], c[0] = madd1(v, x[0], t[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd2(v, x[1], c[1], t[1])
		c[2], t[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd2(v, x[2], c[1], t[2])
		c[2], t[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd2(v, x[3], c[1], t[3])
		c[2], t[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd2(v, x[4], c[1], t[4])
		c[2], t[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd2(v, x[5], c[1], t[5])
		t[5], t[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}
	{
		// round 4
		v := x[4]
		c[1], c[0] = madd1(v, x[0], t[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd2(v, x[1], c[1], t[1])
		c[2], t[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd2(v, x[2], c[1], t[2])
		c[2], t[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd2(v, x[3], c[1], t[3])
		c[2], t[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd2(v, x[4], c[1], t[4])
		c[2], t[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd2(v, x[5], c[1], t[5])
		t[5], t[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}
	{
		// round 5
		v := x[5]
		c[1], c[0] = madd1(v, x[0], t[0])
		m := c[0] * 9940570264628428797
		c[2] = madd0(m, 13402431016077863595, c[0])
		c[1], c[0] = madd2(v, x[1], c[1], t[1])
		c[2], z[0] = madd2(m, 2210141511517208575, c[2], c[0])
		c[1], c[0] = madd2(v, x[2], c[1], t[2])
		c[2], z[1] = madd2(m, 7435674573564081700, c[2], c[0])
		c[1], c[0] = madd2(v, x[3], c[1], t[3])
		c[2], z[2] = madd2(m, 7239337960414712511, c[2], c[0])
		c[1], c[0] = madd2(v, x[4], c[1], t[4])
		c[2], z[3] = madd2(m, 5412103778470702295, c[2], c[0])
		c[1], c[0] = madd2(v, x[5], c[1], t[5])
		z[5], z[4] = madd3(m, 1873798617647539866, c[0], c[2], c[1])
	}

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[5] < 1873798617647539866 || (z[5] == 1873798617647539866 && (z[4] < 5412103778470702295 || (z[4] == 5412103778470702295 && (z[3] < 7239337960414712511 || (z[3] == 7239337960414712511 && (z[2] < 7435674573564081700 || (z[2] == 7435674573564081700 && (z[1] < 2210141511517208575 || (z[1] == 2210141511517208575 && (z[0] < 13402431016077863595))))))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 13402431016077863595, 0)
		z[1], b = bits.Sub64(z[1], 2210141511517208575, b)
		z[2], b = bits.Sub64(z[2], 7435674573564081700, b)
		z[3], b = bits.Sub64(z[3], 7239337960414712511, b)
		z[4], b = bits.Sub64(z[4], 5412103778470702295, b)
		z[5], _ = bits.Sub64(z[5], 1873798617647539866, b)
	}
}

func wadd(z, x, y *wfe) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], y[0], 0)
	z[1], carry = bits.Add64(x[1], y[1], carry)
	z[2], carry = bits.Add64(x[2], y[2], carry)
	z[3], carry = bits.Add64(x[3], y[3], carry)
	z[4], carry = bits.Add64(x[4], y[4], carry)
	z[5], carry = bits.Add64(x[5], y[5], carry)
	z[6], carry = bits.Add64(x[6], y[6], carry)
	z[7], carry = bits.Add64(x[7], y[7], carry)
	z[8], carry = bits.Add64(x[8], y[8], carry)
	z[9], carry = bits.Add64(x[9], y[9], carry)
	z[10], carry = bits.Add64(x[10], y[10], carry)
	z[11], _ = bits.Add64(x[11], y[11], carry)

	if !(z[11] < 1873798617647539866 || (z[11] == 1873798617647539866 && (z[10] < 5412103778470702295 || (z[10] == 5412103778470702295 && (z[9] < 7239337960414712511 || (z[9] == 7239337960414712511 && (z[8] < 7435674573564081700 || (z[8] == 7435674573564081700 && (z[7] < 2210141511517208575 || (z[7] == 2210141511517208575 && (z[6] < 13402431016077863595))))))))))) {
		var b uint64
		z[6], b = bits.Sub64(z[6], 13402431016077863595, 0)
		z[7], b = bits.Sub64(z[7], 2210141511517208575, b)
		z[8], b = bits.Sub64(z[8], 7435674573564081700, b)
		z[9], b = bits.Sub64(z[9], 7239337960414712511, b)
		z[10], b = bits.Sub64(z[10], 5412103778470702295, b)
		z[11], _ = bits.Sub64(z[11], 1873798617647539866, b)
	}
}

func waddAssign(x, y *wfe) {
	wadd(x, x, y)
}

func lwadd(z, x, y *wfe) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], y[0], 0)
	z[1], carry = bits.Add64(x[1], y[1], carry)
	z[2], carry = bits.Add64(x[2], y[2], carry)
	z[3], carry = bits.Add64(x[3], y[3], carry)
	z[4], carry = bits.Add64(x[4], y[4], carry)
	z[5], carry = bits.Add64(x[5], y[5], carry)
	z[6], carry = bits.Add64(x[6], y[6], carry)
	z[7], carry = bits.Add64(x[7], y[7], carry)
	z[8], carry = bits.Add64(x[8], y[8], carry)
	z[9], carry = bits.Add64(x[9], y[9], carry)
	z[10], carry = bits.Add64(x[10], y[10], carry)
	z[11], _ = bits.Add64(x[11], y[11], carry)
}

func lwaddAssign(x, y *wfe) {
	lwadd(x, x, y)
}

func wsub(z, x, y *wfe) {
	var b uint64
	z[0], b = bits.Sub64(x[0], y[0], 0)
	z[1], b = bits.Sub64(x[1], y[1], b)
	z[2], b = bits.Sub64(x[2], y[2], b)
	z[3], b = bits.Sub64(x[3], y[3], b)
	z[4], b = bits.Sub64(x[4], y[4], b)
	z[5], b = bits.Sub64(x[5], y[5], b)
	z[6], b = bits.Sub64(x[6], y[6], b)
	z[7], b = bits.Sub64(x[7], y[7], b)
	z[8], b = bits.Sub64(x[8], y[8], b)
	z[9], b = bits.Sub64(x[9], y[9], b)
	z[10], b = bits.Sub64(x[10], y[10], b)
	z[11], b = bits.Sub64(x[11], y[11], b)
	if b != 0 {
		var c uint64
		z[6], c = bits.Add64(z[6], 13402431016077863595, 0)
		z[7], c = bits.Add64(z[7], 2210141511517208575, c)
		z[8], c = bits.Add64(z[8], 7435674573564081700, c)
		z[9], c = bits.Add64(z[9], 7239337960414712511, c)
		z[10], c = bits.Add64(z[10], 5412103778470702295, c)
		z[11], _ = bits.Add64(z[11], 1873798617647539866, c)
	}
}

func wsubAssign(x, y *wfe) {
	wsub(x, x, y)
}

func lwsub(z, x, y *wfe) {
	var b uint64
	z[0], b = bits.Sub64(x[0], y[0], 0)
	z[1], b = bits.Sub64(x[1], y[1], b)
	z[2], b = bits.Sub64(x[2], y[2], b)
	z[3], b = bits.Sub64(x[3], y[3], b)
	z[4], b = bits.Sub64(x[4], y[4], b)
	z[5], b = bits.Sub64(x[5], y[5], b)
	z[6], b = bits.Sub64(x[6], y[6], b)
	z[7], b = bits.Sub64(x[7], y[7], b)
	z[8], b = bits.Sub64(x[8], y[8], b)
	z[9], b = bits.Sub64(x[9], y[9], b)
	z[10], b = bits.Sub64(x[10], y[10], b)
	z[11], b = bits.Sub64(x[11], y[11], b)
}

func lwsubAssign(x, y *wfe) {
	lwsub(x, x, y)
}

func wdouble(z, x *wfe) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], x[0], 0)
	z[1], carry = bits.Add64(x[1], x[1], carry)
	z[2], carry = bits.Add64(x[2], x[2], carry)
	z[3], carry = bits.Add64(x[3], x[3], carry)
	z[4], carry = bits.Add64(x[4], x[4], carry)
	z[5], carry = bits.Add64(x[5], x[5], carry)
	z[6], carry = bits.Add64(x[6], x[6], carry)
	z[7], carry = bits.Add64(x[7], x[7], carry)
	z[8], carry = bits.Add64(x[8], x[8], carry)
	z[9], carry = bits.Add64(x[9], x[9], carry)
	z[10], carry = bits.Add64(x[10], x[10], carry)
	z[11], _ = bits.Add64(x[11], x[11], carry)

	if !(z[11] < 1873798617647539866 || (z[11] == 1873798617647539866 && (z[10] < 5412103778470702295 || (z[10] == 5412103778470702295 && (z[9] < 7239337960414712511 || (z[9] == 7239337960414712511 && (z[8] < 7435674573564081700 || (z[8] == 7435674573564081700 && (z[7] < 2210141511517208575 || (z[7] == 2210141511517208575 && (z[6] < 13402431016077863595))))))))))) {
		var b uint64
		z[6], b = bits.Sub64(z[6], 13402431016077863595, 0)
		z[7], b = bits.Sub64(z[7], 2210141511517208575, b)
		z[8], b = bits.Sub64(z[8], 7435674573564081700, b)
		z[9], b = bits.Sub64(z[9], 7239337960414712511, b)
		z[10], b = bits.Sub64(z[10], 5412103778470702295, b)
		z[11], _ = bits.Sub64(z[11], 1873798617647539866, b)
	}
}

func wdoubleAssign(x *wfe) {
	wdouble(x, x)
}

func lwdouble(z, x *wfe) {
	var carry uint64

	z[0], carry = bits.Add64(x[0], x[0], 0)
	z[1], carry = bits.Add64(x[1], x[1], carry)
	z[2], carry = bits.Add64(x[2], x[2], carry)
	z[3], carry = bits.Add64(x[3], x[3], carry)
	z[4], carry = bits.Add64(x[4], x[4], carry)
	z[5], carry = bits.Add64(x[5], x[5], carry)
	z[6], carry = bits.Add64(x[6], x[6], carry)
	z[7], carry = bits.Add64(x[7], x[7], carry)
	z[8], carry = bits.Add64(x[8], x[8], carry)
	z[9], carry = bits.Add64(x[9], x[9], carry)
	z[10], carry = bits.Add64(x[10], x[10], carry)
	z[11], _ = bits.Add64(x[11], x[11], carry)
}

func fromWide(c *fe, w *wfe) {
	montRed(c, w)
}

func wmul(w *wfe, a, b *fe) {

	var w0, w1, w2, w3, w4, w5, w6, w7, w8, w9, w10, w11 uint64
	var a0 = a[0]
	var a1 = a[1]
	var a2 = a[2]
	var a3 = a[3]
	var a4 = a[4]
	var a5 = a[5]
	var b0 = b[0]
	var b1 = b[1]
	var b2 = b[2]
	var b3 = b[3]
	var b4 = b[4]
	var b5 = b[5]
	var u, v, c, t uint64

	{
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
		c = u + (v&c|(v|c)&^w3)>>63

		// i = 0, j = 4
		u, v = bits.Mul64(a4, b0)
		w4 = v + c
		c = u + (v&c|(v|c)&^w4)>>63

		// i = 0, j = 5
		u, v = bits.Mul64(a5, b0)
		w5 = v + c
		w6 = u + (v&c|(v|c)&^w5)>>63
	}

	{

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
		c = u + (t&c|(t|c)&^w4)>>63

		// i = 1, j = 4
		u, v = bits.Mul64(a4, b1)
		t = v + w5
		u += (v&w5 | (v|w5)&^t) >> 63
		w5 = t + c
		c = u + (t&c|(t|c)&^w5)>>63

		// i = 1, j = 5
		u, v = bits.Mul64(a5, b1)
		t = v + w6
		u += (v&w6 | (v|w6)&^t) >> 63
		w6 = t + c
		w7 = u + (t&c|(t|c)&^w6)>>63
	}

	{
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
		c = u + (t&c|(t|c)&^w5)>>63

		// i = 2, j = 4
		u, v = bits.Mul64(a4, b2)
		t = v + w6
		u += (v&w6 | (v|w6)&^t) >> 63
		w6 = t + c
		c = u + (t&c|(t|c)&^w6)>>63

		// i = 2, j = 5
		u, v = bits.Mul64(a5, b2)
		t = v + w7
		u += (v&w7 | (v|w7)&^t) >> 63
		w7 = t + c
		w8 = u + (t&c|(t|c)&^w7)>>63
	}

	{
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
		c = u + (t&c|(t|c)&^w6)>>63

		// i = 3, j = 4
		u, v = bits.Mul64(a4, b3)
		t = v + w7
		u += (v&w7 | (v|w7)&^t) >> 63
		w7 = t + c
		c = u + (t&c|(t|c)&^w7)>>63

		// i = 3, j = 5
		u, v = bits.Mul64(a5, b3)
		t = v + w8
		u += (v&w8 | (v|w8)&^t) >> 63
		w8 = t + c
		w9 = u + (t&c|(t|c)&^w8)>>63
	}

	{
		// i = 4, j = 0
		c, v = bits.Mul64(a0, b4)
		t = v + w4
		c += (v&w4 | (v|w4)&^t) >> 63
		w4 = t

		// i = 4, j = 1
		u, v = bits.Mul64(a1, b4)
		t = v + w5
		u += (v&w5 | (v|w5)&^t) >> 63
		w5 = t + c
		c = u + (t&c|(t|c)&^w5)>>63

		// i = 4, j = 2
		u, v = bits.Mul64(a2, b4)
		t = v + w6
		u += (v&w6 | (v|w6)&^t) >> 63
		w6 = t + c
		c = u + (t&c|(t|c)&^w6)>>63

		// i = 4, j = 3
		u, v = bits.Mul64(a3, b4)
		t = v + w7
		u += (v&w7 | (v|w7)&^t) >> 63
		w7 = t + c
		c = u + (t&c|(t|c)&^w7)>>63

		// i = 4, j = 4
		u, v = bits.Mul64(a4, b4)
		t = v + w8
		u += (v&w8 | (v|w8)&^t) >> 63
		w8 = t + c
		c = u + (t&c|(t|c)&^w8)>>63

		// i = 4, j = 5
		u, v = bits.Mul64(a5, b4)
		t = v + w9
		u += (v&w9 | (v|w9)&^t) >> 63
		w9 = t + c
		w10 = u + (t&c|(t|c)&^w9)>>63
	}

	{
		// i = 5, j = 0
		c, v = bits.Mul64(a0, b5)
		t = v + w5
		c += (v&w5 | (v|w5)&^t) >> 63
		w5 = t

		// i = 5, j = 1
		u, v = bits.Mul64(a1, b5)
		t = v + w6
		u += (v&w6 | (v|w6)&^t) >> 63
		w6 = t + c
		c = u + (t&c|(t|c)&^w6)>>63

		// i = 5, j = 2
		u, v = bits.Mul64(a2, b5)
		t = v + w7
		u += (v&w7 | (v|w7)&^t) >> 63
		w7 = t + c
		c = u + (t&c|(t|c)&^w7)>>63

		// i = 5, j = 3
		u, v = bits.Mul64(a3, b5)
		t = v + w8
		u += (v&w8 | (v|w8)&^t) >> 63
		w8 = t + c
		c = u + (t&c|(t|c)&^w8)>>63

		// i = 5, j = 4
		u, v = bits.Mul64(a4, b5)
		t = v + w9
		u += (v&w9 | (v|w9)&^t) >> 63
		w9 = t + c
		c = u + (t&c|(t|c)&^w9)>>63

		// i = 5, j = 5
		u, v = bits.Mul64(a5, b5)
		t = v + w10
		u += (v&w10 | (v|w10)&^t) >> 63
		w10 = t + c
		w11 = u + (t&c|(t|c)&^w10)>>63
	}

	w[0] = w0
	w[1] = w1
	w[2] = w2
	w[3] = w3
	w[4] = w4
	w[5] = w5
	w[6] = w6
	w[7] = w7
	w[8] = w8
	w[9] = w9
	w[10] = w10
	w[11] = w11
}

func montRed(c *fe, w *wfe) {

	// Reduces T as T (R^-1) modp
	// Handbook of Applied Cryptography
	// Hankerson, Menezes, Vanstone
	// Algorithm 14.32 Montgomery reduction

	w0 := w[0]
	w1 := w[1]
	w2 := w[2]
	w3 := w[3]
	w4 := w[4]
	w5 := w[5]
	w6 := w[6]
	w7 := w[7]
	w8 := w[8]
	w9 := w[9]
	w10 := w[10]
	w11 := w[11]
	p0 := modulus[0]
	p1 := modulus[1]
	p2 := modulus[2]
	p3 := modulus[3]
	p4 := modulus[4]
	p5 := modulus[5]

	var e1, e2, el, res uint64
	var t1, t2, u uint64

	{

		// i = 0
		u = w0 * inp
		//
		e1, res = bits.Mul64(u, p0)
		t1 = res + w0
		e1 += (res&w0 | (res|w0)&^t1) >> 63
		//
		e2, res = bits.Mul64(u, p1)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w1
		e2 += (t1&w1 | (t1|w1)&^t2) >> 63
		w1 = t2
		//
		e1, res = bits.Mul64(u, p2)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w2
		e1 += (t1&w2 | (t1|w2)&^t2) >> 63
		w2 = t2
		//
		e2, res = bits.Mul64(u, p3)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w3
		e2 += (t1&w3 | (t1|w3)&^t2) >> 63
		w3 = t2
		//
		e1, res = bits.Mul64(u, p4)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w4
		e1 += (t1&w4 | (t1|w4)&^t2) >> 63
		w4 = t2
		//
		e2, res = bits.Mul64(u, p5)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w5
		e2 += (t1&w5 | (t1|w5)&^t2) >> 63
		w5 = t2
		//
		t1 = w6 + el
		e1 = (w6&el | (w6|el)&^t1) >> 63
		t2 = t1 + e2
		e1 += (t1&e2 | (t1|e2)&^t2) >> 63
		w6 = t2
		el = e1
	}

	{
		// i = 1
		u = w1 * inp
		//
		e1, res = bits.Mul64(u, p0)
		t1 = res + w1
		e1 += (res&w1 | (res|w1)&^t1) >> 63
		//
		e2, res = bits.Mul64(u, p1)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w2
		e2 += (t1&w2 | (t1|w2)&^t2) >> 63
		w2 = t2
		//
		e1, res = bits.Mul64(u, p2)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w3
		e1 += (t1&w3 | (t1|w3)&^t2) >> 63
		w3 = t2
		//
		e2, res = bits.Mul64(u, p3)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w4
		e2 += (t1&w4 | (t1|w4)&^t2) >> 63
		w4 = t2
		//
		e1, res = bits.Mul64(u, p4)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w5
		e1 += (t1&w5 | (t1|w5)&^t2) >> 63
		w5 = t2
		//
		e2, res = bits.Mul64(u, p5)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w6
		e2 += (t1&w6 | (t1|w6)&^t2) >> 63
		w6 = t2
		//
		t1 = w7 + el
		e1 = (w7&el | (w7|el)&^t1) >> 63
		t2 = t1 + e2
		e1 += (t1&e2 | (t1|e2)&^t2) >> 63
		w7 = t2
		el = e1
	}

	{
		// i = 2
		u = w2 * inp
		//
		e1, res = bits.Mul64(u, p0)
		t1 = res + w2
		e1 += (res&w2 | (res|w2)&^t1) >> 63
		//
		e2, res = bits.Mul64(u, p1)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w3
		e2 += (t1&w3 | (t1|w3)&^t2) >> 63
		w3 = t2
		//
		e1, res = bits.Mul64(u, p2)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w4
		e1 += (t1&w4 | (t1|w4)&^t2) >> 63
		w4 = t2
		//
		e2, res = bits.Mul64(u, p3)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w5
		e2 += (t1&w5 | (t1|w5)&^t2) >> 63
		w5 = t2
		//
		e1, res = bits.Mul64(u, p4)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w6
		e1 += (t1&w6 | (t1|w6)&^t2) >> 63
		w6 = t2
		//
		e2, res = bits.Mul64(u, p5)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w7
		e2 += (t1&w7 | (t1|w7)&^t2) >> 63
		w7 = t2
		//
		t1 = w8 + el
		e1 = (w8&el | (w8|el)&^t1) >> 63
		t2 = t1 + e2
		e1 += (t1&e2 | (t1|e2)&^t2) >> 63
		w8 = t2
		el = e1
	}

	{
		// i = 3
		u = w3 * inp
		//
		e1, res = bits.Mul64(u, p0)
		t1 = res + w3
		e1 += (res&w3 | (res|w3)&^t1) >> 63
		//
		e2, res = bits.Mul64(u, p1)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w4
		e2 += (t1&w4 | (t1|w4)&^t2) >> 63
		w4 = t2
		//
		e1, res = bits.Mul64(u, p2)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w5
		e1 += (t1&w5 | (t1|w5)&^t2) >> 63
		w5 = t2
		//
		e2, res = bits.Mul64(u, p3)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w6
		e2 += (t1&w6 | (t1|w6)&^t2) >> 63
		w6 = t2
		//
		e1, res = bits.Mul64(u, p4)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w7
		e1 += (t1&w7 | (t1|w7)&^t2) >> 63
		w7 = t2
		//
		e2, res = bits.Mul64(u, p5)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w8
		e2 += (t1&w8 | (t1|w8)&^t2) >> 63
		w8 = t2
		//
		t1 = w9 + el
		e1 = (w9&el | (w9|el)&^t1) >> 63
		t2 = t1 + e2
		e1 += (t1&e2 | (t1|e2)&^t2) >> 63
		w9 = t2
		el = e1
	}

	{
		// i = 4
		u = w4 * inp
		//
		e1, res = bits.Mul64(u, p0)
		t1 = res + w4
		e1 += (res&w4 | (res|w4)&^t1) >> 63
		//
		e2, res = bits.Mul64(u, p1)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w5
		e2 += (t1&w5 | (t1|w5)&^t2) >> 63
		w5 = t2
		//
		e1, res = bits.Mul64(u, p2)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w6
		e1 += (t1&w6 | (t1|w6)&^t2) >> 63
		w6 = t2
		//
		e2, res = bits.Mul64(u, p3)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w7
		e2 += (t1&w7 | (t1|w7)&^t2) >> 63
		w7 = t2
		//
		e1, res = bits.Mul64(u, p4)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w8
		e1 += (t1&w8 | (t1|w8)&^t2) >> 63
		w8 = t2
		//
		e2, res = bits.Mul64(u, p5)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w9
		e2 += (t1&w9 | (t1|w9)&^t2) >> 63
		w9 = t2
		//
		t1 = w10 + el
		e1 = (w10&el | (w10|el)&^t1) >> 63
		t2 = t1 + e2
		e1 += (t1&e2 | (t1|e2)&^t2) >> 63
		w10 = t2
		el = e1
	}

	{
		// i = 5
		u = w5 * inp
		//
		e1, res = bits.Mul64(u, p0)
		t1 = res + w5
		e1 += (res&w5 | (res|w5)&^t1) >> 63
		//
		e2, res = bits.Mul64(u, p1)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w6
		e2 += (t1&w6 | (t1|w6)&^t2) >> 63
		w6 = t2
		//
		e1, res = bits.Mul64(u, p2)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w7
		e1 += (t1&w7 | (t1|w7)&^t2) >> 63
		w7 = t2
		//
		e2, res = bits.Mul64(u, p3)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w8
		e2 += (t1&w8 | (t1|w8)&^t2) >> 63
		w8 = t2
		//
		e1, res = bits.Mul64(u, p4)
		t1 = res + e2
		e1 += (res&e2 | (res|e2)&^t1) >> 63
		t2 = t1 + w9
		e1 += (t1&w9 | (t1|w9)&^t2) >> 63
		w9 = t2
		//
		e2, res = bits.Mul64(u, p5)
		t1 = res + e1
		e2 += (res&e1 | (res|e1)&^t1) >> 63
		t2 = t1 + w10
		e2 += (t1&w10 | (t1|w10)&^t2) >> 63
		w10 = t2
		//
		t1 = w11 + el
		e1 = (w11&el | (w11|el)&^t1) >> 63
		t2 = t1 + e2
		e1 += (t1&e2 | (t1|e2)&^t2) >> 63
		w11 = t2
	}

	e1--
	c[0] = w6 - ((p0) & ^e1)
	e2 = (^w6&p0 | (^w6|p0)&c[0]) >> 63
	c[1] = w7 - ((p1 + e2) & ^e1)
	e2 = (^w7&p1 | (^w7|p1)&c[1]) >> 63
	c[2] = w8 - ((p2 + e2) & ^e1)
	e2 = (^w8&p2 | (^w8|p2)&c[2]) >> 63
	c[3] = w9 - ((p3 + e2) & ^e1)
	e2 = (^w9&p3 | (^w9|p3)&c[3]) >> 63
	c[4] = w10 - ((p4 + e2) & ^e1)
	e2 = (^w10&p4 | (^w10|p4)&c[4]) >> 63
	c[5] = w11 - ((p5 + e2) & ^e1)

	sub(c, c, &modulus)
}

func fp2Add(c, a, b *fe2) {
	add(&c[0], &a[0], &b[0])
	add(&c[1], &a[1], &b[1])
}

func fp2AddAssign(a, b *fe2) {
	addAssign(&a[0], &b[0])
	addAssign(&a[1], &b[1])
}

func fp2Ladd(c, a, b *fe2) {
	ladd(&c[0], &a[0], &b[0])
	ladd(&c[1], &a[1], &b[1])
}

func fp2LaddAssign(a, b *fe2) {
	laddAssign(&a[0], &b[0])
	laddAssign(&a[1], &b[1])
}

func fp2Double(c, a *fe2) {
	double(&c[0], &a[0])
	double(&c[1], &a[1])
}

func fp2DoubleAssign(a *fe2) {
	doubleAssign(&a[0])
	doubleAssign(&a[1])
}

func fp2Ldouble(c, a *fe2) {
	ldouble(&c[0], &a[0])
	ldouble(&c[1], &a[1])
}

func fp2Sub(c, a, b *fe2) {
	sub(&c[0], &a[0], &b[0])
	sub(&c[1], &a[1], &b[1])
}

func fp2SubAssign(c, a *fe2) {
	subAssign(&c[0], &a[0])
	subAssign(&c[1], &a[1])
}

func mulByNonResidue(c, a *fe2) {
	t := new(fe)
	sub(t, &a[0], &a[1])
	add(&c[1], &a[0], &a[1])
	c[0].set(t)
}

func mulByNonResidueAssign(a *fe2) {
	t := new(fe)
	sub(t, &a[0], &a[1])
	add(&a[1], &a[0], &a[1])
	a[0].set(t)
}

func wfp2Add(c, a, b *wfe2) {
	wadd(&c[0], &a[0], &b[0])
	wadd(&c[1], &a[1], &b[1])
}

func wfp2AddAssign(c, a *wfe2) {
	waddAssign(&c[0], &a[0])
	waddAssign(&c[1], &a[1])
}

func wfp2Ladd(c, a, b *wfe2) {
	lwadd(&c[0], &a[0], &b[0])
	lwadd(&c[1], &a[1], &b[1])
}

func wfp2LaddAssign(a, b *wfe2) {
	lwaddAssign(&a[0], &b[0])
	lwaddAssign(&a[1], &b[1])
}

func wfp2AddMixed(c, a, b *wfe2) {
	wadd(&c[0], &a[0], &b[0])
	lwadd(&c[1], &a[1], &b[1])
}

func wfp2AddMixedAssign(a, b *wfe2) {
	waddAssign(&a[0], &b[0])
	lwaddAssign(&a[1], &b[1])
}

func wfp2Sub(c, a, b *wfe2) {
	wsub(&c[0], &a[0], &b[0])
	wsub(&c[1], &a[1], &b[1])
}

func wfp2SubAssign(a, b *wfe2) {
	wsub(&a[0], &a[0], &b[0])
	wsub(&a[1], &a[1], &b[1])
}

func wfp2SubMixed(c, a, b *wfe2) {
	wsub(&c[0], &a[0], &b[0])
	lwsub(&c[1], &a[1], &b[1])
}

func wfp2SubMixedAssign(a, b *wfe2) {
	wsubAssign(&a[0], &b[0])
	lwsubAssign(&a[1], &b[1])
}

func wfp2Double(c, a *wfe2) {
	wdouble(&c[0], &a[0])
	wdouble(&c[1], &a[1])
}

func wfp2DoubleAssign(a *wfe2) {
	wdoubleAssign(&a[0])
	wdoubleAssign(&a[1])
}

func wfp2MulByNonResidue(c, a *wfe2) {
	wt0 := &wfe{}
	wadd(wt0, &a[0], &a[1])
	wsub(&c[0], &a[0], &a[1])
	c[1].set(wt0)
}

func wfp2MulByNonResidueAssign(a *wfe2) {
	wt0 := &wfe{}
	wadd(wt0, &a[0], &a[1])
	wsub(&a[0], &a[0], &a[1])
	a[1].set(wt0)
}

var wfp2Mul func(c *wfe2, a, b *fe2) = wfp2MulGeneric
var wfp2Square func(c *wfe2, a *fe2) = wfp2SquareGeneric