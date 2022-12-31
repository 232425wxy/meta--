package config

import (
	"strings"
	"text/template"
)

const TemplateToml = `
[basic]
home = "{{ .BasicConfig.Home }}"
key_file = "{{ .BasicConfig.KeyFile }}"
genesis_file = "{{ .BasicConfig.GenesisFile }}"

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
`

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("TemplateToml").Funcs(template.FuncMap{"StringsJoin": strings.Join})
	if configTemplate, err = tmpl.Parse(TemplateToml); err != nil {
		panic(err)
	}
}
