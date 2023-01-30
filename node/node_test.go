package node

import (
	"fmt"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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
	node, err := NewNode(cfg, logger, DefaultProvider())
	if err != nil {
		panic(err)
	}
	return node
}

func TestCreateNode(t *testing.T) {
	node0 := CreateNode(0)
	assert.Nil(t, node0.Start())
	_ = node0
}
