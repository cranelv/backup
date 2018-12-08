package manparams

import (
	"github.com/matrix/go-matrix/common"
)

type stateReader interface {
	GetMatrixStateData(key string) (interface{}, error)
	GetMatrixStateDataByHash(key string, hash common.Hash) (interface{}, error)
	GetMatrixStateDataByNumber(key string, number uint64) (interface{}, error)
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
	return mcfg.stReader.GetMatrixStateData(key)
}

func (mcfg *matrixConfig) getStateDataByHash(key string, hash common.Hash) (interface{}, error) {
	return mcfg.stReader.GetMatrixStateDataByHash(key, hash)
}

func (mcfg *matrixConfig) getStateDataByNumber(key string, number uint64) (interface{}, error) {
	return mcfg.stReader.GetMatrixStateDataByNumber(key, number)
}
