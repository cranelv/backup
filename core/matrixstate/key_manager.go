package matrixstate

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/pkg/errors"
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

func GetDataByState(key string, state *state.StateDB) (interface{}, error) {
	hash := GetKeyHash(key)
	if (hash == common.Hash{}) {
		return nil, errors.Errorf("key(%s) not find", key)
	}
	codec, exist := km.codecMap[key]
	if !exist {
		return nil, errors.Errorf("codec of key(%s) not find", key)
	}

	bytes := state.GetMatrixData(hash)

	return codec.decodeFn(bytes)
}

const (
	matrixStatePrefix = "ms_"
)

var km *keyManager

func init() {
	km = newKeyManager()
}

type keyManager struct {
	keys     map[string]common.Hash
	codecMap map[string]codec
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
		codecMap: make(map[string]codec),
	}
	km.initCodec()
	return km
}
