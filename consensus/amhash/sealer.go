// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package amhash

import (
	"math"
	"math/big"
	//"runtime"
	"sync"

	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/baseinterface"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/consensus/ai"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
)

var (
	aiPictureMaxCount = 64000 // AI图库数量
	aiPictureSize     = 16    // AI选取图片数量
)

func (amhash *Amhash) SealAI(chain consensus.ChainReader, header *types.Header, stop <-chan struct{}) (*types.Header, error) {
	log.Info("amhash sealer", "AI挖矿", "开始", "高度", header.Number.Uint64())
	defer log.Info("amhash sealer", "AI挖矿", "结束", "高度", header.Number.Uint64())

	curHeader := types.CopyHeader(header)
	// start ai mining first
	aiHash, stopped, err := amhash.aiMineProcess(chain, curHeader, stop)
	if err != nil {
		return nil, err
	}
	if stopped {
		return nil, nil
	}

	curHeader.AIHash = aiHash
	curHeader.Nonce = types.BlockNonce{}

	return curHeader, nil
}

func (amhash *Amhash) MineX11(header *types.Header, id int, seed uint64, abort chan struct{}, found chan *types.Header) {
	// Extract some data from the header

	var (
		curHeader         = types.CopyHeader(header)
		mineData          = generateMineData(curHeader)
		target            = new(big.Int).Div(maxUint256, header.Difficulty)
//		basePowerTarget   = new(big.Int).Div(maxUint256, params.BasePowerDifficulty)
//		basePowerFindFlag = false
//		number            = curHeader.Number.Uint64()
	)

	// Start generating random nonces until we abort or find a good one
	log.INFO("SEALER begin mine", "target", target, "isBroadcast", false, "number", curHeader.Number.Uint64(), "diff", header.Difficulty.Uint64())
	defer log.INFO("SEALER stop mine", "number", curHeader.Number.Uint64(), "diff", header.Difficulty.Uint64())
	var (
		nonce    = seed
	)
	logger := log.New("miner", id)
	logger.Trace("Started ethash search for new nonces", "seed", seed)
	//log.INFO("SEALER", "Started ethash search for new nonces seed", seed)
x11search:
	for {
		select {
		case <-abort:
			// Mining terminated, update stats and abort
			logger.Trace("Ethash nonce search aborted", "attempts", nonce-seed)
			return

		default:
			result := x11PowHash(mineData, nonce)
			resultBigInt := new(big.Int).SetBytes(Reverse(result))
			if resultBigInt.Cmp(target) <= 0 {
				// Correct nonce found, create a new header with it
				resultHeader := types.CopyHeader(curHeader)
				resultHeader.Nonce = types.EncodeNonce(nonce)
				// Seal and return a block (if still needed)
				select {
				case found <- resultHeader:
					logger.Trace("amhash sealer", "x11 nonce found and reported, attempts", nonce-seed, "nonce", nonce)
				case <-abort:
					logger.Trace("amhash sealer", "x11 nonce found but discarded, attempts", nonce-seed, "nonce", nonce)
				}
				break x11search
			}
			nonce++
			// We don't have to update hash rate on every nonce, so update after after 2^X nonces
			//			attempts++
			//			if (attempts % (1 << 15)) == 0 {
			//				manash.hashrate.Mark(attempts)
			//				attempts = 0
			//			}
			// Compute the PoW value of this nonce

		}
	}
	// Datasets are unmapped in a finalizer. Ensure that the dataset stays live
	// during sealing so it's not unmapped while being read.
	//	runtime.KeepAlive(dataset)
}

func generateMineData(header *types.Header) []byte {
	data := header.ParentHash.Bytes()
	data = append(data, header.Coinbase.Bytes()...)
	for i := 0; i < 24; i++ {
		data = append(data, byte(0))
	}
	return data
}

func (amhash *Amhash) aiMineProcess(chain consensus.ChainReader, header *types.Header, stop <-chan struct{}) (common.Hash, bool, error) {
	abortCh := make(chan struct{}, 1)
	foundCh := make(chan []byte, 1)
	errCh := make(chan error, 1)

	go amhash.startAIMining(chain, header, abortCh, foundCh, errCh)

	for {
		select {
		case <-stop:
			log.Info("amhash sealer", "Sealer receive stop mine msg", "ai mine stop", "parent hash", header.ParentHash)
			close(abortCh)
			return common.Hash{}, true, nil

		case err := <-errCh:
			log.Warn("amhash sealer", "ai mining err", err)
			return common.Hash{}, false, err

		case result := <-foundCh:
			aiHash := common.BytesToHash(result)
			log.Info("amhash sealer", "aiMineProcess", "get ai digging result", "AIHash", aiHash)
			return aiHash, false, nil
		}
	}
}

func (amhash *Amhash) startAIMining(chain consensus.ChainReader, header *types.Header, abort chan struct{}, found chan []byte, errCh chan error) {
	// get seed
	vrf := baseinterface.NewVrf()
	_, vrfValue, _ := vrf.GetVrfInfoFromHeader(header.VrfValue)
	seed := big.NewInt(0).Add(types.RlpHash(vrfValue).Big(), header.AICoinbase.Big()).Int64()
	ai.Mining(seed, abort, found, errCh)
}

func (amhash *Amhash) sm3MineProcess(chain consensus.ChainReader, header *types.Header, stop <-chan struct{}, resultChan chan<- *types.Header) (*types.Header, bool, error) {
	log.Info("amhash sealer", "sm3 mine process", "begin", "number", header.Number)
	defer log.Info("amhash sealer", "sm3 mine process", "end", "number", header.Number)
	// Create a runner and the multiple search threads it directs
	abort := make(chan struct{})
	found := make(chan *types.Header)
	threads := 1

	var pend sync.WaitGroup
	for i := 0; i < threads; i++ {
		pend.Add(1)
		go func(id int, nonce uint64, abortCh chan struct{}) {
			defer pend.Done()
			amhash.Sm3Mine(header, id, nonce, abortCh, found)
		}(i, uint64(amhash.rand.Int63()), abort)
	}
	// Wait until sealing is terminated or a nonce is found
	var result *types.Header
	var isStop = false
sm3seal:
	for {
		select {
		case <-stop:
			log.Info("amhash sealer", "sm3 sealer receive stop mine", header.Number, "parent hash", header.ParentHash.TerminalString())
			// Outside abort, stop all miner threads
			if nil != abort {
				close(abort)
				abort = nil
			}
			isStop = true
			break sm3seal
		case result = <-found:
			// One of the threads found a block, abort all others
			log.Info("amhash sealer", "successfully sealed new sm3 result", result.Sm3Nonce, "number", result.Number)
			if nil != abort {
				close(abort)
				abort = nil
			}
			break sm3seal

		case <-amhash.update:
			// Thread count was changed on user request, restart
			if nil != abort {
				close(abort)
				abort = nil
			}
			pend.Wait()
			return amhash.sm3MineProcess(chain, header, stop, resultChan)
		}
	}
	// Wait for all miners to terminate and return the block
	pend.Wait()
	return result, isStop, nil
}

func (amhash *Amhash) Sm3Mine(header *types.Header, id int, seed uint64, abort chan struct{}, found chan *types.Header) {
	// Extract some data from the header
	var (
		curHeader     = types.CopyHeader(header)
		mineData      = generateMineData(curHeader)
		sm3Difficulty = big.NewInt(int64(math.Ceil(float64(header.Difficulty.Uint64()) * params.Sm3DifficultyRatio)))
		target        = new(big.Int).Div(maxUint256, sm3Difficulty)
	)
	logger := log.New("sm3 miner", id)
	// Start generating random nonces until we abort or find a good one
	var (
		nonce    = seed
	)
	logger.Trace("amhash sealer", "Started sm3 mine search for new nonces, seed", seed)
sm3search:
	for {
		select {
		case <-abort:
			// Mining terminated, update stats and abort
			logger.Trace("amhash sealer", "pow mine nonce search aborted, attempts", nonce-seed)
			return

		default:
			// We don't have to update hash rate on every nonce, so update after after 2^X nonces
			// Compute the PoW value of this nonce
			result := sm3PowHash(mineData, nonce)

			if new(big.Int).SetBytes(result).Cmp(target) <= 0 {
				// Correct nonce found, create a new header with it
				resultHeader := types.CopyHeader(curHeader)
				resultHeader.Sm3Nonce = types.EncodeNonce(nonce)
				// Seal and return a block (if still needed)
				select {
				case found <- resultHeader:
					logger.Trace("amhash sealer", "sm3 nonce found and reported, attempts", nonce-seed, "nonce", nonce)
				case <-abort:
					logger.Trace("amhash sealer", "sm3 nonce found but discarded, attempts", nonce-seed, "nonce", nonce)
				}
				break sm3search
			}
			nonce++
		}
	}
}
