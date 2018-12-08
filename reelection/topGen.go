// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"errors"
	"math/big"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

//得到随机种子
func (self *ReElection) GetSeed(hash common.Hash) (*big.Int, error) {
	return self.random.GetRandom(hash, "electionseed")
}

func (self *ReElection) ToGenMinerTop(hash common.Hash) ([]mc.ElectNodeInfo, []mc.ElectNodeInfo, []mc.ElectNodeInfo, error) {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash算高度失败 ToGenMinerTop hash", hash, "err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}

	minerDeposit, err := GetAllElectedByHeight(big.NewInt(int64(height)), common.RoleMiner) //
	if err != nil {
		log.ERROR(Module, "获取矿工抵押列表失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	log.INFO(Module, "矿工抵押交易", minerDeposit)

	seed, err := self.GetSeed(hash)
	if err != nil {
		log.ERROR(Module, "获取种子失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	log.Info(Module, "矿工选举种子", seed)

	TopRsp := self.elect.MinerTopGen(&mc.MasterMinerReElectionReqMsg{SeqNum: height, RandSeed: seed, MinerList: minerDeposit})

	return TopRsp.MasterMiner, TopRsp.BackUpMiner, []mc.ElectNodeInfo{}, nil
}

func (self *ReElection) ToGenValidatorTop(hash common.Hash) ([]mc.ElectNodeInfo, []mc.ElectNodeInfo, []mc.ElectNodeInfo, error) {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash算高度失败 ToGenValidatorTop hash", hash.String())
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}

	validatoeDeposit, err := GetAllElectedByHeight(big.NewInt(int64(height)), common.RoleValidator)
	if err != nil {
		log.ERROR(Module, "获取验证者列表失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	log.INFO(Module, "验证者抵押账户", validatoeDeposit)
	foundDeposit := GetFound()

	seed, err := self.GetSeed(hash)
	if err != nil {
		log.ERROR(Module, "获取验证者种子失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	log.INFO(Module, "验证者随机种子", seed)
	TopRsp := self.elect.ValidatorTopGen(&mc.MasterValidatorReElectionReqMsg{SeqNum: height, RandSeed: seed, ValidatorList: validatoeDeposit, FoundationValidatorList: foundDeposit})
	//err = self.writeElectData(common.RoleValidator, hash, ElectMiner{}, ElectValidator{MasterValidator: TopRsp.MasterValidator,
	//	BackUpValidator:    TopRsp.BackUpValidator,
	//	CandidateValidator: TopRsp.CandidateValidator,
	//})
	//return err
	return TopRsp.MasterValidator, TopRsp.BackUpValidator, TopRsp.CandidateValidator, nil

}
func GetFound() []vm.DepositDetail {
	return []vm.DepositDetail{}
}
func GetAllElectedByHeight(Heigh *big.Int, tp common.RoleType) ([]vm.DepositDetail, error) {

	switch tp {
	case common.RoleMiner:
		ans, err := ca.GetElectedByHeightAndRole(Heigh, common.RoleMiner)
		log.INFO("從CA獲取礦工抵押交易", "data", ans, "height", Heigh)
		if err != nil {
			return []vm.DepositDetail{}, errors.New("获取矿工交易身份不对")
		}
		return ans, nil
	case common.RoleValidator:
		ans, err := ca.GetElectedByHeightAndRole(Heigh, common.RoleValidator)
		log.Info("從CA獲取驗證者抵押交易", "data", ans, "height", Heigh)
		if err != nil {
			return []vm.DepositDetail{}, errors.New("获取验证者交易身份不对")
		}
		return ans, nil

	default:
		return []vm.DepositDetail{}, errors.New("获取抵押交易身份不对")
	}
}
