// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

func deleteQueue(address common.Address, allNative support.AllNative) support.AllNative {
	log.INFO(Module, "在缓存中删除节点阶段-开始 地址", address, "缓存", allNative)
	for k, v := range allNative.MasterQ {
		if v == address {
			allNative.MasterQ = append(allNative.MasterQ[:k], allNative.MasterQ[k+1:]...)
			log.INFO(Module, "在缓存中删除节点阶段-master 地址 ", address, "缓存", allNative)
			return allNative
		}
	}
	for k, v := range allNative.BackUpQ {
		if v == address {
			allNative.BackUpQ = append(allNative.BackUpQ[:k], allNative.BackUpQ[k+1:]...)
			log.INFO(Module, "在缓存中删除节点阶段-backup 地址", address, "缓存", allNative)
			return allNative
		}
	}
	for k, v := range allNative.CandidateQ {
		if v == address {
			allNative.CandidateQ = append(allNative.CandidateQ[:k], allNative.CandidateQ[k+1:]...)
			log.INFO(Module, "在缓存中删除节点阶段-candidate 地址", address, "缓存", allNative)
			return allNative
		}
	}

	log.INFO(Module, "在缓存中删除节点阶段-结束-不再任何一个梯队 地址", address, "缓存", allNative)
	return allNative
}

func (self *ReElection) TopoUpdate(allNative support.AllNative, top *mc.TopologyGraph) ([]mc.Alternative,error) {
	elect,err:=self.GetElectPlug(top.Number)
	if err!=nil{
		log.ERROR(Module,"获取选举插件")
		return []mc.Alternative{},err
	}
	return elect.ToPoUpdate(allNative,top),nil
}
