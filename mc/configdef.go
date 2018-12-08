package mc

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"math/big"
)

const (
	MSPBroadcastTx      = "broad_txs"      // 广播交易
	MSPTopologyGraph    = "topology_graph" // 拓扑图
	MSPElectGraph       = "elect_graph"    // 选举图
	MSPElectOnlineState = "elect_state"    // 选举节点在线消息

	//通用
	MSPBroadcastInterval  = "broad_interval" // 广播区块周期
	MSPElectGenTime       = "elect_gen_time"
	MSPMatrixNode         = "matrix_specific_node"
	MSPElectConfigInfo    = "elect_details_info"
	MSPVIPConfig          = "vip_config"
	MSPreBroadcastStateDB = "pre broadcat state db"
)

type ElectGenTimeStruct struct {
	MinerGen           uint16
	MinerNetChange     uint16
	ValidatorGen       uint16
	ValidatorNetChange uint16
	VoteBeforeTime     uint16
}

type MatrixSpecilNode struct {
	BoradcastNode  []common.Address
	InnerMinerNode []common.Address
}
type ElectConfigInfo struct {
	MinerNum           uint16
	ValidatorNum       uint16
	BackValidator      uint16
	MinerElectPlug     string
	ValidatorElectPlug string
	WhiteList          []common.Address
	BlockList          []common.Address
}
type VIPConfig struct {
	MinMoney     *big.Int
	ElectUserNum uint8
	StockScale   uint16 //千分比
}

type PreBroadStateDB struct {
	LastStateDB       *state.StateDB
	BeforeLastStateDb *state.StateDB
}
