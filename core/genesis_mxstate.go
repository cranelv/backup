package core

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

type GenesisMState struct {
	BroadcastNode mc.NodeInfo   `json:"Broadcast"`
	InnerMiner    []mc.NodeInfo `json:"InnerMiner"`
}

func (g *Genesis) setMatrixState(state *state.StateDB) error {
	if err := g.setTopologyToState(state); err != nil {
		return err
	}
	if err := g.setElectToState(state); err != nil {
		return err
	}
	if err := g.setSpecialNodeToState(state); err != nil {
		return err
	}
	return nil
}

func (g *Genesis) setTopologyToState(state *state.StateDB) error {
	if g.NetTopology.Type != common.NetTopoTypeAll {
		return nil
	}
	if len(g.NetTopology.NetTopologyData) == 0 {
		return errors.New("genesis net topology is empty！")
	}

	var newGraph *mc.TopologyGraph = nil
	var err error
	if g.Number == 0 {
		newGraph, err = mc.NewGenesisTopologyGraph(g.Number, g.NetTopology)
		if err != nil {
			return err
		}
	} else {

		data, err := matrixstate.GetDataByState(mc.MSKeyTopologyGraph, state)
		if err != nil {
			return errors.Errorf("get pre topology graph from state err: %v", err)
		}
		preGraph, _ := data.(*mc.TopologyGraph)
		if preGraph == nil {
			return errors.New("pre topology graph is nil")
		}
		newGraph, err = preGraph.Transfer2NextGraph(g.Number, &g.NetTopology)
		if err != nil {
			return err
		}
	}

	if newGraph == nil {
		return errors.New("topology graph is nil")
	}
	return matrixstate.SetDataToState(mc.MSKeyTopologyGraph, newGraph, state)
}

func (g *Genesis) setElectToState(state *state.StateDB) error {
	if len(g.Elect) == 0 {
		return nil
	}

	elect := &mc.ElectGraph{
		Number:    g.Number,
		ElectList: make([]mc.ElectNodeInfo, 0),
		NextElect: make([]mc.ElectNodeInfo, 0),
	}

	minerIndex, backUpMinerIndex, validatorIndex, backUpValidatorIndex := uint16(0), uint16(0), uint16(0), uint16(0)
	for _, item := range g.Elect {
		nodeInfo := mc.ElectNodeInfo{
			Account: item.Account,
			Stock:   item.Stock,
			Type:    item.Type.Transfer2CommonRole(),
		}
		switch item.Type {
		case common.ElectRoleMiner:
			nodeInfo.Position = common.GeneratePosition(minerIndex, item.Type)
			minerIndex++
		case common.ElectRoleMinerBackUp:
			nodeInfo.Position = common.GeneratePosition(backUpMinerIndex, item.Type)
			backUpMinerIndex++
		case common.ElectRoleValidator:
			nodeInfo.Position = common.GeneratePosition(validatorIndex, item.Type)
			validatorIndex++
		case common.ElectRoleValidatorBackUp:
			nodeInfo.Position = common.GeneratePosition(backUpValidatorIndex, item.Type)
			backUpValidatorIndex++
		default:
			nodeInfo.Position = 0
		}
		elect.ElectList = append(elect.ElectList, nodeInfo)
	}

	return matrixstate.SetDataToState(mc.MSKeyElectGraph, elect, state)
}

func (g *Genesis) setSpecialNodeToState(state *state.StateDB) error {
	var specialNodes *mc.MatrixSpecialNode
	if g.Number == 0 {
		if (g.MState.BroadcastNode.Address == common.Address{}) {
			return errors.Errorf("the `broadcast` of genesis is empty")
		}

		specialNodes = &mc.MatrixSpecialNode{}
		specialNodes.BroadcastNode = g.MState.BroadcastNode
		if len(g.MState.InnerMiner) == 0 {
			specialNodes.InnerMinerNode = make([]mc.NodeInfo, 0)
		} else {
			specialNodes.InnerMinerNode = g.MState.InnerMiner
		}
	} else {
		modifyBroad := g.MState.BroadcastNode.Address != common.Address{}
		modifyInner := len(g.MState.InnerMiner) != 0
		if modifyBroad || modifyInner {
			data, err := matrixstate.GetDataByState(mc.MSKeyMatrixNode, state)
			if err != nil {
				return errors.Errorf("get pre special node err: %v", err)
			}
			specialNodes, _ = data.(*mc.MatrixSpecialNode)
			if specialNodes == nil {
				return errors.New("pre special node reflect err")
			}

			if modifyBroad {
				specialNodes.BroadcastNode = g.MState.BroadcastNode
			}
			if modifyInner {
				specialNodes.InnerMinerNode = g.MState.InnerMiner
			}
		}
	}

	if specialNodes != nil {
		return matrixstate.SetDataToState(mc.MSKeyMatrixNode, specialNodes, state)
	} else {
		return nil
	}
}
