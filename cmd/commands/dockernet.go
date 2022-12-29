package commands

import (
	"fmt"
	"github.com/232425wxy/meta--/config"
	"github.com/spf13/cobra"
	"path/filepath"
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

func dockernetFiles(cmd *cobra.Command, args []string) error {
	cfg := config.DefaultConfig()

	for i := 0; i < NodesNum; i++ {
		nodeDirName := fmt.Sprintf("node%d", i)
		nodeDir := filepath.Join(OutputDir, nodeDirName)
		cfg.SetHome(nodeDir)
	}
}
