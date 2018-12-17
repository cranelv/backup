// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/mt19937"

)
//func ValNodesSelected(probVal []Stf, seed int64, M int, P int, J int) ([]Strallyint, []Strallyint, []Strallyint) {
//
//	probnormalized := Normalize(probVal)
//
//	Master,probnormalized:=GetList(probnormalized,M,seed)
//	Backup,probnormalized:=GetList(probnormalized,P,seed)
//	Candid:=[]Strallyint{}
//	for _,v:=range probnormalized{
//		Candid=append(Candid,Strallyint{Addr:v.Addr,Value:1})
//	}
////	fmt.Println("len Master",len(Master),"len Backup",len(Backup),"Cand",len(Candid))
//	return Master,Backup,Candid
//
//}

func GetList(probnormalized []Pnormalized,needNum int,seed int64)([]Strallyint,[]Pnormalized){
	probnormalized =Normalize(probnormalized)
	//fmt.Println("len",len(probnormalized),needNum)
	ans:=[]Strallyint{}
	RemainingProbNormalizedNodes:=[]Pnormalized{}
	if needNum>=len(probnormalized){
		for _,v:=range probnormalized{
			ans=append(ans,Strallyint{Addr:v.Addr,Value:1})
		}
		return ans,[]Pnormalized{}
	}
	rand := mt19937.RandUniformInit(seed)
	dict := make(map[common.Address]int)
	for i := 0; i < MaxSample; i++ {
		node := Sample1NodesInValNodes(probnormalized, float64(rand.Uniform(0.0, 1.0)))

		_, ok := dict[node]
		if ok == true {
			dict[node] = dict[node] + 1
		} else {
			dict[node] = 1
		}

		if len(dict) == (needNum) {
			break
		}
	}
	for _, item := range probnormalized {
		_, ok := dict[item.Addr]
		if ok == false {
			RemainingProbNormalizedNodes = append(RemainingProbNormalizedNodes, Pnormalized{Addr: item.Addr, Value: item.Value})
		} else {
			ans = append(ans, Strallyint{Addr: item.Addr, Value: dict[item.Addr]})
		}
	}
	return ans,RemainingProbNormalizedNodes
}




type Pnormalized struct {
	Value float64
	Addr  common.Address
}

func Normalize(probVal []Pnormalized) []Pnormalized {

	var total float64
	for _, item := range probVal {
		total += item.Value
	}
	var pnormalizedlist []Pnormalized
	for _, item := range probVal {
		var tmp Pnormalized
		tmp.Value = item.Value / total
		tmp.Addr = item.Addr
		pnormalizedlist = append(pnormalizedlist, tmp)
	}
	return pnormalizedlist
}

func Sample1NodesInValNodes(probnormalized []Pnormalized, rand01 float64) common.Address {

	for _, iterm := range probnormalized {
		rand01 -= iterm.Value
		if rand01 < 0 {
			return iterm.Addr
		}
	}
	return probnormalized[0].Addr
}




type SelfNodeInfo struct {
	Address  common.Address
	Stk      float64
	Uptime   int
	Tps      int
	Coef_tps float64
	Coef_stk float64
}

func (self *SelfNodeInfo) TPS_POWER() float64 {
	tps_weight := 1.0
	if self.Tps >= 16000 {
		tps_weight = 5.0
	} else if self.Tps >= 8000 {
		tps_weight = 4.0
	} else if self.Tps >= 4000 {
		tps_weight = 3.0
	} else if self.Tps >= 2000 {
		tps_weight = 2.0
	} else if self.Tps >= 1000 {
		tps_weight = 1.0
	} else {
		tps_weight = 0.0
	}
	return tps_weight
}

func (self *SelfNodeInfo) Last_Time() float64 {
	CandidateTime_weight := 4.0
	if self.Uptime <= 64 {
		CandidateTime_weight = 0.25
	} else if self.Uptime <= 128 {
		CandidateTime_weight = 0.5
	} else if self.Uptime <= 256 {
		CandidateTime_weight = 1
	} else if self.Uptime <= 512 {
		CandidateTime_weight = 2
	} else {
		CandidateTime_weight = 4
	}
	return CandidateTime_weight
}

func (self *SelfNodeInfo) Deposit_stake() float64 {
	stake_weight := 1.0
	if self.Stk >= 40000 {
		stake_weight = 4.5
	} else if self.Stk >= 20000 {
		stake_weight = 2.15
	} else if self.Stk >= 10000 {
		stake_weight = 1.0
	} else {
		stake_weight = 0.0
	}
	return stake_weight
}
