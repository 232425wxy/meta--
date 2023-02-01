package syncer

import (
	"fmt"
	"github.com/232425wxy/meta--/proto/pbsyncer"
	"github.com/cosmos/gogoproto/proto"
)

func EncodeMsg(pb proto.Message) ([]byte, error) {
	msg := &pbsyncer.Message{}

	switch pb := pb.(type) {
	case *pbsyncer.BlockRequest:
		msg.Sum = &pbsyncer.Message_BlockRequest{BlockRequest: pb}
	case *pbsyncer.BlockResponse:
		msg.Sum = &pbsyncer.Message_BlockResponse{BlockResponse: pb}
	case *pbsyncer.NoBlockResponse:
		msg.Sum = &pbsyncer.Message_NoBlockResponse{NoBlockResponse: pb}
	case *pbsyncer.StatusRequest:
		msg.Sum = &pbsyncer.Message_StatusRequest{StatusRequest: pb}
	case *pbsyncer.StatusResponse:
		msg.Sum = &pbsyncer.Message_StatusResponse{StatusResponse: pb}
	default:
		return nil, fmt.Errorf("unknown message type: %T", pb)
	}
	bz, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func DecodeMsg(bz []byte) (proto.Message, error) {
	pb := &pbsyncer.Message{}
	err := proto.Unmarshal(bz, pb)
	if err != nil {
		return nil, err
	}

	switch msg := pb.Sum.(type) {
	case *pbsyncer.Message_BlockRequest:
		return msg.BlockRequest, nil
	case *pbsyncer.Message_BlockResponse:
		return msg.BlockResponse, nil
	case *pbsyncer.Message_NoBlockResponse:
		return msg.NoBlockResponse, nil
	case *pbsyncer.Message_StatusRequest:
		return msg.StatusRequest, nil
	case *pbsyncer.Message_StatusResponse:
		return msg.StatusResponse, nil
	default:
		return nil, fmt.Errorf("unknown message type: %T", msg)
	}
}
