// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package matrixstate

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/pkg/errors"
)

var (
	ErrStateDBNil   = errors.New("state db is nil")
	ErrValueNotFind = errors.New("value not find is state db")
)

type stateReader interface {
	GetHashByNumber(number uint64) common.Hash
	GetHeaderByHash(hash common.Hash) *types.Header
	State() (*state.StateDB, error)
	StateAt(root common.Hash) (*state.StateDB, error)
}

type StateDB interface {
	GetMatrixData(hash common.Hash) (val []byte)
	SetMatrixData(hash common.Hash, val []byte)
}

type PreStateReadFn func(key string) (interface{}, error)
type ProduceMatrixStateDataFn func(block *types.Block, readFn PreStateReadFn) (interface{}, error)
