package types

import "github.com/232425wxy/meta--/proto/pbtypes"

type Data struct {
	Txs            Txs
	MerkleRootHash []byte
}

func (d Data) ToProto() pbtypes.Data {
	pbData := pbtypes.Data{}
	if len(d.Txs) > 0 {
		txs := make([][]byte, len(d.Txs))
		for i := 0; i < len(d.Txs); i++ {
			txs[i] = d.Txs[i]
		}
		pbData.Txs = txs
		pbData.MerkleRootHash = d.Txs.MerkleRootHash()
	}
	return pbData
}
