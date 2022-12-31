package commands

import (
	"fmt"
	mos "github.com/232425wxy/meta--/common/os"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/types"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

var NodesNum int // 网络中节点数量
var OutputDir string

func init() {
	DockerNetCmd.Flags().IntVar(&NodesNum, "n", 4, "number of nodes to initialize in the docker net")
	DockerNetCmd.Flags().StringVar(&OutputDir, "o", ".", "root directory to store everything")
}

var DockerNetCmd = &cobra.Command{
	Use:   "dockernet",
	Short: "Initialize files for a meta-- docker net",
	RunE:  dockernetFiles,
}

func dockernetFiles(cmd *cobra.Command, args []string) (err error) {
	cfg := config.DefaultConfig()
	validators := make([]*types.Validator, 0)

	for i := 0; i < NodesNum; i++ {
		nodeDirName := fmt.Sprintf("node%d", i)
		nodeDir := filepath.Join(OutputDir, nodeDirName)
		cfg.SetHome(nodeDir)
		err = os.MkdirAll(filepath.Join(nodeDir, "config"), 0755)
		if err != nil {
			_ = os.RemoveAll(OutputDir)
			return err
		}
		err = os.MkdirAll(filepath.Join(nodeDir, "data"), 0755)
		if err != nil {
			_ = os.RemoveAll(OutputDir)
			return err
		}

		genesisFilePath := cfg.BasicConfig.GenesisFilePath()
		genesis := &types.Genesis{}
		if mos.FileExists(genesisFilePath) {
			genesis, err = types.GenesisReadFromFile(genesisFilePath)
			if err != nil {
				return err
			}
		} else {
			genesis.GenesisTime = time.Now()
			genesis.InitialHeight = 0
		}
		if err = genesis.SaveAs(genesisFilePath); err != nil {
			return err
		}

		keyFilePath := cfg.BasicConfig.KeyFilePath()

		var key *p2p.NodeKey
		key, err = p2p.LoadOrGenNodeKey(keyFilePath)
		if err != nil {
			_ = os.RemoveAll(OutputDir)
			return err
		}
		if err = key.SaveAs(keyFilePath); err != nil {
			return err
		}
		validator := &types.Validator{
			ID:               key.PublicKey.ToID(),
			PublicKey:        key.PublicKey,
			VotingPower:      10,
			ProposerPriority: 10,
		}
		validators = append(validators, validator)
		cfg.SaveAs(filepath.Join(nodeDir, "config", "config.toml"))
	}

	for i := 0; i < NodesNum; i++ {
		nodeDirName := fmt.Sprintf("node%d", i)
		nodeDir := filepath.Join(OutputDir, nodeDirName)
		cfg.SetHome(nodeDir)
		genesis := &types.Genesis{}
		genesis, err = types.GenesisReadFromFile(cfg.BasicConfig.GenesisFilePath())
		if err != nil {
			return err
		}
		genesis.Validators = validators
		if err = genesis.SaveAs(cfg.BasicConfig.GenesisFilePath()); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully initialized %d validayors files, and initialized genesis file.\n", NodesNum)
	return nil
}
