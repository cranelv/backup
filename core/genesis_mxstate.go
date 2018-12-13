package core

import (
	"encoding/binary"

	"github.com/matrix/go-matrix/params/manparams"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

const (
	RewardFullRate = uint64(10000)
)

type GenesisMState struct {
	Broadcast    *mc.NodeInfo           `json:"Broadcast"`
	Foundation   *mc.NodeInfo           `json:"Foundation"`
	InnerMiners  *[]mc.NodeInfo         `json:"InnerMiners"`
	VIPCfg       *[]mc.VIPConfig        `json:"VIPCfg" gencodec:"required"`
	BCICfg       *mc.BCIntervalInfo     `json:"BroadcastInterval" gencodec:"required"`
	LeaderCfg    *mc.LeaderConfig       `json:"LeaderCfg" gencodec:"required"`
	BlkRewardCfg *mc.BlkRewardCfg       `json:"BlkRewardCfg" gencodec:"required"`
	TxsRewardCfg *mc.TxsRewardCfgStruct `json:"TxsRewardCfg" gencodec:"required"`
	LotteryCfg   *mc.LotteryCfgStruct   `json:"LotteryCfg" gencodec:"required"`
	InterestCfg  *mc.InterestCfgStruct  `json:"InterestCfg" gencodec:"required"`
	SlashCfg     *mc.SlashCfgStruct     `json:"SlashCfg" gencodec:"required"`
	EleTimeCfg   *mc.ElectGenTimeStruct `json:"EleTime" gencodec:"required"`
	EleInfoCfg   *mc.ElectConfigInfo    `json:"EleInfo" gencodec:"required"`
}
type GenesisMState1 struct {
	Broadcast    *mc.NodeInfo1          `json:"Broadcast,omitempty"`
	Foundation   *mc.NodeInfo1          `json:"Foundation,omitempty"`
	InnerMiners  *[]mc.NodeInfo1        `json:"InnerMiners,omitempty"`
	BCICfg       *mc.BCIntervalInfo     `json:"BroadcastInterval" gencodec:"required"`
	VIPCfg       *[]mc.VIPConfig        `json:"VIPCfg" ,omitempty"`
	LeaderCfg    *mc.LeaderConfig       `json:"LeaderCfg" ,omitempty"`
	BlkRewardCfg *mc.BlkRewardCfg       `json:"BlkRewardCfg" ,omitempty"`
	TxsRewardCfg *mc.TxsRewardCfgStruct `json:"TxsRewardCfg" ,omitempty"`
	LotteryCfg   *mc.LotteryCfgStruct   `json:"LotteryCfg" ,omitempty"`
	InterestCfg  *mc.InterestCfgStruct  `json:"InterestCfg" ,omitempty"`
	SlashCfg     *mc.SlashCfgStruct     `json:"SlashCfg" ,omitempty"`
	EleTimeCfg   *mc.ElectGenTimeStruct `json:"EleTime" ,omitempty"`
	EleInfoCfg   *mc.ElectConfigInfo    `json:"EleInfo" ,omitempty"`
}

func (ms *GenesisMState) setMatrixState(state *state.StateDB, netTopology common.NetTopology, elect []common.Elect, num uint64) error {
	if err := ms.setElectTime(state, num); err != nil {
		return err
	}

	if err := ms.setElectInfo(state, num); err != nil {
		return err
	}

	if err := ms.setTopologyToState(state, netTopology, num); err != nil {
		return err
	}

	if err := ms.setElectToState(state, elect, num); err != nil {
		return err
	}
	if err := ms.setSpecialNodeToState(state, num); err != nil {
		return err
	}
	if err := ms.setBlkRewardCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setTxsRewardCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setLotteryCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setInterestCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setSlashCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setVIPCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setLeaderCfgToState(state, num); err != nil {
		return err
	}
	if err := ms.setBCIntervalToState(state, num); err != nil {
		return err
	}

	return nil
}

func (g *GenesisMState) setElectTime(state *state.StateDB, num uint64) error {
	if g.EleTimeCfg == nil {
		if num == 0 {
			return errors.New("选举配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置选举信息", "")
			return nil
		}
	}
	if g.EleTimeCfg.ValidatorGen < g.EleTimeCfg.ValidatorNetChange {
		return errors.New("验证者切换点小于验证者生成点")
	}
	if g.EleTimeCfg.MinerGen < g.EleTimeCfg.MinerNetChange {
		return errors.New("矿工切换点小于矿工生效时间点")
	}
	log.Info("Geneis", "electime", g.EleTimeCfg)
	return matrixstate.SetDataToState(mc.MSKeyElectGenTime, g.EleTimeCfg, state)
}
func (g *GenesisMState) setElectInfo(state *state.StateDB, num uint64) error {
	if g.EleInfoCfg == nil {
		if num == 0 {
			return errors.New("electconfig配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置electconfig信息", "")
			return nil
		}
	}

	log.Info("Geneis", "electconfig", g.EleInfoCfg)
	return matrixstate.SetDataToState(mc.MSKeyElectConfigInfo, g.EleInfoCfg, state)
}

func (g *GenesisMState) setTopologyToState(state *state.StateDB, genesisNt common.NetTopology, num uint64) error {
	if genesisNt.Type != common.NetTopoTypeAll {
		return nil
	}
	if len(genesisNt.NetTopologyData) == 0 {
		return errors.New("genesis net topology is empty！")
	}

	var newGraph *mc.TopologyGraph = nil
	var err error
	if num == 0 {
		newGraph, err = mc.NewGenesisTopologyGraph(num, genesisNt)
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
		newGraph, err = preGraph.Transfer2NextGraph(num, &genesisNt)
		if err != nil {
			return err
		}
	}

	if newGraph == nil {
		return errors.New("topology graph is nil")
	}
	return matrixstate.SetDataToState(mc.MSKeyTopologyGraph, newGraph, state)
}

func (g *GenesisMState) setElectToState(state *state.StateDB, gensisElect []common.Elect, num uint64) error {
	if len(gensisElect) == 0 {
		return nil
	}

	elect := &mc.ElectGraph{
		Number:    num,
		ElectList: make([]mc.ElectNodeInfo, 0),
		NextElect: make([]mc.ElectNodeInfo, 0),
	}

	minerIndex, backUpMinerIndex, validatorIndex, backUpValidatorIndex := uint16(0), uint16(0), uint16(0), uint16(0)
	for _, item := range gensisElect {
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

	err := matrixstate.SetDataToState(mc.MSKeyElectGraph, elect, state)
	if err != nil {
		return err
	}

	electOnlineData := &mc.ElectOnlineStatus{
		Number: elect.Number,
	}
	for _, v := range elect.ElectList {
		tt := v
		tt.Position = common.PosOnline
		electOnlineData.ElectOnline = append(electOnlineData.ElectOnline, tt)
	}

	return matrixstate.SetDataToState(mc.MSKeyElectOnlineState, electOnlineData, state)
}

func (g *GenesisMState) setSpecialNodeToState(state *state.StateDB, num uint64) error {
	var specialNodes *mc.MatrixSpecialAccounts
	if num == 0 {
		if (nil == g.Broadcast || g.Broadcast.Address == common.Address{}) {
			return errors.Errorf("the `broadcast` of genesis is empty")
		}

		specialNodes = &mc.MatrixSpecialAccounts{}
		specialNodes.BroadcastAccount = *g.Broadcast
		if nil != g.Foundation {
			specialNodes.FoundationAccount = *g.Foundation
		}
		if nil != g.InnerMiners {
			if len(*g.InnerMiners) == 0 {
				specialNodes.InnerMinerAccounts = make([]mc.NodeInfo, 0)
			} else {
				specialNodes.InnerMinerAccounts = *g.InnerMiners
			}
		}

	} else {
		modifyBroad := g.Broadcast != nil
		modifyFounda := g.Foundation != nil
		modifyInner := g.InnerMiners != nil
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
				specialNodes.BroadcastAccount = *g.Broadcast
			}
			if modifyFounda {
				specialNodes.BroadcastAccount = *g.Foundation
			}
			if modifyInner {
				specialNodes.InnerMinerAccounts = *g.InnerMiners
			}
		}
	}

	if specialNodes != nil {
		log.Info("Geneis", "specialNodes", specialNodes)
		return matrixstate.SetDataToState(mc.MSKeyMatrixAccount, specialNodes, state)
	} else {
		return nil
	}
}

func (g *GenesisMState) setBlkRewardCfgToState(state *state.StateDB, num uint64) error {

	if g.BlkRewardCfg == nil {
		if num == 0 {
			return errors.New("固定区块配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置固定区块配置信息", "")
			return nil
		}
	}

	rateCfg := g.BlkRewardCfg.RewardRate

	if RewardFullRate != rateCfg.MinerOutRate+rateCfg.ElectedMinerRate+rateCfg.FoundationMinerRate {

		return errors.Errorf("矿工固定区块奖励比例配置错误")
	}
	if RewardFullRate != rateCfg.LeaderRate+rateCfg.ElectedValidatorsRate+rateCfg.FoundationValidatorRate {

		return errors.Errorf("验证者固定区块奖励比例配置错误")
	}

	if RewardFullRate != rateCfg.OriginElectOfflineRate+rateCfg.BackupRewardRate {

		return errors.Errorf("替补固定区块奖励比例配置错误")
	}
	log.Info("Geneis", "BlkRewardCfg", g.BlkRewardCfg)
	return matrixstate.SetDataToState(mc.MSKeyBlkRewardCfg, g.BlkRewardCfg, state)
}

func (g *GenesisMState) setTxsRewardCfgToState(state *state.StateDB, num uint64) error {
	if g.TxsRewardCfg == nil {
		if num == 0 {
			return errors.New("交易费区块配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置交易费区块配置信息", "")
			return nil
		}
	}
	rateCfg := g.TxsRewardCfg.RewardRate

	if RewardFullRate != g.TxsRewardCfg.ValidatorsRate+g.TxsRewardCfg.MinersRate {

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
	log.Info("Geneis", "TxsRewardCfg", g.TxsRewardCfg)
	return matrixstate.SetDataToState(mc.MSKeyTxsRewardCfg, g.TxsRewardCfg, state)
}

func (g *GenesisMState) setLotteryCfgToState(state *state.StateDB, num uint64) error {
	if g.LotteryCfg == nil {
		if num == 0 {
			return errors.New("彩票费配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置彩票费配置信息", "")
			return nil
		}
	}
	log.Info("Geneis", "LotteryCfg", g.LotteryCfg)
	return matrixstate.SetDataToState(mc.MSKeyLotteryCfg, g.LotteryCfg, state)
}

func (g *GenesisMState) setInterestCfgToState(state *state.StateDB, num uint64) error {
	if g.InterestCfg == nil {
		if num == 0 {
			return errors.New("利息配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置利息配置信息", "")
			return nil
		}
	}
	StateCfg := g.InterestCfg

	if StateCfg.PayInterval < StateCfg.CalcInterval {

		return errors.Errorf("配置的发放周期小于计息周期")
	}

	log.Info("Geneis", "InterestCfg", g.InterestCfg)
	return matrixstate.SetDataToState(mc.MSKeyInterestCfg, g.InterestCfg, state)
}

func (g *GenesisMState) setSlashCfgToState(state *state.StateDB, num uint64) error {
	if g.SlashCfg == nil {
		if num == 0 {
			return errors.New("惩罚配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置惩罚配置信息", "")
			return nil
		}
	}

	log.Info("Geneis", "SlashCfg", g.SlashCfg)
	return matrixstate.SetDataToState(mc.MSKeySlashCfg, g.SlashCfg, state)
}

func (g *GenesisMState) setVIPCfgToState(state *state.StateDB, number uint64) error {
	if g.VIPCfg == nil {
		if number == 0 {
			return errors.New("VIP配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置惩VIP配置信息", "")
			return nil
		}
	}
	VIPCfg := *g.VIPCfg

	if nil == g.VIPCfg || 0 == len(VIPCfg) {

		return errors.Errorf("vip 配置为nil")
	}

	log.Info("Geneis", "VIPCfg", g.VIPCfg)
	return matrixstate.SetDataToState(mc.MSKeyVIPConfig, g.VIPCfg, state)
}

func (g *GenesisMState) setLeaderCfgToState(state *state.StateDB, num uint64) error {
	if g.LeaderCfg == nil {
		if num == 0 {
			return errors.New("leader配置信息为nil")
		} else {
			log.INFO("Geneis", "没有配置leader配置信息", "")
			return nil
		}
	}
	cfg := g.LeaderCfg
	if cfg.ParentMiningTime <= 0 {
		return errors.Errorf("`ParentMiningTime`(%d) of leader config illegal", cfg.ParentMiningTime)
	}
	if cfg.PosOutTime <= 0 {
		return errors.Errorf("`PosOutTime`(%d) of leader config illegal", cfg.PosOutTime)
	}
	if cfg.ReelectOutTime <= 0 {
		return errors.Errorf("`ReelectOutTime`(%d) of leader config illegal", cfg.ReelectOutTime)
	}
	if cfg.ReelectHandleInterval <= 0 {
		return errors.Errorf("`ReelectHandleInterval`(%d) of leader config illegal", cfg.ReelectHandleInterval)
	}

	log.Info("Geneis", "LeaderCfg", g.LeaderCfg)
	return matrixstate.SetDataToState(mc.MSKeyLeaderConfig, g.LeaderCfg, state)
}

func (g *GenesisMState) SetSuperBlkToState(state *state.StateDB, extra []byte, num uint64) error {
	var superBlkCfg *mc.SuperBlkCfg
	if num == 0 {
		superBlkCfg = &mc.SuperBlkCfg{Seq: 0, Num: 0}
	} else {
		if len(extra) < 8 {
			return errors.New("没有配置超级区块配置信息")
		}

		seq := uint64(binary.BigEndian.Uint64(extra[:8]))

		superBlkCfg = &mc.SuperBlkCfg{Seq: seq, Num: num}
	}
	log.INFO("Geneis", "超级区块配置", superBlkCfg)
	return matrixstate.SetDataToState(mc.MSKeySuperBlockCfg, superBlkCfg, state)
}
func (g *GenesisMState) setBCIntervalToState(state *state.StateDB, num uint64) error {
	var interval *mc.BCIntervalInfo = nil
	if num == 0 {
		if nil == g.BCICfg {
			return errors.New("广播周期配置信息为nil")
		}
		if g.BCICfg.BCInterval < 20 {
			return errors.Errorf("`BCInterval`(%d) of broadcast interval config illegal", g.BCICfg.BCInterval)
		}

		interval = &mc.BCIntervalInfo{
			LastBCNumber:       0,
			LastReelectNumber:  0,
			BCInterval:         g.BCICfg.BCInterval,
			BackupEnableNumber: 0,
			BackupBCInterval:   0,
		}
	} else {
		if nil == g.BCICfg {
			log.INFO("Geneis", "没有配置广播周期配置信息", "")
			return nil
		}
		if g.BCICfg.BackupBCInterval < 20 {
			return errors.Errorf("`BackupBCInterval`(%d) of broadcast interval config illegal", g.BCICfg.BackupBCInterval)
		}
		if g.BCICfg.BackupEnableNumber < num {
			return errors.Errorf("广播周期生效高度(%d)非法, < 当前高度(%d)", g.BCICfg.BackupEnableNumber, num)
		}

		preData, err := matrixstate.GetDataByState(mc.MSKeyBroadcastInterval, state)
		if err != nil {
			return errors.Errorf("获取前广播周期数据失败(%v)", err)
		}

		bcInterval, err := manparams.NewBCIntervalWithInterval(preData)
		if err != nil {
			return errors.Errorf("前广播周期数据异常(%v)", err)
		}

		if bcInterval.GetBroadcastInterval() == g.BCICfg.BackupBCInterval {
			log.INFO("GenesisMState", "广播周期一致，不配置", g.BCICfg.BackupBCInterval)
			return nil
		}

		if bcInterval.IsReElectionNumber(g.BCICfg.BackupBCInterval) {
			return errors.Errorf("生效高度(%d)必须是选举周期, 上个选举高度(%d), 原广播周期(%d)",
				g.BCICfg.BackupBCInterval, bcInterval.GetLastReElectionNumber(), bcInterval.GetBroadcastInterval())
		}

		bcInterval.SetBackupBCInterval(g.BCICfg.BackupBCInterval, g.BCICfg.BackupEnableNumber)
		interval = bcInterval.ToInfoStu()
	}

	if interval != nil {
		return matrixstate.SetDataToState(mc.MSKeyBroadcastInterval, interval, state)
	}
	return nil
}
