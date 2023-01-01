package pubsub

import (
	"fmt"
	"github.com/232425wxy/meta--/common/pubsub/query"
	"github.com/232425wxy/meta--/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func newServer(t *testing.T, capacity ...int) *Server {
	cap := 0
	if len(capacity) > 0 {
		cap = capacity[0]
	}
	s := NewServer(OptionBufferCapacity(cap))
	logger := log.New("server", "pubsub")
	logger.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	s.SetLogger(logger)
	err := s.Start()
	assert.Nil(t, err)
	return s
}

func outputReceived(clientID string, subscription *Subscription, received chan struct{}) {
	for {
		select {
		case msg := <-subscription.MsgOut():
			fmt.Printf("client %q received message %q for events %q\n", clientID, msg.Data(), msg.events)
			close(received)
			return
		case <-subscription.CancelledWait():
			close(received)
			fmt.Println(subscription.reason)
			return
		}
	}
}

func publish(s *Server, msg interface{}, t *testing.T, published chan struct{}, events ...map[string][]string) {
	if len(events) > 0 {
		err := s.PublishWithEvents(msg, events[0])
		assert.Nil(t, err)
		s.Logger.Info(fmt.Sprintf("publish msg %q", msg), "events", events[0])
		close(published)
	} else {
		err := s.Publish(msg)
		assert.Nil(t, err)
		s.Logger.Info(fmt.Sprintf("publish msg %q", msg))
		close(published)
	}
}

func TestSubscribeMatches(t *testing.T) {
	s := newServer(t)
	defer func() {
		err := s.Stop()
		if err != nil {
			t.Error(err)
		}
	}()

	// 订阅
	q := query.MustParse("block.height > 10")
	clientID := "client-01"
	subscription, err := s.Subscribe(clientID, q)
	assert.Nil(t, err)

	// 客户端等待消息
	received := make(chan struct{})
	go outputReceived(clientID, subscription, received)

	// 服务端发送消息
	published := make(chan struct{})
	go publish(s, "hello", t, published, map[string][]string{"block.height": {"11"}})

	<-published
	<-received
}

func TestSubscribeNotMatches(t *testing.T) {
	s := newServer(t)
	defer func() {
		err := s.Stop()
		if err != nil {
			t.Error(err)
		}
	}()

	// 订阅
	q := query.MustParse("block.height > 10")
	clientID := "client-01"
	subscription, err := s.Subscribe(clientID, q)
	assert.Nil(t, err)

	// 客户端等待消息
	received := make(chan struct{})
	go outputReceived(clientID, subscription, received)

	// 服务端发送消息
	published := make(chan struct{})
	go publish(s, "hello", t, published, map[string][]string{"block.height": {"10"}})
	published2 := make(chan struct{})
	go publish(s, "hello", t, published2)

	<-published
	<-published2
	select {
	case <-received:
		t.Error("client shouldn't receive any message")
	case <-time.After(time.Second * 3):
		t.Log("successful")
	}
}

func TestManyClientsSubscribe(t *testing.T) {
	s := newServer(t, 10)
	defer func() {
		err := s.Stop()
		if err != nil {
			t.Error(err)
		}
	}()

	clientID1 := "client-01"
	clientID2 := "client-02"
	clientID3 := "client-03"
	q1 := query.MustParse("clientID1.power > 10.0")
	q2 := query.MustParse("clientID2.power < 20.0")
	q3 := query.MustParse("clientID3.power > 30.0")

	subscription1, err := s.Subscribe(clientID1, q1, 10)
	assert.Nil(t, err)
	subscription2, err := s.Subscribe(clientID2, q2, 20)
	assert.Nil(t, err)
	subscription3, err := s.Subscribe(clientID3, q3, 30)
	assert.Nil(t, err)

	received1 := make(chan struct{})
	received2 := make(chan struct{})
	received3 := make(chan struct{})

	published1 := make(chan struct{})
	published2 := make(chan struct{})
	published3 := make(chan struct{})

	go publish(s, "client-01 can propose proposal", t, published1)
	go publish(s, "client-02 cannot propose proposal", t, published2, map[string][]string{"clientID2.power": {"19"}})
	go publish(s, "client-03 can propose proposal", t, published3)

	go outputReceived(clientID1, subscription1, received1)
	go outputReceived(clientID2, subscription2, received2)
	go outputReceived(clientID3, subscription3, received3)

	<-published1
	<-published2
	<-published3

	select {
	case <-received1:
		t.Errorf("%q shouldn't receive message", clientID1)
	//case <-received2:
	//	t.Errorf("%q shouldn't receive message", clientID2)
	case <-received3:
		t.Errorf("%q shouldn't receive message", clientID3)
	case <-time.After(time.Second * 3):

	}
	<-received2
}

func TestUnsubscribe(t *testing.T) {
	s := newServer(t)
	defer func() {
		err := s.Stop()
		if err != nil {
			t.Error(err)
		}
	}()

	clientID := "client-01"
	q := query.MustParse("client-01.power > 12.0")

	subscription, err := s.Subscribe(clientID, q)
	assert.Nil(t, err)

	received := make(chan struct{})
	published := make(chan struct{})
	closed := make(chan struct{})

	go publish(s, "client-01 can propose proposal", t, published)
	go outputReceived(clientID, subscription, received)

	<-published

	go func() {
		for {
			select {
			case <-subscription.CancelledWait():
				t.Log(subscription.Reason())
				close(closed)
				return
			}
		}
	}()

	select {
	case <-received:
		t.Errorf("client %q shouldn't receive message", clientID)
	case <-time.After(time.Second * 3):
		err = s.Unsubscribe(clientID, q)
		assert.Nil(t, err)
	}
	<-closed
}

func TestUnsubscribeNum(t *testing.T) {
	s := newServer(t)
	defer func() {
		err := s.Stop()
		if err != nil {
			t.Error(err)
		}
	}()

	clientID1 := "client-01"
	clientID2 := "client-02"
	q11 := query.MustParse("client-01.power > 12.0")
	q12 := query.MustParse("block.height > 3")
	q13 := query.MustParse("consensus.quorum > 66.7")
	q21 := query.MustParse("client-02.power > 12.0")
	subscription11, err := s.Subscribe(clientID1, q11)
	assert.Nil(t, err)
	subscription12, err := s.Subscribe(clientID1, q12)
	assert.Nil(t, err)
	subscription13, err := s.Subscribe(clientID1, q13)
	assert.Nil(t, err)
	subscription21, err := s.Subscribe(clientID2, q21)
	assert.Nil(t, err)
	subscription22, err := s.Subscribe(clientID2, q12)
	assert.Nil(t, err)

	assert.Equal(t, 3, s.NumClientSubscriptions(clientID1))
	assert.Equal(t, 2, s.NumClientSubscriptions(clientID2))
	assert.Equal(t, 2, s.NumClients())

	closed11 := make(chan struct{})
	go outputReceived(clientID1, subscription11, closed11)
	closed12 := make(chan struct{})
	go outputReceived(clientID1, subscription12, closed12)
	closed13 := make(chan struct{})
	go outputReceived(clientID1, subscription13, closed13)
	received21 := make(chan struct{})
	go outputReceived(clientID2, subscription21, received21)
	closed21 := make(chan struct{})
	go outputReceived(clientID2, subscription21, closed21)
	closed22 := make(chan struct{})
	go outputReceived(clientID2, subscription22, closed22)

	err = s.Unsubscribe(clientID1, q11)
	assert.Nil(t, err)
	assert.Equal(t, 2, s.NumClients())
	assert.Equal(t, 2, s.NumClientSubscriptions(clientID1))

	err = s.Unsubscribe(clientID1, q12)
	assert.Nil(t, err)
	assert.Equal(t, 2, s.NumClients())
	assert.Equal(t, 1, s.NumClientSubscriptions(clientID1))

	err = s.Unsubscribe(clientID1, q13)
	assert.Nil(t, err)
	assert.Equal(t, 1, s.NumClients())
	assert.Equal(t, 0, s.NumClientSubscriptions(clientID1))

	err = s.PublishWithEvents("you can propose a new block", map[string][]string{"client-02.power": {"12.01"}})
	assert.Nil(t, err)
	assert.Equal(t, 1, s.NumClients())
	assert.Equal(t, 2, s.NumClientSubscriptions(clientID2))

	err = s.Unsubscribe(clientID2, q21)
	assert.Equal(t, 1, s.NumClients())
	assert.Equal(t, 1, s.NumClientSubscriptions(clientID2))

	err = s.PublishWithEvents("you can propose a new block", map[string][]string{"client-02.power": {"12.01"}})
	assert.Nil(t, err)
	assert.Equal(t, 1, s.NumClients())
	assert.Equal(t, 1, s.NumClientSubscriptions(clientID2))

	err = s.Unsubscribe(clientID2, q12)
	assert.Equal(t, 0, s.NumClients())
	assert.Equal(t, 0, s.NumClientSubscriptions(clientID2))

	<-closed11
	<-closed12
	<-closed13
	<-received21
	<-closed21
	<-closed22
}

func BenchmarkSubscribeMany(b *testing.B) {
	s := NewServer(OptionBufferCapacity(0))
	err := s.Start()
	assert.Nil(b, err)
	defer func() {
		err = s.Stop()
		if err != nil {
			b.Error(err)
		}
	}()

	for i := 0; i < 100; i++ {
		client := fmt.Sprintf("client-%d", i)
		q := query.MustParse(fmt.Sprintf("%s.power > %d", client, i+10))
		subscription, err := s.Subscribe(client, q)
		assert.Nil(b, err)
		go func() {
			for {
				select {
				case <-subscription.MsgOut():
					//b.Log(msg.data, msg.events)
					continue
				case <-subscription.CancelledWait():
					return
				}
			}
		}()
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := i % 100
		err = s.PublishWithEvents("message", map[string][]string{fmt.Sprintf("client-%d.power", id): {fmt.Sprintf("%d", i)}})
		assert.Nil(b, err)
	}
}
