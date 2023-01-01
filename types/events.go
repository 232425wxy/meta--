package types

import (
	"fmt"
	"github.com/232425wxy/meta--/common/pubsub"
	"github.com/232425wxy/meta--/common/pubsub/query"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/proto/pbabci"
)

const (
	EventCompleteProposal = "EVENT_COMPLETE_PROPOSAL"
	EventNewBlock         = "EVENT_NEW_BLOCK"
	EventTx               = "EVENT_TX"
	EventValidatorUpdates = "EVENT_VALIDATOR_UPDATES"
)

const (
	EventKey  = "EVENT_KEY"
	TxHashKey = "TX_HASH_KEY"
	HeightKey = "EVENT_HEIGHT"
)

type EventData interface{}

type EventDataNewBlock struct {
	Block            *Block                     `json:"block"`
	ResultBeginBlock *pbabci.ResponseBeginBlock `json:"result_begin_block"`
	ResultEndBlock   *pbabci.ResponseEndBlock   `json:"result_end_block"`
}

type EventDataTx struct {
	Height            int64 `json:"height"`
	Tx                Tx    `json:"tx"`
	ResponseDeliverTx *pbabci.ResponseDeliverTx
}

type EventDataValidatorUpdates struct {
	ValidatorUpdates []*pbabci.ValidatorUpdate
}

type EventBus struct {
	service.BaseService
	server *pubsub.Server
}

func NewEventBus() *EventBus {
	return &EventBus{
		BaseService: *service.NewBaseService(nil, "EventBus"),
		server:      pubsub.NewServer(pubsub.OptionBufferCapacity(0)),
	}
}

func (bus *EventBus) SetLogger(logger log.Logger) {
	bus.BaseService.SetLogger(logger)
	bus.server.SetLogger(logger.New("module", "pubsub"))
}

func (bus *EventBus) Start() error {
	return bus.server.Start()
}

func (bus *EventBus) Stop() error {
	return bus.server.Stop()
}

func (bus *EventBus) NumClients() int {
	return bus.server.NumClients()
}

func (bus *EventBus) NumClientSubscriptions(clientID string) int {
	return bus.server.NumClientSubscriptions(clientID)
}

func (bus *EventBus) Subscribe(subscriber string, q *query.Query, capacity ...int) (*pubsub.Subscription, error) {
	return bus.server.Subscribe(subscriber, q, capacity...)
}

func (bus *EventBus) SubscribeUnbuffered(subscriber string, q *query.Query) (*pubsub.Subscription, error) {
	return bus.server.SubscribeUnbuffered(subscriber, q)
}

func (bus *EventBus) Unsubscribe(subscriber string, q *query.Query) error {
	return bus.server.Unsubscribe(subscriber, q)
}

func (bus *EventBus) UnsubscribeAll(subscriber string) error {
	return bus.server.UnsubscribeAll(subscriber)
}

func (bus *EventBus) Publish(event string, data EventData) error {
	return bus.server.PublishWithEvents(data, map[string][]string{EventKey: {event}})
}

func (bus *EventBus) PublishEventNewBlock(data EventDataNewBlock) error {
	events := map[string][]string{EventKey: {EventNewBlock}}
	return bus.server.PublishWithEvents(data, events)
}

func (bus *EventBus) PublishEventTx(data EventDataTx) error {
	events := map[string][]string{EventKey: {EventTx}, TxHashKey: {fmt.Sprintf("%x", data.Tx.Hash())}, HeightKey: {fmt.Sprintf("%d", data.Height)}}
	return bus.server.PublishWithEvents(data, events)
}

func (bus *EventBus) PublishEventValidatorUpdates(data EventDataValidatorUpdates) error {
	events := map[string][]string{EventKey: {EventValidatorUpdates}}
	return bus.server.PublishWithEvents(data, events)
}
