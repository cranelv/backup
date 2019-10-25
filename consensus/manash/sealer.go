// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package manash

import (
	"math/big"
	"sync"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
)

type diffiList []*big.Int

func (v diffiList) Len() int {
	return len(v)
}

func (v diffiList) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v diffiList) Less(i, j int) bool {
	if v[i].Cmp(v[j]) == 1 {

		return true
	}
	return false
}

type minerDifficultyList struct {
	lock      sync.RWMutex
	diffiList []*big.Int
	targets   []*big.Int
}

func GetdifficultyListAndTargetList(difficultyList []*big.Int) minerDifficultyList {
	difficultyListAndTargetList := minerDifficultyList{
		diffiList: make([]*big.Int, len(difficultyList)),
		targets:   make([]*big.Int, len(difficultyList)),
		lock:      sync.RWMutex{},
	}
	copy(difficultyListAndTargetList.diffiList, difficultyList)
	var targets = make([]*big.Int, len(difficultyList))
	for i := 0; i < len(difficultyList); i++ {
		targets[i] = new(big.Int).Div(maxUint256, difficultyList[i])

	}
	copy(difficultyListAndTargetList.targets, targets)

	return difficultyListAndTargetList
}


func compareDifflist(result []byte, diffList []*big.Int, targets []*big.Int) (int, bool) {
	for i := 0; i < len(diffList); i++ {
		if new(big.Int).SetBytes(result).Cmp(targets[i]) <= 0 {
			return i, true
		}
	}

	return -1, false
}

// mine is the actual proof-of-work miner that searches for a nonce starting from
// seed that results in correct final block difficulty.
func (manash *Manash) Mine(header *types.Header, id int, seed uint64, abort chan struct{}, found chan *types.Header, aaa []uint64) {
	// Extract some data from the header

	var (
		curHeader = types.CopyHeader(header)
		hash      = curHeader.HashNoNonce().Bytes()
		target    = new(big.Int).Div(maxUint256, header.Difficulty)
//		number    = curHeader.Number.Uint64()
//		dataset   = manash.dataset(number)
//		dataset   = []uint32{}
	)

	// Start generating random nonces until we abort or find a good one
	log.INFO("SEALER begin mine", "target", target, "isBroadcast", false, "number", curHeader.Number.Uint64(), "diff", header.Difficulty.Uint64())
	defer log.INFO("SEALER stop mine", "number", curHeader.Number.Uint64(), "diff", header.Difficulty.Uint64())
	var (
//		attempts = int64(0)
		nonce    = seed
		scratchPad = make([]uint64, 1<<18, 1<<18)
	)
	logger := log.New("miner", id)
	logger.Trace("Started ethash search for new nonces", "seed", seed)
	//log.INFO("SEALER", "Started ethash search for new nonces seed", seed)
search:
	for {
		select {
		case <-abort:
			// Mining terminated, update stats and abort
			logger.Trace("Ethash nonce search aborted", "attempts", nonce-seed)
//			manash.hashrate.Mark(attempts)
			return

		default:
			// We don't have to update hash rate on every nonce, so update after after 2^X nonces
//			attempts++
//			if (attempts % (1 << 15)) == 0 {
//				manash.hashrate.Mark(attempts)
//				attempts = 0
//			}
			// Compute the PoW value of this nonce
			digest, result := newPowHash(hash, nonce, scratchPad, nil)

			//log.Info("sealer","result",new(big.Int).SetBytes(result))
			//log.Info("sealer","target",target)
			if new(big.Int).SetBytes(result).Cmp(target) <= 0 {
				// Correct nonce found, create a new header with it
				header = types.CopyHeader(header)
				header.Nonce = types.EncodeNonce(nonce)
				header.MixDigest = common.BytesToHash(digest)

				// Seal and return a block (if still needed)
				select {
				case found <- header:
					logger.Trace("Ethash nonce found and reported", "attempts", nonce-seed, "nonce", nonce)
				case <-abort:
					logger.Trace("Ethash nonce found but discarded", "attempts", nonce-seed, "nonce", nonce)
				}
				break search
			}
			nonce++
		}
	}
	// Datasets are unmapped in a finalizer. Ensure that the dataset stays live
	// during sealing so it's not unmapped while being read.
//	runtime.KeepAlive(dataset)
}
