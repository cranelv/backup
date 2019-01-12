package core

import (
	"encoding/binary"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
	"sort"
)

const (
	RewardFullRate = uint64(10000)
)

type GenesisMState struct {
	Broadcasts                   *[]common.Address                `json:"Broadcasts"`
	InnerMiners                  *[]common.Address                `json:"InnerMiners"`
	Foundation                   *common.Address                  `json:"Foundation"`
	VersionSuperAccounts         *[]common.Address                `json:"VersionSuperAccounts"`
	BlockSuperAccounts           *[]common.Address                `json:"BlockSuperAccounts"`
	VIPCfg                       *[]mc.VIPConfig                  `json:"VIPCfg" gencodec:"required"`
	BCICfg                       *mc.BCIntervalInfo               `json:"BroadcastInterval" gencodec:"required"`
	LeaderCfg                    *mc.LeaderConfig                 `json:"LeaderCfg" gencodec:"required"`
	BlkRewardCfg                 *mc.BlkRewardCfg                 `json:"BlkRewardCfg" gencodec:"required"`
	TxsRewardCfg                 *mc.TxsRewardCfgStruct           `json:"TxsRewardCfg" gencodec:"required"`
	LotteryCfg                   *mc.LotteryCfgStruct             `json:"LotteryCfg" gencodec:"required"`
	InterestCfg                  *mc.InterestCfgStruct            `json:"InterestCfg" gencodec:"required"`
	SlashCfg                     *mc.SlashCfgStruct               `json:"SlashCfg" gencodec:"required"`
	EleTimeCfg                   *mc.ElectGenTimeStruct           `json:"EleTime" gencodec:"required"`
	EleInfoCfg                   *mc.ElectConfigInfo              `json:"EleInfo" gencodec:"required"`
	ElectMinerNumCfg             *mc.ElectMinerNumStruct          `json:"ElectMinerNum" gencodec:"required"`
	ElectBlackListCfg            *[]common.Address                `json:"ElectBlackList" gencodec:"required"`
	ElectWhiteListCfg            *[]common.Address                `json:"ElectWhiteList" gencodec:"required"`
	CurElect                     *[]common.Elect                  `json:"CurElect"  gencodec:"required"`
	BlockProduceSlashCfg         *mc.BlockProduceSlashCfg         `json:"BlkProduceSlashCfg" gencodec:"required"`
	BlockProduceStats            *mc.BlockProduceStats            `json:"BlkProduceStats" gencodec:"required"`
	BlockProduceSlashBlackList   *mc.BlockProduceSlashBlackList   `json:"BlkProduceBlackList" gencodec:"required"`
	BlockProduceSlashStatsStatus *mc.BlockProduceSlashStatsStatus `json:"BlkProduceStatus" gencodec:"required"`
}
type GenesisMState1 struct {
	Broadcasts                   *[]string                        `json:"Broadcasts,omitempty"`
	InnerMiners                  *[]string                        `json:"InnerMiners,omitempty"`
	Foundation                   *string                          `json:"Foundation,omitempty"`
	VersionSuperAccounts         *[]string                        `json:"VersionSuperAccounts,omitempty"`
	BlockSuperAccounts           *[]string                        `json:"BlockSuperAccounts,omitempty"`
	BCICfg                       *mc.BCIntervalInfo               `json:"BroadcastInterval" gencodec:"required"`
	VIPCfg                       *[]mc.VIPConfig                  `json:"VIPCfg" ,omitempty"`
	LeaderCfg                    *mc.LeaderConfig                 `json:"LeaderCfg" ,omitempty"`
	BlkRewardCfg                 *mc.BlkRewardCfg                 `json:"BlkRewardCfg" ,omitempty"`
	TxsRewardCfg                 *mc.TxsRewardCfgStruct           `json:"TxsRewardCfg" ,omitempty"`
	LotteryCfg                   *mc.LotteryCfgStruct             `json:"LotteryCfg" ,omitempty"`
	InterestCfg                  *mc.InterestCfgStruct            `json:"InterestCfg" ,omitempty"`
	SlashCfg                     *mc.SlashCfgStruct               `json:"SlashCfg" ,omitempty"`
	EleTimeCfg                   *mc.ElectGenTimeStruct           `json:"EleTime" ,omitempty"`
	EleInfoCfg                   *mc.ElectConfigInfo              `json:"EleInfo" ,omitempty"`
	ElectMinerNumCfg             *mc.ElectMinerNumStruct          `json:"ElectMinerNum" gencodec:"required"`
	ElectBlackListCfg            *[]string                        `json:"ElectBlackList" gencodec:"required"`
	ElectWhiteListCfg            *[]string                        `json:"ElectWhiteList" gencodec:"required"`
	CurElect                     *[]common.Elect1                 `json:"curElect"    gencodec:"required"`
	BlockProduceSlashCfg         *mc.BlockProduceSlashCfg         `json:"BlkProduceSlashCfg" gencodec:"required"`
	BlockProduceStats            *mc.BlockProduceStats            `json:"BlkProduceStats" gencodec:"required"`
	BlockProduceSlashBlackList   *mc.BlockProduceSlashBlackList   `json:"BlkProduceBlackList" gencodec:"required"`
	BlockProduceSlashStatsStatus *mc.BlockProduceSlashStatsStatus `json:"BlkProduceStatus" gencodec:"required"`
}

func (ms *GenesisMState) setMatrixState(state *state.StateDB, netTopology common.NetTopology, nextElect []common.Elect, version string, num uint64) error {
	if err := ms.setVersionInfo(state, num, version); err != nil {
		return err
	}

	if err := ms.setElectTime(state, num); err != nil {
		return err
	}

	if err := ms.setElectInfo(state, num); err != nil {
		return err
	}

	if err := ms.setElectMinerNumInfo(state, num); err != nil {
		return err
	}
	if err := ms.setElectBlackListInfo(state, num); err != nil {
		return err
	}
	if err := ms.setElectWhiteListInfo(state, num); err != nil {
		return err
	}

	if err := ms.setTopologyToState(state, netTopology, num); err != nil {
		return err
	}

	if err := ms.setElectToState(state, nextElect, num); err != nil {
		return err
	}

	if err := ms.setBroadcastAccountToState(state, num); err != nil {
		return err
	}
	if err := ms.setInnerMinerAccountsToState(state, num); err != nil {
		return err
	}
	if err := ms.setFoundationAccountToState(state, num); err != nil {
		return err
	}
	if err := ms.setVersionSuperAccountsToState(state, num); err != nil {
		return err
	}
	if err := ms.setBlockSuperAccountsToState(state, num); err != nil {
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

	if err := ms.setBlockProduceSlashStatsStatus(state, num); err != nil {
		return err
	}

	if err := ms.setBlockProduceSlashBlkList(state, num); err != nil {
		return err
	}

	if err := ms.setBlockProduceStats(state, num); err != nil {
		return err
	}

	if err := ms.setBlockProduceSlashCfg(state, num); err != nil {
		return err
	}
	return nil
}

func (g *GenesisMState) setVersionInfo(state *state.StateDB, num uint64, version string) error {
	if len(version) == 0 {
		if num == 0 {
			return errors.New("版本信息为空")
		} else {
			log.INFO("Geneis", "没有配置版本信息", "")
			return nil
		}
	}
	return matrixstate.SetVersionInfo(state, version)
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
	return matrixstate.SetElectGenTime(state, g.EleTimeCfg)
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
	return matrixstate.SetElectConfigInfo(state, g.EleInfoCfg)
}

func (g *GenesisMState) setElectMinerNumInfo(state *state.StateDB, num uint64) error {
	if g.ElectMinerNumCfg == nil {
		if num == 0 {
			return errors.New("electMinerNum为nil")
		} else {
			log.Info("Geneis", "没有配置ElectMinerNumCfg信息", "")
			return nil
		}
	}
	log.Info("Geneis", "ElectMinerNumCfg", g.ElectMinerNumCfg)
	return matrixstate.SetElectMinerNum(state, g.ElectMinerNumCfg)
}

func (g *GenesisMState) setElectWhiteListInfo(state *state.StateDB, num uint64) error {
	var whiteList []common.Address = nil
	if g.ElectWhiteListCfg == nil || *g.ElectWhiteListCfg == nil {
		log.Info("Geneis", "没有配置ElectWhiteListCfg信息", "")
		return nil
	} else {
		whiteList = *g.ElectWhiteListCfg
	}

	log.Info("Geneis", "ElectWhiteListCfg", whiteList)
	return matrixstate.SetElectWhiteList(state, whiteList)
}

func (g *GenesisMState) setElectBlackListInfo(state *state.StateDB, num uint64) error {
	var blackList []common.Address = nil
	if g.ElectBlackListCfg == nil || *g.ElectBlackListCfg == nil {
		log.Info("Geneis", "没有配置ElectBlackListCfg信息", "")
		return nil
	} else {
		blackList = *g.ElectBlackListCfg
	}

	log.Info("Geneis", "ElectBlackListCfg", blackList)
	return matrixstate.SetElectBlackList(state, blackList)
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
		preGraph, err := matrixstate.GetTopologyGraph(state)
		if err != nil {
			return errors.Errorf("get pre topology graph from state err: %v", err)
		}
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
	return matrixstate.SetTopologyGraph(state, newGraph)
}

func (g *GenesisMState) setElectToState(state *state.StateDB, nextElect []common.Elect, num uint64) error {
	var curElect []common.Elect = nil
	if g.CurElect != nil {
		curElect = *g.CurElect
	}

	if len(nextElect) == 0 && len(curElect) == 0 {
		return nil
	}

	elect := &mc.ElectGraph{
		Number:             num,
		ElectList:          make([]mc.ElectNodeInfo, 0),
		NextMinerElect:     make([]mc.ElectNodeInfo, 0),
		NextValidatorElect: make([]mc.ElectNodeInfo, 0),
	}

	minerIndex, validatorIndex, backUpValidatorIndex := uint16(0), uint16(0), uint16(0)
	for _, item := range nextElect {
		nodeInfo := mc.ElectNodeInfo{
			Account: item.Account,
			Stock:   item.Stock,
			Type:    item.Type.Transfer2CommonRole(),
		}
		switch item.Type {
		case common.ElectRoleMiner:
			nodeInfo.Position = common.GeneratePosition(minerIndex, item.Type)
			minerIndex++
			elect.NextMinerElect = append(elect.NextMinerElect, nodeInfo)
		case common.ElectRoleValidator:
			nodeInfo.Position = common.GeneratePosition(validatorIndex, item.Type)
			validatorIndex++
			elect.NextValidatorElect = append(elect.NextValidatorElect, nodeInfo)
		case common.ElectRoleValidatorBackUp:
			nodeInfo.Position = common.GeneratePosition(backUpValidatorIndex, item.Type)
			backUpValidatorIndex++
			elect.NextValidatorElect = append(elect.NextValidatorElect, nodeInfo)
		default:
			nodeInfo.Position = 0
		}
	}

	for _, item := range curElect {
		nodeInfo := mc.ElectNodeInfo{
			Account: item.Account,
			Stock:   item.Stock,
			Type:    item.Type.Transfer2CommonRole(),
		}
		switch item.Type {
		case common.ElectRoleMiner:
			nodeInfo.Position = common.GeneratePosition(minerIndex, item.Type)
			minerIndex++
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

	err := matrixstate.SetElectGraph(state, elect)
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

	return matrixstate.SetElectOnlineState(state, electOnlineData)
}

func (g *GenesisMState) setBroadcastAccountToState(state *state.StateDB, num uint64) error {
	if g.Broadcasts == nil || len(*g.Broadcasts) == 0 {
		if num == 0 {
			return errors.Errorf("the `broadcast` of genesis is empty")
		} else {
			return nil
		}
	}
	return matrixstate.SetBroadcastAccounts(state, *g.Broadcasts)
}

func (g *GenesisMState) setInnerMinerAccountsToState(state *state.StateDB, num uint64) error {
	var innerMiners []common.Address = nil
	if g.InnerMiners == nil || *g.InnerMiners == nil {
		if num == 0 {
			innerMiners = make([]common.Address, 0)
		} else {
			return nil
		}
	} else {
		innerMiners = *g.InnerMiners
	}

	matrixstate.SetInnerMinerAccounts(state, innerMiners)
	return nil
}

func (g *GenesisMState) setFoundationAccountToState(state *state.StateDB, num uint64) error {
	var foundation common.Address
	if g.Foundation == nil || *g.Foundation == (common.Address{}) {
		if num == 0 {
			foundation = common.Address{}
		} else {
			return nil
		}
	} else {
		foundation = *g.Foundation
	}
	matrixstate.SetFoundationAccount(state, foundation)
	return nil
}

func (g *GenesisMState) setVersionSuperAccountsToState(state *state.StateDB, num uint64) error {
	if g.VersionSuperAccounts == nil || len(*g.VersionSuperAccounts) == 0 {
		if num == 0 {
			return errors.Errorf("the version superAccounts of genesis is empty")
		} else {
			return nil
		}
	}
	matrixstate.SetVersionSuperAccounts(state, *g.VersionSuperAccounts)
	return nil
}

func (g *GenesisMState) setBlockSuperAccountsToState(state *state.StateDB, num uint64) error {
	if num != 0 {
		return errors.New("the block superAccounts can't modify")
	}
	if g.BlockSuperAccounts == nil || len(*g.BlockSuperAccounts) == 0 {
		return errors.Errorf("the block superAccounts of genesis is empty")
	}
	matrixstate.SetBlockSuperAccounts(state, *g.BlockSuperAccounts)
	return nil
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
	return matrixstate.SetBlkRewardCfg(state, g.BlkRewardCfg)
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
	return matrixstate.SetTxsRewardCfg(state, g.TxsRewardCfg)
}

func (g *GenesisMState) setLotteryCfgToState(state *state.StateDB, num uint64) error {
	if num == 0 {
		if g.LotteryCfg == nil {
			return errors.New("利息配置信息为nil")
		}
		matrixstate.SetLotteryNum(state, 1)
	} else {
		if g.LotteryCfg == nil {
			log.INFO("Geneis", "没有配置利息配置信息", "")
			return nil
		}
	}
	log.Info("Geneis", "LotteryCfg", g.LotteryCfg)
	return matrixstate.SetLotteryCfg(state, g.LotteryCfg)
}

func (g *GenesisMState) setInterestCfgToState(state *state.StateDB, num uint64) error {
	if num == 0 {
		if g.InterestCfg == nil {
			return errors.New("利息配置信息为nil")
		}
		matrixstate.SetInterestCalcNum(state, 1)
		matrixstate.SetInterestPayNum(state, 1)
	} else {
		if g.InterestCfg == nil {
			log.INFO("Geneis", "没有配置利息配置信息", "")
			return nil
		}
	}
	StateCfg := g.InterestCfg

	if StateCfg.PayInterval < StateCfg.CalcInterval {

		return errors.Errorf("配置的发放周期小于计息周期")
	}

	log.Info("Geneis", "InterestCfg", g.InterestCfg)
	return matrixstate.SetInterestCfg(state, g.InterestCfg)
}

func (g *GenesisMState) setSlashCfgToState(state *state.StateDB, num uint64) error {
	if num == 0 {
		if g.SlashCfg == nil {
			return errors.New("惩罚配置信息为nil")
		}
		matrixstate.SetSlashNum(state, 1)
		matrixstate.SetUpTimeNum(state, 1)
	} else {
		if g.SlashCfg == nil {
			log.INFO("Geneis", "没有配置惩罚配置信息", "")
			return nil
		}
	}
	log.Info("Geneis", "SlashCfg", g.SlashCfg)
	return matrixstate.SetSlashCfg(state, g.SlashCfg)
}

type SortVIPConfig []mc.VIPConfig

func (self SortVIPConfig) Len() int {
	return len(self)
}
func (self SortVIPConfig) Less(i, j int) bool {
	return self[i].MinMoney < self[j].MinMoney
}
func (self SortVIPConfig) Swap(i, j int) {
	temp := self[i]
	self[i] = self[j]
	self[j] = temp
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
	if nil == g.VIPCfg || 0 == len(*g.VIPCfg) {

		return errors.Errorf("vip 配置为nil")
	}
	sort.Sort(SortVIPConfig(*g.VIPCfg))
	if (*g.VIPCfg)[0].MinMoney != uint64(0) {
		return errors.New("vip配置中需包含最小值为0的配置")
	}
	for index := 0; index < len(*g.VIPCfg)-1; index++ {
		if (*g.VIPCfg)[index].MinMoney == (*g.VIPCfg)[index+1].MinMoney {
			return errors.New("vip配置中不能包含最小值相同的配置")
		}
	}

	log.Info("Geneis", "VIPCfg", *g.VIPCfg)
	return matrixstate.SetVIPConfig(state, *g.VIPCfg)
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
	return matrixstate.SetLeaderConfig(state, g.LeaderCfg)
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
	return matrixstate.SetSuperBlockCfg(state, superBlkCfg)
}

func (g *GenesisMState) setBCIntervalToState(st *state.StateDB, num uint64) error {
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

		bcInterval, err := matrixstate.GetBroadcastInterval(st)
		if err != nil || bcInterval == nil {
			return errors.Errorf("获取前广播周期数据失败(%v)", err)
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
		interval = bcInterval
	}

	if interval != nil {
		return matrixstate.SetBroadcastInterval(st, interval)
	}
	return nil
}

func (g *GenesisMState) setBlockProduceSlashCfg(state *state.StateDB, num uint64) error {
	if num == 0 {
		if g.BlockProduceSlashCfg == nil {
			return errors.New("区块生产惩罚配置信息为nil")
		}
	} else {
		if g.BlockProduceSlashCfg == nil {
			log.INFO("Geneis", "未修改区块生产惩罚配置信息为", "")
			return nil
		}
	}
	log.Info("Geneis", "BlockProduceSlashCfg", g.BlockProduceSlashCfg)
	return matrixstate.SetBlockProduceSlashCfg(state, g.BlockProduceSlashCfg)
}
func (g *GenesisMState) setBlockProduceStats(state *state.StateDB, num uint64) error {
	if num == 0 {
		if g.BlockProduceStats == nil {
			return nil
		}
	} else {
		if g.BlockProduceStats == nil {
			log.INFO("Geneis", "未修改区块生产惩罚统计信息", "")
			return nil
		}
	}
	log.Info("Geneis", "BlockProduceStats", g.BlockProduceStats)
	return matrixstate.SetBlockProduceStats(state, g.BlockProduceStats)
}
func (g *GenesisMState) setBlockProduceSlashBlkList(state *state.StateDB, num uint64) error {
	if num == 0 {
		if g.BlockProduceSlashBlackList == nil {
			return nil
		}
	} else {
		if g.BlockProduceSlashBlackList == nil {
			log.INFO("Geneis", "未修改区块生产惩黑名单", "")
			return nil
		}
	}
	log.Info("Geneis", "BlockProduceBlackList", g.BlockProduceSlashBlackList)
	return matrixstate.SetBlockProduceBlackList(state, g.BlockProduceSlashBlackList)
}
func (g *GenesisMState) setBlockProduceSlashStatsStatus(state *state.StateDB, num uint64) error {
	if num == 0 {
		if g.BlockProduceSlashStatsStatus == nil {
			return errors.New("区块生产惩状态信息为nil")
		}
	} else {
		if g.BlockProduceSlashStatsStatus == nil {
			log.INFO("Geneis", "未修改区块生产状态信息", "")
			return nil
		}
	}
	log.Info("Geneis", "BlockProduceSlashStatsStatus", g.BlockProduceSlashStatsStatus)
	return matrixstate.SetBlockProduceStatsStatus(state, g.BlockProduceSlashStatsStatus)
}
