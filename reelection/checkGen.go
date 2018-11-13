// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"github.com/matrix/go-matrix/common"

	"github.com/matrix/go-matrix/log"
)

func (self *ReElection) boolTopStatus(hash common.Hash, types common.RoleType) bool {
	if _, _, err := self.readElectData(types, hash); err != nil {
		return false
	}
	return true
}
func (self *ReElection) checkTopGenStatus(hash common.Hash) error {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		return err
	}

	if ok := self.boolTopStatus(hash, common.RoleMiner); ok == false {
		log.Warn(Module, "矿工拓扑图需要重新算 hash", hash.String())
		if height == 0 {
			return nil
		}
		if err := self.ToGenMinerTop(hash); err != nil {
			return err
		}

	}

	if ok := self.boolTopStatus(hash, common.RoleValidator); ok == false {
		log.Warn(Module, "验证者拓扑图需要重新算 高度", height)
		if height == 0 {
			return nil
		}
		if err := self.ToGenValidatorTop(hash); err != nil {
			return err
		}
	}
	return nil
}

func (self *ReElection) checkUpdateStatus(height uint64) error {
	if common.IsReElectionNumber(height) {
		if ok := self.boolNativeStatus(height); ok == false { //无该初选列表
			log.INFO(Module, "检查初选列表时出错", "重新计算", "高度", height)
			return self.GetNativeFromDB(height)
		}
		return nil
	}

	lastPoint := common.GetLastReElectionNumber(height)

	log.INFO(Module, "检查初选列表阶段 上一个点", lastPoint, "现在高度", height, "状态", "开始")
	for index := lastPoint; index <= height; index++ {
		if self.boolNativeStatus(index) == true {
			continue
		}
		log.INFO(Module, "检查初选列表阶段 该位置没数据 需要重新算 index", index)
		if common.IsReElectionNumber(index) {
			self.GetNativeFromDB(index)
			continue
		}
		native, err := self.readNativeData(index - 1)
		if err != nil {
			log.Error(Module, "检查更新阶段 获取上一个初选列表失败 高度", index-1)
			return err
		}
		if err = self.UpdateNative(index, native); err != nil {
			log.Error(Module, "检查更新阶段", "更新初选列表失败 高度", index)
			return err
		}
	}
	log.INFO(Module, "检查初选列表阶段 上一个点", lastPoint, "现在高度", height, "状态", "结束")
	return nil

}
