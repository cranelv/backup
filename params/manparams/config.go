package manparams

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/pkg/errors"
)

type stateReader interface {
	State() (*state.StateDB, error)
	StateAt(root common.Hash) (*state.StateDB, error)
	GetStateByHash(hash common.Hash) (*state.StateDB, error)
	GetMatrixStateData(key string, state *state.StateDB) (interface{}, error)
}

type matrixConfig struct {
	stReader stateReader
}

var mtxCfg = newMatrixConfig()

func newMatrixConfig() *matrixConfig {
	return &matrixConfig{
		stReader: nil,
	}
}

func SetStateReader(stReader stateReader) {
	mtxCfg.stReader = stReader
}

func (mcfg *matrixConfig) getStateData(key string) (interface{}, error) {
	state, err := mcfg.stReader.State()
	if err != nil {
		return nil, errors.Errorf("获取state失败(%v)", err)
	}
	return mcfg.stReader.GetMatrixStateData(key, state)
}

func (mcfg *matrixConfig) getStateDataByHash(key string) (interface{}, error) {
	state, err := mcfg.stReader.State()
	if err != nil {
		return nil, errors.Errorf("获取state失败(%v)", err)
	}
	return mcfg.stReader.GetMatrixStateData(key, state)
}
