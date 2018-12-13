// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

type TopGenStatus struct {
	//V_State bool
	MastV []mc.ElectNodeInfo
	BackV []mc.ElectNodeInfo
	CandV []mc.ElectNodeInfo

	//M_State bool
	MastM []mc.ElectNodeInfo
	BackM []mc.ElectNodeInfo
	CandM []mc.ElectNodeInfo
}

func (self *ReElection) HandleTopGen(hash common.Hash) (TopGenStatus, error) {
	topGenStatus := TopGenStatus{}

	if self.IsMinerTopGenTiming(hash) { //矿工生成时间 240
		log.INFO(Module, "是矿工生成时间点 hash", hash.String())
		MastM, BackM, CandM, err := self.ToGenMinerTop(hash)
		if err != nil {
			log.ERROR(Module, "矿工拓扑生成错误 err", err)
			return topGenStatus, err
		}
		topGenStatus.MastM = append(topGenStatus.MastM, MastM...)
		topGenStatus.BackM = append(topGenStatus.BackM, BackM...)
		topGenStatus.CandM = append(topGenStatus.CandM, CandM...)
	}

	if self.IsValidatorTopGenTiming(hash) { //验证者生成时间 260
		log.INFO(Module, "是验证者生成时间点 height", hash)
		MastV, BackV, CandV, err := self.ToGenValidatorTop(hash)
		if err != nil {
			log.ERROR(Module, "验证者拓扑生成错误 err", err)
			return topGenStatus, err
		}
		topGenStatus.MastV = append(topGenStatus.MastV, MastV...)
		topGenStatus.BackV = append(topGenStatus.BackV, BackV...)
		topGenStatus.CandV = append(topGenStatus.CandV, CandV...)
	}
	return topGenStatus, nil

}

//是不是矿工拓扑生成时间段
func (self *ReElection) IsMinerTopGenTiming(hash common.Hash) bool {

	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.ERROR(Module, "判断是否是矿工生成点错误 hash", hash.String(), "err", err)
		return false
	}

	bcInterval, err := self.GetBroadcastIntervalByHash(hash)
	if err != nil {
		log.ERROR(Module, "获取广播区间失败 err", err)
		return false
	}

	genData, err := self.GetElectGenTimes(height)
	if err != nil {
		log.ERROR(Module, "获取配置错误 高度", height,"err",err)
		return false
	}

	if bcInterval.IsReElectionNumber(height + 1 + uint64(genData.MinerNetChange)) {
		log.ERROR(Module, "是矿工生成点 高度", height, "MinerNetChange", genData.MinerNetChange,"换届周期", bcInterval.GetReElectionInterval())
		return true
	}

	log.ERROR(Module, "不是矿工生成点", "高度",height,uint64(genData.MinerNetChange), "换届周期", bcInterval.GetReElectionInterval())
	return false
}

//是不是验证者拓扑生成时间段
func (self *ReElection) IsValidatorTopGenTiming(hash common.Hash) bool {

	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.ERROR(Module, "判断是否是验证者生成点错误 height", height, "err", err)
		return false
	}

	bcInterval, err := self.GetBroadcastIntervalByHash(hash)
	if err != nil {
		log.ERROR(Module, "获取广播区间失败 err", err)
		return false
	}

	genData, err := self.GetElectGenTimes(height)
	if err != nil {
		log.ERROR(Module, "获取配置错误 高度", height,"err",err)
		return false
	}
	if bcInterval.IsReElectionNumber(height + 1 + uint64(genData.ValidatorNetChange)) {
		log.ERROR(Module, "是验证者生成点 height", height, "ValidatorNetChange", genData.ValidatorNetChange,"换届周期", bcInterval.GetReElectionInterval())
		return true
	}
	log.ERROR(Module, "不是验证者生成点", height, "err", genData.ValidatorNetChange,"换届周期", bcInterval.GetReElectionInterval())
	return false
}
