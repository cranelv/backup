// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package stock

import (
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
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
	return support.MinerTopGen(mmrerm)
}

func (self *StockElect) ValidatorTopGen(mvrerm *mc.MasterValidatorReElectionReqMsg) *mc.MasterValidatorReElectionRsq {
	MaxValidator := int(mvrerm.ElectConfig.ValidatorNum)
	MaxBackValidator := int(mvrerm.ElectConfig.BackValidator)
	log.INFO("选举种子", "验证者拓扑生成", len(mvrerm.ValidatorList))
	validatorList := support.CheckData(mvrerm.ValidatorList)

	var master, backup, candiate []support.Strallyint
	var value []support.Stf
	if len(mvrerm.FoundationValidatorList) == 0 {
		value = support.CalcAllValueFunction(validatorList)
		master, backup, candiate = support.ValNodesSelected(value, mvrerm.RandSeed.Int64(), MaxValidator, MaxBackValidator, 0) //mvrerm.RandSeed.Int64(), 11, 5, 0) //0x12217)
	} else {
		value = support.CalcAllValueFunction(validatorList)
		valuefound := support.CalcAllValueFunction(mvrerm.FoundationValidatorList)
		master, backup, candiate = support.ValNodesSelected(value, mvrerm.RandSeed.Int64(), MaxValidator, MaxBackValidator, len(mvrerm.FoundationValidatorList)) //0x12217)
		master = support.CommbineFundNodesAndPricipal(value, valuefound, master, 0.25, 4.0)
	}
	return support.MakeValidatoeTopGenAns(mvrerm.SeqNum, []support.Strallyint{}, master, backup, candiate)
}

func (self *StockElect) ToPoUpdate(allNative support.AllNative, topoG *mc.TopologyGraph) []mc.Alternative {

	return support.ToPoUpdate(allNative, topoG)
}

func (self *StockElect) PrimarylistUpdate(Q0, Q1, Q2 []mc.TopologyNodeInfo, online mc.TopologyNodeInfo, flag int) ([]mc.TopologyNodeInfo, []mc.TopologyNodeInfo, []mc.TopologyNodeInfo) {
	return support.PrimarylistUpdate(Q0, Q1, Q2, online, flag)
}
