package node

import (
	"fmt"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func ReadConfigFile(path string) *config.Config {
	viper.AddConfigPath(filepath.Join(path, "config"))
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	cfg := &config.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		panic(err)
	}
	viper.Reset()
	return cfg
}

func AdjustHomePath(cfg *config.Config) {
	cfg.BasicConfig.Home = fmt.Sprintf("/root/lab/code/go/src/meta--/%s", cfg.BasicConfig.Home)
	cfg.P2PConfig.Home = fmt.Sprintf("/root/lab/code/go/src/meta--/%s", cfg.P2PConfig.Home)
	cfg.ConsensusConfig.Home = fmt.Sprintf("/root/lab/code/go/src/meta--/%s", cfg.ConsensusConfig.Home)
	cfg.TxsPoolConfig.Home = fmt.Sprintf("/root/lab/code/go/src/meta--/%s", cfg.TxsPoolConfig.Home)
}

func CreateNode(i int) *Node {
	dir := fmt.Sprintf("../node%d", i)
	cfg := ReadConfigFile(dir)
	AdjustHomePath(cfg)
	logger := log.New("node", i)
	logger.SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stdout, log.TerminalFormat(true))))
	log.PrintOrigins(true)
	node, err := NewNode(cfg, logger, DefaultProvider())
	if err != nil {
		panic(err)
	}
	return node
}

func TestCreateAndStartNode(t *testing.T) {
	nodes := make([]*Node, 4)
	nodes[0] = CreateNode(0)
	nodes[1] = CreateNode(1)
	nodes[2] = CreateNode(2)
	nodes[3] = CreateNode(3)

	for i := 0; i < len(nodes); i++ {
		go func(i int) { assert.Nil(t, nodes[i].Start()) }(i)
	}

	time.Sleep(time.Second * 5)

	tx := []byte(fmt.Sprintf("number=%d", 1))
	err := nodes[1%4].txsPool.CheckTx(tx, nodes[1%4].nodeInfo.ID())
	assert.Nil(t, err)
	//time.Sleep(time.Millisecond * 2)

	time.Sleep(time.Second * 5)
	//fmt.Println("第二阶段...")
	//for i := 0; i < 5; i++ {
	//	tx := []byte(fmt.Sprintf("number=%d", i+1000))
	//	err := nodes[i%4].txsPool.CheckTx(tx, nodes[i%4].nodeInfo.ID())
	//	assert.Nil(t, err)
	//	time.Sleep(time.Second * 2)
	//}

	fmt.Println("修改前")
	for _, n := range nodes {
		fmt.Println(n.blockStore.LoadBlockByHeight(1))
	}

	time.Sleep(time.Second * 3)

	nodes[0].State().RedactBlock(1, 0, []byte("学校"), []byte("信息工程大学"))

	time.Sleep(time.Second * 4)

	fmt.Println("修改后")
	for _, n := range nodes {
		fmt.Println(n.blockStore.LoadBlockByHeight(1))
	}

	nodes[0].State().RedactBlock(1, 0, []byte("学校"), []byte("西北工业大学"))

	time.Sleep(time.Second * 4)

	fmt.Println("再次修改后")
	for _, n := range nodes {
		fmt.Println(n.blockStore.LoadBlockByHeight(1))
	}

	//fmt.Println("第三阶段...")
	//for i := 0; i < 3; i++ {
	//	tx := []byte(fmt.Sprintf("number=%d", i+1005))
	//	err := nodes[i%4].txsPool.CheckTx(tx, nodes[i%4].nodeInfo.ID())
	//	assert.Nil(t, err)
	//	time.Sleep(time.Second * 3)
	//}
	//time.Sleep(time.Second * 10)
	//fmt.Println("第四阶段...")
	//for i := 0; i < 2; i++ {
	//	tx := []byte(fmt.Sprintf("number=%d", i+1008))
	//	err := nodes[i%4].txsPool.CheckTx(tx, nodes[i%4].nodeInfo.ID())
	//	assert.Nil(t, err)
	//	time.Sleep(time.Second * 4)
	//}

	select {}
}
