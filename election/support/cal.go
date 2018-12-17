// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package support



import (
 	"math/big"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/common"
)


type NodeElect struct {
	NodeList []NodeInfo
	SeqNum      uint64
	RandSeed    *big.Int
	ElectConfig mc.ElectConfigInfo
}


func NewNodeElect()*NodeElect{
	return &NodeElect{}
}

func (self *NodeElect)SetNodeList(vmList []vm.DepositDetail){
	self.NodeList=[]NodeInfo{}
	for _,v:=range vmList{
		self.NodeList=append(self.NodeList,NodeInfo{
			Addr:v.Address,
			WithDraw :v.WithdrawH,
			OnlineTime:v.OnlineTime,
			Deposit :v.Deposit,
			Usable:true,
		})
	}
	self.CheckNodeList()
}
func (self *NodeElect)CheckNodeList(){
	for k,v:=range self.NodeList{
		if nil==v.Deposit{
			self.NodeList[k].Deposit=big.NewInt(DefaultDeposit)
		}
		if nil==v.OnlineTime{
			self.NodeList[k].OnlineTime=big.NewInt(DefaultOnlineTime)
		}
		if nil==v.WithDraw{
			self.NodeList[k].WithDraw=big.NewInt(DefaultWithdrawH)
		}
	}
}
func (self *NodeElect)SetSeqNum(seqNum uint64){
	self.SeqNum=seqNum
}
func (self *NodeElect)SetRandom(randomSeed *big.Int)  {
	self.RandSeed=randomSeed
}
func (self *NodeElect)SetElectConfig(electConfig mc.ElectConfigInfo){
	self.ElectConfig=electConfig
}
func (self *NodeElect)CalcValue()[]Pnormalized {
	return CalcAllValueFunction(self.NodeList)
}

func (self *NodeElect)MakeMinerAns(chosed []Strallyint)*mc.MasterMinerReElectionRsp{
	minerResult:=&mc.MasterMinerReElectionRsp{
	}
	minerResult.SeqNum=self.SeqNum
	for k,v:=range chosed{
		minerResult.MasterMiner=append(minerResult.MasterMiner,MakeElectNode(v.Addr,k,v.Value,common.RoleMiner))
	}
	return minerResult
}



func CalcAllValueFunction(nodelist []NodeInfo) []Pnormalized { //nodelist []Mynode) map[string]float32 {
	var CapitalMap []Pnormalized

	for _, item := range nodelist {
		self := SelfNodeInfo{Address: item.Addr, Stk: float64(item.Deposit.Uint64()), Uptime: int(item.OnlineTime.Uint64()), Tps: int(item.WithDraw.Uint64()), Coef_tps: 0.2, Coef_stk: 0.25}
		value := self.Last_Time() * (self.TPS_POWER()*self.Coef_tps + self.Deposit_stake()*self.Coef_stk)
		flot := float64(value)
		CapitalMap = append(CapitalMap, Pnormalized{Addr: self.Address, Value: flot})
	}
	return CapitalMap
}

func MakeElectNode(address common.Address, Pos int, Stock int, Type common.RoleType) mc.ElectNodeInfo {
	return mc.ElectNodeInfo{
		Account:  address,
		Position: uint16(Pos),
		Stock:    uint16(Stock),
		Type:     Type,
	}
}
func MakeValidatoeTopGenAns(seqnum uint64, VIPNode []Strallyint, master []Strallyint, backup []Strallyint, candiate []Strallyint) *mc.MasterValidatorReElectionRsq {
	ans := &mc.MasterValidatorReElectionRsq{
		SeqNum: seqnum,
	}
	for _, v := range VIPNode {
		ans.MasterValidator = append(ans.MasterValidator, MakeElectNode(v.Addr, len(ans.MasterValidator), DefaultStock, common.RoleValidator))
	}
	for _, v := range master {
		ans.MasterValidator = append(ans.MasterValidator, MakeElectNode(v.Addr, len(ans.MasterValidator), v.Value, common.RoleValidator))
	}
	for _, v := range backup {
		ans.BackUpValidator = append(ans.BackUpValidator, MakeElectNode(v.Addr, len(ans.BackUpValidator), v.Value, common.RoleBackupValidator))
	}
	for _, v := range candiate {
		ans.CandidateValidator = append(ans.CandidateValidator, MakeElectNode(v.Addr, len(ans.CandidateValidator), v.Value, common.RoleCandidateValidator))
	}
	return ans
}

func CheckData(data []vm.DepositDetail) []vm.DepositDetail {
	ans := []vm.DepositDetail{}
	ans = append(ans, data...)
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
