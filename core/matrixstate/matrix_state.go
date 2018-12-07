package matrixstate

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/pkg/errors"
	"sync"
)

type keyInfo struct {
	keyHash      common.Hash
	dataProducer ProduceMatrixStateDataFn
}

func genKeyMap() (keyMap map[string]*keyInfo) {
	keyMap = make(map[string]*keyInfo)
	for key, hash := range km.keys {
		keyMap[key] = &keyInfo{hash, nil}
	}
	return
}

type MatrixState struct {
	mu     sync.RWMutex
	keyMap map[string]*keyInfo
}

func NewMatrixState() *MatrixState {
	return &MatrixState{
		keyMap: genKeyMap(),
	}
}

func (ms *MatrixState) RegisterProducer(key string, producer ProduceMatrixStateDataFn) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	info, err := ms.findKeyInfo(key)
	if err != nil {
		return err
	}
	info.dataProducer = producer
	return nil
}

func (ms *MatrixState) ProcessMatrixState(block *types.Block, state *state.StateDB) error {
	if block == nil || state == nil {
		return errors.New("param is nil")
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	readFn := func(key string) ([]byte, error) {
		info, err := ms.findKeyInfo(key)
		if err != nil {
			return nil, err
		}
		return state.GetMatrixData(info.keyHash), nil
	}

	dataMap := make(map[common.Hash][]byte)
	for key, info := range ms.keyMap {
		if info == nil || info.dataProducer == nil {
			continue
		}

		data, err := info.dataProducer(block, readFn)
		if err != nil {
			return errors.Errorf("key(%s) produce matrix state data err(%v)", key, err)
		}
		if data != nil {
			dataMap[info.keyHash] = data
		}
	}

	for keyHash, data := range dataMap {
		state.SetMatrixData(keyHash, data)
	}

	return nil
}

func (ms *MatrixState) GetMatrixStateData(key string, state *state.StateDB) ([]byte, error) {
	if state == nil {
		return nil, errors.New("state is nil")
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()
	info, err := ms.findKeyInfo(key)
	if err != nil {
		return nil, err
	}
	return state.GetMatrixData(info.keyHash), nil
}

func (ms *MatrixState) findKeyInfo(key string) (*keyInfo, error) {
	info, OK := ms.keyMap[key]
	if !OK {
		return nil, errors.Errorf("key(%s) is illegal", key)
	}
	if info == nil {
		return nil, errors.Errorf("CRITICAL the info of key(%s) is nil in map", key)
	}
	return info, nil
}
