package p2p

import (
	"encoding/json"
	"fmt"
	mos "github.com/232425wxy/meta--/common/os"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/crypto"
	"os"
	"sync"
	"time"
)

type addrBook struct {
	service.BaseService
	mu             sync.RWMutex
	ourAddresses   map[string]struct{} // id@ip:port -> struct{}{}
	bucket         map[crypto.ID]*knownAddress
	filePath       string
	fileSaveTicker *time.Ticker
}

func NewAddrBook(filePath string) *addrBook {
	return &addrBook{
		BaseService:    *service.NewBaseService(nil, "AddrBook"),
		ourAddresses:   make(map[string]struct{}),
		bucket:         make(map[crypto.ID]*knownAddress),
		filePath:       filePath,
		fileSaveTicker: time.NewTicker(defaultSaveToFileDur),
	}
}

func (a *addrBook) Start() error {
	a.loadFromFile()
	go a.saveRoutine()
	return a.BaseService.Start()
}

func (a *addrBook) Stop() error {
	a.fileSaveTicker.Stop()
	return a.BaseService.Stop()
}

func (a *addrBook) AddOurAddress(addr *NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Logger.Debug("add our address to book", "address", addr.DialString())
	a.ourAddresses[addr.String()] = struct{}{}
}

func (a *addrBook) AddAddress(addr *NetAddress) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.ourAddresses[addr.String()]; ok {
		return false
	}
	a.bucket[addr.ID] = newKnownAddress(addr)
	return true
}

func (a *addrBook) RemoveAddress(addr *NetAddress) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.bucket, addr.ID)
}

func (a *addrBook) MarkAttempt(addr *NetAddress) {
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
func (a *addrBook) saveRoutine() {
	for {
		select {
		case <-a.fileSaveTicker.C:
			a.saveToFile()
		case <-a.WaitStop():
			return
		}
	}
}

func (a *addrBook) saveToFile() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	a.Logger.Debug("saving addresses from book to the disk")
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

func (a *addrBook) loadFromFile() bool {
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
	defaultSaveToFileDur = 2 * time.Minute
)
