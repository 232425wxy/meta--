package config

import (
	"strings"
	"text/template"
)

const TemplateToml = `
[basic]
home = "{{ .BasicConfig.Home }}"
key_file = "{{ .BasicConfig.KeyFile }}"
chameleon_key_file = "{{ .BasicConfig.ChameleonKeyFile }}"
genesis_file = "{{ .BasicConfig.GenesisFile }}"
db_backend = "{{ .BasicConfig.DBBackend }}"
db_dir = "{{ .BasicConfig.DBDir }}"
app = "{{ .BasicConfig.App }}"

[p2p]
home = "{{ .P2PConfig.Home }}"
listen_address = "{{ .P2PConfig.ListenAddress }}"
addr_book = "{{ .P2PConfig.AddrBook }}"
flush_duration = "{{ .P2PConfig.FlushDuration }}"
max_packet_msg_payload_size = "{{ .P2PConfig.MaxPacketMsgPayloadSize }}"
send_rate = "{{ .P2PConfig.SendRate }}"
recv_rate = "{{ .P2PConfig.RecvRate }}"
pong_timeout = "{{ .P2PConfig.PongTimeout }}"
ping_interval = "{{ .P2PConfig.PingInterval }}"
neighbours = "{{ .P2PConfig.Neighbours }}"

[txs_pool]
home = "{{ .TxsPoolConfig.Home }}"
max_size = "{{ .TxsPoolConfig.MaxSize }}"
max_tx_bytes = "{{ .TxsPoolConfig.MaxTxBytes }}"

[consensus]
home = "{{ .ConsensusConfig.Home }}"
timeout_prepare = "{{ .ConsensusConfig.TimeoutPrepare }}"
timeout_pre_commit = "{{ .ConsensusConfig.TimeoutPreCommit }}"
timeout_commit = "{{ .ConsensusConfig.TimeoutCommit }}"
timeout_decide = "{{ .ConsensusConfig.TimeoutDecide }}"
timeout_consensus = "{{ .ConsensusConfig.TimeoutConsensus }}"
`

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("TemplateToml").Funcs(template.FuncMap{"StringsJoin": strings.Join})
	if configTemplate, err = tmpl.Parse(TemplateToml); err != nil {
		panic(err)
	}
}
