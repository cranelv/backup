// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package commonsupport

import (
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
)


type stateReader interface {
	GetHashByNumber(number uint64) common.Hash
	GetHeaderByHash(hash common.Hash) *types.Header
	GetMatrixStateData(key string) (interface{}, error)
	GetMatrixStateDataByHash(key string, hash common.Hash) (interface{}, error)
	GetMatrixStateDataByNumber(key string, number uint64) (interface{}, error)
}
func GetElectGenTimes (stateReader stateReader,height uint64)(*mc.ElectGenTimeStruct,error){
	data,err:=stateReader.GetMatrixStateDataByNumber(mc.MSKeyElectGenTime,height)
	if err!=nil{
		log.Error("baseinterface","获取选举基础信息失败 err",err)
		return nil,err
	}
	electGenConfig, OK := data.(*mc.ElectGenTimeStruct)
	if OK == false || electGenConfig == nil {
		log.ERROR("baseinterface", "ElectGenTimeStruct 非法", "反射失败")
		return nil,err
	}
	return electGenConfig,nil
}