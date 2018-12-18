// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package commonsupport

import (
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

type stateReader interface {
	GetMatrixStateDataByNumber(key string, number uint64) (interface{}, error)
}

func GetElectGenTimes(stateReader stateReader, height uint64) (*mc.ElectGenTimeStruct, error) {
	data, err := stateReader.GetMatrixStateDataByNumber(mc.MSKeyElectGenTime, height)
	if err != nil {
		log.Error("random-commonsupport", "获取选举基础信息失败 err", err)
		return nil, err
	}
	electGenConfig, OK := data.(*mc.ElectGenTimeStruct)
	if OK == false || electGenConfig == nil {
		log.ERROR("random-commonsupport", "ElectGenTimeStruct 非法", "反射失败")
		return nil, err
	}
	return electGenConfig, nil
}
