package core

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

const (
	RewardFullRate = uint64(10000)
)

type GenesisMState struct {
	Broadcast    mc.NodeInfo           `json:"Broadcast"`
	Foundation   mc.NodeInfo           `json:"Foundation"`
	InnerMiners  []mc.NodeInfo         `json:"InnerMiners"`
	VIPCfg       []mc.VIPConfig        `json:"VIPCfg" gencodec:"required"`
	BlkRewardCfg mc.BlkRewardCfg       `json:"BlkRewardCfg" gencodec:"required"`
	TxsRewardCfg mc.TxsRewardCfgStruct `json:"TxsRewardCfg" gencodec:"required"`
	LotteryCfg   mc.LotteryCfgStruct   `json:"LotteryCfg" gencodec:"required"`
	InterestCfg  mc.InterestCfgStruct  `json:"InterestCfg" gencodec:"required"`
	SlashCfg     mc.SlashCfgStruct     `json:"SlashCfg" gencodec:"required"`
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
	if err := g.setBlkRewardCfgToState(state); err != nil {
		return err
	}
	if err := g.setTxsRewardCfgToState(state); err != nil {
		return err
	}
	if err := g.setLotteryCfgToState(state); err != nil {
		return err
	}
	if err := g.setInterestCfgToState(state); err != nil {
		return err
	}
	if err := g.setSlashCfgToState(state); err != nil {
		return err
	}
	if err := g.setVIPCfgToState(state); err != nil {
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
	var specialNodes *mc.MatrixSpecialAccounts
	if g.Number == 0 {
		if (g.MState.Broadcast.Address == common.Address{}) {
			return errors.Errorf("the `broadcast` of genesis is empty")
		}

		specialNodes = &mc.MatrixSpecialAccounts{}
		specialNodes.BroadcastAccount = g.MState.Broadcast
		specialNodes.FoundationAccount = g.MState.Foundation
		if len(g.MState.InnerMiners) == 0 {
			specialNodes.InnerMinerAccounts = make([]mc.NodeInfo, 0)
		} else {
			specialNodes.InnerMinerAccounts = g.MState.InnerMiners
		}
	} else {
		modifyBroad := g.MState.Broadcast.Address != common.Address{}
		modifyFounda := g.MState.Foundation.Address != common.Address{}
		modifyInner := len(g.MState.InnerMiners) != 0
		if modifyBroad || modifyFounda || modifyInner {
			data, err := matrixstate.GetDataByState(mc.MSKeyMatrixAccount, state)
			if err != nil {
				return errors.Errorf("get pre special node err: %v", err)
			}
			specialNodes, _ = data.(*mc.MatrixSpecialAccounts)
			if specialNodes == nil {
				return errors.New("pre special node reflect err")
			}

			if modifyBroad {
				specialNodes.BroadcastAccount = g.MState.Broadcast
			}
			if modifyFounda {
				specialNodes.BroadcastAccount = g.MState.Foundation
			}
			if modifyInner {
				specialNodes.InnerMinerAccounts = g.MState.InnerMiners
			}
		}
	}

	if specialNodes != nil {
		return matrixstate.SetDataToState(mc.MSKeyMatrixAccount, specialNodes, state)
	} else {
		return nil
	}
}

func (g *Genesis) setBlkRewardCfgToState(state *state.StateDB) error {
	rateCfg := g.MState.BlkRewardCfg.RewardRate

	if RewardFullRate != rateCfg.MinerOutRate+rateCfg.ElectedMinerRate+rateCfg.FoundationMinerRate {

		return errors.Errorf("矿工固定区块奖励比例配置错误")
	}
	if RewardFullRate != rateCfg.LeaderRate+rateCfg.ElectedValidatorsRate+rateCfg.FoundationValidatorRate {

		return errors.Errorf("验证者固定区块奖励比例配置错误")
	}

	if RewardFullRate != rateCfg.OriginElectOfflineRate+rateCfg.BackupRewardRate {

		return errors.Errorf("替补固定区块奖励比例配置错误")
	}
	return matrixstate.SetDataToState(mc.MSKeyBlkRewardCfg, g.MState.BlkRewardCfg, state)
}

func (g *Genesis) setTxsRewardCfgToState(state *state.StateDB) error {
	rateCfg := g.MState.TxsRewardCfg.RewardRate

	if RewardFullRate != g.MState.TxsRewardCfg.ValidatorsRate+g.MState.TxsRewardCfg.MinersRate {

		return errors.Errorf("交易奖励比例配置错误")
	}

	if RewardFullRate != rateCfg.MinerOutRate+rateCfg.ElectedMinerRate+rateCfg.FoundationMinerRate {

		return errors.Errorf("矿工固定区块奖励比例配置错误")
	}
	if RewardFullRate != rateCfg.LeaderRate+rateCfg.ElectedValidatorsRate+rateCfg.FoundationValidatorRate {

		return errors.Errorf("验证者固定区块奖励比例配置错误")
	}

	if RewardFullRate != rateCfg.OriginElectOfflineRate+rateCfg.BackupRewardRate {

		return errors.Errorf("替补固定区块奖励比例配置错误")
	}
	return matrixstate.SetDataToState(mc.MSKeyTxsRewardCfg, g.MState.TxsRewardCfg, state)
}

func (g *Genesis) setLotteryCfgToState(state *state.StateDB) error {
	return matrixstate.SetDataToState(mc.MSKeyLotteryCfg, g.MState.LotteryCfg, state)
}

func (g *Genesis) setInterestCfgToState(state *state.StateDB) error {
	StateCfg := g.MState.InterestCfg

	if StateCfg.PayInterval < StateCfg.CalcInterval {

		return errors.Errorf("配置的发放周期小于计息周期")
	}

	return matrixstate.SetDataToState(mc.MSKeyInterestCfg, g.MState.InterestCfg, state)
}

func (g *Genesis) setSlashCfgToState(state *state.StateDB) error {
	return matrixstate.SetDataToState(mc.MSKeySlashCfg, g.MState.SlashCfg, state)
}

func (g *Genesis) setVIPCfgToState(state *state.StateDB) error {
	VIPCfg := g.MState.VIPCfg

	if 0 == len(VIPCfg) {

		return errors.Errorf("vip 配置为nil")
	}

	return matrixstate.SetDataToState(mc.MSKeyVIPConfig, g.MState.VIPCfg, state)
}
