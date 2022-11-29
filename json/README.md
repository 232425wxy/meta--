# JSON序列化

稍微对`github.com/tendermint/tendermint/libs/json`里的代码做了些改动，得到了当前包。

## 1. 编码数字

**int、int8、int16、int32、int64**

2 -> 2

0 -> 0

-3 -> 3

\[1, 2, 3] -> \[1,2,3]

**uint、uint8、uint16、uint32、uint64**

0 -> 0

6 -> 6

\[1, 2, 3] -> \[1,2,3]

**float32、float64**

3.14 -> 3.14E+00

-3.14 -> -3.14E+00

## 2. 编码字符串

"hello, golang" -> \`"hello, golang"`

## 3. 与encoding/json不同的地方

**问题描述**

我们看下面这个例子，这里定义一个`Animal`接口和一个`Dog`结构体，注意，`Dog`实现了`Animal`接口：

```go
type Animal interface {
	Eat()
}

type Dog struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}
```

现在我们实例化一个`Dog`对象，然后用`encoding/json`包里定义的方法对其进行序列化：

```go
d := Dog{Name: "tick", Age: 11}
bz, err := json.Marshal(d)
```

这里得到的序列化结果如下所示：

>string(bz): {"Name":"tick","Age":12}

现在实例化一个指向`Animal`的指针，然后再利用`encoding/json`包里的方法将`bz`里的内容反序列化到刚刚创建的指针所指向的存储结构里：

```go
a := new(Animal)
err := json.Unmarshal(bz, a)
```

执行下来会发现产生了如下错误：

>&json.UnmarshalTypeError{Value:"object", Type:(*reflect.rtype)(0x5eba40), Offset:1, Struct:"", Field:""}

**解决办法**

本包提供了一个`RegisterType(x interface{}, name string)`方法接口，允许我们对自己定义的结构体进行注册：

```go
RegisterType(Dog{}, "Animal/Dog")
```

然后我们实例化一个`Dog`对象，然后用本包里定义的`Encode()`方法对其进行序列化：

```go
d := Dog{Name: "tick", Age: 11}
bz, err := Encode(d)
```

得到的序列化结果与前面不一样：

>{"type":"Animal/Dog","value":{"Name":"tick","Age":12}}

发现在序列化结果里多了`type`信息，就是这个`type`信息帮助我们能在反序列化的时候找到对应的类型，将序列化结果反序列化到对应类型所指向的空间里：

```go
a := new(Animal)
err := Decode(bz, a)
```

输出`a`，得到结果：

>{tick 12}
> 