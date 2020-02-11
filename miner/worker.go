// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package miner

import (
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/event"
	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/MatrixAINetwork/go-matrix/msgsend"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/consensus/manash"
	"github.com/pkg/errors"
	"github.com/MatrixAINetwork/go-matrix/consensus/amhash"
)

const (
	resultQueueSize   = 10
	chainHeadChanSize = 10
)

// Agent can register themself with the worker
type Agent interface {
	Work() chan<- *Work
	SetReturnCh(chan<- *types.Header)
	Stop()
	Start()
	GetHashRate() int64
}
type mineTaskType int

const (
	mineTaskTypeX11 mineTaskType = 1
	mineTaskTypeAI               = 2
	mineTaskTypePow              = 3
)
// Work is the workers current environment and holds
// all of the current state information
type Work struct {
	config *params.ChainConfig
	signer types.Signer

	state *state.StateDBManage // apply state changes here
	Block *types.Block         // the new block

	header   *types.Header
	txs      []types.SelfTransaction
	receipts []*types.Receipt

	createdAt time.Time

	isBroadcastNode bool

	mineType mineTaskType
}

type Result struct {
	Difficulty *big.Int
	Header     *types.Header
}

// worker is the main object which takes care of applying messages to the new state
type worker struct {
	config *params.ChainConfig
	manhash *manash.Manash
	amhash  *amhash.Amhash
	mu sync.Mutex

	// update loop
	mux *event.TypeMux

	agents map[Agent]struct{}
	recv   chan *types.Header

	extra []byte

	currentMu sync.Mutex
	current   *Work
	// atomic status counters
	mining int32
	atWork int32

	quitCh                chan struct{}
	roleUpdateCh		  chan *types.Header
	roleUpdateSub         event.Subscription
	miningRequestCh       chan *mc.HD_MiningReqMsg
	miningRequestSub      event.Subscription
	localMiningRequestCh  chan *mc.BlockGenor_BroadcastMiningReqMsg
//	localMiningRequestSub event.Subscription
	mineReqCtrl           *mineReqCtrl
	hd                    *msgsend.HD
	mineResultSender      *common.ResendMsgCtrl
}

type ChainReader interface {
	Config() *params.ChainConfig
	Engine(version []byte) consensus.Engine
	DPOSEngine(version []byte) consensus.DPOSEngine
	VerifyHeader(header *types.Header) error
	GetCurrentHash() common.Hash
	GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error)
	GetBroadcastAccounts(blockHash common.Hash) ([]common.Address, error)
	GetInnerMinerAccounts(blockHash common.Hash) ([]common.Address, error)
	GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBroadcastIntervalByHash(blockHash common.Hash) (*mc.BCIntervalInfo, error)
	GetA0AccountFromAnyAccount(account common.Address, blockHash common.Hash) (common.Address, common.Address, error)
	CurrentHeader() *types.Header
	// GetBlock retrieves a block from the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header
}

func newWorker(config *params.ChainConfig, manh *manash.Manash,amhash *amhash.Amhash, mux *event.TypeMux, hd *msgsend.HD) (*worker, error) {
	worker := &worker{
		config: config,
		manhash:     manh,
		amhash:		amhash,
		mux:    mux,

		agents:               make(map[Agent]struct{}),
		quitCh:               make(chan struct{}),
		miningRequestCh:      make(chan *mc.HD_MiningReqMsg, 100),
		roleUpdateCh:         make(chan *types.Header, 2),
		recv:                 make(chan *types.Header, resultQueueSize),
		localMiningRequestCh: make(chan *mc.BlockGenor_BroadcastMiningReqMsg, 100),
		mineReqCtrl:          newMinReqCtrl(nil),
		hd:                   hd,
		mineResultSender:     nil,
	}

	atomic.StoreInt32(&worker.mining, 0)

	err := worker.init_SubscribeEvent()
	if err != nil {
		log.Error(ModuleMiner, "worker创建失败", err)
		return nil, err
	}
	go worker.update()
	go worker.wait()
	log.INFO(ModuleMiner, "worker创建成功", err)
	return worker, nil
}
func (self *worker) init_SubscribeEvent() error {
	var err error
//	self.roleUpdateSub, err = mc.SubscribeEvent(mc.NewBlockMessage, self.roleUpdateCh)
//	self.localMiningRequestSub, err = mc.SubscribeEvent(mc.HD_BroadcastMiningReq, self.localMiningRequestCh) //广播节点
//	if err != nil {
//		log.Error(ModuleMiner, "广播节点挖矿请求订阅失败", err)
//		return err
//	} else {
//		log.INFO(ModuleMiner, "广播节点挖矿请求订阅成功", "")
//	}

//	self.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, self.roleUpdateCh) //身份到达
//	if err != nil {
//		log.Error(ModuleMiner, "身份更新订阅失败", err)
//		return err
//	} else {
//		log.INFO(ModuleMiner, "身份更新订阅成功", "")
//	}

	self.miningRequestSub, err = mc.SubscribeEvent(mc.HD_V2_MiningReq, self.miningRequestCh) //挖矿请求
	if err != nil {
		log.Error(ModuleMiner, "普通矿工挖矿请求订阅失败", err)
		return err
	} else {
		log.INFO(ModuleMiner, "普通矿工挖矿请求订阅成功", err)
	}
	return nil

}

func (self *worker) Getmining() int32 { return atomic.LoadInt32(&self.mining) }

func (self *worker) update() {
	defer func() {
		if self.miningRequestSub != nil {
			self.miningRequestSub.Unsubscribe()
		}
		if self.roleUpdateSub != nil {
			self.roleUpdateSub.Unsubscribe()
		}
		self.StopAgent()
		self.stopMineResultSender()
		log.INFO("矿工节点退出成功")
	}()
	self.StartAgent()
	self.beginMine()

	for {
		select {
		case roleData := <-self.roleUpdateCh:
			self.RoleUpdatedMsgHandler(roleData.Number.Uint64())

		case minerReqData := <-self.miningRequestCh:
			self.MiningRequestHandle(minerReqData)

		case data := <-self.localMiningRequestCh:
			self.BroadcastHashLocalMiningReqMsgHandle(data.BlockMainData)

		case <-self.miningRequestSub.Err():
			return
		case <-self.quitCh:
			return
		}
	}
}
func (self *worker) RoleUpdatedMsgHandler(blockNum uint64) {
//	if data.SuperSeq > self.mineReqCtrl.curSuperSeq {
//		self.StopAgent()
//		self.stopMineResultSender()
//		self.mineReqCtrl.Clear()
//		self.mineReqCtrl.curSuperSeq = data.SuperSeq
//	}

//	if blockNum+1 > self.mineReqCtrl.curNumber {
//		self.stopMineResultSender()
//	}
	/*
	canMining := self.mineReqCtrl.CanMining()
	if canMining {
		self.StartAgent()
		self.processMineReq()
	} else {
		self.StopAgent()
	}
	*/
}

func (self *worker) MiningRequestHandle(data *mc.HD_MiningReqMsg) {
	if nil == data || nil == data.Header {
		log.ERROR(ModuleMiner, "挖矿请求Msg", "nil")
		return
	}
	reqData, err := self.mineReqCtrl.AddMineReq(data.Header, data.From, false,data.IsRemote)
	if err != nil {
		log.ERROR(ModuleMiner, "缓存挖矿请求", err)
		return
	}
	if reqData != nil {
		self.processAppointedMineReq(reqData)

	}
}
func (self *worker) MineAI(header *types.Header, miner common.Address){
	bcInterval, err := GetBroadcastIntervalByHash(header.ParentHash)
	if err != nil || bcInterval == nil {
		return
	}


	if IsAIBlock(header.Number.Uint64(),bcInterval.GetBroadcastInterval()) == false {
		return
	}


	aiMiningNumber := GetNextAIBlockNumber(header.Number.Uint64(), bcInterval.GetBroadcastInterval())
	mineHash := header.HashNoSignsAndNonce()
	var task *aiMineTask
	if bcInterval.IsReElectionNumber(aiMiningNumber - 1) {
		task = nil
		return
	} else {
		task = newAIMineTask(mineHash, header, aiMiningNumber, bcInterval)
		task.aiMiner = miner
		task.mineHeader.AICoinbase = miner
	}
	stop:=  make(chan struct{}, 1)
	result, err := self.amhash.SealAI(nil, task.mineHeader, stop)
	if err == nil && result != nil{
		task.minedAI = true
		task.aiMiner = result.AICoinbase
		task.aiHash = result.AIHash

		sendData := &aiMineTask{
			mineHash:       task.mineHash,
			mineHeader:     result,
			minedAI:        true,
			aiMiningNumber: task.aiMiningNumber,
			aiMiner:        task.aiMiner,
			aiHash:         task.aiHash,
		}

		_, err := common.NewResendMsgCtrl(sendData, self.sendAIMineResultFunc, 100, 100)
		if err != nil {
			log.ERROR(ModuleMiner, "创建挖矿结果发送器", "失败", "err", err)
			return
		}
	}
}
func (self *worker) sendAIMineResultFunc(data interface{}, times uint32) {
	resultData, OK := data.(*aiMineTask)
	if !OK {
		log.Error(ModuleMiner, "发出AI挖矿结果", "反射消息失败", "次数", times)
		return
	}

	if nil == resultData {
		log.Error(ModuleMiner, "发出AI挖矿结果", "入参错误", "次数", times)
		return
	}

	rsp := &mc.HD_V2_AIMiningRspMsg{
		Number:     resultData.aiMiningNumber,
		BlockHash:  resultData.mineHash,
		AIHash:     resultData.aiHash,
		AICoinbase: resultData.aiMiner,
	}

	self.hd.SendNodeMsg(mc.HD_V2_AIMiningRsp, rsp,resultData.aiMiner, common.RoleValidator, nil)
	log.Trace(ModuleMiner, "AI挖矿结果", "发送", "parent hash", rsp.BlockHash.TerminalString(), "次数", times, "高度", rsp.Number, "AIHash", rsp.AIHash.TerminalString(), "AIMiner", rsp.AICoinbase.Hex())
}

func (self *worker) BroadcastHashLocalMiningReqMsgHandle(req *mc.BlockData) {
	if nil == req || nil == req.Header {
		log.ERROR(ModuleMiner, "广播挖矿请求Msg", "nil")
		return
	}
	log.Trace(ModuleMiner, "广播请求", req.Header.Number)
	reqData, err := self.mineReqCtrl.AddMineReq(req.Header, common.Address{}, true,false)
	if err != nil {
		log.ERROR(ModuleMiner, "缓存请求Err", err)
		return
	}

	if reqData != nil {
		self.processAppointedMineReq(reqData)
	}
}

func (self *worker) Stop() {
	close(self.quitCh)
}

func (self *worker) wait() {
	for {
		for header := range self.recv {
			atomic.AddInt32(&self.atWork, -1)

			if header == nil {
				continue
			}
			self.foundHandle(header)
		}
	}
}

func (self *worker) foundHandle(header *types.Header) {
	cache, err := self.mineReqCtrl.SetMiningResult(header,header.ParentHash)
	if err != nil {
		log.ERROR(ModuleMiner, "结果保存失败", err)
		return
	}
//	log.Info("SSSSSSSSSSSSSSSSSSSSSSSSS","number",header.Number,"coinbase",base58.Base58EncodeToString("MAN",header.Coinbase))
	self.startMineResultSender(cache)
	for agent := range self.agents {
		if ch := agent.Work(); ch != nil {
			ch <- nil
		}
	}
}
func GetBroadcastIntervalByHash(hash common.Hash) (*mc.BCIntervalInfo, error) {
	return &mc.BCIntervalInfo{LastBCNumber: 0, LastReelectNumber: 0, BCInterval: 100}, nil
}

func CreateMinePowerTask(mineHeader *types.Header) *powMineTask {
	bcInterval, err := GetBroadcastIntervalByHash(mineHeader.ParentHash)
	if err != nil || bcInterval == nil {
		return nil
	}


	if IsAIBlock(mineHeader.Number.Uint64(),bcInterval.GetBroadcastInterval()) == false {
		return nil
	}


	powMiningNumber := mineHeader.Number.Uint64() + PowBlockPeriod - 1

	mineHash := mineHeader.HashNoSignsAndNonce()
	difficulty := mineHeader.Difficulty

	powTask := newPowMineTask(mineHash, mineHeader, powMiningNumber, bcInterval, difficulty)
	return powTask
}
func (self *worker) createMineTask(mineHeader *types.Header, verify bool) (powTask *powMineTask, aiTask *aiMineTask, returnErr error) {
	bcInterval, err := GetBroadcastIntervalByHash(mineHeader.ParentHash)
	if err != nil || bcInterval == nil {
		return nil, nil, errors.Errorf("get broadcast interval err: %v", err)
	}


	if IsAIBlock(mineHeader.Number.Uint64(),bcInterval.GetBroadcastInterval()) == false {
		return nil, nil, errors.Errorf("mine header is not ai header")
	}


	powMiningNumber := mineHeader.Number.Uint64() + PowBlockPeriod - 1
	aiMiningNumber := GetNextAIBlockNumber(mineHeader.Number.Uint64(), bcInterval.GetBroadcastInterval())

	mineHash := mineHeader.HashNoSignsAndNonce()
	difficulty := mineHeader.Difficulty

	powTask = newPowMineTask(mineHash, mineHeader, powMiningNumber, bcInterval, difficulty)
	if bcInterval.IsReElectionNumber(aiMiningNumber - 1) {
		aiTask = nil
	} else {
		aiTask = newAIMineTask(mineHash, mineHeader, aiMiningNumber, bcInterval)
	}

	return powTask, aiTask, nil
}
func (self *worker) setExtra(extra []byte) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.extra = extra
}

func (self *worker) pending() (*types.Block, *state.StateDBManage) {
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	tx, rx := types.GetCoinTXRS(self.current.txs, self.current.receipts)
	cb := types.MakeCurencyBlock(tx, rx, nil)
	if atomic.LoadInt32(&self.mining) == 0 {
		return types.NewBlock(
			self.current.header,
			cb,
			nil,
		), self.current.state.Copy()
	}
	return self.current.Block, self.current.state.Copy()
}

func (self *worker) pendingBlock() *types.Block {
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	if atomic.LoadInt32(&self.mining) == 0 {
		tx, rx := types.GetCoinTXRS(self.current.txs, self.current.receipts)
		cb := types.MakeCurencyBlock(tx, rx, nil)
		return types.NewBlock(
			self.current.header,
			cb,
			nil,
		)
	}

	self.currentMu.Lock()
	defer self.currentMu.Unlock()
	return self.current.Block
}

func (self *worker) StartAgent() {
	self.mu.Lock()
	defer self.mu.Unlock()
	atomic.StoreInt32(&self.mining, 1)

	// spin up agents
	for agent := range self.agents {
		agent.Start()
	}
}

/*
func (self *worker) StopMiner() {
	self.mu.Lock()
	defer self.mu.Unlock()
	if atomic.LoadInt32(&self.mining) != 0 {
		for agent := range self.agents {
			agent.Stop()
		}
	}
	atomic.StoreInt32(&self.atWork, 0)
}
*/
func (self *worker) StopAgent() {
	self.mu.Lock()
	defer self.mu.Unlock()
	if atomic.LoadInt32(&self.mining) != 0 {
		for agent := range self.agents {
			agent.Stop()
		}
	}
	atomic.StoreInt32(&self.mining, 0)
	atomic.StoreInt32(&self.atWork, 0)
}

func (self *worker) Register(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.agents[agent] = struct{}{}
	agent.SetReturnCh(self.recv)
}

func (self *worker) Unregister(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	delete(self.agents, agent)
	agent.Stop()
}

// push sends a new work task to currently live miner agents.
func (self *worker) push(work *Work) {
	if atomic.LoadInt32(&self.mining) != 1 {
		return
	}
	atomic.AddInt32(&self.atWork, 1)
	for agent := range self.agents {
		if ch := agent.Work(); ch != nil {
			ch <- work
		}
	}
}


func (self *worker) CommitNewWork(reqData MinerRequestInterface) {
	work := reqData.CreateWork()
	log.INFO(ModuleMiner, "挖矿", "开始")

	//// Create the current work task and check any fork transitions needed
	self.current = work
	self.push(work)
}

func (self *worker) processAppointedMineReq(reqData MinerRequestInterface) {
	if nil == reqData {
		return
	}
//	if self.mineReqCtrl.CanMining() == false {
//		return
//	}

	if reqData.IsMined() {
		log.Trace(ModuleMiner, "请求已完成，直接发送结果")
		self.sendMineResultFunc(reqData, 0)
	} else {
		log.Trace(ModuleMiner, "接收请求，开始处理")
		self.beginMine()
	}
}

func (self *worker) beginMine() {
	maxReq := self.mineReqCtrl.getBeginMine()
	if maxReq != nil{
		self.CommitNewWork(maxReq)
		req := maxReq.(*powMineTask)
		go self.MineAI(maxReq.RequstHeader(),req.coinbase)
	}
//	self.StopMiner()
}

func (self *worker) startMineResultSender(data MinerRequestInterface) {
	self.stopMineResultSender()
	sender, err := common.NewResendMsgCtrl(data, self.sendMineResultFunc, 100, 100)
	if err != nil {
		log.ERROR(ModuleMiner, "创建挖矿结果发送器", "失败", "err", err)
		return
	}
	self.mineResultSender = sender
}

func (self *worker) stopMineResultSender() {
	if self.mineResultSender == nil {
		return
	}
	self.mineResultSender.Close()
	self.mineResultSender = nil
	log.Trace(ModuleMiner, "挖矿结果发送器", "停止")
}

func (self *worker) sendMineResultFunc(data interface{}, times uint32) {
	switch data.(type) {
	case *mineReqData:
		self.sendMineOldResultFunc(data, times)
	case *powMineTask:
		self.sendMineX11ResultFunc(data,times)
	}
}
func (self *worker) sendMineOldResultFunc(data interface{}, times uint32) {
	resultData, OK := data.(*mineReqData)
	if !OK {
		log.ERROR(ModuleMiner, "发出挖矿结果", "反射消息失败", "次数", times)
		return
	}

	if nil == resultData || nil == resultData.header || resultData.mined == false {
		log.ERROR(ModuleMiner, "发出挖矿结果", "入参错误", "次数", times)
		return
	}

	if err := resultData.ResendMineResult(time.Now().UnixNano()); err != nil {
		log.ERROR(ModuleMiner, "发出挖矿结果", "发送挖矿结果失败", "原因", err, "次数", times)
		return
	}

	if resultData.isBroadcastReq {
	} else {
		rsp := &mc.HD_MiningRspMsg{
			BlockHash:  resultData.headerHash,
			Difficulty: resultData.mineDiff,
			Number:     resultData.header.Number.Uint64(),
			Nonce:      resultData.header.Nonce,
			Coinbase:   resultData.header.Coinbase,
			MixDigest:  resultData.header.MixDigest,
			Signatures: resultData.header.Signatures}
			if !resultData.isFriend {
				self.hd.SendNodeMsg(mc.HD_MiningRsp, rsp,resultData.header.Coinbase, common.RoleValidator, nil)
			}else{
				self.hd.SendNodeMsg(mc.HD_MiningRsp, rsp,resultData.header.Coinbase, common.RoleBroadcast, nil)
			}
//		log.Trace(ModuleMiner, "挖矿结果", "发送", "hash", rsp.BlockHash.TerminalString(), "Difficulty",rsp.Difficulty,"次数", times, "高度", rsp.Number,"coinbase",rsp.Coinbase, "Nonce", rsp.Nonce)
	}
}
func (self *worker) sendMineX11ResultFunc(data interface{}, times uint32) {
	resultData, OK := data.(*powMineTask)
	if !OK {
		log.Error(ModuleMiner, "发出POW挖矿结果", "反射消息失败", "次数", times)
		return
	}

	if nil == resultData {
		log.Error(ModuleMiner, "发出POW挖矿结果", "入参错误", "次数", times)
		return
	}

	rsp := &mc.HD_V2_PowMiningRspMsg{
		Number:     resultData.powMiningNumber,
		BlockHash:  resultData.mineHash,
		Difficulty: resultData.powMiningDifficulty,
		Nonce:      resultData.mineHeader.Nonce,
		Coinbase:   resultData.coinbase,
		MixDigest:  resultData.mineHeader.MixDigest,
		Sm3Nonce:   resultData.mineHeader.Sm3Nonce,
	}

	if !resultData.isFriend {
		self.hd.SendNodeMsg(mc.HD_V2_PowMiningRsp, rsp,resultData.coinbase, common.RoleValidator, nil)
	}else{
		self.hd.SendNodeMsg(mc.HD_V2_PowMiningRsp, rsp,resultData.coinbase, common.RoleBroadcast, nil)
	}

}
