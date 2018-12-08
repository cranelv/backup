// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package manparams

import (
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
)

func IsBroadcastNumber(number uint64) bool {
	if number%broadcastInterval == 0 {
		return true
	}
	return false
}

func IsReElectionNumber(number uint64) bool {
	if number%reelectionInterval == 0 {
		return true
	}
	return false
}

func GetLastBroadcastNumber(number uint64) uint64 {
	if IsBroadcastNumber(number) {
		return number
	}
	ans := (number / broadcastInterval) * broadcastInterval
	return ans
}

func GetLastReElectionNumber(number uint64) uint64 {
	if IsReElectionNumber(number) {
		return number
	}
	ans := (number / reelectionInterval) * reelectionInterval
	return ans
}

func GetNextBroadcastNumber(number uint64) uint64 {
	if IsBroadcastNumber(number) {
		return number
	}
	ans := (number/broadcastInterval + 1) * broadcastInterval
	return ans
}

func GetNextReElectionNumber(number uint64) uint64 {
	if IsReElectionNumber(number) {
		return number
	}
	ans := (number/reelectionInterval + 1) * reelectionInterval
	return ans
}

func GetBroadcastInterval() uint64 {
	interval, err := mtxCfg.getStateData(mc.MSPBroadcastInterval)
	if err != nil {
		log.Crit("config", "get broadcast interval from state err", err)
	}
	return interval.(uint64)
}

func GetReElectionInterval() uint64 {
	interval, err := mtxCfg.getStateData(mc.MSPBroadcastInterval)
	if err != nil {
		log.Crit("config", "get broadcast interval from state err", err)
	}
	return interval.(uint64) * 3
}
