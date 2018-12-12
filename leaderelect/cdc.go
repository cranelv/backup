// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/pkg/errors"
)

type cdc struct {
	state            stateDef
	number           uint64
	role             common.RoleType
	curConsensusTurn uint32
	consensusLeader  common.Address
	curReelectTurn   uint32
	reelectMaster    common.Address
	isMaster         bool
	leaderCal        *leaderCalculator
	bcInterval       *manparams.BCInterval
	parentState      *state.StateDB
	turnTime         *turnTimes
	chain            *core.BlockChain
	logInfo          string
}

func newCDC(number uint64, chain *core.BlockChain, logInfo string) *cdc {
	dc := &cdc{
		state:            stIdle,
		number:           number,
		role:             common.RoleNil,
		curConsensusTurn: 0,
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

	dc.leaderCal = newLeaderCalculator(chain, dc)
	return dc
}

func (dc *cdc) AnalysisState(preHash common.Hash, preIsSupper bool, preLeader common.Address, parentState *state.StateDB) error {
	if parentState == nil {
		return errors.New("parent state is nil")
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

	if err := dc.leaderCal.SetValidatorsAndSpecials(preHash, preIsSupper, preLeader, validators, specials, bcInterval); err != nil {
		return err
	}

	consensusLeader, err := dc.GetLeader(dc.curConsensusTurn, bcInterval)
	if err != nil {
		return err
	}
	if dc.curReelectTurn != 0 {
		reelectLeader, err := dc.GetLeader(dc.curConsensusTurn+dc.curReelectTurn, bcInterval)
		if err != nil {
			return err
		}
		dc.reelectMaster.Set(reelectLeader)
	} else {
		dc.reelectMaster.Set(common.Address{})
	}
	if err := dc.turnTime.SetTimeConfig(config); err != nil {
		log.Error(dc.logInfo, "设置时间配置参数失败", err)
	}
	dc.bcInterval = bcInterval
	dc.consensusLeader.Set(consensusLeader)
	dc.parentState = parentState
	dc.role = role
	return nil
}

func (dc *cdc) SetConsensusTurn(consensusTurn uint32) error {
	consensusLeader, err := dc.GetLeader(consensusTurn, dc.bcInterval)
	if err != nil {
		return errors.Errorf("获取共识leader错误(%v), 共识轮次(%d)", err, consensusTurn)
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
	master, err := dc.GetLeader(dc.curConsensusTurn+reelectTurn, dc.bcInterval)
	if err != nil {
		return errors.Errorf("获取master错误(%v), 重选轮次(%d), 共识轮次(%d)", err, reelectTurn, dc.curConsensusTurn)
	}
	dc.reelectMaster.Set(master)
	dc.curReelectTurn = reelectTurn
	return nil
}

func (dc *cdc) GetLeader(turn uint32, bcInterval *manparams.BCInterval) (common.Address, error) {
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
	leaders, err := dc.leaderCal.GetLeader(dc.curConsensusTurn+dc.curReelectTurn, dc.bcInterval)
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

func (dc *cdc) GetSpecialAccounts(blockHash common.Hash) (*mc.MatrixSpecialAccounts, error) {
	if (blockHash == common.Hash{}) {
		return nil, errors.New("输入hash为空")
	}
	if blockHash == dc.leaderCal.preHash {
		return dc.leaderCal.specials, nil
	}
	return dc.chain.GetSpecialAccounts(blockHash)
}

func (dc *cdc) GetBroadcastInterval(blockHash common.Hash) (*mc.BCIntervalInfo, error) {
	if (blockHash == common.Hash{}) {
		return nil, errors.New("输入hash为空")
	}
	if blockHash == dc.leaderCal.preHash {
		if dc.bcInterval == nil {
			return nil, errors.New("缓存中不存在广播周期信息")
		}
		return dc.bcInterval.ToInfoStu(), nil
	}
	return dc.chain.GetBroadcastInterval(blockHash)
}

func (dc *cdc) GetAuthAccount(addr common.Address, hash common.Hash) (common.Address, error) {
	if hash.Equal(common.Hash{}) {
		log.Error("cdc", "GetSignAccount", "输入hash为空")
		return common.Address{}, errors.New("输入hash为空")
	}
	if hash == dc.leaderCal.preHash {
		if nil == dc.parentState {
			log.Error(dc.logInfo, "GetSignAccount", "parentStateDB为空")
			return common.Address{}, errors.New("cdc state can't find")
		}
		preHeight := dc.number - 1
		authAddr, err := dc.chain.GetAuthAddr(addr, preHeight, dc.parentState)
		log.Info("cdc", "preHeight", preHeight, "addr", addr, "signAddr", authAddr, "err", err)
		return authAddr, err
	}
	return dc.chain.GetAuthAccount(addr, hash)
}

func (dc *cdc) readValidatorsAndRoleFromState(state *state.StateDB) ([]mc.TopologyNodeInfo, common.RoleType, error) {
	topology, _, err := dc.chain.GetGraphByState(state)
	if err != nil {
		return nil, common.RoleNil, err
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
	selfAccount := ca.GetAddress()
	for _, v := range TopologyGraph.NodeList {
		if v.Account == selfAccount {
			return v.Type
		}
	}
	return common.RoleNil
}

func (dc *cdc) readSpecialAccountsFromState(state *state.StateDB) (*mc.MatrixSpecialAccounts, error) {
	data, err := matrixstate.GetDataByState(mc.MSKeyMatrixAccount, state)
	if err != nil {
		return nil, err
	}
	specials, OK := data.(*mc.MatrixSpecialAccounts)
	if OK == false {
		return nil, errors.New("反射MatrixSpecialAccounts失败")
	}
	if specials == nil {
		return nil, errors.New("MatrixSpecialAccounts == nil")
	}
	return specials, nil
}

func (dc *cdc) readLeaderConfigFromState(state *state.StateDB) (*mc.LeaderConfig, error) {
	data, err := matrixstate.GetDataByState(mc.MSKeyLeaderConfig, state)
	if err != nil {
		return nil, err
	}
	config, OK := data.(*mc.LeaderConfig)
	if OK == false {
		return nil, errors.New("反射LeaderConfig失败")
	}
	if config == nil {
		return nil, errors.New("LeaderConfig == nil")
	}
	return config, nil
}

func (dc *cdc) readBroadCastIntervalFromState(state *state.StateDB) (*manparams.BCInterval, error) {
	data, err := matrixstate.GetDataByState(mc.MSKeyBroadcastInterval, state)
	if err != nil {
		return nil, err
	}
	return manparams.NewBCIntervalWithInterval(data)
}
