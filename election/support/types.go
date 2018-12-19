// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"fmt"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/mc"
	"math/big"
	"math/rand"
	"sort"
)

const (
	MaxSample    = 1000 //配置参数,采样最多发生1000次,是一个离P+M较远的值
	J            = 0    //基金会验证节点个数tps_weight
	DefaultStock = 1
)

const (
	DefaultNodeConfig       = 0
	MaxVipEleLevelNum       = 2
	DefaultRatio            = 1000
	DefaultRatioDenominator = 1000
)

type AllNative struct {
	Master    []mc.ElectNodeInfo //验证者主节点
	BackUp    []mc.ElectNodeInfo //验证者备份
	Candidate []mc.ElectNodeInfo //验证者候选

	MasterQ    []common.Address //第一梯队候选
	BackUpQ    []common.Address //第二梯队候选
	CandidateQ []common.Address //第三梯队候选
	ElectInfo  *mc.ElectConfigInfo
}

type Strallyint struct {
	Value int
	Addr  common.Address
}

type Node struct {
	Address    common.Address
	Deposit    *big.Int
	WithdrawH  *big.Int
	OnlineTime *big.Int
	Ratio      uint16
	vipLevel   int
	index      int
	Usable     bool
}

type Electoion struct {
	SeqNum        uint64
	RandSeed      *big.Int
	MaxLevelNum   int
	VipLevelCfg   []mc.VIPConfig
	NodeList      []Node
	EleCfg        mc.ElectConfigInfo
	WhiteNodeInfo []Strallyint
}

func (node *Node) SetUsable(status bool) {
	node.Usable = status
}

func (node *Node) SetIndex(index int) {
	node.index = index
}
func (node *Node) SetVipLevelInfo(VipLevelCfg []mc.VIPConfig) {
	for index := 0; index < len(VipLevelCfg); index++ {
		temp := big.NewInt(0).Set(node.Deposit)
		deposMan := temp.Div(temp, common.ManValue).Uint64()
		if deposMan >= VipLevelCfg[index].MinMoney {
			node.vipLevel = index
			node.Ratio = VipLevelCfg[index].StockScale
			return
		}
	}
	maxLevel := len(VipLevelCfg)
	node.Ratio = DefaultRatio
	node.vipLevel = maxLevel
}

func (node *Node) SetDepositInfo(depsit vm.DepositDetail) {
	node.Address = depsit.Address
	node.OnlineTime = depsit.OnlineTime
	node.WithdrawH = depsit.WithdrawH
	node.Deposit = depsit.Deposit

	if nil == depsit.Deposit {
		node.Deposit = big.NewInt(DefaultNodeConfig)
	}
	if nil == depsit.WithdrawH {
		node.WithdrawH = big.NewInt(DefaultNodeConfig)
	}
	if nil == depsit.OnlineTime {
		node.OnlineTime = big.NewInt(DefaultNodeConfig)
	}
}

func NewElelection(VipLevelCfg []mc.VIPConfig, vm []vm.DepositDetail, EleCfg mc.ElectConfigInfo, randseed *big.Int, seqNum uint64) *Electoion {
	var vip Electoion
	vip.SeqNum = seqNum
	vip.RandSeed = randseed
	vip.MaxLevelNum = len(VipLevelCfg) + 1
	vip.EleCfg = EleCfg
	vip.VipLevelCfg = VipLevelCfg
	vip.WhiteNodeInfo = make([]Strallyint, 0)
	for i := 0; i < len(vm); i++ {
		vip.NodeList = append(vip.NodeList, Node{})
	}
	for i := 0; i < len(vm); i++ {
		vip.NodeList[i].SetDepositInfo(vm[i])
		vip.NodeList[i].SetVipLevelInfo(VipLevelCfg)
		vip.NodeList[i].SetIndex(i)
		vip.NodeList[i].SetUsable(true)

	}
	return &vip
}

func FindAddress(addr common.Address, addrList []common.Address) bool {
	for _, v := range addrList {
		if v.Equal(addr) == true {
			return true
		}
	}
	return false
}
func (vip *Electoion) DisPlayNode() {
	for _, v := range vip.NodeList {
		fmt.Println(v.Address, v.Deposit, v.WithdrawH, v.OnlineTime, v.vipLevel, v.index, "Ratio", v.Ratio)
	}
}
func (vip *Electoion) ProcessBlackNode() {
	for k, v := range vip.NodeList {
		if FindAddress(v.Address, vip.EleCfg.BlackList) {
			vip.NodeList[k].SetUsable(false)
		}
	}
}
func (vip *Electoion) ProcessWhiteNode() {
	for k, v := range vip.NodeList {
		if v.Usable == false {
			continue
		}
		if FindAddress(v.Address, vip.EleCfg.WhiteList) {
			vip.WhiteNodeInfo = append(vip.WhiteNodeInfo, Strallyint{Addr: v.Address, Value: DefaultStock})
			vip.NodeList[k].SetUsable(false)
		}
	}
}
func (vip *Electoion) GetNodeByLevel(level int) []Node {
	specialNode := make([]Node, 0)
	for i := 0; i < len(vip.NodeList); i++ {
		if vip.NodeList[i].Usable == false {
			continue
		}
		if level == vip.NodeList[i].vipLevel {
			specialNode = append(specialNode, vip.NodeList[i])
		}
	}
	return specialNode
}

func (vip *Electoion) GetNodeIndexByLevel(level int) []int {
	specialNode := make([]int, 0)
	for i := 0; i < len(vip.NodeList); i++ {
		if level == vip.NodeList[i].vipLevel {
			specialNode = append(specialNode, i)
		}
	}
	return specialNode
}

func (vip *Electoion) GetLastNode() []Node {
	var remainNodeList = make([]Node, 0)
	for i := 0; i < len(vip.NodeList); i++ {
		if vip.NodeList[i].Usable == false {
			continue
		}
		remainNodeList = append(remainNodeList, vip.NodeList[i])
	}
	return remainNodeList
}

func (vip *Electoion) GetWeight() []Pnormalized {
	lastnode := vip.GetLastNode()
	var CapitalMap []Pnormalized
	for _, item := range lastnode {
		self := SelfNodeInfo{Address: item.Address, Stk: item.Deposit, Uptime: int(item.OnlineTime.Uint64()), Tps: 1000, Coef_tps: 0.2, Coef_stk: 0.25}
		value := self.Last_Time() * (self.TPS_POWER()*self.Coef_tps + self.Deposit_stake()*self.Coef_stk)
		value = value * (float64(item.Ratio) / float64(DefaultRatioDenominator))
		CapitalMap = append(CapitalMap, Pnormalized{Addr: self.Address, Value: float64(value)})
	}
	return CapitalMap
}

func (vip *Electoion) Disorder() {
	vip.NodeList = Knuth_Fisher_Yates_Algorithm(vip.NodeList, vip.RandSeed)
}
func (vip *Electoion) Sort() {
	sort.Sort(SortNodeList(vip.NodeList))
}
func Knuth_Fisher_Yates_Algorithm(nodeList []Node, randSeed *big.Int) []Node {
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
func (vip *Electoion) GetIndex(addr common.Address) (int, bool) {
	for k, v := range vip.NodeList {
		if v.Address.Equal(addr) {
			return k, true
		}
	}
	return 0, false
}
func (vip *Electoion) VipElection(nodeList []Node, maxNum int) []Node {

	nodeList = Knuth_Fisher_Yates_Algorithm(nodeList, vip.RandSeed)
	sort.Sort(SortNodeList(nodeList))
	var vipElected = make([]Node, 0)
	for _, v := range nodeList {
		vipElected = append(vipElected, v)
		index, flag := vip.GetIndex(v.Address)
		if flag == false {
			continue
		}
		vip.NodeList[index].SetUsable(false)
		if len(vipElected) >= maxNum {
			return vipElected
		}
	}

	return vipElected
}

type SortNodeList []Node

func (self SortNodeList) Len() int {
	return len(self)
}
func (self SortNodeList) Less(i, j int) bool {
	if self[i].Deposit.Cmp(self[j].Deposit) == 0 {
		return self[i].OnlineTime.Cmp(self[j].OnlineTime) > 0
	}
	return self[i].Deposit.Cmp(self[j].Deposit) > 0
}
func (self SortNodeList) Swap(i, j int) {
	temp := self[i]
	self[i] = self[j]
	self[j] = temp
}
