package p2p

import (
	"encoding/json"
	"fmt"
	mos "github.com/232425wxy/meta--/common/os"
	"github.com/232425wxy/meta--/crypto"
	"os"
	"sync"
	"time"
)

type AddrBook struct {
	mu             sync.RWMutex
	ourAddresses   map[string]struct{} // id@ip:port -> struct{}{}
	bucket         map[crypto.ID]*knownAddress
	filePath       string
	fileSaveTicker *time.Ticker
	quit           chan struct{}
}

func NewAddrBook(filePath string) *AddrBook {
	return &AddrBook{
		ourAddresses:   make(map[string]struct{}),
		bucket:         make(map[crypto.ID]*knownAddress),
		filePath:       filePath,
		fileSaveTicker: time.NewTicker(defaultSaveToFileDur),
		quit:           make(chan struct{}),
	}
}

func (a *AddrBook) Start() {
	a.loadFromFile()
	go a.saveRoutine()
}

func (a *AddrBook) Close() {
	a.fileSaveTicker.Stop()
	close(a.quit)
}

func (a *AddrBook) AddOurAddress(addr *NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ourAddresses[addr.String()] = struct{}{}
}

func (a *AddrBook) IsOurAddress(addr *NetAddress) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	_, ok := a.ourAddresses[addr.String()]
	return ok
}

func (a *AddrBook) AddAddress(addr *NetAddress) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.ourAddresses[addr.String()]; ok {
		return false
	}
	if _, ok := a.bucket[addr.ID]; ok {
		return true
	}
	a.bucket[addr.ID] = newKnownAddress(addr)
	return true
}

func (a *AddrBook) RemoveAddress(addr *NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.bucket, addr.ID)
}

func (a *AddrBook) MarkAttempt(addr *NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()
	ka, ok := a.bucket[addr.ID]
	if !ok {
		return
	}
	ka.LastAttempt = time.Now()
	ka.Attempts++
}

// saveRoutine ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// saveRoutine 默认情况下，每隔两分钟将地址簿里的所有地址存储到硬盘中。
func (a *AddrBook) saveRoutine() {
	for {
		select {
		case <-a.fileSaveTicker.C:
			a.saveToFile()
		case <-a.quit:
			return
		}
	}
}

func (a *AddrBook) saveToFile() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	aJSON := &addrBookJSON{Addresses: make([]*knownAddress, 0)}
	for _, ka := range a.bucket {
		aJSON.Addresses = append(aJSON.Addresses, ka)
	}
	bz, err := json.MarshalIndent(aJSON, "", "\t")
	if err != nil {
		panic(fmt.Sprintf("failed to encode address book: %q", err))
	}
	mos.MustWriteFile(a.filePath, bz, 0644)
}

func (a *AddrBook) loadFromFile() bool {
	_, err := os.Stat(a.filePath)
	if os.IsNotExist(err) {
		return false
	}
	r, err := os.Open(a.filePath)
	if err != nil {
		panic(fmt.Sprintf("failed to open address book: %q", err))
	}
	defer func() {
		_ = r.Close()
	}()
	aJSON := &addrBookJSON{}
	decoder := json.NewDecoder(r)
	err = decoder.Decode(aJSON)
	if err != nil {
		panic(fmt.Sprintf("failed to decode content in the address book: %q", err))
	}
	for _, ka := range aJSON.Addresses {
		a.bucket[ka.NodeID()] = ka
	}
	return true
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量

type addrBookJSON struct {
	Addresses []*knownAddress `json:"addresses"`
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 地址簿里已知的节点网络地址

type knownAddress struct {
	Addr        *NetAddress `json:"addr"`
	Attempts    int         `json:"attempts"`
	LastAttempt time.Time   `json:"lastAttempt"`
}

func newKnownAddress(addr *NetAddress) *knownAddress {
	return &knownAddress{
		Addr:     addr,
		Attempts: 0,
	}
}

func (ka *knownAddress) NodeID() crypto.ID {
	return ka.Addr.ID
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 常量

const (
	defaultSaveToFileDur = 2 * time.Second
)
