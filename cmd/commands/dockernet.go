package commands

import (
	"fmt"
	mos "github.com/232425wxy/meta--/common/os"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/stch"
	"github.com/232425wxy/meta--/types"
	"github.com/spf13/cobra"
	"net"
	"os"
	"path/filepath"
	"time"
)

var NodesNum int // 网络中节点数量
var OutputDir string
var IP string
var Port int

func init() {
	DockerNetCmd.Flags().IntVar(&NodesNum, "n", 4, "number of nodes to initialize in the docker net")
	DockerNetCmd.Flags().StringVar(&OutputDir, "o", ".", "root directory to store everything")
	DockerNetCmd.Flags().StringVar(&IP, "ip", "127.0.0.1", "ip address")
	DockerNetCmd.Flags().IntVar(&Port, "port", 26656, "p2p listen port")
}

var DockerNetCmd = &cobra.Command{
	Use:   "dockernet",
	Short: "Initialize files for a meta-- docker net",
	RunE:  dockernetFiles,
}

func dockernetFiles(cmd *cobra.Command, args []string) (err error) {
	cfg := config.DefaultConfig()
	validators := make([]*types.Validator, 0)
	var genesisExists = make(map[int]bool)
	var neighbours = make([]string, 0)

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
			genesisExists[i] = true
			genesis, err = types.GenesisReadFromFile(genesisFilePath)
			if err != nil {
				return err
			}
		} else {
			genesis.GenesisTime = time.Now()
			genesis.InitialHeight = 1
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
			ID:             key.PublicKey.ToID(),
			PublicKey:      key.PublicKey,
			VotingPower:    10,
			LeaderPriority: 10,
		}

		kp := stch.NewKP(NodesNum)
		kp.Save(cfg.BasicConfig.ChameleonKeyFilePath())

		validators = append(validators, validator)

		ip := net.ParseIP(IP)
		ip = ip.To4()
		//for j := 0; j < i; j++ {
		//	ip[3]++
		//}
		neighbour := p2p.IDAddressString(key.PublicKey.ToID(), fmt.Sprintf("%s:%d", ip.String(), Port+i))
		neighbours = append(neighbours, neighbour)
	}

	for i := 0; i < NodesNum; i++ {
		cfg.P2PConfig.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", 26656+i)
		nodeDirName := fmt.Sprintf("node%d", i)
		nodeDir := filepath.Join(OutputDir, nodeDirName)
		cfg.SetHome(nodeDir)
		cfg.P2PConfig.SetNeighbours(neighbours)
		cfg.SaveAs(filepath.Join(nodeDir, "config", "config.toml"))
		if !genesisExists[i] {
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
	}

	fmt.Printf("Successfully initialized %d validayors files, and initialized genesis file.\n", NodesNum)
	return nil
}
