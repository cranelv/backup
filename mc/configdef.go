package mc

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/p2p/discover"
)

const (
	MSKeyBroadcastTx      = "broad_txs"      // 广播交易
	MSKeyTopologyGraph    = "topology_graph" // 拓扑图
	MSKeyElectGraph       = "elect_graph"    // 选举图
	MSKeyElectOnlineState = "elect_state"    // 选举节点在线消息

	//通用
	MSKeyBroadcastInterval = "broad_interval" // 广播区块周期
	MSKeyElectGenTime      = "elect_gen_time"
	MSKeyMatrixAccount     = "matrix_specific_account"
	MSKeyElectConfigInfo   = "elect_details_info"
	MSKeyVIPConfig         = "vip_config"
	MSKeyPreBroadcastRoot  = "pre_broadcast_Root"

	MSKeyBlkRewardCfg = "blk_reward"
	MSKeyTxsRewardCfg = "txs_reward"
	MSKeyInterestCfg  = "interest_reward" //利息状态
	MSKeyLotteryCfg   = "lottery_reward"
	MSKeySlashCfg     = "slash_reward"
	MSKeyMultiCoin    = "coin_reward"
)

type ElectGenTimeStruct struct {
	MinerGen           uint16
	MinerNetChange     uint16
	ValidatorGen       uint16
	ValidatorNetChange uint16
	VoteBeforeTime     uint16
}

type NodeInfo struct {
	NodeID  discover.NodeID
	Address common.Address
}

type MatrixSpecialAccounts struct {
	BroadcastAccount   NodeInfo
	FoundationAccount  NodeInfo
	InnerMinerAccounts []NodeInfo
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
	MinMoney     uint64
	InterestRate uint64 //(分母待定为1000w)
	ElectUserNum uint8
	StockScale   uint16 //千分比
}

type PreBroadStateDB struct {
	LastStateDB       *state.StateDB
	BeforeLastStateDb *state.StateDB
}

type RewardRateCfg struct {
	MinerOutRate        uint64 //出块矿工奖励
	ElectedMinerRate    uint64 //当选矿工奖励
	FoundationMinerRate uint64 //基金会网络奖励

	LeaderRate              uint64 //出块验证者（leader）奖励
	ElectedValidatorsRate   uint64 //当选验证者奖励
	FoundationValidatorRate uint64 //基金会网络奖励

	OriginElectOfflineRate uint64 //初选下线验证者奖励
	BackupRewardRate       uint64 //当前替补验证者奖励
}

type BlkRewardCfg struct {
	MinerMount     uint64 //矿工奖励单位man
	MinerHalf      uint64 //矿工折半周期
	ValidatorMount uint64 //验证者奖励 单位man
	ValidatorHalf  uint64 //验证者折半周期
	RewardRate     RewardRateCfg
}

type TxsRewardCfgStruct struct {
	MinersRate     uint64 //矿工网络奖励
	ValidatorsRate uint64 //验证者网络奖励
	RewardRate     RewardRateCfg
}

type LotteryInfo struct {
	PrizeLevel uint8  //奖励级别
	PrizeNum   uint64 //奖励名额
	PrizeMoney uint64 //奖励金额 单位man
}

type LotteryCfgStruct struct {
	LotteryCalc string
	LotteryInfo []LotteryInfo
}

type InterestCfgStruct struct {
	CalcInterval uint64
	PayInterval  uint64
}

type SlashCfgStruct struct {
	SlashRate uint64
}
