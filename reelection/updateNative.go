// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)


func (self *ReElection) TopoUpdate(allNative support.AllNative, top *mc.TopologyGraph, height uint64) ([]mc.Alternative, error) {
	elect, err := self.GetElectPlug(height)
	if err != nil {
		log.ERROR(Module, "获取选举插件")
		return []mc.Alternative{}, err
	}
	return elect.ToPoUpdate(allNative, top), nil
}
