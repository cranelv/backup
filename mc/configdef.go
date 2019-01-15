package mc

import (
	"github.com/matrix/go-matrix/common"
	"math/big"
)

const (
	MSKeyVersionInfo      = "version_info"   // 版本信息
	MSKeyBroadcastTx      = "broad_txs"      // 广播交易
	MSKeyTopologyGraph    = "topology_graph" // 拓扑图
	MSKeyElectGraph       = "elect_graph"    // 选举图
	MSKeyElectOnlineState = "elect_state"    // 选举节点在线信息

	//通用
	MSKeyBroadcastInterval    = "broad_interval"         // 广播区块周期
	MSKeyElectGenTime         = "elect_gen_time"         // 选举生成时间
	MSKeyElectMinerNum        = "elect_miner_num"        // 选举矿工数量
	MSKeyElectConfigInfo      = "elect_details_info"     // 选举配置
	MSKeyElectBlackList       = "elect_black_list"       // 选举黑名单
	MSKeyElectWhiteList       = "elect_white_list"       // 选举白名单
	MSKeyAccountBroadcasts    = "account_broadcasts"     // 广播账户 []common.Address
	MSKeyAccountInnerMiners   = "account_inner_miners"   // 基金会矿工 []common.Address
	MSKeyAccountFoundation    = "account_foundation"     // 基金会账户 common.Address
	MSKeyAccountVersionSupers = "account_version_supers" // 版本签名账户 []common.Address
	MSKeyAccountBlockSupers   = "account_block_supers"   // 超级区块签名账户 []common.Address
	MSKeyVIPConfig            = "vip_config"             // VIP配置信息
	MSKeyPreBroadcastRoot     = "pre_broadcast_Root"     // 前广播区块root信息
	MSKeyLeaderConfig         = "leader_config"          // leader服务配置信息
	MSKeyMinHash              = "pre_100_min_hash"       // 最小hash
	MSKeySuperBlockCfg        = "super_block_config"     // 超级区块配置

	//奖励配置
	MSKeyBlkRewardCfg      = "blk_reward"         // 区块奖励配置
	MSKeyTxsRewardCfg      = "txs_reward"         // 交易奖励配置
	MSKeyInterestCfg       = "interest_reward"    // 利息配置
	MSKeyLotteryCfg        = "lottery_reward"     // 彩票配置
	MSKeySlashCfg          = "slash_reward"       // 惩罚配置
	MSKeyPreMinerBlkReward = "preMiner_blkreward" // 上一矿工区块奖励金额
	MSKeyPreMinerTxsReward = "preMiner_txsreward" // 上一矿工交易奖励金额
	MSKeyUpTimeNum         = "upTime_num"         // upTime状态
	MSKeyLotteryNum        = "lottery_num"        // 彩票状态
	MSKeyLotteryAccount    = "lottery_from"       // 彩票候选账户
	MSKeyInterestCalcNum   = "interest_calc_num"  // 利息计算状态
	MSKeyInterestPayNum    = "interest_pay_num"   // 利息支付状态
	MSKeySlashNum          = "slash_num"          // 惩罚状态

	//未出块选举惩罚配置相关
	MSKeyBlockProduceStatsStatus = "block_produce_stats_status" //
	MSKeyBlockProduceSlashCfg    = "block_produce_slash_cfg"    //
	MSKeyBlockProduceStats       = "block_produce_stats"        //
	MSKeyBlockProduceBlackList   = "block_produce_blacklist"    //
)

type BCIntervalInfo struct {
	LastBCNumber       uint64 // 最后的广播区块高度
	LastReelectNumber  uint64 // 最后的选举区块高度
	BCInterval         uint64 // 广播周期
	BackupEnableNumber uint64 // 预备广播周期启用高度
	BackupBCInterval   uint64 // 预备广播周期
}

type ElectGenTimeStruct struct {
	MinerGen           uint16
	MinerNetChange     uint16
	ValidatorGen       uint16
	ValidatorNetChange uint16
	VoteBeforeTime     uint16
}

type ElectMinerNumStruct struct {
	MinerNum uint16
}

type ElectConfigInfo_All struct {
	MinerNum      uint16
	ValidatorNum  uint16
	BackValidator uint16
	ElectPlug     string
	WhiteList     []common.Address
	BlackList     []common.Address
}
type ElectConfigInfo struct {
	ValidatorNum  uint16
	BackValidator uint16
	ElectPlug     string
}

type VIPConfig struct {
	MinMoney     uint64
	InterestRate uint64 //(分母待定为1000w)
	ElectUserNum uint8
	StockScale   uint16 //千分比
}

type LeaderConfig struct {
	ParentMiningTime      int64 // 预留父区块挖矿时间
	PosOutTime            int64 // 区块POS共识超时时间
	ReelectOutTime        int64 // 重选超时时间
	ReelectHandleInterval int64 // 重选处理间隔时间
}

type PreBroadStateRoot struct {
	LastStateRoot       common.Hash
	BeforeLastStateRoot common.Hash
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
	BlkRewardCalc  string
	MinerMount     uint64 //矿工奖励单位man
	MinerHalf      uint64 //矿工折半周期
	ValidatorMount uint64 //验证者奖励 单位man
	ValidatorHalf  uint64 //验证者折半周期
	RewardRate     RewardRateCfg
}

type TxsRewardCfgStruct struct {
	TxsRewardCalc  string
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
	InterestCalc string
	CalcInterval uint64
	PayInterval  uint64
}

type SlashCfgStruct struct {
	SlashCalc string
	SlashRate uint64
}

type SuperBlkCfg struct {
	Seq uint64
	Num uint64
}

type MinerOutReward struct {
	Reward big.Int
}

type LotteryFrom struct {
	From []common.Address
}

type RandomInfoStruct struct {
	MinHash  common.Hash
	MaxNonce uint64
}

type BlockProduceSlashCfg struct {
	Switcher         bool
	LowTHR           uint16
	ProhibitCycleNum uint16
}

type UserBlockProduceNum struct {
	Address    common.Address
	ProduceNum uint16
}

type BlockProduceStats struct {
	StatsList []UserBlockProduceNum
}

type UserBlockProduceSlash struct {
	Address              common.Address
	ProhibitCycleCounter uint16
}

type BlockProduceSlashBlackList struct {
	BlackList []UserBlockProduceSlash
}

type BlockProduceSlashStatsStatus struct {
	Number uint64
}
