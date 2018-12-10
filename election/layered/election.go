// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package layered

import (
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/common"
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
	return support.MinerTopGen(mmrerm)

}

func (self *layered) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	log.INFO("分层方案", "验证者拓扑生成", len(mvrerm.ValidatorList))
	eleCfg:=support.GetElectCfg()
	vipEle := NewVipElelection(common.EchelonArrary, mvrerm.ValidatorList, eleCfg, mvrerm.RandSeed)
	//vipEle.DisPlayNode()

	//VIP特殊选举的层数
	var maxVipEleLevelNum = MaxVipEleLevelNum
	if maxVipEleLevelNum > len(vipEle.VipLevelCfg){
		maxVipEleLevelNum = len(vipEle.VipLevelCfg)
	}

	var dispatchedList = make([]vip_node, 0)
	for vipEleLoop := 0; vipEleLoop < maxVipEleLevelNum; vipEleLoop++{
		if vipEle.VipLevelCfg[vipEleLoop].MaxNum <= 0{
			continue
		}
		nodeList := vipEle.GetNodeByLevel(vipEleLoop)
		electedNode := vipElection(nodeList, vipEle.randSeed, vipEle.VipLevelCfg[vipEleLoop].MaxNum)
		dispatchedList = append(dispatchedList, electedNode ...)
	}
	//fmt.Println("dispatch",dispatchedList)
	lastNode := vipEle.GetLastNode(dispatchedList)
//	fmt.Println("lastNode",lastNode)
	weight:=vipEle.GetWeight(lastNode)
	Master,Backup,Candidate := support.ValNodesSelected(weight, mvrerm.RandSeed.Int64(), int(eleCfg.MaxValidatorNum)-len(dispatchedList), int(eleCfg.MaxBackUpValidatorNum), 0) //mvrerm.RandSeed.Int64(), 11, 5, 0) //0x12217)

	vipNode:=TransVIPNode(dispatchedList)
	return support.MakeValidatoeTopGenAns(mvrerm.SeqNum,vipNode,Master,Backup,Candidate)

}


func TransVIPNode(vipnode []vip_node)[]support.Strallyint{
	ans:=[]support.Strallyint{}
	for _,v:=range vipnode{
		ans=append(ans,support.Strallyint{Value:DefaultStock,Addr:v.Address})
	}
	return ans
}
func (self *layered) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {
	return support.ToPoUpdate(allNative, topoG)
}

func (self *layered) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
