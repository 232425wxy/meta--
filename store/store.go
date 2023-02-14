package store

import (
	"fmt"
	"github.com/232425wxy/meta--/database"
	"github.com/232425wxy/meta--/proto/pbstate"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"github.com/232425wxy/meta--/types"
	"github.com/cosmos/gogoproto/proto"
	"sync"
)

var StoreBlockKey = []byte("meta--/store-block")

/**********************************************************************************************************************/

type BlockStore struct {
	db     database.DB
	mu     sync.RWMutex
	height int64
}

func NewStoreBlock(db database.DB) *BlockStore {
	sb := new(BlockStore)
	bz, err := db.Get(StoreBlockKey)
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		sb.height = 0
		sb.db = db
		return sb
	}
	pb := new(pbstate.StoreBlock)
	if err = proto.Unmarshal(bz, pb); err != nil {
		panic(err)
	}
	return &BlockStore{db: db, height: pb.Height}
}

// Height
//
// Height 反映当前区块链的高度和区块数量。
func (sb *BlockStore) Height() int64 {
	sb.mu.RLock()
	sb.mu.RUnlock()
	return sb.height
}

func (sb *BlockStore) LoadBlockByHeight(height int64) *types.Block {
	pb := &pbtypes.Block{}
	bz, err := sb.db.Get(calcBlockHeightKey(height))
	if err != nil {
		return nil
	}
	if err = proto.Unmarshal(bz, pb); err != nil {
		return nil
	}
	block := &types.Block{
		Header:        types.HeaderFromProto(pb.Header),
		Body:          types.DataFromProto(pb.Body),
		ChameleonHash: types.ChameleonHashFromProto(pb.ChameleonHash),
	}
	return block
}

func (sb *BlockStore) LoadBlockByHash(hash []byte) *types.Block {
	pb := &pbtypes.BlockHeight{}
	bz, err := sb.db.Get(calcBlockHashKey(hash))
	if err != nil {
		panic(err)
	}
	if err = proto.Unmarshal(bz, pb); err != nil {
		panic(err)
	}
	return sb.LoadBlockByHeight(pb.Height)
}

func (sb *BlockStore) SaveBlock(block *types.Block) {
	if block == nil {
		panic("cannot save nil block")
	}
	pb := block.ToProto()
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	bh := &pbtypes.BlockHeight{Height: block.Header.Height}
	bzh, err := proto.Marshal(bh)
	if err != nil {
		panic(err)
	}
	if err = sb.db.SetSync(calcBlockHashKey(block.ChameleonHash.Hash), bzh); err != nil {
		panic(err)
	}
	if err = sb.db.SetSync(calcBlockHeightKey(block.Header.Height), bz); err != nil {
		panic(err)
	}
}

func calcBlockHeightKey(height int64) []byte {
	return append([]byte("block-height:"), fmt.Sprintf("%d", height)...)
}

func calcBlockHashKey(hash []byte) []byte {
	return append([]byte("block-hash:"), hash...)
}

func (sb *BlockStore) DB() database.DB {
	return sb.db
}
