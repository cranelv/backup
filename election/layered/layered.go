// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package layered

import (
	"sort"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"math/big"
	"math/rand"
	"github.com/matrix/go-matrix/election/support"
	"fmt"
)

const (
	DefaultStock=0
	DefaultNodeConfig = 0
	MaxVipEleLevelNum = 2
	DefaultRatio=1
	DefaultRatioDenominator=1000
)

type vip_node struct{
	Address    common.Address
	Deposit    *big.Int
	WithdrawH  *big.Int
	OnlineTime *big.Int
	Ratio uint16
	vipLevel int
	index int
}

type VIP_Electoion struct{
	randSeed * big.Int
	MaxLevelNum int
	VipLevelCfg []common.Echelon
	VipNodeInfo []vip_node
	EleCfg  support.Eletion_cfg
}

func (node * vip_node)SetIndex(index int){
	node.index = index
}
func (node * vip_node)SetVipLevelInfo(VipLevelCfg []common.Echelon){

	for vipLevel, vipConfigInfo := range VipLevelCfg {
		if node.Deposit.Cmp(vipConfigInfo.MinMoney) >= 0 {
			node.vipLevel = vipLevel
			node.Ratio=vipConfigInfo.Ratio
			return
		}
	}

	maxLevel := len(VipLevelCfg)
	node.Ratio=DefaultRatio
	node.vipLevel = maxLevel
}

func (node  * vip_node)SetDepositInfo(depsit vm.DepositDetail){
	node.Address = depsit.Address
	node.OnlineTime=depsit.OnlineTime
	node.WithdrawH=depsit.WithdrawH
	node.Deposit=depsit.Deposit

	if nil==depsit.Deposit{
		node.Deposit=big.NewInt(DefaultNodeConfig)
	}
	if nil==depsit.WithdrawH{
		node.WithdrawH=big.NewInt(DefaultNodeConfig)
	}
	if nil==depsit.OnlineTime{
		node.OnlineTime=big.NewInt(DefaultNodeConfig)
	}
}

func NewVipElelection(VipLevelCfg []common.Echelon, vm []vm.DepositDetail, EleCfg  support.Eletion_cfg, randseed * big.Int) * VIP_Electoion{
	var vip VIP_Electoion

	vip.randSeed = randseed
	vip.MaxLevelNum = len(VipLevelCfg) + 1
	vip.EleCfg = EleCfg
	vip.VipLevelCfg = VipLevelCfg

	for i:=0;i<len(vm);i++{
		vip.VipNodeInfo=append(vip.VipNodeInfo,vip_node{})
	}

	for i := 0; i < len(vm); i++{

		vip.VipNodeInfo[i].SetDepositInfo(vm[i])
		vip.VipNodeInfo[i].SetVipLevelInfo(VipLevelCfg)
		vip.VipNodeInfo[i].SetIndex(i)
	}

	return &vip
}


func (vip *VIP_Electoion)DisPlayNode(){
	for _,v:=range vip.VipNodeInfo{
		fmt.Println(v.Address,v.Deposit,v.WithdrawH,v.OnlineTime,v.vipLevel,v.index,v.Ratio)
	}
}
func (vip * VIP_Electoion)GetNodeByLevel(level int) []vip_node{

	specialNode := make([]vip_node,0)

	for i:=0; i < len(vip.VipNodeInfo);i++{
		if level == vip.VipNodeInfo[i].vipLevel{
			specialNode = append(specialNode, vip.VipNodeInfo[i])
		}
	}
	return specialNode
}


func (vip * VIP_Electoion)GetNodeIndexByLevel(level int) []int{
	specialNode := make([]int,0)

	for i:=0; i < len(vip.VipNodeInfo);i++{
		if level == vip.VipNodeInfo[i].vipLevel{
			specialNode = append(specialNode, i)
		}
	}

	return specialNode
}

func (vip * VIP_Electoion)GetLastNode(nodelist []vip_node) []vip_node{
	var originIdx = make([]int , len(vip.VipNodeInfo))
	for i:= 0; i < len(vip.VipNodeInfo); i++{
		originIdx[i] = i
	}

	for i:=0; i < len(nodelist); i++{
		originIdx[nodelist[i].index] = -1
	}

	var remainNodeList = make([]vip_node, 0)
//	fmt.Println("len",len(vip.VipNodeInfo) - len(nodelist))
	for i := 0; i < len(vip.VipNodeInfo); i++{
		//fmt.Println("i",i,"vip",vip.VipNodeInfo[i])
		if originIdx[i] == -1{
			continue
		}
		//fmt.Println("vip.VipNodeInfo[originIdx[i]]",vip.VipNodeInfo[originIdx[i]])
		remainNodeList = append(remainNodeList, vip.VipNodeInfo[originIdx[i]])
		//fmt.Println("len",len(remainNodeList),remainNodeList)
	}
	//fmt.Println("remainBodeList",remainNodeList)
	return remainNodeList
}

func (vip *VIP_Electoion)GetWeight(lastnode []vip_node)[]support.Stf{
	var CapitalMap []support.Stf

	for _, item := range lastnode {
		self := support.SelfNodeInfo{Address: item.Address, Stk: float64(item.Deposit.Uint64()), Uptime: int(item.OnlineTime.Uint64()), Tps: 1000, Coef_tps: 0.2, Coef_stk: 0.25}
		value := self.Last_Time() * (self.TPS_POWER()*self.Coef_tps + self.Deposit_stake()*self.Coef_stk)
		value=value*float64(item.Ratio/DefaultRatioDenominator)
		CapitalMap = append(CapitalMap, support.Stf{Addr: self.Address, Flot: float64(value)})
	}
	return CapitalMap
}

type VipNodeList []vip_node
func (self VipNodeList) Len() int {
	return len(self)
}

func (self VipNodeList) Less(i, j int) bool {
	if self[i].Deposit.Cmp(self[j].Deposit) == 0{
		return self[i].OnlineTime.Cmp(self[j].OnlineTime) > 0
	}

	return self[i].Deposit.Cmp(self[j].Deposit) > 0
}
func (self  VipNodeList) Swap(i, j int) {
	temp := self[i]
	self[i] = self[j]
	self[j] = temp
}

func Knuth_Fisher_Yates_Algorithm( nodeList []vip_node, randSeed *big.Int) []vip_node {
	//高纳德置乱算法
	rand.Seed(randSeed.Int64())
	for index := len(nodeList) - 1; index > 0; index-- {
		aimIndex := rand.Intn(index + 1)
		t := nodeList[index]
		nodeList[index] = nodeList[aimIndex]
		nodeList[aimIndex] = t
	}
	return nodeList
}
func vipElection( nodeList []vip_node, random *big.Int, maxNum int) []vip_node{

	nodeList = Knuth_Fisher_Yates_Algorithm(nodeList, random)

	sort.Sort(VipNodeList(nodeList))

	var vipElected = make([]vip_node, 0)
	if len(nodeList) <= maxNum{
		copy(vipElected, nodeList)
	}else{
		copy(vipElected, nodeList[:maxNum-1])
	}

	return vipElected
}
