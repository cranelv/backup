// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/mc"
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
	if err := dc.leaderCal.SetValidators(preHash, preIsSupper, preLeader, validators); err != nil {
		return err
	}

	consensusLeader, err := dc.GetLeader(dc.curConsensusTurn)
	if err != nil {
		return err
	}
	if dc.curReelectTurn != 0 {
		reelectLeader, err := dc.GetLeader(dc.curConsensusTurn + dc.curReelectTurn)
		if err != nil {
			return err
		}
		dc.reelectMaster.Set(reelectLeader)
	} else {
		dc.reelectMaster.Set(common.Address{})
	}
	dc.consensusLeader.Set(consensusLeader)
	dc.parentState = parentState
	dc.role = role
	return nil
}

func (dc *cdc) readValidatorsAndRoleFromState(state *state.StateDB) ([]mc.TopologyNodeInfo, common.RoleType, error) {
	topology, _, err := dc.chain.GetGraphByState(state)
	if err != nil {
		return nil, common.RoleNil, err
	}

	if topology.Number+1 != dc.number {
		return nil, common.RoleNil, errors.Errorf("state中的拓扑图高度不匹配，state number（%d） + 1 != local number(%d)", topology.Number, dc.number)
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

func (dc *cdc) SetConsensusTurn(consensusTurn uint32) error {
	consensusLeader, err := dc.GetLeader(consensusTurn)
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
	master, err := dc.GetLeader(dc.curConsensusTurn + reelectTurn)
	if err != nil {
		return errors.Errorf("获取master错误(%v), 重选轮次(%d), 共识轮次(%d)", err, reelectTurn, dc.curConsensusTurn)
	}
	dc.reelectMaster.Set(master)
	dc.curReelectTurn = reelectTurn
	return nil
}

func (dc *cdc) GetLeader(turn uint32) (common.Address, error) {
	leaders, err := dc.leaderCal.GetLeader(turn)
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
	leaders, err := dc.leaderCal.GetLeader(dc.curConsensusTurn + dc.curReelectTurn)
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
