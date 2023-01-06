package state

import (
	"errors"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/proto/pbabci"
	"github.com/232425wxy/meta--/proxy"
	"github.com/232425wxy/meta--/txspool"
	"github.com/232425wxy/meta--/types"
)

type BlockExecutor struct {
	store          *StoreState
	proxyConsensus *proxy.AppConnConsensus
	txsPool        *txspool.TxsPool
	eventBus       *types.EventBus
	cfg            *config.TxsPoolConfig
	logger         log.Logger
}

func NewBlockExecutor(store *StoreState, consensus *proxy.AppConnConsensus, txsPool *txspool.TxsPool, logger log.Logger) *BlockExecutor {
	return &BlockExecutor{
		store:          store,
		proxyConsensus: consensus,
		txsPool:        txsPool,
		logger:         logger,
	}
}

func (be *BlockExecutor) SetEventBUs(bus *types.EventBus) {
	be.eventBus = bus
}

func (be *BlockExecutor) CreateProposalBlock(height int64, state *State, proposer crypto.ID, lastBlockHash []byte) *types.Block {
	txs := be.txsPool.ReapMaxBytes(be.cfg.MaxTxBytes * be.cfg.MaxSize)
	return state.MakeBlock(height, txs, proposer, lastBlockHash)
}

func (be *BlockExecutor) ApplyBlock(state *State, block *types.Block) (*State, error) {
	responses, err := execBlockOnProxyConsensus(be.proxyConsensus, block, be.logger)
	if err != nil {
		return state, err
	}
	be.txsPool.Lock()
	defer be.txsPool.Unlock()
	// TODO 这里图方便直接将区块里的交易数据从交易池里删除了
	be.txsPool.Update(block.Header.Height, block.Body.Txs)
	if err = be.store.SaveState(state); err != nil {
		return state, err
	}
	if err = be.eventBus.PublishEventNewBlock(types.EventDataNewBlock{
		Block:            block,
		ResultBeginBlock: responses.BeginBlock,
		ResultEndBlock:   responses.EndBlock,
	}); err != nil {
		be.logger.Error("failed to publish new block", "err", err)
	}
	for i, tx := range block.Body.Txs {
		if err = be.eventBus.PublishEventTx(types.EventDataTx{
			Height:            block.Header.Height,
			Tx:                tx,
			ResponseDeliverTx: responses.DeliverTxs[i],
		}); err != nil {
			be.logger.Error("failed to publish event TX", "err", err)
		}
	}
	validatorUpdates := responses.EndBlock.ValidatorUpdates
	if len(validatorUpdates) > 0 {
		be.logger.Info("update validators", "num_validators_update", len(validatorUpdates))
		if err = be.eventBus.PublishEventValidatorUpdates(types.EventDataValidatorUpdates{ValidatorUpdates: validatorUpdates}); err != nil {
			be.logger.Error("failed to publish event validator updates", "err", err)
		}
	}
	updateState(state, validatorUpdates, block)
	return state, nil
}

func execBlockOnProxyConsensus(proxyConsensus *proxy.AppConnConsensus, block *types.Block, logger log.Logger) (*pbabci.ABCIResponses, error) {
	var validTxs, invalidTxs = 0, 0
	responses := new(pbabci.ABCIResponses)
	responses.DeliverTxs = make([]*pbabci.ResponseDeliverTx, len(block.Body.Txs))

	pbHeader := block.Header.ToProto()
	if pbHeader == nil {
		return nil, errors.New("empty block header")
	}
	beginBlock := proxyConsensus.BeginBlock(pbabci.RequestBeginBlock{
		Evidences: nil,
		Height:    block.Header.Height,
	})
	responses.BeginBlock = &beginBlock
	for i, tx := range block.Body.Txs {
		res := proxyConsensus.DeliverTx(pbabci.RequestDeliverTx{Tx: tx})
		responses.DeliverTxs[i] = &res
		if !res.OK {
			invalidTxs++
		} else {
			validTxs++
		}
	}
	endBlock := proxyConsensus.EndBlock(pbabci.RequestEndBlock{Height: block.Header.Height})
	responses.EndBlock = &endBlock
	logger.Info("executed block", "height", block.Header.Height, "num_valid_txs", validTxs, "num_invalid_txs", invalidTxs)
	return responses, nil
}

func updateState(state *State, validatorUpdates []*pbabci.ValidatorUpdate, block *types.Block) {
	if len(validatorUpdates) > 0 {
		state.Validators.Update(validatorUpdates)
	}
	state.PreviousBlock = block
	state.LastBlockHeight = block.Header.Height
	state.LastBlockTime = block.Header.Timestamp
}
