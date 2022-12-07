package main

import "fmt"

func main() {
	var s int
	var n int
	//fmt.Scan(&s)
	//fmt.Scan(&n)
	s = 4
	n = 7
	var x, y *int
	x = new(int)
	y = new(int)
	*x = 0
	*y = 0
	var d int = Exgcd(s, n, x, y)
	var m int = Mod_reverse(s, n)
	fmt.Println(d)
	fmt.Println(m)
}
func Exgcd(a, b int, x, y *int) int {

	var d int
	if b == 0 {
		*x = 1
		*y = 0
		return a
	}
	d = Exgcd(b, a%b, y, x)
	fmt.Printf("div: %v, x: %v, y: %v, mu: %v\n", a/b, *x, *y, (a/b)**x)
	*y = *y - (a/b)**x
	fmt.Printf("y: %v\n", *y)
	return d
}
func Mod_reverse(a, mod int) int {
	var d int
	var x, y *int
	x = new(int)
	y = new(int)
	d = Exgcd(a, mod, x, y)
	if d == 1 {
		if *x%mod <= 0 {
			return *x%mod + mod
		} else {
			return *x % mod
		}
	} else {
		return -1 //表示无解
	}
}
