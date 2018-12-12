package reelection

import (
	"errors"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
)

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

	bciData, err := readFn(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Error(Module, "ProducePreAllTopData read broadcast interval err", err)
		return nil, err
	}
	bcInterval, err := manparams.NewBCIntervalWithInterval(bciData)
	if err != nil {
		log.Error(Module, "ProducePreAllTopData create broadcast interval err", err)
	}
	if bcInterval.IsReElectionNumber(block.NumberU64() + 1) {
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

	bciData, err := readFn(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Error(Module, "ProducePreAllTopData read broadcast interval err", err)
		return nil, err
	}
	bcInterval, err := manparams.NewBCIntervalWithInterval(bciData)
	if err != nil {
		log.Error(Module, "ProducePreAllTopData create broadcast interval err", err)
	}

	if bcInterval.IsReElectionNumber(height + 1) {
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
	bciData, err := readFn(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Error(Module, "ProducePreAllTopData read broadcast interval err", err)
		return nil, err
	}
	bcInterval, err := manparams.NewBCIntervalWithInterval(bciData)
	if err != nil {
		log.Error(Module, "ProducePreAllTopData create broadcast interval err", err)
	}
	height := block.Header().Number.Uint64()
	if height == 1 {
		firstData := &mc.PreBroadStateRoot{
			LastStateRoot:       common.Hash{},
			BeforeLastStateRoot: common.Hash{},
		}
		return firstData, nil
	}

	if bcInterval.IsBroadcastNumber(height-1) == false {
		return nil, nil
	}
	data, err := readFn(mc.MSKeyPreBroadcastRoot)
	if err != nil {
		log.ERROR(Module, "readFn 失败 key", mc.MSKeyPreBroadcastRoot, "err", err)
		return nil, err
	}
	preBroadcast, OK := data.(*mc.PreBroadStateRoot)
	if OK == false || preBroadcast == nil {
		log.ERROR(Module, "PreBroadStateRoot 非法", "反射失败")
		return nil, err
	}
	header := self.bc.GetHeaderByHash(block.ParentHash())
	if header == nil {
		log.ERROR(Module, "根据hash算区块头失败 高度", block.Number().Uint64())
		return nil, errors.New("header is nil")
	}

	preBroadcast.BeforeLastStateRoot = preBroadcast.LastStateRoot
	preBroadcast.LastStateRoot = header.Root
	return preBroadcast, nil

}
func (self *ReElection) ProduceMinHashData(block *types.Block, readFn matrixstate.PreStateReadFn) (interface{}, error) {
	if err := CheckBlock(block); err != nil {
		log.ERROR(Module, "ProducePreBroadcastStateData CheckBlock err ", err)
		return []byte{}, err
	}
	bciData, err := readFn(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Error(Module, "ProducePreAllTopData read broadcast interval err", err)
		return nil, err
	}
	bcInterval, err := manparams.NewBCIntervalWithInterval(bciData)
	if err != nil {
		log.Error(Module, "ProducePreAllTopData create broadcast interval err", err)
	}
	height := block.Number().Uint64()
	if bcInterval.IsBroadcastNumber(height - 1) {
		return mc.MinHashStruct{MinHash: block.ParentHash()}, nil
	}
	data, err := readFn(mc.MSKeyMinHash)
	if err != nil {
		log.ERROR(Module, "readFn 失败 key", mc.MSKeyMinHash, "err", err)
		return nil, err
	}
	preMinHash, OK := data.(*mc.MinHashStruct)
	if OK == false || preMinHash == nil {
		log.ERROR(Module, "PreBroadStateRoot 非法", "反射失败")
		return nil, err
	}
	header := self.bc.GetHeaderByHash(block.ParentHash())
	if header == nil {
		log.ERROR(Module, "根据hash算区块头失败 高度", block.Number().Uint64())
		return nil, errors.New("header is nil")
	}

	nowHash := header.Hash().Big()
	if nowHash.Cmp(preMinHash.MinHash.Big()) < 0 {
		preMinHash.MinHash = header.Hash()
	}
	return preMinHash, nil
}

func (self *ReElection) ProducePreAllTopData(block *types.Block, readFn matrixstate.PreStateReadFn) (interface{}, error) {
	if err := CheckBlock(block); err != nil {
		log.ERROR(Module, "ProducePreAllTopData CheckBlock err ", err)
		return []byte{}, err
	}
	bciData, err := readFn(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Error(Module, "ProducePreAllTopData read broadcast interval err", err)
		return nil, err
	}
	bcInterval, err := manparams.NewBCIntervalWithInterval(bciData)
	if err != nil {
		log.Error(Module, "ProducePreAllTopData create broadcast interval err", err)
	}
	height := block.Header().Number.Uint64()
	if bcInterval.IsReElectionNumber(height+1) == false {
		return nil, nil
	}
	data, err := readFn(mc.MSKeyPerAllTop)
	if err != nil {
		log.ERROR(Module, "ProducePreAllTopData readFn 失败 key", mc.MSKeyPreBroadcastRoot, "err", err)
		return nil, err
	}
	preAllTop, OK := data.(*mc.PreAllTopStruct)
	if OK == false || preAllTop == nil {
		log.ERROR(Module, "ProducePreAllTopData 非法", "反射失败")
		return nil, err
	}
	header := self.bc.GetHeaderByHash(block.ParentHash())
	if header == nil {
		log.ERROR(Module, "根据hash算区块头失败 高度", block.Number().Uint64())
		return nil, errors.New("header is nil")
	}

	preAllTop.PreAllTopRoot = header.Root
	return preAllTop, nil
}

func (self *ReElection) ProducePreMinerData(block *types.Block, readFn matrixstate.PreStateReadFn) (interface{}, error) {
	if err := CheckBlock(block); err != nil {
		log.ERROR(Module, "ProducePreMinerData CheckBlock err ", err)
		return nil, err
	}

	bciData, err := readFn(mc.MSKeyBroadcastInterval)
	if err != nil {
		log.Error(Module, "ProducePreMinerData read broadcast interval err", err)
		return nil, err
	}
	bcInterval, err := manparams.NewBCIntervalWithInterval(bciData)
	if err != nil {
		log.Error(Module, "ProducePreMinerData create broadcast interval err", err)
	}

	height := block.Header().Number.Uint64()
	if bcInterval.IsBroadcastNumber(height) {
		return nil, nil
	}
	data, err := readFn(mc.MSKeyPreMiner)
	if err != nil {
		log.ERROR(Module, "readFn 失败 key", mc.MSKeyPreMiner, "err", err)
		return nil, err
	}
	preMiner, OK := data.(*mc.PreMinerStruct)
	if OK == false || preMiner == nil {
		log.ERROR(Module, "PreBroadStateRoot 非法", "反射失败")
		return nil, err
	}
	header := self.bc.GetHeaderByHash(block.ParentHash())
	if header == nil {
		log.ERROR(Module, "根据hash算区块头失败 高度", block.Number().Uint64())
		return nil, errors.New("header is nil")
	}

	preMiner.PreMiner = header.Coinbase
	return preMiner, nil
}
