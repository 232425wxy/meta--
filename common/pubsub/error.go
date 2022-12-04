package pubsub

import (
	"fmt"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 实例化错误用的项目级全局函数

func NewErrAlreadySubscribed(clientID, query string) ErrAlreadySubscribed {
	return ErrAlreadySubscribed{clientID: clientID, query: query}
}

func NewErrSubscriptionNotFound(clientID, query string) ErrSubscriptionNotFound {
	return ErrSubscriptionNotFound{clientID: clientID, query: query}
}

func NewErrOutOfCapacity(clientID string) ErrOutOfCapacity {
	return ErrOutOfCapacity{clientID: clientID}
}

func NewErrUnsubscribed(clientID string, query string) ErrUnSubscribed {
	return ErrUnSubscribed{clientID: clientID, query: query}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义一堆错误

type ErrAlreadySubscribed struct {
	clientID string
	query    string
}

func (err ErrAlreadySubscribed) Error() string {
	return fmt.Sprintf("pubsub: client %q has already subscribed %q", err.clientID, err.query)
}

type ErrSubscriptionNotFound struct {
	clientID string
	query    string
}

func (err ErrSubscriptionNotFound) Error() string {
	return fmt.Sprintf("pubsub: cannot find subscription %q for client %q", err.query, err.clientID)
}

type ErrOutOfCapacity struct {
	clientID string
}

func (err ErrOutOfCapacity) Error() string {
	return fmt.Sprintf("pubsub: client %q is too slow to pull message from subscription' message channel", err.clientID)
}

type ErrUnSubscribed struct {
	clientID string
	query    string
}

func (err ErrUnSubscribed) Error() string {
	return fmt.Sprintf("pubsub: client %q unsubscribed query %q", err.clientID, err.query)
}
