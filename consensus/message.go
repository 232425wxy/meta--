package consensus

import (
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"github.com/232425wxy/meta--/types"
	"github.com/cosmos/gogoproto/proto"
)

type Message interface {
	ValidateBasic() error
}

type MessageInfo struct {
	Msg    Message   `json:"msg"`
	NodeID crypto.ID `json:"node_id"`
}

func MustEncode(msg Message) []byte {
	if msg == nil {
		panic("consensus: message is nil")
	}
	pb := &pbtypes.Message{}
	switch message := msg.(type) {
	case *types.NextView:
		pb = &pbtypes.Message{
			Msg: &pbtypes.Message_NextView{
				NextView: message.ToProto(),
			},
		}
	case *types.Prepare:
		pb = &pbtypes.Message{
			Msg: &pbtypes.Message_Prepare{
				Prepare: message.ToProto(),
			},
		}
	case *types.PrepareVote:
		pb = &pbtypes.Message{
			Msg: &pbtypes.Message_PrepareVote{
				PrepareVote: &pbtypes.PrepareVote{
					Vote: message.Vote.ToProto(),
				},
			},
		}
	case *types.PreCommit:
		pb = &pbtypes.Message{
			Msg: &pbtypes.Message_PreCommit{
				PreCommit: message.ToProto(),
			},
		}
	case *types.PreCommitVote:
		pb = &pbtypes.Message{
			Msg: &pbtypes.Message_PreCommitVote{
				PreCommitVote: &pbtypes.PreCommitVote{
					Vote: message.Vote.ToProto(),
				},
			},
		}
	case *types.Commit:
		pb = &pbtypes.Message{
			Msg: &pbtypes.Message_Commit{
				Commit: message.ToProto(),
			},
		}
	case *types.CommitVote:
		pb = &pbtypes.Message{
			Msg: &pbtypes.Message_CommitVote{
				CommitVote: &pbtypes.CommitVote{
					Vote: message.Vote.ToProto(),
				},
			},
		}
	case *types.Decide:
		pb = &pbtypes.Message{
			Msg: &pbtypes.Message_Decide{
				Decide: message.ToProto(),
			},
		}
	default:
		panic(fmt.Sprintf("unknown message type: %T", msg))
	}
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return bz
}

func MustDecode(bz []byte) (msg Message) {
	pb := &pbtypes.Message{}
	var err error
	if err = proto.Unmarshal(bz, pb); err != nil {
		panic(err)
	}
	switch m := pb.Msg.(type) {
	case *pbtypes.Message_NextView:
		msg = types.NextViewFromProto(m.NextView)
	case *pbtypes.Message_Prepare:
		msg = types.PrepareFromProto(m.Prepare)
	case *pbtypes.Message_PrepareVote:
		msg = types.PrepareVoteFromProto(m.PrepareVote)
	case *pbtypes.Message_PreCommit:
		msg = types.PreCommitFromProto(m.PreCommit)
	case *pbtypes.Message_PreCommitVote:
		msg = types.PreCommitVoteFromProto(m.PreCommitVote)
	case *pbtypes.Message_Commit:
		msg = types.CommitFromProto(m.Commit)
	case *pbtypes.Message_CommitVote:
		msg = types.CommitVoteFromProto(m.CommitVote)
	case *pbtypes.Message_Decide:
		msg = types.DecideFromProto(m.Decide)
	default:
		panic(fmt.Sprintf("unknown message type: %T", pb.Msg))
	}
	return msg
}
