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
	"fmt"
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
	vipEle := support.NewElelection(nil, mmrerm.MinerList, mmrerm.ElectConfig, mmrerm.RandSeed, mmrerm.SeqNum,common.RoleMiner)

	vipEle.ProcessBlackNode()
	vipEle.ProcessWhiteNode()


	nodeList := vipEle.GetNodeByLevel(common.VIP_Nil)
	value:=support.CalcValue(nodeList, common.RoleMiner)

	Chosed, value := support.GetList(value, vipEle.NeedNum, vipEle.RandSeed.Int64())


	return support.MakeMinerAns(Chosed, vipEle.SeqNum)

}

func (self *layered) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	log.INFO("分层方案", "验证者拓扑生成", mvrerm.ValidatorList)
	vipEle := support.NewElelection(mvrerm.VIPList, mvrerm.ValidatorList, mvrerm.ElectConfig, mvrerm.RandSeed, mvrerm.SeqNum,common.RoleValidator)
	vipEle.ProcessBlackNode()
	vipEle.ProcessWhiteNode()
	vipEle.DisPlayNode()

	for vipEleLoop := 0; vipEleLoop < len(vipEle.VipLevelCfg); vipEleLoop++ {
		if vipEle.VipLevelCfg[vipEleLoop].ElectUserNum <= 0 &&vipEleLoop!=len(vipEle.VipLevelCfg)-1{
			continue
		}
		nodeList := vipEle.GetNodeByLevel(common.GetVIPLevel(vipEleLoop,len(vipEle.VipLevelCfg)))

		value:=support.CalcValue(nodeList, common.RoleValidator)
		curNeed:=0
		fmt.Println(vipEle.NeedNum,vipEle.ChosedNum,vipEle.VipLevelCfg[vipEleLoop].ElectUserNum)
		if vipEleLoop==len(vipEle.VipLevelCfg)-1{
			fmt.Println("111")
			curNeed=vipEle.NeedNum-vipEle.ChosedNum
		}else{
			fmt.Println("12122")
			curNeed=int(vipEle.VipLevelCfg[vipEleLoop].ElectUserNum)
		}
		fmt.Println("len(nodelist",len(nodeList),"level",common.GetVIPLevel(vipEleLoop,len(vipEle.VipLevelCfg)),"curNeed",curNeed)
		Chosed, value := support.GetList(value, curNeed, vipEle.RandSeed.Int64())
		fmt.Println("vipEleLoop","value",len(value),vipEleLoop,"Chosed",len(Chosed),"curNeed",curNeed,"vipEle.ChoseNum",vipEle.ChosedNum)
		for _,v:=range Chosed{
			fmt.Println("---",v.Addr.String(),v.Value)
		}
		vipEle.SetChosed(Chosed)


	}

	fmt.Println("+++++++++++++")
	for k,v:=range vipEle.HasChosedNode{
		for _,vv:=range v{
			fmt.Println("k",k,"vvvvv",vv.Addr.String())
		}
	}
	fmt.Println("++++++++++++++++++")


	Master:=[]support.Strallyint{}
	Backup:=[]support.Strallyint{}
	Candidate:=[]support.Strallyint{}

	for k,v:=range vipEle.HasChosedNode{
		for _,vv:=range v{
			temp:=support.Strallyint{}
			if k==len(vipEle.HasChosedNode)-1{
				temp=support.Strallyint{Addr:vv.Addr,Value:vv.Value,VIPLevel:common.VIP_Nil}
			}else{
				temp=support.Strallyint{Addr:vv.Addr,Value:vipEle.GetVipStock(vv.Addr),VIPLevel:common.GetVIPLevel(k,len(vipEle.VipLevelCfg))}
			}

			if len(Master)<int(vipEle.EleCfg.ValidatorNum){
				Master=append(Master,temp)
				continue
			}
			if len(Backup)<int(vipEle.EleCfg.BackValidator){
				Backup=append(Backup,temp)
				continue
			}
		}
	}

	lastNode:=vipEle.GetLastNode()
	value:=support.CalcValue(lastNode, common.RoleValidator)
	for _,v:=range lastNode{
		if len(Candidate)<=int(4*vipEle.EleCfg.ValidatorNum-vipEle.EleCfg.BackValidator){
			Candidate=append(Candidate,support.Strallyint{Addr:v.Address,Value:1})
		}
	}

	Candidate, _ = support.GetList(value, int(4*vipEle.EleCfg.ValidatorNum-vipEle.EleCfg.BackValidator), vipEle.RandSeed.Int64())



	return support.MakeValidatoeTopGenAns(mvrerm.SeqNum, Master, Backup, Candidate)
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
