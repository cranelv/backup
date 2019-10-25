// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package miner

import (
	"sync"

	"sync/atomic"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/consensus/manash"
	"runtime"
	"math/big"
	"math/rand"
	crand "crypto/rand"
	"math"
	"github.com/MatrixAINetwork/go-matrix/consensus/amhash"
)

type CpuAgent struct {
	mu sync.Mutex

	workCh chan *Work
	stop   chan struct{}

	quitCurrentOp chan struct{}
	returnCh      chan<- *types.Header

	rand     *rand.Rand    // Properly seeded random source for nonces
	manhash *manash.Manash
	amhash  *amhash.Amhash
	sealTreads []*SealThread
	isMining int32 // isMining indicates whether the agent is currently mining
}

func NewCpuAgent(manhash *manash.Manash,amhash *amhash.Amhash) *CpuAgent {
	miner := &CpuAgent{
		manhash:  manhash,
		amhash:	  amhash,
		stop:   make(chan struct{}, 1),
		workCh: make(chan *Work, 1),
	}
	miner.WaitingSeal()
	return miner
}
func (self *CpuAgent) WaitingSeal() error{
	threads := runtime.NumCPU()
	//	if isBroadcastNode {
	//		threads = 1
	//	}
	if self.rand == nil {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			return err
		}
		self.rand = rand.New(rand.NewSource(seed.Int64()))
	}
	self.sealTreads = make([]*SealThread,threads)
	for i := 0; i < threads; i++ {
		self.sealTreads[i] = &SealThread{
			id : i,
			seed : self.rand.Uint64(),
			mineCh : make(chan mineInfo,5),
			manHash: self.manhash,
			amHash:	 self.amhash,
			scratchPad:make([]uint64, 1<<18, 1<<18),
		}
		go self.sealTreads[i].waitSeal()
	}
	return nil
}
func (self *CpuAgent) Work() chan<- *Work                  { return self.workCh }
func (self *CpuAgent) SetReturnCh(ch chan<- *types.Header) { self.returnCh = ch }

func (self *CpuAgent) Stop() {
	if !atomic.CompareAndSwapInt32(&self.isMining, 1, 0) {
		return // agent already stopped
	}
	self.stop <- struct{}{}
done:
	// Empty work channel
	for {
		select {
		case <-self.workCh:
		default:
			break done
		}
	}
}

func (self *CpuAgent) Start() {
	if !atomic.CompareAndSwapInt32(&self.isMining, 0, 1) {
		return // agent already started
	}

	go self.update()
}

func (self *CpuAgent) update() {
out:
	for {
		select {
		case work := <-self.workCh:
			self.mu.Lock()
			if self.quitCurrentOp != nil {
				close(self.quitCurrentOp)
				self.quitCurrentOp = nil
			}
			if work != nil{
				self.quitCurrentOp = make(chan struct{})
				go self.mine(work, self.quitCurrentOp)
			}
			self.mu.Unlock()
		case <-self.stop:
			self.mu.Lock()
			if self.quitCurrentOp != nil {
				close(self.quitCurrentOp)
				self.quitCurrentOp = nil
			}
			self.mu.Unlock()
			log.Info("miner", "CpuAgent Stop Minning", "")
			break out
		}
	}
}
func (self *CpuAgent) SealPowOld(header *types.Header, stop <-chan struct{}, isBroadcastNode bool) (*types.Header, error) {
	log.INFO("seal", "挖矿", "开始", "高度", header.Number.Uint64())
	defer log.INFO("seal", "挖矿", "结束", "高度", header.Number.Uint64())

	mineinfo := mineInfo{
		abort:  make(chan struct{}),
		found:  make(chan *types.Header, len(self.sealTreads)),
		header: types.CopyHeader(header),
		powType:PowOld,
	}

	for _, thread := range self.sealTreads {
		thread.mineCh <- mineinfo
	}
	// Wait until sealing is terminated or a nonce is found
	var result *types.Header
	select {
	case <-stop:
		//		log.INFO("SEALER", "Sealer receive stop mine, curHeader", curHeader.HashNoSignsAndNonce().TerminalString())
		// Outside abort, stop all miner threads
		close(mineinfo.abort)
	case result = <-mineinfo.found:
		// One of the threads found a block, abort all others
		close(mineinfo.abort)
	}
	return result, nil
}
// Seal implements consensus.Engine, attempting to find a nonce that satisfies
// the block's difficulty requirements.
func (self *CpuAgent) SealPowX11( header *types.Header, stop <-chan struct{}, isBroadcastNode bool) (*types.Header, error) {
	log.INFO("seal", "挖矿", "开始", "高度", header.Number.Uint64())
	defer log.INFO("seal", "挖矿", "结束", "高度", header.Number.Uint64())


	mineinfo := mineInfo{
		abort:make(chan struct{}),
		found:make(chan *types.Header,len(self.sealTreads)),
		header:types.CopyHeader(header),
		powType : PowX11,
	}

	for _,thread := range self.sealTreads {
		thread.mineCh <- mineinfo
	}
	// Wait until sealing is terminated or a nonce is found
	var result *types.Header
	select {
	case <-stop:
		//		log.INFO("SEALER", "Sealer receive stop mine, curHeader", curHeader.HashNoSignsAndNonce().TerminalString())
		// Outside abort, stop all miner threads
		close(mineinfo.abort)
	case result = <-mineinfo.found:
		// One of the threads found a block, abort all others
		close(mineinfo.abort)
	}
	if result != nil{
		mineinfo := mineInfo{
			abort:make(chan struct{}),
			found:make(chan *types.Header,len(self.sealTreads)),
			header:types.CopyHeader(header),
			powType : PowSm3,
		}
		for _,thread := range self.sealTreads {
			thread.mineCh <- mineinfo
		}
		// Wait until sealing is terminated or a nonce is found
		result = nil
		select {
		case <-stop:
			//		log.INFO("SEALER", "Sealer receive stop mine, curHeader", curHeader.HashNoSignsAndNonce().TerminalString())
			// Outside abort, stop all miner threads
			close(mineinfo.abort)
		case result = <-mineinfo.found:
			// One of the threads found a block, abort all others
			close(mineinfo.abort)
		}
		return result,nil
	}
	return nil,nil
}
func (self *CpuAgent) mine(work *Work, stop <-chan struct{}) {
	if work.mineType == mineTaskTypePow {
		if result, err := self.SealPowOld(work.header, stop, work.isBroadcastNode); result != nil {
			log.Info("Successfully sealed new block", "number", result.Number)
			self.returnCh <- result
		} else {
			if err != nil {
				log.Warn("Block sealing failed", "err", err)
			}
			self.returnCh <- nil
		}
	}else{
		if result, err := self.SealPowX11(work.header, stop, work.isBroadcastNode); result != nil {
			log.Info("Successfully sealed new block", "number", result.Number)
			self.returnCh <- result
		} else {
			if err != nil {
				log.Warn("Block sealing failed", "err", err)
			}
			self.returnCh <- nil
		}

	}
}

func (self *CpuAgent) GetHashRate() int64 {
	return int64(self.manhash.Hashrate())
	//todo：从状态树获取
//	if pow, ok := self.manhash.Engine([]byte(manparams.VersionAlpha)).(consensus.PoW); ok {
//		return int64(pow.Hashrate())
//	}
	return 0
}
