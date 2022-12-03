# 查询字符串

可以查询区块高度、区块的时间辍、区块里的验证者节点等数据。

## 使用案例

**比较数字**

1. 判断区块高度是否大于12

实例化一个查询句柄：

```go
query, err := New("block.height > 12")
```

根据库里给的信息判断区块高度大于12是否正确，给的数据类型必须是`map[string][]string`类型的，例如：

```go
query.Matches(map[string][]string{"block.height": []string{"2", "12", "13"}})
```

从库里查询到区块的高度有2、12、13这三个情况，其中13大于12,所以匹配成功。

2. 判断验证者节点的投票权是否小于等于10.0

实例化一个查询句柄：

```go
query, err := New("validator1.power <= 10.0")
```

根据库里给的信息判断验证者节点的投票权小于等于10.0是否正确，同样，给的数据类型必须是`map[string][]string`类型的，例如：

```go
query.Matches(map[string][]string{"validator1.power": []string{"12"}, "validator2.power": []string{"9.3"}})
```

从库里查询到有两个验证者节点的投票权，其中validator1的投票权大于10.0,尽管validator2的投票权小于10.0,但是它不是我们查询的对象，所以上述匹配失败。

**比较字符串**

1. 判断验证者节点是否存在

实例化一个查询句柄：

```go
query, err := New("block.validator1 EXISTS AND block.height = 12")
```

根据库里给的信息判断高度12的区块的验证者节点集合中是否存在验证者节点validator1，同样的，给的数据类型必须是`map[string][]string`
类型的，例如：

```go
query.Matches(map[string][]string{"block.validator1": []string{"10.0"}, "block.height": []string{"12"}})
```

从库里查询到有高度为12的区块，且存在validator1节点，所以匹配成功，但是倘若我们从库里查询到的信息是下面这样的，匹配则会失败：

```go
query.Matches(map[string][]string{"block.validator1.power": []string{"10.0"}, "block.height": []string{"12"}})
```

原因在于`block.validator1`里面含有一个`.`，这样程序就会认为这是一个完整的查询对象，而不是依靠前缀匹配规则区查询，这样的话，自然无法在库提供的信息里找到合适的匹配，详见匹配源码：

```go
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
```

2. 判断区块里是否包含指定的节点

实例化一个查询句柄：

```go
query, err := New("block.validators CONTAINS 'validator1'")
```

根据库里给的信息区块的验证者节点集合中是否存在节点validator1，同样的，给的数据类型必须是`map[string][]string`类型的，例如：

```go
query.Matches(map[string][]string{"block.validators": []string{"validator2", "validator3"}})
```

由于给定的信息里不含validator1，所以匹配失败。

**比较时间和日期**

比较时间要用到关键词`TIME`，而比较日期则要用到关键词`DATE`，下面给出两个例子，分别判断区块时间是否早于指定时间：

```go
query1, err := New("block.time < TIME 2022-12-03T00:00:40+08:00")
query2, err := New("block.date < DATE 2022-11-09")
```

根据给定的信息来判断：

```go
query1.Matches(map[string][]string{"block.time": []string{time.Now().Format(time.RFC3339)}})
query2.Matches(map[string][]string{"block.date": []string{"2022-11-08"}})
```

上面的匹配结果分别是成功和失败。

## 注意事项

1. 如果要对小数做比较，那么小数的整数部分不可以是**0**。
2. 无法查询负数。
3. 如果要查询字符串，需要在字符串的两端加上引号。
4. 数字、时间、日期可以作："<=" | ">=" | ">" | "<" | "="这五种比较。
5. 字符串可以作："=" | "CONTAINS" 两种比较。
6. "EXISTS" 用来判断属性是否存在。

**peg工具在这里：** https://github.com/pointlander/peg