# Pub和Sub

这个包实现了订阅某个事件，然后接收和特定事件相关的消息的功能。

## 使用案例

1. **初始化并启动服务**

首先我们需要初始化一个服务，一切事件的订阅都是在服务这个地方完成的：

```go
s := pubsub.NewServer(options ...pubsub.Option)
s.Start()
```

`options`参数可以让我们定制服务，例如设置服务中命令通道的大小，一般来讲，不需要特定设置。初始化结束后，就需要调用`Start()`方法来启动服务，这里会
开启一个协程，不停的监听命令通道里是否有新的命令出现，如果有的话就会去执行，一般来讲，这些命令无非就是创建新的订阅，或者取消某个订阅，或者给订阅相关
事件的客户发送消息。

2. **在服务处订阅一个事件**

```go
clientID := "client-xx"
event := query.MustParse("block.height > 10")
subscription, err := s.Subscribe(clientID, event)
```

在上面代码中，我们创建了一个客户端，用字符串表示，然后让该字符串在服务端订阅了一个事件，这个事件的含义是*区块的高度大于10*，也就是说，将来只有服务端发布
与*区块高度大于10*这个事件相关的消息，客户端才能收到这个消息，之后，客户端通过以下代码来时刻监听订阅服务里的新消息：

```go
for {
	select {
	case msg := <-subscription.MsgOut():
		// handle msg
	case <-subscription.CancelledWait():
		// 订阅服务在服务端被取消了
		return
        }
}
```

3. **服务端发布消息**

服务端利用以下代码发布与*区块高度大于10*事件相关的消息：

```go
err := s.PublishWithEvents(msg, map[string][]string{"block.height": {"11"}})
```

上面代码里除了11，任何大于10的小数或者整数都可以，客户端将来能够收到服务端发布的消息，如果上面不是11，而是一个比10小的数字，那么客户端将收不到消息。

4. **取消订阅**

服务端调用以下代码来取消指定客户端在本地的订阅：

```go
clientID := "client-xx"
event := query.MustParse("block.height > 10")
err := s.Unsubscribe(clientID, event)
```

上面代码执行完毕后，客户端那里就会侦听到自己的订阅被取消了：

```go
case <-subscription.CancelledWait():
```

从此，客户端都不会受到和*区块高度大于10*事件相关的消息了。