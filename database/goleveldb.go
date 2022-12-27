package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"path/filepath"
)

type GoLevelDB struct {
	db *leveldb.DB
}

func (g *GoLevelDB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errKeyEmpty
	}
	res, err := g.db.Get(key, nil)
	if err != nil {
		if err == errors.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return res, nil
}

func (g *GoLevelDB) Has(key []byte) (bool, error) {
	val, err := g.Get(key)
	if err != nil {
		return false, err
	}
	return val != nil, nil
}

func (g *GoLevelDB) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if len(value) == 0 {
		return errValueEmpty
	}
	if err := g.db.Put(key, value, nil); err != nil {
		return err
	}
	return nil
}

func (g *GoLevelDB) SetSync(key []byte, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if len(value) == 0 {
		return errValueEmpty
	}
	if err := g.db.Put(key, value, &opt.WriteOptions{Sync: true}); err != nil {
		return err
	}
	return nil
}

func (g *GoLevelDB) Delete(key []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if err := g.db.Delete(key, nil); err != nil {
		return err
	}
	return nil
}

func (g *GoLevelDB) DeleteSync(key []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if err := g.db.Delete(key, &opt.WriteOptions{Sync: true}); err != nil {
		return err
	}
	return nil
}

func (g *GoLevelDB) Iterator(start, end []byte) (Iterator, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GoLevelDB) ReverseIterator(start, end []byte) (Iterator, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GoLevelDB) Close() error {
	return g.db.Close()
}

func (g *GoLevelDB) NewBatch() Batch {
	//TODO implement me
	panic("implement me")
}

func (g *GoLevelDB) Stats() map[string]string {
	keys := []string{
		"leveldb.num-files-at-level{n}",
		"leveldb.stats",
		"leveldb.sstables",
		"leveldb.blockpool",
		"leveldb.cachedblock",
		"leveldb.openedtables",
		"leveldb.alivesnaps",
		"leveldb.aliveiters",
	}

	stats := make(map[string]string)
	for _, key := range keys {
		str, err := g.db.GetProperty(key)
		if err == nil {
			stats[key] = str
		}
	}
	return stats
}

// NewGoLevelDB ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// NewGoLevelDB 接收的第一个参数name表示数据库的文件名："name.db"，dir表示存储数据库文件的目录地址。
func NewGoLevelDB(name string, dir string) (*GoLevelDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}
	return &GoLevelDB{db: db}, nil
}

var _ DB = (*GoLevelDB)(nil)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 流处理

type goLevelBatch struct {
	db    *GoLevelDB
	batch *leveldb.Batch
}
