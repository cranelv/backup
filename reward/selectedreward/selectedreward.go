package selectedreward

import (
	"github.com/matrix/go-matrix/params/manparams"
	"math/big"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
)

const (
	PackageName = "参与奖励"
)

type SelectedReward struct {
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
	NewTopologyGraph(header *types.Header) (*mc.TopologyGraph, error)
}

func (sr *SelectedReward) SetSelectedRewards(reward *big.Int, chain ChainReader, topRewards map[common.Address]*big.Int, roleType common.RoleType, header *types.Header, rate uint64) {

	//计算选举的拓扑图的高度

	var eleNum uint64
	num := header.Number
	if num.Uint64() < common.GetReElectionInterval() {
		eleNum = 0
	} else {
		eleNum = common.GetLastReElectionNumber(num.Uint64()) - 1
	}

	if reward.Cmp(big.NewInt(0)) <= 0 {
		log.WARN(PackageName, "奖励金额不合法", reward)
		return
	}
	originElectNodes, err := ca.GetTopologyByNumber(roleType, eleNum)
	if err != nil {
		log.Error(PackageName, "获取初选拓扑图错误", err)
		return
	}

	if 0 == len(originElectNodes.NodeList) {
		log.Error(PackageName, "get获取初选列表为空", "")
		return
	}
	newGraph, err :=  ca.GetTopologyByNumber(roleType, header.Number.Uint64()-1)

	if err != nil {
		log.Error(PackageName, "获取当前拓扑图错误", err)
		return
	}

	if 0 == len(newGraph.NodeList) {
		log.Error(PackageName, "当前拓扑图是 空", "")
		return
	}
	log.INFO(PackageName, "参与奖励大家共发放", reward)
	selectedNodesDeposit := sr.caclSelectedDeposit(newGraph, originElectNodes, num, roleType, rate)
	if 0 == len(selectedNodesDeposit) {
		log.Error(PackageName, "获取参与的抵押列表错误", "")
		return
	}
	util.CalcDepositRate(reward, selectedNodesDeposit, topRewards)
	return

}

func (sr *SelectedReward) caclSelectedDeposit(newGraph *mc.TopologyGraph, originElectNodes *mc.TopologyGraph, num *big.Int, roleType common.RoleType, rewardRate uint64) (map[common.Address]*big.Int) {
	NodesRewardMap := make(map[common.Address]uint64, 0)
	for _, nodelist := range newGraph.NodeList {
		NodesRewardMap[nodelist.Account] = rewardRate
		log.INFO(PackageName,"当前节点",nodelist.Account.Hex())
	}
	for _, electList := range originElectNodes.NodeList {
		if _, ok := NodesRewardMap[electList.Account]; ok {
			NodesRewardMap[electList.Account] = util.RewardFullRate
		} else {
			NodesRewardMap[electList.Account] = util.RewardFullRate - rewardRate
		}
		log.INFO(PackageName,"初选节点",electList.Account.Hex(),"比例",NodesRewardMap[electList.Account] )
	}

	selectedNodesDeposit := make(map[common.Address]*big.Int, 0)
	var depositNum uint64
	if num.Uint64() < common.GetReElectionInterval(){
		depositNum = 0
	}else{
		if common.RoleValidator == common.RoleValidator&roleType {
			depositNum = common.GetLastReElectionNumber(num.Uint64()) - manparams.VerifyTopologyGenerateUpTime
		}else{
			depositNum = common.GetLastReElectionNumber(num.Uint64()) - manparams.MinerTopologyGenerateUpTime
		}
	}

	depositNodes, err:= ca.GetElectedByHeightAndRole(new(big.Int).SetUint64(depositNum), roleType)
	if nil != err {
		log.ERROR(PackageName, "获取抵押列表错误", err)
		return nil
	}
	if 0 == len(depositNodes) {
		log.ERROR(PackageName, "获取抵押列表为空", "")
		return nil
	}
	for _, v := range depositNodes {

		if depositRate, ok := NodesRewardMap[v.Address]; ok {
			if v.Deposit.Cmp(big.NewInt(0)) < 0 {
				log.ERROR(PackageName, "获取抵押值错误，抵押", v.Deposit, "账户", v.Address.Hex())
				return nil
			}
			deposit := util.CalcRateReward(v.Deposit, depositRate)
			selectedNodesDeposit[v.Address] = deposit
			log.INFO(PackageName,"计算抵押总额,账户",v.Address.Hex(),"抵押",deposit)
		}
	}
	return selectedNodesDeposit
}
