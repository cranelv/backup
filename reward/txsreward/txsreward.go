package txsreward

import (
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/reward"
	"github.com/matrix/go-matrix/reward/rewardexec"

	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/reward/cfg"
)

const (
	PackageName = "交易奖励"

	//todo: 分母10000， 加法做参数检查
	ValidatorsTxsRewardRate = uint64(util.RewardFullRate) //验证者交易奖励比例100%
	MinerTxsRewardRate      = uint64(0)                   //矿工交易奖励比例0%
	FoundationTxsRewardRate = uint64(0)                   //基金会交易奖励比例0%

	MinerOutRewardRate     = uint64(4000) //出块矿工奖励40%
	ElectedMinerRewardRate = uint64(6000) //当选矿工奖励60%

	LeaderRewardRate            = uint64(4000) //出块验证者（leader）奖励40%
	ElectedValidatorsRewardRate = uint64(6000) //当选验证者奖励60%

	OriginElectOfflineRewardRate = uint64(5000) //初选下线验证者奖励50%
	BackupRate                   = uint64(5000) //当前替补验证者奖励50%

)

type TxsReward struct {
	blockReward *rewardexec.BlockReward
}

func New(chain util.ChainReader, st util.StateDB) reward.Reward {

	Rewardcfg, err := matrixstate.GetDataByState(mc.MSKeyTxsRewardCfg, st)
	if nil != err {
		log.ERROR(PackageName, "获取状态树配置错误")
		return nil
	}
	rate := Rewardcfg.(*mc.TxsRewardCfgStruct).RewardRate

	cfg := cfg.New(&mc.BlkRewardCfg{RewardRate: rate}, nil)
	if util.RewardFullRate != Rewardcfg.(*mc.TxsRewardCfgStruct).ValidatorsRate+Rewardcfg.(*mc.TxsRewardCfgStruct).MinersRate {
		log.ERROR(PackageName, "交易费奖励比例配置错误", "")
		return nil
	}

	return rewardexec.New(chain, cfg, st)

}
