package cfg

import (
	"math/big"

	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/reward/leaderreward"
	"github.com/matrix/go-matrix/reward/mineroutreward"
	"github.com/matrix/go-matrix/reward/selectedreward"

	"github.com/matrix/go-matrix/core/types"

	"github.com/matrix/go-matrix/common"
)

const (
	PackageName = "奖励"

	//todo: 分母10000， 加法做参数检查
	MinersBlockRewardRate     = uint64(5000) //矿工网络奖励50%
	ValidatorsBlockRewardRate = uint64(5000) //验证者网络奖励50%

	MinerOutRewardRate        = uint64(4000) //出块矿工奖励40%
	ElectedMinerRewardRate    = uint64(5000) //当选矿工奖励50%
	FoundationMinerRewardRate = uint64(1000) //基金会网络奖励10%

	LeaderRewardRate               = uint64(4000) //出块验证者（leader）奖励40%
	ElectedValidatorsRewardRate    = uint64(5000) //当选验证者奖励60%
	FoundationValidatorsRewardRate = uint64(1000) //基金会网络奖励10%

	OriginElectOfflineRewardRate = uint64(5000) //初选下线验证者奖励50%
	BackupRate                   = uint64(5000) //当前替补验证者奖励50%
)

type RewardStateCfg struct {
	MinersRate     uint64 //矿工网络奖励
	ValidatorsRate uint64 //验证者网络奖励

	MinerOutRate        uint64 //出块矿工奖励
	ElectedMinerRate    uint64 //当选矿工奖励
	FoundationMinerRate uint64 //基金会网络奖励

	LeaderRate              uint64 //出块验证者（leader）奖励
	ElectedValidatorsRate   uint64 //当选验证者奖励
	FoundationValidatorRate uint64 //基金会网络奖励

	OriginElectOfflineRate uint64 //初选下线验证者奖励
	BackupRewardRate       uint64 //当前替补验证者奖励
}

type RewardCfg struct {
	RewardMount *mc.RewardRateCfg
	SetReward   SetRewardsExec
}
type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	GetBlockByNumber(number uint64) *types.Block

	// GetBlock retrieves a block sfrom the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)
	State() (*state.StateDB, error)
}
type SetRewardsExec interface {
	SetLeaderRewards(reward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int
	SetMinerOutRewards(reward *big.Int, chain ChainReader, num uint64) map[common.Address]*big.Int
	GetSelectedRewards(reward *big.Int, roleType common.RoleType, number uint64, rate uint64) map[common.Address]*big.Int //todo 金额
}
type DefaultSetRewards struct {
	leader   leaderreward.LeaderReward
	miner    mineroutreward.MinerOutReward
	selected selectedreward.SelectedReward
}

func DefaultSetRewardNew() *DefaultSetRewards {
	return &DefaultSetRewards{
		leader:   leaderreward.LeaderReward{},
		miner:    mineroutreward.MinerOutReward{},
		selected: selectedreward.SelectedReward{},
	}

}

func (str *DefaultSetRewards) SetLeaderRewards(reward *big.Int, Leader common.Address, num uint64) map[common.Address]*big.Int {
	if common.IsBroadcastNumber(num) {
		log.WARN(PackageName, "leader奖励高度错误", num)
		return nil
	}
	return str.leader.SetLeaderRewards(reward, Leader, num)
}
func (str *DefaultSetRewards) GetSelectedRewards(reward *big.Int, roleType common.RoleType, number uint64, rate uint64) map[common.Address]*big.Int {
	if common.IsBroadcastNumber(number) {
		return nil
	}
	return str.selected.GetSelectedRewards(reward, roleType, number, rate)
}
func (str *DefaultSetRewards) SetMinerOutRewards(reward *big.Int, chain ChainReader, num uint64) map[common.Address]*big.Int {
	if common.IsBroadcastNumber(num) {
		log.WARN(PackageName, "矿工奖励高度错误", num)
		return nil
	}
	return str.miner.SetMinerOutRewards(reward, chain, num)
}

func New(RewardMount *mc.RewardRateCfg, SetReward SetRewardsExec) *RewardCfg {

	//默认配置
	if nil == RewardMount {
		RewardMount = &mc.RewardRateCfg{
			MinersRate:     MinersBlockRewardRate,
			ValidatorsRate: ValidatorsBlockRewardRate,

			MinerOutRate:        MinerOutRewardRate,
			ElectedMinerRate:    ElectedMinerRewardRate,
			FoundationMinerRate: FoundationMinerRewardRate,

			LeaderRate:              LeaderRewardRate,
			ElectedValidatorsRate:   ElectedValidatorsRewardRate,
			FoundationValidatorRate: FoundationValidatorsRewardRate,

			OriginElectOfflineRate: OriginElectOfflineRewardRate,
			BackupRewardRate:       BackupRate,
		}

	}
	//默认配置
	if nil == SetReward {

		SetReward = DefaultSetRewardNew()
	}

	return &RewardCfg{
		RewardMount: RewardMount,
		SetReward:   SetReward,
	}
}
