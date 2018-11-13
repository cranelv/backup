// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

func (self *ReElection) ToNativeValidatorStateUpdate(height uint64, allNative support.AllNative) (support.AllNative, error) {

	header := self.bc.GetHeaderByNumber(height)
	if header == nil {
		log.ERROR(Module, "获取指定高度的区块头失败 高度", height)
		return support.AllNative{}, errors.New("获取指定高度的区块头失败")
	}
	DiffFromBlock := header.NetTopology

	TopoGrap, err := GetCurrentTopology(height-1, common.RoleValidator|common.RoleBackupValidator)
	log.INFO(Module, "更新初选列表信息 拓扑的高度", height-1, "拓扑值", TopoGrap, "diff", DiffFromBlock)
	if err != nil {
		log.ERROR(Module, "从ca获取验证者拓扑图失败", err)
		return allNative, err
	}

	allNative = self.CalOnline(DiffFromBlock, TopoGrap, allNative)
	log.INFO(Module, "更新上下线状态", "结束", "高度", height)

	return allNative, nil
}

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

func addQueue(address common.Address, allNative support.AllNative) support.AllNative {
	log.INFO(Module, "在缓存中增加节点阶段-开始 地址", address, "allNative", allNative)
	for _, v := range allNative.Master {
		if v.Account == address {

			allNative.MasterQ = append(allNative.MasterQ, address)
			log.INFO(Module, "在缓存中增加节点阶段-master 地址", address, "allNative", allNative)
			return allNative
		}
	}
	for _, v := range allNative.BackUp {
		if v.Account == address {
			allNative.BackUpQ = append(allNative.BackUpQ, address)
			log.INFO(Module, "在缓存中增加节点阶段-backup 地址", address, "allNative", allNative)
			return allNative
		}
	}
	for _, v := range allNative.Candidate {
		if v.Account == address {
			allNative.CandidateQ = append(allNative.CandidateQ, address)
			log.INFO(Module, "在缓存中增加节点阶段-candidate 地址", address, "allNative", allNative)
			return allNative
		}
	}
	log.INFO(Module, "在缓存中增加节点阶段-结束 地址-不在任何一个梯队", address, "allNative", allNative)
	return allNative
}
func (self *ReElection) CalOnline(diff common.NetTopology, top *mc.TopologyGraph, allNative support.AllNative) support.AllNative {

	log.INFO(Module, "更新上下线阶段 拓扑差值-开始", diff.NetTopologyData, "allNative", allNative)

	for _, v := range diff.NetTopologyData {
		if v.Position == common.PosOnline {
			allNative = addQueue(v.Account, allNative)
		} else {
			allNative = deleteQueue(v.Account, allNative)
		}
	}
	log.INFO(Module, "更新上下线阶段 拓扑差值-结束", diff.NetTopologyData, "allNative", allNative)
	return allNative

}

func (self *ReElection) writeNativeData(height uint64, data support.AllNative) error {
	key := MakeNativeDBKey(height)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = self.ldb.Put([]byte(key), jsonData, nil)
	log.INFO(Module, "数据库 初选列表 err", err, "高度", height, "key", key)
	return err
}

func (self *ReElection) readNativeData(height uint64) (support.AllNative, error) {

	key := MakeNativeDBKey(height)
	ans, err := self.ldb.Get([]byte(key), nil)
	if err != nil {
		return support.AllNative{}, err
	}
	var realAns support.AllNative
	err = json.Unmarshal(ans, &realAns)
	if err != nil {
		return support.AllNative{}, err
	}

	return realAns, nil

}
func MakeNativeDBKey(height uint64) string {
	t := big.NewInt(int64(height))
	ss := t.String() + "---" + "Native"
	return ss
}
func needReadFromGenesis(height uint64) bool {
	if height == 0 {
		return true
	}
	return false
}
func (self *ReElection) wirteNativeFromGeneis() error {
	preBroadcast := support.AllNative{}
	block := self.bc.GetBlockByNumber(0)
	if block == nil {
		return errors.New("第0块区块拿不到")
	}
	header := block.Header()
	if header == nil {
		return errors.New("第0块区块头拿不到")
	}
	for _, v := range header.NetTopology.NetTopologyData {
		switch common.GetRoleTypeFromPosition(v.Position) {
		case common.RoleValidator:
			temp := mc.TopologyNodeInfo{
				Account:  v.Account,
				Position: v.Position,
				Type:     common.RoleValidator,
			}
			preBroadcast.Master = append(preBroadcast.Master, temp)
		case common.RoleBackupValidator:
			temp := mc.TopologyNodeInfo{
				Account:  v.Account,
				Position: v.Position,
				Type:     common.RoleBackupValidator,
			}
			preBroadcast.BackUp = append(preBroadcast.BackUp, temp)
		}
	}
	log.INFO(Module, "第0块到达处理阶段 更新初选列表", "从0的区块头中获取", "初选列表", preBroadcast)
	err := self.writeNativeData(0, preBroadcast)
	log.INFO(Module, "第0块到达处理阶段 更新初选列表", "从0的区块头中获取 写数据到数据库", "err", err)
	return err
}
func (self *ReElection) GetNativeFromDB(height uint64) error {
	if needReadFromGenesis(height) {
		return self.wirteNativeFromGeneis()
	}
	log.INFO(Module, "GetNativeFromDB", height)

	hash := self.bc.GetHashByNumber(height)

	if err := self.checkTopGenStatus(hash); err != nil {
		log.ERROR(Module, "检查top生成出错 err", err)
	}
	_, validatorElect, err := self.readElectData(common.RoleValidator, hash)
	if err != nil {
		return err
	}
	preBroadcast := support.AllNative{
		Master:     validatorElect.MasterValidator,
		BackUp:     validatorElect.BackUpValidator,
		Candidate:  validatorElect.CandidateValidator,
		MasterQ:    []common.Address{},
		BackUpQ:    []common.Address{},
		CandidateQ: []common.Address{},
	}

	for _, v := range preBroadcast.Candidate {
		preBroadcast.CandidateQ = append(preBroadcast.CandidateQ, v.Account)
	}

	err = self.writeNativeData(height, preBroadcast)
	log.INFO(Module, "writeNativeData", height, "err", err)
	return err
}
