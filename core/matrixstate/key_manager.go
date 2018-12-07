package matrixstate

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
)

const (
	MSPBroadcastInterval = "broad_interval" // 广播区块周期
	MSPBroadcastTx       = "broad_txs"      // 广播交易
	MSPTopologyGraph     = "topology_graph" // 拓扑图
	MSPElectGraph        = "elect_graph"    // 选举图
	MSPElectOnlineState  = "elect_state"    // 选举节点在线消息
)

func GetKeyHash(key string) common.Hash {
	hash, OK := km.keys[key]
	if OK {
		return hash
	} else {
		return common.Hash{}
	}
}

const (
	matrixStatePrefix = "ms_"
)

var km *keyManager

func init() {
	km = newKeyManager()
}

type keyManager struct {
	keys map[string]common.Hash
}

func newKeyManager() *keyManager {
	km := &keyManager{
		keys: map[string]common.Hash{
			MSPBroadcastInterval: types.RlpHash(matrixStatePrefix + MSPBroadcastInterval),
			MSPBroadcastTx:       types.RlpHash(matrixStatePrefix + MSPBroadcastTx),
			MSPTopologyGraph:     types.RlpHash(matrixStatePrefix + MSPTopologyGraph),
			MSPElectGraph:        types.RlpHash(matrixStatePrefix + MSPElectGraph),
			MSPElectOnlineState:  types.RlpHash(matrixStatePrefix + MSPElectOnlineState),
		},
	}
	return km
}
