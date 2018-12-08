// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package manparams

import (
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

type PeriodInterval struct {
	broadcastInterval uint64
	electionInterval  uint64
}

func NewPeriodInterval(stateNumber uint64) (*PeriodInterval, error) {
	bi, err := GetBroadcastIntervalByNumber(stateNumber)
	if err != nil {
		return nil, err
	}

	return &PeriodInterval{
		broadcastInterval: bi,
		electionInterval:  bi * 3,
	}, nil
}

func (period *PeriodInterval) GetBroadcastInterval() uint64 {
	return period.broadcastInterval
}

func (period *PeriodInterval) IsBroadcastNumber(number uint64) bool {
	if number%period.broadcastInterval == 0 {
		return true
	}
	return false
}

func (period *PeriodInterval) IsReElectionNumber(number uint64) bool {
	if number%period.electionInterval == 0 {
		return true
	}
	return false
}

func IsBroadcastNumber(number uint64, stateNumber uint64) bool {
	broadcastInterval, err := GetBroadcastIntervalByNumber(stateNumber)
	if err != nil {
		log.Error("config", "获取广播区块周期失败", err, "stateNumber", stateNumber)
		return false
	}
	if number%broadcastInterval == 0 {
		return true
	}
	return false
}

func IsReElectionNumber(number uint64, stateNumber uint64) bool {
	broadcastInterval, err := GetBroadcastIntervalByNumber(stateNumber)
	if err != nil {
		log.Error("config", "获取广播区块周期失败", err, "stateNumber", stateNumber)
		return false
	}
	if number%(broadcastInterval*3) == 0 {
		return true
	}
	return false
}

func GetBroadcastInterval() uint64 {
	interval, err := mtxCfg.getStateData(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Crit("config", "get broadcast interval from state err", err)
	}
	return interval.(uint64)
}

func GetBroadcastIntervalByNumber(number uint64) (uint64, error) {
	interval, err := mtxCfg.getStateDataByNumber(mc.MSKeyBroadcastInterval, number)
	if err != nil {
		return 0, err
	}
	return interval.(uint64), nil
}
