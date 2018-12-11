package mineroutreward

import (
	"math/big"

	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
)

type MinerOutReward struct {
}

const (
	PackageName = "矿工挖矿奖励"
)

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

func (mr *MinerOutReward) SetMinerOutRewards(reward *big.Int, state util.StateDB, chain ChainReader, num uint64) map[common.Address]*big.Int {
	//后一块给前一块的矿工发钱，广播区块不发钱， 广播区块下一块给广播区块前一块发钱

	if num == 1 {
		log.WARN(PackageName, "初始化矿工状态：", num)
		matrixstate.SetNumByState(mc.MSKEYMinerPayNum, state, num)
		return nil
	}

	if common.IsBroadcastNumber(num) {
		log.WARN(PackageName, "挖矿奖励高度错误：", num)
		return nil
	}
	var coinBase common.Address
	latestNum, err := matrixstate.GetNumByState(mc.MSKEYMinerPayNum, state)
	if nil != err {
		log.WARN(PackageName, "获取状态树错误", err)
		return nil
	}

	if latestNum >= num {
		log.WARN(PackageName, "奖励已发放", "")
		return nil
	}
	matrixstate.SetNumByState(mc.MSKEYMinerPayNum, state, num)

	if reward.Cmp(big.NewInt(0)) <= 0 {
		log.WARN(PackageName, "奖励金额不合法", reward)
		return nil
	}

	rewards := make(map[common.Address]*big.Int)
	for i := latestNum + 1; i < num+1; i++ {
		if common.IsBroadcastNumber(i - 1) {
			coinBase = chain.GetHeaderByNumber(i - 2).Coinbase
		} else {
			coinBase = chain.GetHeaderByNumber(i - 1).Coinbase
		}
		if coinBase.Equal(common.Address{}) {
			log.ERROR(PackageName, "矿工奖励的地址非法", coinBase.Hex())
			return nil
		}

		util.SetAccountRewards(rewards, coinBase, reward)
		log.Info(PackageName, "出块矿工账户：", coinBase.String(), "发放奖励高度", i, "奖励金额", reward)
	}

	return rewards
}
