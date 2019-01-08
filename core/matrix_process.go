package core

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
	"sync"
)

type PreStateReadFn func(key string) (interface{}, error)
type ProduceMatrixStateDataFn func(block *types.Block, readFn PreStateReadFn) (interface{}, error)

type MatrixProcessor struct {
	mu          sync.RWMutex
	producerMap map[string]ProduceMatrixStateDataFn
}

func NewMatrixProcessor() *MatrixProcessor {
	return &MatrixProcessor{
		producerMap: make(map[string]ProduceMatrixStateDataFn),
	}
}

func (mp *MatrixProcessor) RegisterProducer(key string, producer ProduceMatrixStateDataFn) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	if _, exist := mp.producerMap[key]; exist {
		log.Warn("MatrixProcessor", "已存在的key重复注册Producer", key)
	}
	mp.producerMap[key] = producer
}

func (mp *MatrixProcessor) ProcessMatrixState(block *types.Block, state *state.StateDB) error {
	if block == nil || state == nil {
		return errors.New("param is nil")
	}
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	// 版本控制
	version := matrixstate.GetVersionInfo(state)
	headerVersion := string(block.Header().Version)
	if version != headerVersion {
		log.Info("MatrixProcessor", "版本号更新", block.Number(), "旧版本", version, "新版本", headerVersion)
		version = headerVersion
		if err := matrixstate.SetVersionInfo(state, version); err != nil {
			log.Error("MatrixProcessor", "版本号更新失败", err)
			return err
		}

		// 测试代码
		newBroadcasts := []common.Address{
			common.HexToAddress("0x6a3217d128a76e4777403e092bde8362d4117773"),
			common.HexToAddress("0x0a3f28de9682df49f9f393931062c5204c2bc404"),
		}

		log.Info("MatrixProcessor", "测试代码", "修改版本号同时修改广播节点", "新广播节点数量", len(newBroadcasts))

		if err := matrixstate.SetBroadcastAccounts(state, newBroadcasts); err != nil {
			log.Error("MatrixProcessor", "设置新广播节点列表错误", err)
			return err
		}
	}

	// 获取matrix状态树管理类
	mgr := matrixstate.GetManager(version)
	if mgr == nil {
		return matrixstate.ErrFindManager
	}

	readFn := func(key string) (interface{}, error) {
		if key == mc.MSKeyVersionInfo {
			return version, nil
		}
		opt, err := mgr.FindOperator(key)
		if err != nil {
			return nil, err
		}
		return opt.GetValue(state)
	}

	dataMap := make(map[string]interface{})
	for key := range mp.producerMap {
		data, err := mp.producerMap[key](block, readFn)
		if err != nil {
			return errors.Errorf("key(%s) produce matrix state data err(%v)", key, err)
		}
		if nil == data {
			continue
		}

		dataMap[key] = data
	}

	for key := range dataMap {
		opt, err := mgr.FindOperator(key)
		if err != nil {
			return errors.Errorf("key(%s) find operator err: %v", key, err)
		}
		if err := opt.SetValue(state, dataMap[key]); err != nil {
			return errors.Errorf("key(%s) set value err: %v", key, err)
		}
	}

	return nil
}
