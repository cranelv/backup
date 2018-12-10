package blkreward

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/reward"
	"github.com/matrix/go-matrix/reward/cfg"
	"github.com/matrix/go-matrix/reward/rewardexec"
	"github.com/matrix/go-matrix/reward/util"
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
	Genesis() *types.Block
	GetSuperBlockHash() common.Hash
}

type blkreward struct {
	blockReward *rewardexec.BlockReward
	chain       ChainReader
	state       util.StateDB
}

func New(chain ChainReader, st util.StateDB) reward.Reward {
	//todo:从状态树读取配置.

	Rewardcfg, err := matrixstate.GetDataByState(mc.MSKeyBlkRewardCfg, st)
	if nil != err {
		log.ERROR("固定区块奖励", "获取状态树配置错误")
		return nil
	}

	rewardCfg := cfg.New(Rewardcfg.(*mc.BlkRewardCfg), nil)
	return rewardexec.New(chain, rewardCfg, st)
}

//func (tr *blkreward) CalcNodesRewards(blockReward *big.Int, Leader common.Address, header *types.Header) map[common.Address]*big.Int {
//	return tr.blockReward.CalcNodesRewards(blockReward, Leader, header)
//}
