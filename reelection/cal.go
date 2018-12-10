// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"errors"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	//"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/baseinterface"
)


func (self *ReElection)GetElectGenTimes (height uint64)(*mc.ElectGenTimeStruct,error){
	data,err:=self.bc.GetMatrixStateDataByNumber(mc.MSKeyElectGenTime,height)
	if err!=nil{
		log.Error("GetElectGenTimes","获取选举时间点信息失败 err",err)
		return nil,err
	}
	electGenConfig, OK := data.(*mc.ElectGenTimeStruct)
	if OK == false || electGenConfig == nil {
		log.ERROR("GetElectGenTimes", "ElectGenTimeStruct 非法", "反射失败","高度",height)
		return nil,errors.New("反射失败")
	}
	return electGenConfig,nil
}
func (self *ReElection)GetElectConfig(height uint64)(*mc.ElectConfigInfo,error){
	data,err:=self.bc.GetMatrixStateDataByNumber(mc.MSKeyElectConfigInfo,height)
	if err!=nil{
		log.ERROR("GetElectInfo","获取选举基础信息失败 err",err)
		return nil,err
	}
	electInfo,OK:=data.(*mc.ElectConfigInfo)
	if OK==false||electInfo==nil{
		log.ERROR("GetElectInfo","GetElectInfo ","反射失败","高度",height)
		return nil,errors.New("反射失败")
	}
	return electInfo,nil
}
func (self *ReElection)GetViPList(height uint64)([]mc.VIPConfig,error){
	data,err:=self.bc.GetMatrixStateDataByNumber(mc.MSKeyVIPConfig,height)
	if err!=nil{
		log.ERROR("GetElectInfo","获取选举基础信息失败 err",err)
		return nil,err
	}
	vipList,OK:=data.([]mc.VIPConfig)
	if OK==false||vipList==nil{
		log.ERROR("GetElectInfo","GetElectInfo ","反射失败","高度",height)
		return nil,errors.New("反射失败")
	}
	return vipList,nil
}

func (self *ReElection)GetElectPlug(height uint64)(baseinterface.ElectionInterface,error){
	data,err:=self.bc.GetMatrixStateDataByNumber(mc.MSKeyElectConfigInfo,height)
	if err!=nil{
		log.ERROR("GetElectInfo","获取选举基础信息失败 err",err)
		return nil,err
	}
	electInfo,OK:=data.(*mc.ElectConfigInfo)
	if OK==false||electInfo==nil{
		log.ERROR("GetElectInfo","GetElectInfo ","反射失败","高度",height)
		return nil,errors.New("反射失败")
	}
	return baseinterface.NewElect(electInfo.ElectPlug),nil
}

func (self *ReElection) TransferToElectionStu(info *ElectReturnInfo) []common.Elect {
	result := make([]common.Elect, 0)

	srcMap := make(map[common.ElectRoleType][]mc.ElectNodeInfo)
	srcMap[common.ElectRoleMiner] = info.MasterMiner
	//srcMap[common.ElectRoleMinerBackUp] = info.BackUpMiner
	srcMap[common.ElectRoleValidator] = info.MasterValidator
	srcMap[common.ElectRoleValidatorBackUp] = info.BackUpValidator
	orderIndex := []common.ElectRoleType{common.ElectRoleValidator, common.ElectRoleValidatorBackUp, common.ElectRoleMiner}

	for _, role := range orderIndex {
		src := srcMap[role]
		for _, node := range src {
			e := common.Elect{
				Account: node.Account,
				Stock:   node.Stock,
				Type:    role,
			}

			result = append(result, e)
		}
	}

	return result
}

func (self *ReElection) TransferToNetTopologyAllStu(info *ElectReturnInfo) *common.NetTopology {
	result := &common.NetTopology{
		Type:            common.NetTopoTypeAll,
		NetTopologyData: make([]common.NetTopologyData, 0),
	}

	srcMap := make(map[common.ElectRoleType][]mc.ElectNodeInfo)
	srcMap[common.ElectRoleMiner] = info.MasterMiner
	//srcMap[common.ElectRoleMinerBackUp] = info.BackUpMiner
	srcMap[common.ElectRoleValidator] = info.MasterValidator
	srcMap[common.ElectRoleValidatorBackUp] = info.BackUpValidator
	orderIndex := []common.ElectRoleType{common.ElectRoleMiner, common.ElectRoleValidator, common.ElectRoleValidatorBackUp}

	for _, role := range orderIndex {
		src := srcMap[role]
		for i, node := range src {
			data := common.NetTopologyData{
				Account:  node.Account,
				Position: common.GeneratePosition(uint16(i), role),
			}
			result.NetTopologyData = append(result.NetTopologyData, data)
		}
	}

	return result
}

func (self *ReElection) TransferToNetTopologyChgStu(alterInfo []mc.Alternative) *common.NetTopology {
	result := &common.NetTopology{
		Type:            common.NetTopoTypeChange,
		NetTopologyData: make([]common.NetTopologyData, 0),
	}

	for _, alter := range alterInfo {
		data := common.NetTopologyData{
			Account:  alter.A,
			Position: alter.Position,
		}
		result.NetTopologyData = append(result.NetTopologyData, data)
	}

	return result
}

func (self *ReElection) GetNumberByHash(hash common.Hash) (uint64, error) {
	tHeader := self.bc.GetHeaderByHash(hash)
	if tHeader == nil {
		log.Error(Module, "GetNumberByHash 根据hash算header失败 hash", hash.String())
		return 0, errors.New("根据hash算header失败")
	}
	if tHeader.Number == nil {
		log.Error(Module, "GetNumberByHash header 内的高度获取失败", hash.String())
		return 0, errors.New("header 内的高度获取失败")
	}
	return tHeader.Number.Uint64(), nil
}

func (self *ReElection) GetHeaderHashByNumber(hash common.Hash, height uint64) (common.Hash, error) {
	AimHash, err := self.bc.GetAncestorHash(hash, height)
	if err != nil {
		log.Error(Module, "获取祖先hash失败 hash", hash.String(), "height", height, "err", err)
		return common.Hash{}, err
	}
	return AimHash, nil
}

func GetCurrentTopology(hash common.Hash, reqtypes common.RoleType) (*mc.TopologyGraph, error) {
	return ca.GetTopologyByHash(reqtypes, hash)
	//return ca.GetTopologyByNumber(reqtypes, height)
}

func CheckBlock(block *types.Block) error {
	if block == nil {
		return errors.New("block为空")
	}
	if block.Header() == nil {
		return errors.New("block.Header()为空")
	}
	if block.Header().Number == nil {
		return errors.New("block.Header.Number为空 ")
	}
	return nil
}

func SloveElectStatus(electStates *mc.ElectGraph) (interface{}, error) {

	log.INFO("上树信息", "electStates 高度", electStates.Number)
	for _, v := range electStates.ElectList {
		log.INFO("上树信息", "当前拓扑图类型", v.Type, "账户", v.Account.String())
	}
	for _, v := range electStates.NextElect {
		log.INFO("上树信息", "下届拓扑图 类型", v.Type, "账户", v.Account.String())
	}
	return electStates, nil
}
func SloveOnlineStatus(electonline *mc.ElectOnlineStatus) (interface{}, error) {
	log.INFO("上树信息", "electonline 高度", electonline.Number)
	for _, v := range electonline.ElectOnline {
		log.INFO("上树信息", "当前上下线状态", v.Position, "账户", v.Account.String())
	}
	return electonline, nil
}
