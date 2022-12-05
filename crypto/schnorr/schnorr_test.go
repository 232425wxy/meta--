package schnorr

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"math/big"
	"strconv"
	"testing"
)

func TestName(t *testing.T) {
	main()
}

/*Here we are taking SPEC256k1 curve */
var a = big.NewInt(0)
var b = big.NewInt(7)

type Point struct {
	x, y *big.Int
}

// find order//
func GetOrder() *big.Int {
	var ord big.Int
	ord.SetString("115792089237316195423570985008687907852837564279074904382605163141518161494337", 10)
	return &ord
}

// Get prime number(modulo)//
func GetPrime() big.Int {
	var prime big.Int
	prime.SetString("115792089237316195423570985008687907853269984665640564039457584007908834671663", 10)
	return prime
}

/*
check whether the given curve is valid by calculating 4*a^3+27*b^2 and it should not be equal to zero.

	Indicates that there are no repeated roots
*/
func CheckValidCurve() bool {
	s0 := big.NewInt(0).Mul(big.NewInt(0).Mul(b, b), big.NewInt(27))
	s1 := big.NewInt(0).Mul(big.NewInt(0).Mul(big.NewInt(0).Mul(a, a), a), big.NewInt(4))
	s := big.NewInt(0).Add(s0, s1)

	if s == big.NewInt(0) {
		return false
	}
	return true
}

/*Here our curve is in the form  y^2=x^3+ax+b. Given X-coordinate of a point then find Y-coordinate.*/

func FindY(x *big.Int) Point {
	//var prime big.Int
	//prime.SetString("115792089237316195423570985008687907853269984665640564039457584007908834671663",10)
	prime := GetPrime()
	var y1, x0, x1, x2, x4 big.Int
	//var y *big.Int
	x0.Mul(a, x)
	x1.Add(&x0, b)
	x2.Mul(x, x)
	x2.Mul(&x2, x)
	x4.Add(&x1, &x2)
	y1.ModSqrt(&x4, &prime)
	y1.Sub(&prime, &y1)
	//y=y1
	return Point{x, &y1}
}

/*
This method adds two points. Where slope m = (y2-y1)/(x2-x1); new point r = (x3,y3),

	x3 = m^2-(x1+x2), y3 = -m(x3-x1)-y1
*/
func PointAddition(p Point, q Point) Point {

	var numerator, denominator, m, x3, y3 big.Int
	prime := GetPrime()
	zero := big.NewInt(0)

	if p.x.Cmp(q.x) == 0 && p.y.Cmp(q.y) == 0 {
		if p.y.Cmp(zero) == 0 {
			return Point{big.NewInt(0), big.NewInt(0)}
		}
		return PointDoubling(p)
	}
	if p.x.Cmp(q.x) == 0 {
		return Point{big.NewInt(0), big.NewInt(0)}
	}
	if q.x.Cmp(zero) == 0 && q.y.Cmp(zero) == 0 {
		return Point{p.x, p.y}
	}
	if p.x.Cmp(zero) == 0 && p.y.Cmp(zero) == 0 {
		return Point{q.x, q.y}
	}

	numerator.Mod(numerator.Sub(q.y, p.y), &prime)

	denominator.Mod(denominator.Sub(q.x, p.x), &prime)
	denominator.ModInverse(&denominator, &prime)

	m.Mul(&numerator, &denominator)
	x3.Mul(&m, &m)
	x3.Mod(x3.Sub(&x3, big.NewInt(0).Add(p.x, q.x)), &prime)
	y3.Mod(y3.Sub(&x3, p.x), &prime)
	y3.Mul(&y3, big.NewInt(0).Mul(&m, big.NewInt(-1)))
	y3.Mod(y3.Sub(&y3, p.y), &prime)
	r := Point{&x3, &y3}
	return r

}

/*
For pointDoubling slope m = (3*x1^2 + a)/2y1, new point r=(x3,y3),

	x3 = m^2-(x1+x2), y3 = -m(x3-x1)-y1
*/
func PointDoubling(p Point) Point {
	var numerator, denominator, m, x3, y3 big.Int
	prime := GetPrime()

	numerator.Mul(numerator.Mul(p.x, p.x), big.NewInt(3))
	numerator.Mod(numerator.Add(&numerator, a), &prime)

	denominator.Mod(denominator.Mul(p.y, big.NewInt(2)), &prime)
	denominator.ModInverse(&denominator, &prime)
	m.Mul(&numerator, &denominator)
	x3.Mul(&m, &m)
	x3.Mod(x3.Sub(&x3, big.NewInt(0).Add(p.x, p.x)), &prime)
	y3.Mod(y3.Sub(&x3, p.x), &prime)
	y3.Mul(&y3, big.NewInt(0).Mul(&m, big.NewInt(-1)))
	y3.Mod(y3.Sub(&y3, p.y), &prime)
	r := Point{&x3, &y3}
	return r
}

/*PointMultiplication is the scalar multiplication of the given point. Let say pp = (2,3), then find out
[2]pp, [3]pp etc.
We are using double and add algorithm for point multiplication.
For more details you can visit https://sefiks.com/2016/03/27/double-and-add-method/
* */

func PointMultiplication(p Point, n *big.Int) Point {

	bigStr := fmt.Sprintf("%b", n)
	//fmt.Println(bigStr)
	var i int
	var r Point = p
	for i = 1; i < len(bigStr); i++ {

		currentBit, err := strconv.Atoi(string(bigStr[i : i+1]))
		fmt.Println(currentBit)
		fmt.Println(err)
		r = PointAddition(r, r)
		if r.y == big.NewInt(0) {
			j := 2 * (i + 1)
			if int64(j) == n.Int64() {
				return Point{big.NewInt(0), big.NewInt(0)}
			}
		}
		if currentBit == 1 {
			r = PointAddition(r, p)
			if r.y == big.NewInt(0) {
				j := 2 * (i + 1)
				if int64(j) == n.Int64() {
					return Point{big.NewInt(0), big.NewInt(0)}
				}
			}
		}
	}
	return r
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

var pp, qq Point
var c = big.NewInt(0)

/* step 1)Choose a random point pp on elliptical curve
 *     2)Choose a random integer 'a' from range[1, r],
 *       where r is the order of point pp.
 *     3) calculate Q=[a]pp
 *     4) Output: public key Pk:(pp,[a]pp) , secret key Sk=(a, Pk) */
func KeyGeneraton() Point {
	var x big.Int
	//var Reader io.Reader
	x.SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	P := FindY(&x)
	Q := Point{big.NewInt(0), big.NewInt(0)}
	a := big.NewInt(0)
	var err error
	curveOrder := GetOrder()
	for Q.x.Cmp(big.NewInt(0)) == 0 && Q.y.Cmp(big.NewInt(0)) == 0 {
		if !(P.x == big.NewInt(0)) || !(P.x == big.NewInt(0)) {

			a, err = rand.Int(rand.Reader, curveOrder)
			if err != nil {
				//error handling
			}

			fmt.Println("Random value a: ", a)

			Q = PointMultiplication(P, a)

		}
	}
	pp = P
	qq = Q
	c = a
	return Q
}

type PreSigObj struct {
	R Point
	k *big.Int
}

type SigObj struct {
	R Point
	S *big.Int
	m string
}

/*
Calculating Point R :
* 1)Choose a random integer 'k' from range[1, r],where r is the order of point pp.
 2. Calculate Point R= [k]pp
*/
func OffLineCalculation() PreSigObj {
	curveOrder := GetOrder()
	R := Point{big.NewInt(0), big.NewInt(0)}
	k := big.NewInt(0)
	var err error

	for R.x.Cmp(big.NewInt(0)) == 0 && R.y.Cmp(big.NewInt(0)) == 0 {
		if !(pp.x == big.NewInt(0)) || !(pp.x == big.NewInt(0)) {

			k, err = rand.Int(rand.Reader, curveOrder)
			if err != nil {
				//error handling
			}

			fmt.Println("Random value k: ", k)
			R = PointMultiplication(pp, k)

		}
	}
	result := PreSigObj{R, k}

	return result
}

func Signature(obj PreSigObj, m string) SigObj {
	ae := big.NewInt(0)
	s := big.NewInt(0)
	e := big.NewInt(0)
	R := obj.R
	rx := obj.R.x.String()
	ry := obj.R.y.String()
	concatenation := m + rx + ry
	h := sha1.New()
	h.Write([]byte(concatenation))
	bs := h.Sum([]byte{})
	e.SetBytes(bs)
	e.Mod(e, GetOrder())
	ae.Mul(c, e)
	// ae.Mod(ae,getOrder())
	s.Add(ae, obj.k)
	s.Mod(s, GetOrder())
	result := SigObj{R, s, m}
	fmt.Println("e at sig", e)
	return result
}

/*In batch verification we have given list of signatures. In order to validate all signatures at once
* i)Get s value from each signature in the list and add them to get 'S'
* ii)Calculate  e = H(message || R) from each signature and sum to get 'E'
* iii)Get point R from each signature and add to get Rs
 *iv) Finally calculate, if Rs+[E]Q = [S]pp then output true else output false.*/
func BatchVerification(sigList [500]SigObj) bool {
	//func batchVerification(sigList [1]SigObj) bool{
	var sig SigObj
	var m string
	var R Point

	Rs := Point{big.NewInt(0), big.NewInt(0)}
	S := big.NewInt(0)
	E := big.NewInt(0)
	e := big.NewInt(0)
	for i := 0; i < len(sigList); i++ {
		sig = sigList[i]
		m = sig.m
		R = sig.R
		S.Add(S, sig.S)
		S.Mod(S, GetOrder())

		Rs = PointAddition(Rs, R)
		rx := sig.R.x.String()
		ry := sig.R.y.String()
		concatenation := m + rx + ry
		h := sha1.New()
		h.Write([]byte(concatenation))
		bs := h.Sum([]byte{})
		e.SetBytes(bs)
		e.Mod(e, GetOrder())
		E.Add(E, e)
		E.Mod(E, GetOrder())
	}
	EtimesQ := PointMultiplication(qq, E)
	leftSide := PointAddition(Rs, EtimesQ)
	rightSide := PointMultiplication(pp, S)

	if leftSide.x.Cmp(rightSide.x) == 0 && leftSide.y.Cmp(rightSide.y) == 0 {
		return true
	} else {
		return false
	}
}

func main() {
	result := KeyGeneraton()
	fmt.Println("Point Q is: ", result.x, result.y)
	const size int = 500
	var arr [size]PreSigObj
	var arr1 [size]SigObj
	for i := 0; i < size; i++ {
		result1 := OffLineCalculation()
		arr[i] = result1
	}

	for i := 0; i < size; i++ {
		result1 := arr[i]
		result2 := Signature(result1, "hello world")
		arr1[i] = result2
		fmt.Println("Point R", result2.R.x, result2.R.y)
		fmt.Println("s, m are: ", result2.S, result2.m)
	}

	fmt.Println("All signatures are valid: ", BatchVerification(arr1))

}
