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
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	BroadCastInterval     = common.GetBroadcastInterval()
	MinerAccount          = common.GetReElectionInterval() - manparams.MinerTopologyGenerateUpTime
	MinerTopGenTiming     = common.GetReElectionInterval() - manparams.MinerNetChangeUpTime
	ValidatorAccount      = common.GetReElectionInterval() - manparams.VerifyTopologyGenerateUpTime
	ValidatorTopGenTiming = common.GetReElectionInterval() - manparams.VerifyNetChangeUpTime
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

	elect  baseinterface.ElectionInterface
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
	reelection.elect = baseinterface.NewElect()
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
func (self *ReElection) GetTopoChange(hash common.Hash, offline []common.Address, online []common.Address) ([]mc.Alternative, error) {
	//todo 从hash获取state， 得更换信息

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

	ElectGraphBytes := stateDB.GetMatrixData(matrixstate.GetKeyHash(matrixstate.MSPElectGraph))
	var electState mc.ElectGraph
	if err := json.Unmarshal(ElectGraphBytes, &electState); err != nil {
		log.ERROR(Module, "GetElection Unmarshal err", err)
		return []mc.Alternative{}, err
	}
	ElectOnlineBytes := stateDB.GetMatrixData(matrixstate.GetKeyHash(matrixstate.MSPElectOnlineState))
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
	DiffValidatot := self.TopoUpdate(antive, TopoGrap)
	log.INFO(Module, "获取拓扑改变 end ", DiffValidatot)

	return []mc.Alternative{}, nil

	//height, err := self.GetNumberByHash(hash)
	//if err != nil {
	//	return []mc.Alternative{}, errors.New("根据hash获取高度失败")
	//}
	//height = height + 1
	//self.lock.Lock()
	//defer self.lock.Unlock()
	//if common.IsReElectionNumber(height) {
	//	log.INFO(Module, "是换届区块", "无差值")
	//	return []mc.Alternative{}, nil
	//}
	//
	//log.INFO(Module, "获取拓扑改变 start height", height, "offline", offline)
	//lastHash, err := self.GetHeaderHashByNumber(hash, height-1)
	//if err != nil {
	//	log.Error(Module, "根据hash获取高度失败 err", err)
	//	return []mc.Alternative{}, err
	//}
	//self.checkUpdateStatus(lastHash)
	//antive, err := self.readNativeData(lastHash)
	//if err != nil {
	//	log.Error(Module, "获取上一个高度的初选列表失败 height-1", height-1)
	//	return []mc.Alternative{}, err
	//}
	//
	////aim := 0x04 + 0x08
	//TopoGrap, err := GetCurrentTopology(lastHash, common.RoleBackupValidator|common.RoleValidator)
	//if err != nil {
	//	log.Error(Module, "获取CA当前拓扑图失败 err", err)
	//	return []mc.Alternative{}, err
	//}
	//
	//log.Info(Module, "获取拓扑变化 start 上一个高度缓存allNative-M", antive.MasterQ, "B", antive.BackUpQ, "Can", antive.CandidateQ)
	//DiffValidatot := self.TopoUpdate(offline, antive, TopoGrap)
	//log.INFO(Module, "获取拓扑改变 end ", DiffValidatot)
	//return DiffValidatot, nil

}

func (self *ReElection) GetElection(state *state.StateDB, hash common.Hash) (*ElectReturnInfo, error) {
	// todo 从状态树中获取elect
	preElectGraphBytes := state.GetMatrixData(matrixstate.GetKeyHash(matrixstate.MSPElectGraph))
	var electState mc.ElectGraph
	if err := json.Unmarshal(preElectGraphBytes, &electState); err != nil {
		log.ERROR(Module, "GetElection Unmarshal err", err)
		return nil, err
	}
	log.INFO(Module, "开始获取选举信息 hash", hash.String())
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.Error(Module, "GetElection", "获取hash的高度失败")
		return nil, err
	}
	if common.IsReElectionNumber(height + 1 + manparams.MinerNetChangeUpTime) {
		log.Error(Module, "是矿工网络生成切换时间点 height", height)

		resultM := &ElectReturnInfo{}
		nextElect := electState.NextElect
		for _, v := range nextElect {
			types := common.GetRoleTypeFromPosition(v.Position)
			switch types {
			case common.RoleMiner:
				resultM.MasterMiner = append(resultM.MasterMiner, v)
			}
		}
		return resultM, nil
	} else if common.IsReElectionNumber(height + 1 + manparams.VerifyNetChangeUpTime) {
		log.Error(Module, "是验证者网络切换时间点 height", height)
		resultV := &ElectReturnInfo{}
		for _, v := range electState.NextElect {
			types := common.GetRoleTypeFromPosition(v.Position)
			switch types {
			case common.RoleValidator:
				resultV.MasterValidator = append(resultV.MasterValidator, v)
			case common.RoleBackupValidator:
				resultV.BackUpValidator = append(resultV.BackUpValidator, v)

			}
		}
		return resultV, nil
	}
	log.INFO(Module, "不是任何网络切换时间点 height", height)
	temp := &ElectReturnInfo{}
	return temp, nil
}

func LastMinerGenTimeStamp(height uint64, types common.RoleType) uint64 {
	switch types {
	case common.RoleMiner:
		return common.GetNextReElectionNumber(height) - manparams.MinerNetChangeUpTime
	default:
		return common.GetNextReElectionNumber(height) - manparams.VerifyNetChangeUpTime
	}

}
func (self *ReElection) GetTopNodeInfo(hash common.Hash, types common.RoleType) ([]mc.ElectNodeInfo, []mc.ElectNodeInfo, []mc.ElectNodeInfo, error) {
	height, err := self.GetNumberByHash(hash)
	if err != nil {
		log.ERROR(Module, "根据hash获取高度失败 err", err)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	heightPos := LastMinerGenTimeStamp(height, types)

	hashPos, err := self.GetHeaderHashByNumber(hash, heightPos)
	if err != nil {
		log.ERROR(Module, "根据hash算父header失败 hash", hashPos)
		return []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, []mc.ElectNodeInfo{}, err
	}
	headerPos := self.bc.GetHeaderByHash(hashPos)
	stateDB, err := self.bc.StateAt(headerPos.Root)
	ElectGraphBytes := stateDB.GetMatrixData(matrixstate.GetKeyHash(matrixstate.MSPElectGraph))
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

/*
type ElectGraph struct {
	Number        uint64
	ElectList     []ElectNodeInfo
	CandidateList []ElectNodeInfo
	NextElect     []ElectNodeInfo
}
type ElectOnlineStatus struct {
	Number  uint64
	MasterV []common.Address
	BackV   []common.Address
	CandV   []common.Address
}
*/
func (self *ReElection) ProduceElectGraphData(block *types.Block, readFn matrixstate.PreStateReadFn) ([]byte, error) {
	if err := CheckBlock(block); err != nil {
		log.ERROR(Module, "ProduceElectGraphData CheckBlock err ", err)
		return []byte{}, err
	}
	data, err := readFn(matrixstate.MSPTopologyGraph)
	if err != nil {
		log.ERROR(Module, "readFn 失败 key", matrixstate.MSPTopologyGraph, "err", err)
		return []byte{}, err
	}
	var electStates mc.ElectGraph
	err = json.Unmarshal(data, &electStates)
	if err != nil {
		log.ERROR(Module, "ElectStates Unmarshal失败 err", err)
		return []byte{}, err
	}
	electStates.Number = block.Header().Number.Uint64()

	currentHash := block.Hash()
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
			types := common.GetRoleTypeFromPosition(v.Position)
			switch types {
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

func (self *ReElection) ProduceElectOnlineStateData(block *types.Block, readFn matrixstate.PreStateReadFn) ([]byte, error) {
	if err := CheckBlock(block); err != nil {
		log.ERROR(Module, "ProduceElectGraphData CheckBlock err ", err)
		return []byte{}, err
	}
	height := block.Header().Number.Uint64()

	if common.IsReElectionNumber(height + 1) {
		electOnline := mc.ElectOnlineStatus{
			Number: height,
		}
		masterV, backupV, CandV, err := self.GetTopNodeInfo(block.Header().Hash(), common.RoleValidator)
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
		return SloveOnlineStatus(electOnline)
	}

	header := self.bc.GetHeaderByHash(block.Header().ParentHash)
	data, err := readFn(matrixstate.MSPElectOnlineState)
	if err != nil {
		log.ERROR(Module, "readFn 失败 key", matrixstate.MSPTopologyGraph, "err", err)
		return []byte{}, err
	}
	var electStates mc.ElectOnlineStatus
	err = json.Unmarshal(data, &electStates)
	if err != nil {
		log.ERROR(Module, "ElectStates Unmarshal失败 err", err)
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
