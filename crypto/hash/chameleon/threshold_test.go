// 利用拉格朗日插值法来计算(t,n)门限里的秘密值，定义7个3次多项式，
// 门限值t设为4，n为7，任意大于或等于4个参与者可以计算出秘密值。

package chameleon

import (
	"testing"
)

// 用户1：y = x^3 + x^2 + x + 1，11

func f1(x float64) float64 {
	return x*x*x + x*x + x + 1.0
}

// 用户2：y = 2*x^3 + 3*x^2 + 5*x + 2，12
func f2(x float64) float64 {
	return 2.0*x*x*x + 3.0*x*x + 5.0*x + 2.0
}

// 用户3：y = 3*x^3 + 6*x + 3，13
func f3(x float64) float64 {
	return 3.0*x*x*x + 6.0*x + 3.0
}

// 用户4：y = 4*x^3 + 7*x^2 + 4，14
func f4(x float64) float64 {
	return 4.0*x*x*x + 7.0*x*x + 4.0
}

// 用户5：y = 5*x^3 + 6*x + 5，15
func f5(x float64) float64 {
	return 5.0*x*x*x + 6.0*x + 5.0
}

// 用户6：y = 6*x^3 + 3*x^2 + x + 6，16
func f6(x float64) float64 {
	return 6.0*x*x*x + 3.0*x*x + x + 6.0
}

// 用户7：y = 7*x^3 + 5*x^2 + 3*x + 7，17
func f7(x float64) float64 {
	return 7.0*x*x*x + 5.0*x*x + 3.0*x + 7.0
}

func F(x float64) float64 {
	return 28.0*x*x*x + 19.0*x*x + 22.0*x + 28.0
}

type f func(float64) float64

var users = map[int]float64{1: 11.0, 2: 12.0, 3: 13.0, 4: 14.0, 5: 15.0, 6: 16.0, 7: 17.0}

var fs = []f{f1, f2, f3, f4, f5, f6, f7}

func aggregation(ids ...int) float64 {
	CalcLambdaF := func(id int) float64 {
		lambda := 0.0
		for _, fn := range fs {
			lambda += fn(users[id])
		}
		return lambda
	}
	CalcPartition := func(lambda float64, id int) float64 {
		var res float64 = 1.0
		for _, _id := range ids {
			if id != _id {
				res = res * (-users[_id]) / (users[id] - users[_id])
			}
		}
		return res * lambda
	}
	CompleteSK := 0.0
	for _, id := range ids {
		lambda := CalcLambdaF(id)
		CompleteSK += CalcPartition(lambda, id)
	}
	return CompleteSK
}

func TestExample(t *testing.T) {
	t.Logf("完整陷门(1,2,3,4)：%.0f\n", aggregation(1, 2, 3, 4))
	t.Logf("完整陷门(1,2,3,4,5)：%.0f\n", aggregation(1, 2, 3, 4, 5))
	t.Logf("完整陷门(2,4,6,7)：%.0f\n", aggregation(2, 4, 6, 7))
	t.Logf("完整陷门(3,4,6,1)：%.0f\n", aggregation(3, 4, 6, 1))
	t.Logf("完整陷门(1, 2, 3)：%.0f\n", aggregation(1, 2, 3))
	t.Logf("完整陷门(1, 2, 3, 4, 5, 6, 7)：%.0f\n", aggregation(1, 2, 3, 4, 5, 6, 7))
}
