// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/mc"
)

func MinerTopGen(mmrerm *mc.MasterMinerReElectionReqMsg) *mc.MasterMinerReElectionRsp {
	for i, item := range mmrerm.MinerList {
		if item.Deposit == nil {
			mmrerm.MinerList[i].Deposit = big.NewInt(DefaultDeposit)
		}
		if item.WithdrawH == nil {
			mmrerm.MinerList[i].WithdrawH = big.NewInt(DefaultWithdrawH)
		}
		if item.OnlineTime == nil {
			mmrerm.MinerList[i].OnlineTime = big.NewInt(DefaultOnlineTime)
		}
	}

	value := CalcAllValueFunction(mmrerm.MinerList)
	Master, _ := MinerNodesSelected(value, mmrerm.RandSeed.Int64(), mmrerm.ElectConfig ) //Ele.Engine(value, mmrerm.RandSeed.Int64()) //0x12217)


	MinerEleRs:=&mc.MasterMinerReElectionRsp{
		SeqNum:mmrerm.SeqNum,
	}

	for index, item := range Master {
		MinerEleRs.MasterMiner = append(MinerEleRs.MasterMiner, MakeElectNode(item.Addr,index,item.Value,common.RoleMiner))
	}

	return MinerEleRs
}

func CalcAllValueFunction(nodelist []vm.DepositDetail) []Stf { //nodelist []Mynode) map[string]float32 {
	var CapitalMap []Stf

	for _, item := range nodelist {
		self := SelfNodeInfo{Address: item.Address, Stk: float64(item.Deposit.Uint64()), Uptime: int(item.OnlineTime.Uint64()), Tps: int(item.WithdrawH.Uint64()), Coef_tps: 0.2, Coef_stk: 0.25}
		value := self.Last_Time() * (self.TPS_POWER()*self.Coef_tps + self.Deposit_stake()*self.Coef_stk)
		CapitalMap = append(CapitalMap, Stf{Addr: self.Address, Flot: float64(value)})
	}
	return CapitalMap
}

func MakeElectNode(address common.Address,Pos int,Stock int,Type common.RoleType)mc.ElectNodeInfo{
	return mc.ElectNodeInfo{
		Account:address,
		Position:uint16(Pos),
		Stock:uint16(Stock),
		Type:Type,
	}
}
func MakeValidatoeTopGenAns(seqnum uint64,VIPNode []Strallyint,master []Strallyint, backup []Strallyint, candiate []Strallyint )*mc.MasterValidatorReElectionRsq{
	ans:=&mc.MasterValidatorReElectionRsq{
		SeqNum:seqnum,
	}
	for _,v:=range VIPNode{
		ans.MasterValidator=append(ans.MasterValidator,MakeElectNode(v.Addr,len(ans.MasterValidator),DefaultStock,common.RoleValidator))
	}
	for _,v:=range master{
		ans.MasterValidator=append(ans.MasterValidator,MakeElectNode(v.Addr,len(ans.MasterValidator),v.Value,common.RoleValidator))
	}
	for _,v:=range backup{
		ans.BackUpValidator=append(ans.BackUpValidator,MakeElectNode(v.Addr,len(ans.BackUpValidator),v.Value,common.RoleBackupValidator))
	}
	for _,v:=range candiate{
		ans.CandidateValidator=append(ans.CandidateValidator,MakeElectNode(v.Addr,len(ans.CandidateValidator),v.Value,common.RoleCandidateValidator))
	}
	return ans
}

func CheckData(data []vm.DepositDetail)[]vm.DepositDetail{
	ans:=[]vm.DepositDetail{}
	ans=append(ans,data...)
	for i, item := range ans {
		if item.Deposit == nil {
			ans[i].Deposit = big.NewInt(DefaultDeposit)
		}
		if item.WithdrawH == nil {
			ans[i].WithdrawH = big.NewInt(DefaultWithdrawH)
		}
		if item.OnlineTime == nil {
			ans[i].OnlineTime = big.NewInt(DefaultOnlineTime)
		}
	}
	return ans
}
