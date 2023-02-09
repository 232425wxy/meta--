package events

import (
	"fmt"
	"github.com/232425wxy/meta--/proto/pbevents"
	"github.com/cosmos/gogoproto/proto"
)

type Message interface {
	ValidateBasic() error
}

func MustEncode(msg Message) []byte {
	if msg == nil {
		panic("message is nil")
	}
	var pb *pbevents.Event
	switch message := msg.(type) {
	case *EventDataNewStep:
		pb = &pbevents.Event{
			Data: &pbevents.Event_EventDataNewStep{EventDataNewStep: message.ToProto()},
		}
	default:
		panic(fmt.Sprintf("unknown message type: %T", message))
	}
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return bz
}

func MustDecode(bz []byte) (msg Message) {
	pb := &pbevents.Event{}
	var err error
	if err = proto.Unmarshal(bz, pb); err != nil {
		panic(err)
	}
	switch m := pb.Data.(type) {
	case *pbevents.Event_EventDataNewStep:
		msg = EventDataNewStepFromProto(m.EventDataNewStep)
	default:
		panic(fmt.Sprintf("unkonwn message type: %T", m))
	}
	return msg
}
