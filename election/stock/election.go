// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package stock

import (
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"fmt"
)

type StockElect struct {
}

func init() {
	baseinterface.RegElectPlug("stock", RegInit)
}

func RegInit() baseinterface.ElectionInterface {

	return &StockElect{}
}

func (self *StockElect) MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	log.INFO("选举种子", "矿工拓扑生成", len(mmrerm.MinerList))
	nodeElect:=support.NewNodeElect()
	nodeElect.SetNodeList(mmrerm.MinerList)
	nodeElect.SetSeqNum(mmrerm.SeqNum)
	nodeElect.SetRandom(mmrerm.RandSeed)
	nodeElect.SetElectConfig(mmrerm.ElectConfig)

	value:=nodeElect.CalcValue()
	Master,_:=support.GetList(value,int(nodeElect.ElectConfig.MinerNum),nodeElect.RandSeed.Int64())

	return nodeElect.MakeMinerAns(Master)
}

func (self *StockElect) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	log.INFO("选举种子", "验证者拓扑生成", len(mvrerm.ValidatorList))
	nodeElect:=support.NewNodeElect()
	nodeElect.SetNodeList(mvrerm.ValidatorList)
	nodeElect.SetSeqNum(mvrerm.SeqNum)
	nodeElect.SetRandom(mvrerm.RandSeed)
	nodeElect.SetElectConfig(mvrerm.ElectConfig)

	value:=nodeElect.CalcValue()
	for _,v:=range value{
		fmt.Println(v.Value,v.Addr.String())
	}
	Master,value:=support.GetList(value,int(nodeElect.ElectConfig.ValidatorNum),nodeElect.RandSeed.Int64())

	BackUp,value:=support.GetList(value,int(nodeElect.ElectConfig.BackValidator),nodeElect.RandSeed.Int64())

	Candid,value:=support.GetList(value,len(value),nodeElect.RandSeed.Int64())

	return support.MakeValidatoeTopGenAns(mvrerm.SeqNum, []support.Strallyint{}, Master, BackUp, Candid)

}

func (self *StockElect) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {

	return support.ToPoUpdate(allNative, topoG)
}

func (self *StockElect) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
