package commands

import (
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/p2p"
	"github.com/spf13/cobra"
)

func initFiles(cmd *cobra.Command, args []string) error {

}

func initFilesWithConfig(cfg *config.Config) error {
	// 初始化节点的密钥文件
	keyFile := cfg.BasicConfig.KeyFilePath()
	key, err := p2p.LoadOrGenNodeKey(keyFile)
	if err != nil {
		panic(err)
	}
	err = key.SaveAs(keyFile)
	if err != nil {
		panic(err)
	}

	// 初始化区块链的初始文件
	genesisFile := cfg.BasicConfig.GenesisFilePath()

}
