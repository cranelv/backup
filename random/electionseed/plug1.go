// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package electionseed

import (
	"math/big"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/random/commonsupport"
)

func init() {
	electSeedPlug1 := &ElectSeedPlug1{privatekey: big.NewInt(0)}
	RegisterElectSeedPlugs("Minhash&Key", electSeedPlug1)
}

type ElectSeedPlug1 struct {
	privatekey *big.Int
}

func (self *ElectSeedPlug1) Prepare(height uint64, support baseinterface.RandomChainSupport) error {
	log.INFO(ModuleElectSeed, "生成随机种子准备阶段", "开始", "height", height)
	defer log.INFO(ModuleElectSeed, "生成随机种子准备阶段", "结束", "height", height)

	data, err := commonsupport.GetElectGenTimes(support.BlockChain(), height)
	if err != nil {
		log.ERROR(ModuleElectSeed, "获取通用配置失败 err", err)
		return err
	}
	voteBeforeTime := uint64(data.VoteBeforeTime)
	bcInterval := manparams.NewBCInterval()
	if bcInterval.IsBroadcastNumber(height+voteBeforeTime) == false {
		log.INFO(ModuleElectSeed, "RoleUpdateMsgHandle", "当前不是投票点,忽略")
		return nil
	}
	if NeedVote(height) == false {
		log.WARN(ModuleElectSeed, "不需要投票  账户不存在抵押交易 高度", height)
		return nil
	}
	privatekey, publickeySend, err := commonsupport.Getkey()
	privatekeySend := common.BigToHash(self.privatekey).Bytes()
	if err != nil {
		log.INFO(ModuleElectSeed, "获取公私钥失败 err", err)
		return err
	}

	log.INFO(ModuleElectSeed, "公钥 高度", (height + voteBeforeTime))
	log.INFO(ModuleElectSeed, "私钥 高度", (height + voteBeforeTime))
	mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.Publickey, Height: big.NewInt(int64(height + voteBeforeTime)), Data: publickeySend})
	mc.PublishEvent(mc.SendBroadCastTx, mc.BroadCastEvent{Txtyps: mc.Privatekey, Height: big.NewInt(int64(height + voteBeforeTime)), Data: privatekeySend})

	self.privatekey = privatekey
	return nil
}

func (self *ElectSeedPlug1) CalcSeed(hash common.Hash, support baseinterface.RandomChainSupport) (*big.Int, error) {
	ans, err := commonsupport.GetCurrentKeys(hash, support)
	if err != nil {
		log.ERROR(ModuleElectSeed, "计算阶段", "", "获取有效私钥出错 err", err)
		return nil, err
	}
	minHash := commonsupport.GetMinHash(hash, support)
	ans.Add(ans, minHash.Big())
	log.INFO(ModuleElectSeed, "计算阶段", "", "计算结果为", ans, "高度hash", hash.String())
	return ans, nil
}

func NeedVote(height uint64) bool {
	ans, err := ca.GetElectedByHeightAndRole(big.NewInt(int64(height)), common.RoleValidator)
	if err != nil {
		log.Error(ModuleElectSeed, "投票失敗", "获取验证者身份列表失败", "高度", height)
		return false
	}
	selfAddress := ca.GetAddress()
	for _, v := range ans {
		if v.Address == selfAddress {
			log.INFO(ModuleElectSeed, "具备投票身份 账户", selfAddress)
			return true
		}
	}
	log.Error(ModuleElectSeed, "不具备投票身份,不存在抵押列表里 账户", selfAddress)
	return false
}
