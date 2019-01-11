// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

type cdc struct {
	state            stateDef
	number           uint64
	selfAddr         common.Address
	role             common.RoleType
	curConsensusTurn mc.ConsensusTurnInfo
	consensusLeader  common.Address
	curReelectTurn   uint32
	reelectMaster    common.Address
	isMaster         bool
	leaderCal        *leaderCalculator
	bcInterval       *mc.BCIntervalInfo
	parentState      StateReader
	turnTime         *turnTimes
	chain            *core.BlockChain
	logInfo          string
}

func newCDC(number uint64, chain *core.BlockChain, logInfo string) *cdc {
	dc := &cdc{
		state:            stIdle,
		number:           number,
		selfAddr:         common.Address{},
		role:             common.RoleNil,
		curConsensusTurn: mc.ConsensusTurnInfo{},
		consensusLeader:  common.Address{},
		curReelectTurn:   0,
		reelectMaster:    common.Address{},
		isMaster:         false,
		bcInterval:       nil,
		parentState:      nil,
		turnTime:         newTurnTimes(),
		chain:            chain,
		logInfo:          logInfo,
	}

	dc.leaderCal = newLeaderCalculator(chain, dc.number, dc.logInfo)
	return dc
}

func (dc *cdc) SetSelfAddress(addr common.Address) {
	dc.selfAddr = addr
}

func (dc *cdc) AnalysisState(parentHeader *types.Header, preIsSupper bool, parentState StateReader) error {
	if parentState == nil || parentHeader == nil {
		return errors.New("parent state or parentHeader is nil")
	}

	validators, role, err := dc.readValidatorsAndRoleFromState(parentState)
	if err != nil {
		return err
	}
	specials, err := dc.readSpecialAccountsFromState(parentState)
	if err != nil {
		return err
	}
	config, err := dc.readLeaderConfigFromState(parentState)
	if err != nil {
		return err
	}
	bcInterval, err := dc.readBroadCastIntervalFromState(parentState)
	if err != nil {
		return err
	}

	if err := dc.leaderCal.SetValidatorsAndSpecials(parentHeader, preIsSupper, validators, specials, bcInterval); err != nil {
		return err
	}

	consensusIndex := dc.curConsensusTurn.TotalTurns()
	consensusLeader, err := dc.GetLeader(consensusIndex, bcInterval)
	if err != nil {
		return err
	}
	if dc.curReelectTurn != 0 {
		reelectLeader, err := dc.GetLeader(consensusIndex+dc.curReelectTurn, bcInterval)
		if err != nil {
			return err
		}
		dc.reelectMaster.Set(reelectLeader)
	} else {
		dc.reelectMaster.Set(common.Address{})
	}
	if err := dc.turnTime.SetTimeConfig(config); err != nil {
		log.Error(dc.logInfo, "turnTime设置时间配置参数失败", err)
		return err
	}
	dc.bcInterval = bcInterval
	dc.consensusLeader.Set(consensusLeader)
	dc.parentState = parentState
	dc.role = role

	return nil
}

func (dc *cdc) SetConsensusTurn(consensusTurn mc.ConsensusTurnInfo) error {
	consensusLeader, err := dc.GetLeader(consensusTurn.TotalTurns(), dc.bcInterval)
	if err != nil {
		return errors.Errorf("获取共识leader错误(%v), 共识轮次: %s", err, consensusTurn.String())
	}

	dc.consensusLeader.Set(consensusLeader)
	dc.curConsensusTurn = consensusTurn
	dc.reelectMaster.Set(common.Address{})
	dc.curReelectTurn = 0
	return nil
}

func (dc *cdc) SetReelectTurn(reelectTurn uint32) error {
	if dc.curReelectTurn == reelectTurn {
		return nil
	}
	if reelectTurn == 0 {
		dc.reelectMaster.Set(common.Address{})
		dc.curReelectTurn = 0
		return nil
	}
	master, err := dc.GetLeader(dc.curConsensusTurn.TotalTurns()+reelectTurn, dc.bcInterval)
	if err != nil {
		return errors.Errorf("获取master错误(%v), 重选轮次(%d), 共识轮次(%d)", err, reelectTurn, dc.curConsensusTurn.String())
	}
	dc.reelectMaster.Set(master)
	dc.curReelectTurn = reelectTurn
	return nil
}

func (dc *cdc) GetLeader(turn uint32, bcInterval *mc.BCIntervalInfo) (common.Address, error) {
	leaders, err := dc.leaderCal.GetLeader(turn, bcInterval)
	if err != nil {
		return common.Address{}, err
	}
	return leaders.leader, nil
}

func (dc *cdc) GetConsensusLeader() common.Address {
	return dc.consensusLeader
}

func (dc *cdc) GetReelectMaster() common.Address {
	return dc.reelectMaster
}

func (dc *cdc) PrepareLeaderMsg() (*mc.LeaderChangeNotify, error) {
	leaders, err := dc.leaderCal.GetLeader(dc.curConsensusTurn.TotalTurns()+dc.curReelectTurn, dc.bcInterval)
	if err != nil {
		return nil, err
	}

	return &mc.LeaderChangeNotify{
		PreLeader:      dc.leaderCal.preLeader,
		Leader:         leaders.leader,
		NextLeader:     leaders.nextLeader,
		ConsensusTurn:  dc.curConsensusTurn,
		ReelectTurn:    dc.curReelectTurn,
		Number:         dc.number,
		ConsensusState: dc.state != stReelect,
		TurnBeginTime:  dc.turnTime.GetBeginTime(dc.curConsensusTurn),
		TurnEndTime:    dc.turnTime.GetPosEndTime(dc.curConsensusTurn),
	}, nil
}

func (dc *cdc) readValidatorsAndRoleFromState(state StateReader) ([]mc.TopologyNodeInfo, common.RoleType, error) {
	topology, err := matrixstate.GetTopologyGraph(state)
	if err != nil {
		return nil, common.RoleNil, err
	}
	if topology == nil {
		return nil, common.RoleNil, errors.New("topology data is nil")
	}

	role := dc.getRoleFromTopology(topology)

	validators := make([]mc.TopologyNodeInfo, 0)
	for _, node := range topology.NodeList {
		if node.Type == common.RoleValidator {
			validators = append(validators, node)
		}
	}
	return validators, role, nil
}

func (dc *cdc) getRoleFromTopology(TopologyGraph *mc.TopologyGraph) common.RoleType {
	for _, v := range TopologyGraph.NodeList {
		if v.Account == dc.selfAddr {
			return v.Type
		}
	}
	return common.RoleNil
}

func (dc *cdc) readSpecialAccountsFromState(state StateReader) (*specialAccounts, error) {
	broadcasts, err := matrixstate.GetBroadcastAccounts(state)
	if err != nil {
		return nil, err
	}

	versionSupers, err := matrixstate.GetVersionSuperAccounts(state)
	if err != nil {
		return nil, err
	}

	blockSupers, err := matrixstate.GetBlockSuperAccounts(state)
	if err != nil {
		return nil, err
	}

	return &specialAccounts{
		broadcasts:    broadcasts,
		versionSupers: versionSupers,
		blockSupers:   blockSupers,
	}, nil
}

func (dc *cdc) readLeaderConfigFromState(state StateReader) (*mc.LeaderConfig, error) {
	config, err := matrixstate.GetLeaderConfig(state)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("LeaderConfig == nil")
	}
	return config, nil
}

func (dc *cdc) readBroadCastIntervalFromState(state StateReader) (*mc.BCIntervalInfo, error) {
	interval, err := matrixstate.GetBroadcastInterval(state)
	if err != nil {
		return nil, err
	}
	if interval == nil {
		return nil, errors.New("broadcast interval is nil")
	}
	return interval, nil
}

//////////////////////////////////////////////////////////////////////////////////////////
//提供共识引擎调用，获取数据的接口
func (dc *cdc) GetCurrentHash() common.Hash {
	return dc.leaderCal.preHash
}

func (dc *cdc) GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error) {
	if (hash == common.Hash{}) {
		return nil, nil, errors.New("输入hash为空")
	}
	if hash == dc.leaderCal.preHash {
		return dc.chain.GetGraphByState(dc.parentState)
	}
	return dc.chain.GetGraphByHash(hash)
}

func (dc *cdc) GetBroadcastAccounts(blockHash common.Hash) ([]common.Address, error) {
	if (blockHash == common.Hash{}) {
		return nil, errors.New("输入hash为空")
	}
	if blockHash == dc.leaderCal.preHash {
		return dc.leaderCal.specialAccounts.broadcasts, nil
	}
	return dc.chain.GetBroadcastAccounts(blockHash)
}

func (dc *cdc) GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	if (blockHash == common.Hash{}) {
		return nil, errors.New("输入hash为空")
	}
	if blockHash == dc.leaderCal.preHash {
		return dc.leaderCal.specialAccounts.versionSupers, nil
	}
	return dc.chain.GetVersionSuperAccounts(blockHash)
}

func (dc *cdc) GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error) {
	if (blockHash == common.Hash{}) {
		return nil, errors.New("输入hash为空")
	}
	if blockHash == dc.leaderCal.preHash {
		return dc.leaderCal.specialAccounts.blockSupers, nil
	}
	return dc.chain.GetBlockSuperAccounts(blockHash)
}

func (dc *cdc) GetBroadcastIntervalByHash(blockHash common.Hash) (*mc.BCIntervalInfo, error) {
	if (blockHash == common.Hash{}) {
		return nil, errors.New("输入hash为空")
	}
	if blockHash == dc.leaderCal.preHash {
		if dc.bcInterval == nil {
			return nil, errors.New("缓存中不存在广播周期信息")
		}
		return dc.bcInterval, nil
	}
	return dc.chain.GetBroadcastIntervalByHash(blockHash)
}

func (dc *cdc) GetSignAccountPassword(signAccounts []common.Address) (common.Address, string, error) {
	return dc.chain.GetSignAccountPassword(signAccounts)
}
func (dc *cdc) GetSignAccounts(authFrom common.Address, blockHash common.Hash) ([]common.Address, error) {
	if blockHash.Equal(common.Hash{}) {
		log.Error(common.SignLog, "获取签名账户阶段", "cdc 最终结果", "输入数据err", "区块hash为空")
		return nil, errors.New("cdc:输入hash为空")
	}

	if blockHash != dc.leaderCal.preHash {
		log.Info(common.SignLog, "获取签名账户阶段", "cdc 最终结果", "调blockchain接口", "")
		return dc.chain.GetSignAccounts(authFrom, blockHash)
	}

	if common.TopAccountType == common.TopAccountA0 {
		//TODO 暂定根据ca提供的接口获取委托账户
	}

	if nil == dc.parentState {
		log.Info(common.SignLog, "获取签名账户阶段", "cdc 最终结果", "err", "dc.parentState是空")
		return nil, errors.New("cdc: parent stateDB is nil, can't reader data")
	}

	height := dc.number - 1
	ans := dc.parentState.GetEntrustFrom(authFrom, height)
	if len(ans) == 0 {
		ans = append(ans, authFrom)
	}
	return ans, nil
}

func (dc *cdc) GetAuthAccount(signAccount common.Address, hash common.Hash) (common.Address, error) {
	if hash.Equal(common.Hash{}) {
		log.ERROR(common.SignLog, "获取委托账户阶段", "cdc 最终结果", "输入的hash err", "hash为空")
		return common.Address{}, errors.New("cdc: 输入hash为空")
	}
	if hash == dc.leaderCal.preHash {
		if nil == dc.parentState {
			log.ERROR(common.SignLog, "获取委托账户阶段", "cdc 最终结果", "dc.parentState err", "preentState is nil")
			return common.Address{}, errors.New("cdc: parent stateDB is nil, can't reader data")
		}

		preHeight := dc.number - 1
		addr := dc.parentState.GetAuthFrom(signAccount, preHeight)
		if addr.Equal(common.Address{}) {
			addr = signAccount
		}
		if common.TopAccountType == common.TopAccountA0 {
			//TODO 利用CA接口将A1转换为A0
		}

		log.Info(common.SignLog, "获取委托账户阶段", "cdc 最终结果", "高度", preHeight, "签名账户", signAccount, "真实账户", addr)
		return addr, nil
	}
	log.Warn(common.SignLog, "获取委托账户阶段", "cdc 最终结果", "采用blockchain的接口 hash", hash.String())
	return dc.chain.GetAuthAccount(signAccount, hash)
}
