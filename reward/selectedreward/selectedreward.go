package selectedreward

import (
	"errors"
	"github.com/matrix/go-matrix/core/vm"
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

func (sr *SelectedReward)getTopAndDeposit(chain util.ChainReader,currentNum uint64,roleType common.RoleType)( *mc.TopologyGraph,  *mc.TopologyGraph,  []vm.DepositDetail, error){

	var eleNum uint64

	if currentNum < common.GetReElectionInterval() {
		eleNum = 0
	} else {
		eleNum = common.GetLastReElectionNumber(currentNum) - 1
	}

	originElectNodes, err := ca.GetTopologyByNumber(roleType, eleNum)
	if err != nil {
		log.Error(PackageName, "获取初选拓扑图错误", err)
		return nil,nil,nil,errors.New("获取初选拓扑图错误")
	}

	if 0 == len(originElectNodes.NodeList) {
		log.Error(PackageName, "get获取初选列表为空", "")
		return nil,nil,nil,errors.New("get获取初选列表为空")
	}
	currentTop, err :=  ca.GetTopologyByNumber(roleType, currentNum-1)

	if err != nil {
		log.Error(PackageName, "获取当前拓扑图错误", err)
		return nil,nil,nil,errors.New("获取当前拓扑图错误")
	}

	if 0 == len(currentTop.NodeList) {
		log.Error(PackageName, "当前拓扑图是 空", "")
		return nil,nil,nil,errors.New("当前拓扑图是 空")
	}



	var depositNum uint64
	chain.GetMatrixStateData(mc.MSKeyElectGenTime)
	originInfo,err:=chain.GetMatrixStateDataByNumber(mc.MSKeyElectGenTime, currentNum-1)

	if nil!=err{
		return nil,nil,nil,errors.New("获取选举信息的出错")
	}
	if currentNum < common.GetReElectionInterval(){
		depositNum = 0
	}else{
		if common.RoleValidator == common.RoleValidator&roleType {
			depositNum = common.GetLastReElectionNumber(currentNum) - uint64(originInfo.(*mc.ElectGenTimeStruct).ValidatorGen)
		}else{
			depositNum = common.GetLastReElectionNumber(currentNum) - uint64(originInfo.(*mc.ElectGenTimeStruct).MinerGen)
		}
	}

	depositNodes, err:= ca.GetElectedByHeightAndRole(new(big.Int).SetUint64(depositNum), roleType)
	if nil != err {
		log.ERROR(PackageName, "获取抵押列表错误", err)
		return nil,nil,nil,errors.New("获取抵押列表错误 ")
	}
	if 0 == len(depositNodes) {
		log.ERROR(PackageName, "获取抵押列表为空", "")
		return nil,nil,nil,errors.New("获取抵押列表为空 ")
	}
	return  currentTop,originElectNodes,depositNodes,nil
}

func (sr *SelectedReward) GetSelectedRewards(reward *big.Int, chain util.ChainReader,roleType common.RoleType, currentNum uint64, rate uint64) map[common.Address]*big.Int{

	//计算选举的拓扑图的高度
	if reward.Cmp(big.NewInt(0)) <= 0 {
		log.WARN(PackageName, "奖励金额不合法", reward)
		return nil
	}
	log.INFO(PackageName, "参与奖励大家共发放", reward)

	currentTop,originElectNodes,depositNodes,err:=sr.getTopAndDeposit(chain,currentNum,roleType)
	if nil!=err{
	        return nil
	}

	selectedNodesDeposit := sr.caclSelectedDeposit(currentTop, originElectNodes, depositNodes, rate)
	if 0 == len(selectedNodesDeposit) {
		log.Error(PackageName, "获取参与的抵押列表错误", "")
		return nil
	}

	return util.CalcDepositRate(reward, selectedNodesDeposit)

}

func (sr *SelectedReward) caclSelectedDeposit(newGraph *mc.TopologyGraph, originElectNodes *mc.TopologyGraph, depositNodes []vm.DepositDetail,rewardRate uint64) (map[common.Address]*big.Int) {
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
