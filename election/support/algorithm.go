// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package support

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/common/mt19937"
)

func GetList(probnormalized []Pnormalized, needNum int, seed int64) ([]Strallyint, []Pnormalized) {
	probnormalized = Normalize(probnormalized)
	if len(probnormalized)==0{
		return []Strallyint{},probnormalized
	}
	if needNum>len(probnormalized){
		needNum=len(probnormalized)
	}
	ChoseNode:=[]Strallyint{}
	RemainingProbNormalizedNodes:=[]Pnormalized{}
	rand := mt19937.RandUniformInit(seed)
	dict := make(map[common.Address]int)
	orderAddress:=[]common.Address{}
	for i := 0; i < MaxSample; i++ {
		node := Sample1NodesInValNodes(probnormalized, float64(rand.Uniform(0.0, 1.0)))
		_, ok := dict[node]
		if ok == true {
			dict[node] = dict[node] + 1
		} else {
			dict[node] = 1
			orderAddress=append(orderAddress,node)
		}
		if len(dict) == (needNum) {
			break
		}
	}


	for _,v:=range orderAddress{
		ChoseNode=append(ChoseNode,Strallyint{Addr:v,Value:dict[v]})
	}

	for _, item := range probnormalized {
		if _,ok:=dict[item.Addr];ok==true{
			continue
		}
		if len(ChoseNode)<needNum{
			ChoseNode=append(ChoseNode,Strallyint{Addr:item.Addr,Value:1})
		}else{
			RemainingProbNormalizedNodes=append(RemainingProbNormalizedNodes,item)
		}
	}

	return ChoseNode, RemainingProbNormalizedNodes
}

func Normalize(probVal []Pnormalized) []Pnormalized {
	var pnormalizedlist []Pnormalized

 	total:=0.0
	for _, item := range probVal {
		pnormalizedlist=append(pnormalizedlist,Pnormalized{Addr:item.Addr,Value:total})
		total += item.Value
	}
	for index:=0;index<len(probVal);index++{
		pnormalizedlist[index].Value /= total
	}


	return pnormalizedlist
}

func Sample1NodesInValNodes(probnormalized []Pnormalized, rand01 float64) common.Address {
	len:=len(probnormalized)
	for index:=len-1;index>=0;index--{
		if rand01>=probnormalized[index].Value{
			return probnormalized[index].Addr
		}
	}
	return common.Address{}
}
