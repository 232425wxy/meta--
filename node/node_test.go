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
	logger := log.New()
	logger.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
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

	select {}
}
