package matrixstate

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

const (
	MSPBroadcastTx      = "broad_txs"      // 广播交易
	MSPTopologyGraph    = "topology_graph" // 拓扑图
	MSPElectGraph       = "elect_graph"    // 选举图
	MSPElectOnlineState = "elect_state"    // 选举节点在线消息

	//通用
	MSPBroadcastInterval = "broad_interval" // 广播区块周期
	ElectGenTime         = "elect_gen_time"
	//奖励
	MinerRewardCfg     = "miner_reward"
	ValidatorRewardCfg = "validator_reward"
	TxsRewardCfg       = "txs_reward"
	InterestCfg        = "interest_reward" //利息状态
	LotteryCfg         = "lottery_reward"
	SlashCfg           = "slash_reward"
	MultiCoin          = "coin_reward"
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
			mc.MSPBroadcastInterval: types.RlpHash(matrixStatePrefix + mc.MSPBroadcastInterval),
			mc.MSPBroadcastTx:       types.RlpHash(matrixStatePrefix + mc.MSPBroadcastTx),
			mc.MSPTopologyGraph:     types.RlpHash(matrixStatePrefix + mc.MSPTopologyGraph),
			mc.MSPElectGraph:        types.RlpHash(matrixStatePrefix + mc.MSPElectGraph),
			mc.MSPElectOnlineState:  types.RlpHash(matrixStatePrefix + mc.MSPElectOnlineState),
			mc.MSPElectGenTime:      types.RlpHash(matrixStatePrefix + mc.MSPElectGenTime),
			mc.MSPMatrixNode:        types.RlpHash(matrixStatePrefix + mc.MSPMatrixNode),
			mc.MSPElectConfigInfo:   types.RlpHash(matrixStatePrefix + mc.MSPElectConfigInfo),
			mc.MSPVIPConfig:         types.RlpHash(matrixStatePrefix + mc.MSPVIPConfig),
			mc.MinerRewardCfg:       types.RlpHash(matrixStatePrefix + mc.MinerRewardCfg),
			mc.ValidatorRewardCfg:   types.RlpHash(matrixStatePrefix + mc.ValidatorRewardCfg),
			mc.TxsRewardCfg:         types.RlpHash(matrixStatePrefix + mc.TxsRewardCfg),
			mc.InterestCfg:          types.RlpHash(matrixStatePrefix + mc.InterestCfg),
			mc.LotteryCfg:           types.RlpHash(matrixStatePrefix + mc.LotteryCfg),
			mc.SlashCfg:             types.RlpHash(matrixStatePrefix + mc.SlashCfg),
			mc.MultiCoin:            types.RlpHash(matrixStatePrefix + mc.MultiCoin),
		},
		codecMap: make(map[string]codec),
	}
	km.initCodec()
	return km
}
