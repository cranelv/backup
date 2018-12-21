// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package layered

import (
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
)

type layered struct {
}

func init() {
	baseinterface.RegElectPlug(manparams.ElectPlug_layerd, RegInit)
}

func RegInit() baseinterface.ElectionInterface {
	return &layered{}
}

func (self *layered) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	log.INFO("分层方案", "矿工拓扑生成", len(mmrerm.MinerList))
	nodeElect := support.NewElelection(nil, mmrerm.MinerList, mmrerm.ElectConfig, mmrerm.RandSeed, mmrerm.SeqNum)
	nodeElect.Disorder()
	nodeElect.Sort()
	nodeElect.ProcessBlackNode()
	nodeElect.ProcessWhiteNode()

	value := nodeElect.GetWeight(common.RoleMiner)

	Master, value := support.GetList(value, int(nodeElect.EleCfg.MinerNum)-len(nodeElect.WhiteNodeInfo), nodeElect.RandSeed.Int64())

	Master = append(Master, nodeElect.WhiteNodeInfo...)

	return support.MakeMinerAns(Master, nodeElect.SeqNum)

}

func (self *layered) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	log.INFO("分层方案", "验证者拓扑生成", len(mvrerm.ValidatorList))
	vipEle := support.NewElelection(mvrerm.VIPList, mvrerm.ValidatorList, mvrerm.ElectConfig, mvrerm.RandSeed, mvrerm.SeqNum)
	vipEle.ProcessBlackNode()
	vipEle.ProcessWhiteNode()

	var maxVipEleLevelNum = support.MaxVipEleLevelNum
	if maxVipEleLevelNum > len(vipEle.VipLevelCfg) {
		maxVipEleLevelNum = len(vipEle.VipLevelCfg)
	}
	var MasterList = make([]support.Node, 0)
	for vipEleLoop := 0; vipEleLoop < maxVipEleLevelNum; vipEleLoop++ {
		if vipEle.VipLevelCfg[vipEleLoop].ElectUserNum <= 0 {
			continue
		}
		nodeList := vipEle.GetNodeByLevel(vipEleLoop)
		electedNode := vipEle.VipElection(nodeList, int(vipEle.VipLevelCfg[vipEleLoop].ElectUserNum))
		MasterList = append(MasterList, electedNode...)
	}

	MasterChosed := TransVIPNode(MasterList)
	MasterChosed = append(MasterChosed, vipEle.WhiteNodeInfo...)

	Master, Backup, Candidate := vipEle.ValidatorTopGen(int(vipEle.EleCfg.ValidatorNum)-len(MasterChosed), int(vipEle.EleCfg.BackValidator))

	return support.MakeValidatoeTopGenAns(mvrerm.SeqNum, MasterChosed, Master, Backup, Candidate)
}

func TransVIPNode(vipnode []support.Node) []support.Strallyint {
	ans := []support.Strallyint{}
	for _, v := range vipnode {
		ans = append(ans, support.Strallyint{Value: support.DefaultStock, Addr: v.Address})
	}
	return ans
}
func (self *layered) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	return support.ToPoUpdate(allNative, topoG)
}

func (self *layered) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
