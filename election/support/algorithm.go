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
	ans := []Strallyint{}
	RemainingProbNormalizedNodes := []Pnormalized{}
	if needNum >= len(probnormalized) {
		for _, v := range probnormalized {
			ans = append(ans, Strallyint{Addr: v.Addr, Value: 1})
		}
		return ans, []Pnormalized{}
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
	return ans, RemainingProbNormalizedNodes
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
