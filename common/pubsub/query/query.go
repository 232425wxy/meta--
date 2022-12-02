package query

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 项目级全局函数

// New ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// New 解析给定的字符串，如果给定的字符串不合法会返回错误，如果一切顺利，就返回一个解析器 *Query。
func New(s string) (*Query, error) {
	parser := &QueryParser{Buffer: fmt.Sprintf(`"%s"`, s)}
	if err := parser.Init(); err != nil {
		return nil, err
	}
	if err := parser.Parse(); err != nil {
		return nil, err
	}
	return &Query{str: s, parser: parser}, nil
}

// MustParse ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// MustParse 方法接受一个字符串，然后将该字符串传给 New 方法，调用 New 方法，如果出错则直接panic，
// 如果一切顺利，就返回一个解析器 *Query。
func MustParse(s string) *Query {
	q, err := New(s)
	if err != nil {
		panic(fmt.Sprintf("pubsub/query: new query error: %q", err))
	}
	return q
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量

// Condition ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Condition 这个结构体包含了条件判断所需的三个元素：对象|操作符|值，分别对应下面三个字段：
//   - CompositeKey：例如 "block.height"
//   - Op：例如 "="
//   - Operand：例如 "21"
type Condition struct {
	CompositeKey string
	Op           Operator    // 可以是"<=" | ">=" | ">" | "<" | "=" | "CONTAINS" | "EXISTS"这七种判断条件
	Operand      interface{} // 可以是字符串、数字、日期和时间这四类数据
}

// Query ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Query 解析查询请求
type Query struct {
	str    string
	parser *QueryParser
}

// String ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// String 方法直接返回要解析的原始字符串。
func (q *Query) String() string {
	return q.str
}

// Conditions ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Conditions 解析查询字符串，从中获取所有查询条件。
func (q *Query) Conditions() ([]Condition, error) {
	var (
		compositeKey string
		op           Operator
		operand      interface{}
	)
	conditions := make([]Condition, 0)
	buffer, begin, end := q.parser.Buffer, 0, 0

	// tokens 里元素的顺序如下：tag\compositeKey -> op -> operand
	for _, token := range q.parser.Tokens() {
		switch token.pegRule {

		// 从这里开始获取操作对象
		case rulePegText:
			begin, end = int(token.begin), int(token.end)
		case ruletag:
			compositeKey = buffer[begin:end]

		// 从这里开始获取操作符
		case rulele:
			op = OpLessEqual
		case rulege:
			op = OpGreaterEqual
		case rulel:
			op = OpLess
		case ruleg:
			op = OpGreater
		case ruleequal:
			op = OpEqual
		case rulecontains:
			op = OpContains
		case ruleexists:
			op = OpExists
			// 判断一个tag/compositeKey存不存在，不需要operand
			conditions = append(conditions, Condition{CompositeKey: compositeKey, Op: op, Operand: nil})

		// 从这里开始获取操作数
		case rulevalue:
			// 字符串数据被单引号包围了起来，需要将单引号去掉
			operand = buffer[begin+1 : end-1] // "'block.height'" -> "block.height"
			conditions = append(conditions, Condition{CompositeKey: compositeKey, Op: op, Operand: operand})
		case rulenumber:
			number := buffer[begin:end]
			if strings.Contains(number, ".") {
				// 如果是小数，那就要把它转换为64位浮点数
				value, err := strconv.ParseFloat(number, 64)
				if err != nil {
					return nil, fmt.Errorf("pubsub/query: failed to parse %q to 64 bit float: %q", number, err)
				}
				conditions = append(conditions, Condition{CompositeKey: compositeKey, Op: op, Operand: value})
			} else {
				// 如果不是小数，是一个整数，那就把它转换为64位的10进制整数
				value, err := strconv.ParseInt(number, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("pubsub/query: failed to parse %q to 64 bit integer: %q", number, err)
				}
				conditions = append(conditions, Condition{CompositeKey: compositeKey, Op: op, Operand: value})
			}
		case ruletime:
			value, err := time.Parse(TimeLayout, buffer[begin:end])
			if err != nil {
				return nil, fmt.Errorf("pubsub/query: failed to parse time %q under layout %q: %q", buffer[begin:end], TimeLayout, err)
			}
			conditions = append(conditions, Condition{CompositeKey: compositeKey, Op: op, Operand: value})
		case ruledate:
			value, err := time.Parse(DateLayout, buffer[begin:end])
			if err != nil {
				return nil, fmt.Errorf("pubsub/query: failed parse date %q under layout %q: %q", buffer[begin:end], DateLayout, err)
			}
			conditions = append(conditions, Condition{CompositeKey: compositeKey, Op: op, Operand: value})
		}
	}
	return conditions, nil
}

// Matches ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Matches 给定参数events是一个map[string][]string类型的，该方法测试 Query 里待查询的字符串是否能和给定的events匹配。
func (q *Query) Matches(events map[string][]string) (bool, error) {
	if len(events) == 0 {
		return false, nil
	}

	var (
		compositeKey string
		op           Operator
	)

	buffer, begin, end := q.parser.Buffer, 0, 0

	for _, token := range q.parser.Tokens() {
		switch token.pegRule {
		// 从这里开始获取操作对象
		case rulePegText:
			begin, end = int(token.begin), int(token.end)
		case ruletag:
			compositeKey = buffer[begin:end]

		// 从这里开始获取操作符
		case rulele:
			op = OpLessEqual
		case rulege:
			op = OpGreaterEqual
		case rulel:
			op = OpLess
		case ruleg:
			op = OpGreater
		case ruleequal:
			op = OpEqual
		case rulecontains:
			op = OpContains
		case ruleexists:
			op = OpExists
			if strings.Contains(compositeKey, ".") {
				// 如果查询对象含有"."的话，则代表该对象是一个完整的事件属性
				if _, ok := events[compositeKey]; !ok {
					// 给的map里不存在这个事件属性
					return false, nil
				}
			} else {
				foundEvent := false
				for event := range events {
					if strings.Index(event, compositeKey) == 0 {
						// 假如event是"block.height"，compositeKey是"block"这样的，就能够匹配成功。
						foundEvent = true
						break
					}
				}
				if !foundEvent {
					return false, nil
				}
			}
		case rulevalue:
			// 去掉字符串两端的单引号
			operand := buffer[begin+1 : end-1]
			res, err := match(compositeKey, op, reflect.ValueOf(operand), events)
			if err != nil {
				return false, err
			}
			if !res {
				return false, nil
			}
		case rulenumber:
			number := buffer[begin:end]
			if strings.Contains(number, ".") {
				value, err := strconv.ParseFloat(number, 64)
				if err != nil {
					return false, fmt.Errorf("pubsub/query: failed to parse %q to 64 bit float: %q", number, err)
				}
				res, err := match(compositeKey, op, reflect.ValueOf(value), events)
				if err != nil {
					return false, err
				}
				if !res {
					return false, nil
				}
			} else {
				value, err := strconv.ParseInt(number, 10, 64)
				if err != nil {
					return false, fmt.Errorf("pubsub/query: failed to parse %q to 64 bit integer: %q", number, err)
				}
				res, err := match(compositeKey, op, reflect.ValueOf(value), events)
				if err != nil {
					return false, err
				}
				if !res {
					return false, nil
				}
			}
		case ruletime:
			value, err := time.Parse(TimeLayout, buffer[begin:end])
			if err != nil {
				return false, fmt.Errorf("pubsub/query: failed to parse time %q under layout %q: %q", buffer[begin:end], TimeLayout, err)
			}
			res, err := match(compositeKey, op, reflect.ValueOf(value), events)
			if err != nil {
				return false, err
			}
			if !res {
				return false, nil
			}
		case ruledate:
			value, err := time.Parse(DateLayout, buffer[begin:end])
			if err != nil {
				return false, fmt.Errorf("pubsub/query: failed parse date %q under layout %q: %q", buffer[begin:end], DateLayout, err)
			}
			res, err := match(compositeKey, op, reflect.ValueOf(value), events)
			if err != nil {
				return false, err
			}
			if !res {
				return false, nil
			}
		}
	}
	return true, nil
}

// match ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// match 给定events的建：compositeKey，找到events中compositeKey对应的[]string，也就是values，然后
// 逐个判断values里的value是否满足"value op operand"这个关系，如果有满足的，就返回true，如果都不满足，
// 就返回false。
//
//	"value op operand"：例如 "block.height = 21"
func match(compositeKey string, op Operator, operand reflect.Value, events map[string][]string) (bool, error) {
	values, ok := events[compositeKey]
	if !ok {
		return false, nil
	}
	for _, value := range values {
		res, err := matchValue(value, op, operand)
		if err != nil {
			return false, err
		}
		if res {
			return true, nil
		}
	}
	return false, nil
}

// matchValue ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// matchValue 给定比较关系里的三个元素：比较对象value、比较符op、被比较对象operand，然后判断比较关系是否成立，
// value是events里的值。
func matchValue(value string, op Operator, operand reflect.Value) (bool, error) {
	switch operand.Kind() {
	case reflect.Struct:
		// 时间 time.Time 是结构体类型
		operandAsTime := operand.Interface().(time.Time)
		var (
			v   time.Time
			err error
		)
		if strings.Contains(value, "T") {
			v, err = time.Parse(TimeLayout, value)
		} else {
			v, err = time.Parse(DateLayout, value)
		}
		if err != nil {
			return false, fmt.Errorf("pubsub/query: failed to parse time value %q: %q", value, err)
		}
		switch op {
		case OpLessEqual:
			return v.Before(operandAsTime) || v.Equal(operandAsTime), nil
		case OpGreaterEqual:
			return v.After(operandAsTime) || v.Equal(operandAsTime), nil
		case OpLess:
			return v.Before(operandAsTime), nil
		case OpGreater:
			return v.After(operandAsTime), nil
		case OpEqual:
			return v.Equal(operandAsTime), nil
		}
	case reflect.Float64:
		var v float64
		operandAsFloat64 := operand.Interface().(float64)
		filteredValue := numRegex.FindString(value)
		v, err := strconv.ParseFloat(filteredValue, 64)
		if err != nil {
			return false, fmt.Errorf("pubsub/query: failed to parse %q to 64 bit float: %q", filteredValue, err)
		}
		switch op {
		case OpLessEqual:
			return v <= operandAsFloat64, nil
		case OpGreaterEqual:
			return v >= operandAsFloat64, nil
		case OpLess:
			return v < operandAsFloat64, nil
		case OpGreater:
			return v > operandAsFloat64, nil
		case OpEqual:
			return v == operandAsFloat64, nil
		}
	case reflect.Int64:
		var (
			v   int64
			_v  float64
			err error
		)
		operandAsInt64 := operand.Interface().(int64)
		filteredValue := numRegex.FindString(value)
		if strings.Contains(filteredValue, ".") {
			_v, err = strconv.ParseFloat(filteredValue, 64)
			if err != nil {
				return false, fmt.Errorf("pubsub/query: failed to parse %q to 64 bit float: %q", filteredValue, err)
			}
			v = int64(_v)
		} else {
			v, err = strconv.ParseInt(filteredValue, 10, 64)
			if err != nil {
				return false, fmt.Errorf("pubsub/query: failed to parse %q to 64 bit integer: %q", filteredValue, err)
			}
		}
		switch op {
		case OpLessEqual:
			return v <= operandAsInt64, nil
		case OpGreaterEqual:
			return v >= operandAsInt64, nil
		case OpLess:
			return v < operandAsInt64, nil
		case OpGreater:
			return v > operandAsInt64, nil
		case OpEqual:
			return v == operandAsInt64, nil
		}
	case reflect.String:
		switch op {
		case OpEqual:
			return value == operand.String(), nil
		case OpContains:
			return strings.Contains(value, operand.String()), nil
		}
	default:
		return false, fmt.Errorf("pubsub/query: unknown kind of operand: %q", operand.Kind())
	}
	return false, nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义项目级全局变量

// Operator ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Operator 逻辑操作符："<=" | ">=" | ">" | "<" | "=" | "CONTAINS" | "EXISTS"
type Operator uint8

const (
	OpLessEqual Operator = iota
	OpGreaterEqual
	OpGreater
	OpLess
	OpEqual
	OpContains // 用来判断一个字符串是否含有给定的子字符串
	OpExists   // 用来判断一个事件属性是否存在
)

const (
	DateLayout = "2006-01-02" // 定义日期的格式："2006-01-02"
	TimeLayout = time.RFC3339 // 定义时间格式："2006-01-02T15:04:05Z07:00"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级全局变量，定义数字的正则表达式

// numRegex ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// numRegex 可以匹配小数、整数等数字
var numRegex = regexp.MustCompile(`([0-9\.]+)`)
