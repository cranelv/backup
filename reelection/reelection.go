// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package reelection

import (
	"sync"
	"time"

	"encoding/json"
	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/election/support"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var (

	Time_Out_Limit        = 2 * time.Second
	ChanSize              = 10
)

const (
	Module = "换届服务"
)

// Backend wraps all methods required for mining.
type Backend interface {
	AccountManager() *accounts.Manager
	BlockChain() *core.BlockChain
	TxPool() *core.TxPool
	ChainDb() mandb.Database
}

type ElectMiner struct {
	MasterMiner []mc.ElectNodeInfo
	BackUpMiner []mc.ElectNodeInfo
}

type ElectValidator struct {
	MasterValidator    []mc.ElectNodeInfo
	BackUpValidator    []mc.ElectNodeInfo
	CandidateValidator []mc.ElectNodeInfo
}

type ElectReturnInfo struct {
	MasterMiner     []mc.ElectNodeInfo
	BackUpMiner     []mc.ElectNodeInfo
	MasterValidator []mc.ElectNodeInfo
	BackUpValidator []mc.ElectNodeInfo
}
type ReElection struct {
	bc  *core.BlockChain //eth实例：生成种子时获取一周期区块的最小hash
	ldb *leveldb.DB      //本都db数据库

	roleUpdateCh    chan *mc.RoleUpdatedMsg //身份变更信息通道
	roleUpdateSub   event.Subscription
	minerGenCh      chan *mc.MasterMinerReElectionRsp //矿工主节点生成消息通道
	minerGenSub     event.Subscription
	validatorGenCh  chan *mc.MasterValidatorReElectionRsq //验证者主节点生成消息通道
	validatorGenSub event.Subscription
	electionSeedCh  chan *mc.ElectionEvent //选举种子请求消息通道
	electionSeedSub event.Subscription

	//allNative AllNative

	currentID common.RoleType //当前身份

//	elect  baseinterface.ElectionInterface
	random *baseinterface.Random
	lock   sync.Mutex
}

func New(bc *core.BlockChain, dbDir string, random *baseinterface.Random) (*ReElection, error) {
	reelection := &ReElection{
		bc:             bc,
		roleUpdateCh:   make(chan *mc.RoleUpdatedMsg, ChanSize),
		minerGenCh:     make(chan *mc.MasterMinerReElectionRsp, ChanSize),
		validatorGenCh: make(chan *mc.MasterValidatorReElectionRsq, ChanSize),
		electionSeedCh: make(chan *mc.ElectionEvent, ChanSize),
		random:         random,

		currentID: common.RoleDefault,
	}
	var err error
	dbDir = dbDir + "/reElection"
	reelection.ldb, err = leveldb.OpenFile(dbDir, nil)
	if err != nil {
		return nil, err
	}
	err = reelection.initSubscribeEvent()
	if err != nil {
		return nil, err
	}
	go reelection.update()
	return reelection, nil
}

func (self *ReElection) initSubscribeEvent() error {
	var err error

	self.roleUpdateSub, err = mc.SubscribeEvent(mc.CA_RoleUpdated, self.roleUpdateCh)

	if err != nil {
		return err
	}
	log.INFO(Module, "CA_RoleUpdated", "订阅成功")
	return nil
}
func (self *ReElection) update() {
	defer func() {
		if self.roleUpdateSub != nil {
			self.roleUpdateSub.Unsubscribe()
		}

	}()
	for {
		select {
		case roleData := <-self.roleUpdateCh:
			log.INFO(Module, "roleData", roleData)
			//go self.roleUpdateProcess(roleData)
		}
	}
}

func GetAllNativeDataForUpdate(electstate mc.ElectGraph, electonline mc.ElectOnlineStatus, top *mc.TopologyGraph) support.AllNative {
	mapTopStatus := make(map[common.Address]common.RoleType, 0)
	for _, v := range top.NodeList {
		mapTopStatus[v.Account] = v.Type
	}
	native := support.AllNative{}
	mapELectStatus := make(map[common.Address]common.RoleType, 0)
	for _, v := range electstate.ElectList {
		mapELectStatus[v.Account] = v.Type
		switch v.Type {
		case common.RoleValidator:
			native.Master = append(native.Master, v)
		case common.RoleBackupValidator:
			native.BackUp = append(native.BackUp, v)
		case common.RoleCandidateValidator:
			native.Candidate = append(native.Candidate, v)
		}
	}
	for _, v := range electonline.ElectOnline {
		if v.Position != common.PosOnline { //过滤在线的
			continue
		}
		if _, ok := mapTopStatus[v.Account]; ok == true { //过滤当前不在拓扑图中的
			continue
		}
		if _, ok := mapELectStatus[v.Account]; ok == true { //在初选列表中的
			switch mapELectStatus[v.Account] {
			case common.RoleValidator:
				native.MasterQ = append(native.MasterQ, v.Account)
			case common.RoleBackupValidator:
				native.BackUpQ = append(native.BackUpQ, v.Account)
			case common.RoleCandidateValidator:
				native.CandidateQ = append(native.CandidateQ, v.Account)
			}
		}
	}
	return native
}
func GetOnlineAlter(offline []common.Address, online []common.Address, electonline mc.ElectOnlineStatus) []mc.Alternative {
	ans := []mc.Alternative{}
	mappOnlineStatus := make(map[common.Address]uint16)
	for _, v := range electonline.ElectOnline {
		mappOnlineStatus[v.Account] = v.Position
	}
	for _, v := range offline {
		if _, ok := mappOnlineStatus[v]; ok == false {
			log.ERROR(Module, "计算下线节点的alter时 下线节点不在初选列表中 账户", v.String())
			continue
		}
		if mappOnlineStatus[v] == common.PosOffline {
			log.ERROR(Module, "该节点已处于下线阶段 不需要上块 账户", v.String())
			continue
		}
		temp := mc.Alternative{
			A:        v,
			Position: common.PosOffline,
		}
		ans = append(ans, temp)
	}

	for _, v := range online {
		if _, ok := mappOnlineStatus[v]; ok == false {
			log.ERROR(Module, "计算上线节点的alter时 上线节点不在初选列表中 账户", v.String())
			continue
		}
		if mappOnlineStatus[v] == common.PosOnline {
			log.ERROR(Module, "该节点已处于上线阶段，不需要上块 账户", v.String())
			continue
		}
		temp := mc.Alternative{
			A:        v,
			Position: common.PosOnline,
		}
		ans = append(ans, temp)
	}
	log.INFO(Module, "计算上下线节点结果 online", online, "offline", offline, "ans", ans)
	return ans
}
func (self *ReElection) GetTopoChange(hash common.Hash, offline []common.Address, online []common.Address) ([]mc.Alternative, error) {
	//todo 从hash获取state， 得更换信息
	log.INFO(Module, "GetTopoChange", "start", "hash", hash, "online", online, "offline", offline)
	defer log.INFO(Module, "GetTopoChange", "end", "hash", hash, "online", online, "offline", offline)
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash获取高度失败 err", err)
		return []mc.Alternative{}, err
	}
	if common.IsReElectionNumber(height + 1) {
		log.ERROR(Module, "当前是广播区块 无差值", "height", height+1)
		return []mc.Alternative{}, err
	}
	lastHash, err := self.GetHeaderHashByNumber(hash, height-1)
	if err != nil {
		log.ERROR(Module, "根据hash找高度失败 hash ", hash, "高度", height-1)
		return []mc.Alternative{}, err
	}

	headerPos := self.bc.GetHeaderByHash(hash)
	stateDB, err := self.bc.StateAt(headerPos.Root)

	ElectGraphBytes := stateDB.GetMatrixData(matrixstate.GetKeyHash(mc.MSKeyElectGraph))
	var electState mc.ElectGraph
	if err := json.Unmarshal(ElectGraphBytes, &electState); err != nil {
		log.ERROR(Module, "GetElection Unmarshal err", err)
		return []mc.Alternative{}, err
	}
	ElectOnlineBytes := stateDB.GetMatrixData(matrixstate.GetKeyHash(mc.MSKeyElectOnlineState))
	var electOnlineState mc.ElectOnlineStatus
	if err := json.Unmarshal(ElectOnlineBytes, &electOnlineState); err != nil {
		log.ERROR(Module, "GetElection Unmarshal err", err)
		return []mc.Alternative{}, err
	}

	TopoGrap, err := GetCurrentTopology(lastHash, common.RoleBackupValidator|common.RoleValidator)
	if err != nil {
		log.Error(Module, "获取CA当前拓扑图失败 err", err)
		return []mc.Alternative{}, err
	}
	antive := GetAllNativeDataForUpdate(electState, electOnlineState, TopoGrap)
	DiffValidatot,err := self.TopoUpdate(antive, TopoGrap)
	if err!=nil{
		log.ERROR(Module,"拓扑更新失败 err",err,"高度",height)
	}

	olineStatus := GetOnlineAlter(offline, online, electOnlineState)
	DiffValidatot = append(DiffValidatot, olineStatus...)
	log.INFO(Module, "获取拓扑改变 end ", DiffValidatot)
	return DiffValidatot, nil
}

func (self *ReElection) GetElection(state *state.StateDB, hash common.Hash) (*ElectReturnInfo, error) {
	// todo 从状态树中获取elect
	log.INFO(Module, "GetElection", "start", "hash", hash)
	defer log.INFO(Module, "GetElection", "end", "hash", hash)
	preElectGraphBytes := state.GetMatrixData(matrixstate.GetKeyHash(mc.MSKeyElectGraph))
	var electState mc.ElectGraph
	if err := json.Unmarshal(preElectGraphBytes, &electState); err != nil {
		log.ERROR(Module, "GetElection Unmarshal err", err)
		return nil, err
	}
	log.INFO(Module, "开始获取选举信息 hash", hash.String())
	height, err := self.GetNumberByHash(hash)
	log.INFO(Module, "electStatte", electState, "高度", height, "err", err)
	if err != nil {
		log.Error(Module, "GetElection", "获取hash的高度失败")
		return nil, err
	}
	topStatus, err := self.HandleTopGen(hash)
	if err != nil {
		log.ERROR(Module, "GetElection err", err)
		return nil, err
	}
	data := &ElectReturnInfo{}

	if self.IsMinerTopGenTiming(hash) {
		log.INFO(Module, "GetElection", "IsMinerTopGenTiming", "高度", height)
		data.MasterMiner = append(data.MasterMiner, topStatus.MastM...)
		data.BackUpMiner = append(data.BackUpMiner, topStatus.BackM...)

	}
	if self.IsValidatorTopGenTiming(hash) {
		log.INFO(Module, "GetElection", "IsValidatorTopGenTiming", "高度", height)
		data.MasterValidator = append(data.MasterValidator, topStatus.MastV...)
		data.BackUpValidator = append(data.BackUpValidator, topStatus.BackV...)
	}

	log.INFO(Module, "不是任何网络切换时间点 height", height)

	return data, nil
}

func (self *ReElection)LastMinerGenTimeStamp(height uint64, types common.RoleType) (uint64,error) {

	data,err:=self.GetElectGenTimes(height)
	if err!=nil{
		log.ERROR(Module,"获取配置文件失败 err",err)
		return 0,err
	}
	minerGenTime:=uint64(data.MinerNetChange)
	validatorGenTime:=uint64(data.ValidatorNetChange)

	switch types {
	case common.RoleMiner:
		return common.GetNextReElectionNumber(height) - minerGenTime,nil
	default:
		return common.GetNextReElectionNumber(height) - validatorGenTime,nil
	}

}
func (self *ReElection) GetTopNodeInfo(hash common.Hash, types common.RoleType) ([]mc.ElectNodeInfo, []mc.ElectNodeInfo, []mc.ElectNodeInfo, error) {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash获取高度失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	heightPos,err := self.LastMinerGenTimeStamp(height, types)
	if err!=nil{
		log.ERROR(Module,"根据生成点高度失败",height,"types",types)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}

	hashPos, err := self.GetHeaderHashByNumber(hash, heightPos)
	log.INFO(Module, "GetTopNodeInfo pos", heightPos)
	if err != nil {
		log.ERROR(Module, "根据hash算父header失败 hash", hashPos)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	headerPos := self.bc.GetHeaderByHash(hashPos)
	stateDB, err := self.bc.StateAt(headerPos.Root)
	ElectGraphBytes := stateDB.GetMatrixData(matrixstate.GetKeyHash(mc.MSKeyElectGraph))
	var electState mc.ElectGraph
	if err := json.Unmarshal(ElectGraphBytes, &electState); err != nil {
		log.ERROR(Module, "GetElection Unmarshal err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	master := []mc.ElectNodeInfo{}
	backup := []mc.ElectNodeInfo{}
	cand := []mc.ElectNodeInfo{}

	switch types {
	case common.RoleMiner:
		for _, v := range electState.NextElect {
			switch v.Type {
			case common.RoleMiner:
				master = append(master, v)
			}
		}
	case common.RoleValidator:
		for _, v := range electState.NextElect {
			switch v.Type {
			case common.RoleValidator:
				master = append(master, v)
			case common.RoleBackupValidator:
				backup = append(backup, v)
			case common.RoleCandidateValidator:
				cand = append(cand, v)

			}
		}
	}
	return master, backup, cand, nil
}
func (self *ReElection) GetNetTopologyAll(hash common.Hash) (*ElectReturnInfo, error) {

	result := &ElectReturnInfo{}
	//todo 从hash获取state， 得全拓扑
	height, err := self.GetNumberByHash(hash)
	log.INFO(Module, "GetNetTopologyAll", "start", "height", height)
	defer log.INFO(Module, "GetNetTopologyAll", "end", "height", height)
	if err != nil {
		log.ERROR(Module, "根据hash获取高度失败 err", err)
		return nil, err
	}
	if common.IsReElectionNumber(height + 2) {
		masterV, backupV, _, err := self.GetTopNodeInfo(hash, common.RoleValidator)
		if err != nil {
			log.ERROR(Module, "获取验证者全拓扑图失败 err", err)
			return nil, err
		}
		masterM, backupM, _, err := self.GetTopNodeInfo(hash, common.RoleMiner)
		if err != nil {
			log.ERROR(Module, "获取矿工全拓扑图失败 err", err)
			return nil, err
		}

		result = &ElectReturnInfo{
			MasterMiner:     masterM,
			BackUpMiner:     backupM,
			MasterValidator: masterV,
			BackUpValidator: backupV,
		}
		log.INFO(Module, "是299 height", height)
		return result, nil

	}
	log.Info(Module, "不是广播区间前一块 不处理 height", height)
	return result, nil
}

func (self *ReElection) ProduceElectGraphData(block *types.Block, readFn matrixstate.PreStateReadFn) (interface{}, error) {
	log.INFO(Module, "ProduceElectGraphData", "start", "height", block.Header().Number.Uint64())
	defer log.INFO(Module, "ProduceElectGraphData", "end", "height", block.Header().Number.Uint64())
	if err := CheckBlock(block); err != nil {
		log.ERROR(Module, "ProduceElectGraphData CheckBlock err ", err)
		return nil, err
	}
	data, err := readFn(mc.MSKeyElectGraph)
	log.INFO(Module, "data", data, "err", err)
	if err != nil {
		log.ERROR(Module, "readFn 失败 key", mc.MSKeyElectGraph, "err", err)
		return nil, err
	}
	electStates, OK := data.(*mc.ElectGraph)
	if OK == false || electStates == nil {
		log.ERROR(Module, "ElectStates 非法", "反射失败")
		return nil, err
	}
	electStates.Number = block.Header().Number.Uint64()

	currentHash := block.ParentHash()
	topState, err := self.HandleTopGen(currentHash)
	if self.IsMinerTopGenTiming(currentHash) {
		//electStates.NextElect = []mc.ElectNodeInfo{}
		electStates.NextElect = append(electStates.NextElect, topState.MastM...)
		electStates.NextElect = append(electStates.NextElect, topState.BackM...)
		electStates.NextElect = append(electStates.NextElect, topState.CandM...)
	}
	if self.IsValidatorTopGenTiming(currentHash) {
		electStates.NextElect = append(electStates.NextElect, topState.MastV...)
		electStates.NextElect = append(electStates.NextElect, topState.BackV...)
		electStates.NextElect = append(electStates.NextElect, topState.CandV...)
	}
	if common.IsReElectionNumber(block.Header().Number.Uint64() + 1) {
		nextElect := electStates.NextElect
		electList := []mc.ElectNodeInfo{}
		for _, v := range nextElect {

			switch v.Type {
			case common.RoleBackupValidator:
				electList = append(electList, v)
			case common.RoleValidator:
				electList = append(electList, v)
			case common.RoleMiner:
				electList = append(electList, v)
			case common.RoleCandidateValidator:
				electList = append(electList, v)
			}
		}
		electStates.ElectList = []mc.ElectNodeInfo{}
		electStates.ElectList = append(electStates.ElectList, electList...)
		electStates.NextElect = []mc.ElectNodeInfo{}
	}

	return SloveElectStatus(electStates)
}

func (self *ReElection) ProduceElectOnlineStateData(block *types.Block, readFn matrixstate.PreStateReadFn) (interface{}, error) {
	log.INFO(Module, "ProduceElectOnlineStateData", "start", "height", block.Header().Number.Uint64())
	defer log.INFO(Module, "ProduceElectOnlineStateData", "end", "height", block.Header().Number.Uint64())
	if err := CheckBlock(block); err != nil {
		log.ERROR(Module, "ProduceElectGraphData CheckBlock err ", err)
		return []byte{}, err
	}
	height := block.Header().Number.Uint64()

	if common.IsReElectionNumber(height + 1) {
		electOnline := mc.ElectOnlineStatus{
			Number: height,
		}
		masterV, backupV, CandV, err := self.GetTopNodeInfo(block.Header().ParentHash, common.RoleValidator)
		if err != nil {
			log.ERROR(Module, "获取验证者全拓扑图失败 err", err)
			return nil, err
		}
		for _, v := range masterV {
			tt := v
			tt.Position = common.PosOnline
			electOnline.ElectOnline = append(electOnline.ElectOnline, tt)
		}
		for _, v := range backupV {
			tt := v
			tt.Position = common.PosOnline
			electOnline.ElectOnline = append(electOnline.ElectOnline, tt)
		}
		for _, v := range CandV {
			tt := v
			tt.Position = common.PosOnline
			electOnline.ElectOnline = append(electOnline.ElectOnline, tt)
		}
		return SloveOnlineStatus(&electOnline)
	}

	header := self.bc.GetHeaderByHash(block.Header().ParentHash)
	data, err := readFn(mc.MSKeyElectOnlineState)
	log.INFO(Module, "data", data, "err", err)
	if err != nil {
		log.ERROR(Module, "readFn 失败 key", mc.MSKeyElectOnlineState, "err", err)
		return []byte{}, err
	}
	electStates, OK := data.(*mc.ElectOnlineStatus)
	if OK == false || electStates == nil {
		log.ERROR(Module, "ElectStates 非法", "反射失败")
		return []byte{}, err
	}
	mappStatus := make(map[common.Address]uint16)
	for _, v := range header.NetTopology.NetTopologyData {
		switch v.Position {
		case common.PosOnline:
			mappStatus[v.Account] = common.PosOnline
		case common.PosOffline:
			mappStatus[v.Account] = common.PosOffline
		}
	}
	for k, v := range electStates.ElectOnline {
		if _, ok := mappStatus[v.Account]; ok == false {
			continue
		}
		electStates.ElectOnline[k].Position = mappStatus[v.Account]
	}

	return SloveOnlineStatus(electStates)
}

func (self *ReElection) ProducePreBroadcastStateData(block *types.Block, readFn matrixstate.PreStateReadFn) (interface{}, error) {
	if err := CheckBlock(block); err != nil {
		log.ERROR(Module, "ProducePreBroadcastStateData CheckBlock err ", err)
		return []byte{}, err
	}
	height := block.Header().Number.Uint64()
	if common.IsBroadcastNumber(height-1) == false {
		return nil, nil
	}
	data, err := readFn(mc.MSKeyPreBroadcastRoot)
	if err != nil {
		log.ERROR(Module, "readFn 失败 key", mc.MSKeyPreBroadcastRoot, "err", err)
		return nil, err
	}
	preBroadcast, OK := data.(*mc.PreBroadStateDB)
	if OK == false || preBroadcast == nil {
		log.ERROR(Module, "PreBroadStateDB 非法", "反射失败")
		return nil, err
	}
	header := self.bc.GetHeaderByHash(block.ParentHash())
	if header == nil {
		log.ERROR(Module, "根据hash算区块头失败 高度", block.Number().Uint64())
		return nil, errors.New("header is nil")
	}
	stateDB, err := self.bc.StateAt(header.Root)
	if err != nil {
		log.ERROR(Module, "根据高度获取state失败", header.Number.Uint64())
		return nil, nil
	}
	preBroadcast.BeforeLastStateDb = preBroadcast.LastStateDB
	preBroadcast.LastStateDB = stateDB
	return preBroadcast, nil

}
