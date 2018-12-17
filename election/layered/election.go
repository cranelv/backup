// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package layered

import (
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"fmt"
)

type layered struct {
}

func init() {
	baseinterface.RegElectPlug("layered", RegInit)
}

func RegInit() baseinterface.ElectionInterface {
	return &layered{}
}

func (self *layered) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	log.INFO("分层方案", "矿工拓扑生成", len(mmrerm.MinerList))
	nodeElect:=support.NewNodeElect()
	nodeElect.SetNodeList(mmrerm.MinerList)
	nodeElect.SetSeqNum(mmrerm.SeqNum)
	nodeElect.SetRandom(mmrerm.RandSeed)
	nodeElect.SetElectConfig(mmrerm.ElectConfig)

	value:=nodeElect.CalcValue()
	Master,value:=support.GetList(value,int(nodeElect.ElectConfig.MinerNum),nodeElect.RandSeed.Int64())

	return nodeElect.MakeMinerAns(Master)

}

func (self *layered) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	log.INFO("分层方案", "验证者拓扑生成", len(mvrerm.ValidatorList))
	vipEle := NewVipElelection(mvrerm.VIPList, mvrerm.ValidatorList, mvrerm.ElectConfig, mvrerm.RandSeed)
	vipEle.ProcessBlackNode()
	vipEle.ProcessWhiteNode()
	//vipEle.DisPlayNode()

	var maxVipEleLevelNum = MaxVipEleLevelNum
	if maxVipEleLevelNum > len(vipEle.VipLevelCfg) {
		maxVipEleLevelNum = len(vipEle.VipLevelCfg)
	}
	var MasterList = make([]vip_node, 0)
	for vipEleLoop := 0; vipEleLoop < maxVipEleLevelNum; vipEleLoop++ {
		if vipEle.VipLevelCfg[vipEleLoop].ElectUserNum <= 0 {
			continue
		}
		nodeList := vipEle.GetNodeByLevel(vipEleLoop)
		electedNode := vipEle.vipElection(nodeList, int(vipEle.VipLevelCfg[vipEleLoop].ElectUserNum))
		MasterList = append(MasterList, electedNode...)
		vipEle.LastMasterNum -= len(electedNode)
	}

	MasterChosed := make([]vip_node, 0)
	MasterChosed = append(MasterChosed, vipEle.WhiteNodeInfo...)
	MasterChosed = append(MasterChosed, MasterList...)
	lastNode := vipEle.GetLastNode(MasterChosed)
	weight := vipEle.GetWeight(lastNode)
	//fmt.Println("len",len(weight))
	for _,v:=range MasterChosed{
		fmt.Println(v.Address.String())
	}



	Master,weight:=support.GetList(weight,vipEle.LastMasterNum,vipEle.randSeed.Int64())

	Backup,weight:=support.GetList(weight,int(vipEle.EleCfg.BackValidator),vipEle.randSeed.Int64())

	Candidate,weight:=support.GetList(weight,len(weight),vipEle.randSeed.Int64())
	vipNode := TransVIPNode(MasterChosed)
	return support.MakeValidatoeTopGenAns(mvrerm.SeqNum, vipNode, Master, Backup, Candidate)

}

func TransVIPNode(vipnode []vip_node) []support.Strallyint {
	ans := []support.Strallyint{}
	for _, v := range vipnode {
		ans = append(ans, support.Strallyint{Value: DefaultStock, Addr: v.Address})
	}
	return ans
}
func (self *layered) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	return support.ToPoUpdate(allNative, topoG)
}

func (self *layered) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
