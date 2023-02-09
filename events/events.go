package events

import (
	"fmt"
	"github.com/232425wxy/meta--/common/pubsub"
	"github.com/232425wxy/meta--/common/pubsub/query"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/proto/pbabci"
	"github.com/232425wxy/meta--/proto/pbevents"
	"github.com/232425wxy/meta--/types"
)

const (
	EventNewBlock = "EVENT_NEW_BLOCK"
	EventTx       = "EVENT_TX"
)

const (
	EventKey  = "EVENT_KEY"
	TxHashKey = "TX_HASH_KEY"
	HeightKey = "EVENT_HEIGHT"
)

type EventData interface{}

type EventDataNewBlock struct {
	Block            *types.Block               `json:"block"`
	ResultBeginBlock *pbabci.ResponseBeginBlock `json:"result_begin_block"`
	ResultEndBlock   *pbabci.ResponseEndBlock   `json:"result_end_block"`
}

type EventDataTx struct {
	Height            int64    `json:"height"`
	Tx                types.Tx `json:"tx"`
	ResponseDeliverTx *pbabci.ResponseDeliverTx
}

type EventDataNewStep struct {
	Height int64 `json:"height"`
	Round  int16 `json:"round"`
	Step   int8  `json:"step"`
}

func (e *EventDataNewStep) ValidateBasic() error {
	return nil
}

func (e *EventDataNewStep) ToProto() *pbevents.EventDataNewStep {
	if e == nil {
		return nil
	}
	return &pbevents.EventDataNewStep{
		Height: e.Height,
		Round:  int32(e.Round),
		Step:   int32(e.Step),
	}
}

func EventDataNewStepFromProto(pb *pbevents.EventDataNewStep) *EventDataNewStep {
	if pb == nil {
		return nil
	}
	return &EventDataNewStep{
		Height: pb.Height,
		Round:  int16(pb.Round),
		Step:   int8(pb.Step),
	}
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

func CreateAndStartEventBus(logger log.Logger) (*EventBus, error) {
	bus := NewEventBus()
	bus.SetLogger(logger)
	if err := bus.Start(); err != nil {
		return nil, err
	}
	return bus, nil
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

func (bus *EventBus) PublishEventNewBlock(data EventDataNewBlock) error {
	events := map[string][]string{EventKey: {EventNewBlock}}
	return bus.server.PublishWithEvents(data, events)
}

func (bus *EventBus) PublishEventTx(data EventDataTx) error {
	events := map[string][]string{EventKey: {EventTx}, TxHashKey: {fmt.Sprintf("%x", data.Tx.Hash())}, HeightKey: {fmt.Sprintf("%d", data.Height)}}
	return bus.server.PublishWithEvents(data, events)
}
