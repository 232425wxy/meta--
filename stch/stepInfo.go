package stch

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/types"
)

const defaultSize = 10

type Rss struct {
	id  crypto.ID
	rss *ReplicaSchnorrSig
}

type stepInfo struct {
	redactMission       map[string]*Task
	leaderRedact        map[crypto.ID]*LeaderSchnorrSig
	replicaRedacts      map[crypto.ID]*ReplicaSchnorrSig
	randomVerifications map[crypto.ID]*RandomVerification
	redactBlock         *types.Block
	rssChan             chan *Rss
	randomChan          chan *RandomVerification
	mu                  sync.Mutex
}

func newStepInfo() *stepInfo {
	return &stepInfo{
		redactMission:       make(map[string]*Task),
		leaderRedact:        make(map[crypto.ID]*LeaderSchnorrSig),
		replicaRedacts:      make(map[crypto.ID]*ReplicaSchnorrSig),
		randomVerifications: make(map[crypto.ID]*RandomVerification),
		rssChan:             make(chan *Rss, defaultSize),
		randomChan:          make(chan *RandomVerification, 1),
		redactBlock:         nil,
	}
}

func (si *stepInfo) reset() {
	si.mu.Lock()
	defer si.mu.Unlock()
	for id := range si.redactMission {
		delete(si.redactMission, id)
	}
	for id := range si.leaderRedact {
		delete(si.leaderRedact, id)
	}
	for id := range si.replicaRedacts {
		delete(si.replicaRedacts, id)
	}
	for id := range si.randomVerifications {
		delete(si.randomVerifications, id)
	}
	si.rssChan = nil
	si.rssChan = make(chan *Rss, defaultSize)
	si.redactBlock = nil
}

func (si *stepInfo) addLeaderRedact(peerID crypto.ID, lss *LeaderSchnorrSig, n int) (bool, error) {
	si.mu.Lock()
	defer si.mu.Unlock()

	redactName := redactHash(lss.BlockHeight, lss.TxIndex, lss.NewTx)
	if len(si.redactMission) > 0 {
		for redact, task := range si.redactMission {
			return false, fmt.Errorf("last redact %s mission is not finished, task: %v", redact, *task)
		}
	}
	kvs := bytes.Split(lss.NewTx, []byte("="))
	key, _ := hex.DecodeString(string(kvs[0]))
	value, _ := hex.DecodeString(string(kvs[1]))
	si.redactMission[redactName] = &Task{
		BlockHeight: lss.BlockHeight,
		TxIndex:     lss.TxIndex,
		Key:         key,
		Value:       value,
	}

	if si.leaderRedact[peerID] != nil {
		return false, fmt.Errorf("leader %s already has already sent a redact mission to me", peerID)
	}
	if len(si.leaderRedact) > 0 {
		for id, redact := range si.leaderRedact {
			return false, fmt.Errorf("other leader %s has already sent another redact mission to me, mission is (height: %d, txIndex:%d, newTx: %x)", id, redact.BlockHeight, redact.TxIndex, redact.NewTx)
		}
	}
	si.leaderRedact[peerID] = lss
	if (len(si.replicaRedacts) + len(si.leaderRedact)) == n {
		return true, nil
	} else {
		return false, nil
	}
}

func (si *stepInfo) removeLeaderRedact(peerID crypto.ID) {
	delete(si.leaderRedact, peerID)
}

func (si *stepInfo) addReplicaRedact(peerID crypto.ID, sig *ReplicaSchnorrSig, n int) (bool, error) {
	si.mu.Lock()
	defer si.mu.Unlock()

	redactName := redactHash(sig.BlockHeight, sig.TxIndex, sig.NewTx)
	if len(si.redactMission) > 0 && si.redactMission[redactName] == nil {
		return false, fmt.Errorf("hasn't receive redact mission from leader, please wait for a moment")
	}

	isExist := si.replicaRedacts[peerID]
	if isExist != nil && sig.BlockHeight == isExist.BlockHeight && sig.TxIndex == isExist.TxIndex && bytes.Equal(sig.NewTx, isExist.NewTx) {
		return false, fmt.Errorf("replica %s has already sent a segment of threshold key to me", peerID)
	} else if isExist != nil {
		return false, fmt.Errorf("replica %s has already sent a segment of threshold key about different redact mission to me", peerID)
	}
	si.replicaRedacts[peerID] = sig
	if (len(si.replicaRedacts) + len(si.leaderRedact)) == n {
		return true, nil
	} else {
		return false, nil
	}
}

func (si *stepInfo) removeReplicaRedact(peerID crypto.ID) {
	delete(si.replicaRedacts, peerID)
}

func (si *stepInfo) addRandomVerification(peerID crypto.ID, rv *RandomVerification, n int) (bool, error) {
	si.mu.Lock()
	defer si.mu.Unlock()

	if len(si.redactMission) > 0 && si.redactMission[rv.RedactName] == nil {
		return false, fmt.Errorf("doesn't have the specified redact mission: %s", rv.RedactName)
	}

	isExist := si.randomVerifications[peerID]

	if isExist != nil && isExist.GSigmaExpSK.Cmp(rv.GSigmaExpSK) == 0 && isExist.R2.Cmp(rv.R2) == 0 {
		return false, fmt.Errorf("peer %s has already sent information to me to verify randomness", peerID)
	} else if isExist != nil {
		return false, fmt.Errorf("peer %s has already sent information to me to verify different randomness", peerID)
	}

	si.randomVerifications[peerID] = rv

	if len(si.randomVerifications) == n {
		return true, nil
	} else {
		return false, nil
	}
}

func (si *stepInfo) removeRandomVerification(peerID crypto.ID) {
	delete(si.randomVerifications, peerID)
}

func (si *stepInfo) hasRedactMission(redactName string) bool {
	si.mu.Lock()
	defer si.mu.Unlock()
	return si.redactMission[redactName] != nil
}
