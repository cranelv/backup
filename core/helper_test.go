// Copyright (c) 2008 The MATRIX Authors 
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
// Copyright 2014 The go-matrix Authors
// This file is part of the go-matrix library.
//
// The go-matrix library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-matrix library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-matrix library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"container/list"

	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/event"
)

// Implement our EthTest Manager
type TestManager struct {
	// stateManager *StateManager
	eventMux *event.TypeMux

	db         mandb.Database
	txPool     *TxPool
	blockChain *BlockChain
	Blocks     []*types.Block
}

func (tm *TestManager) IsListening() bool {
	return false
}

func (tm *TestManager) IsMining() bool {
	return false
}

func (tm *TestManager) PeerCount() int {
	return 0
}

func (tm *TestManager) Peers() *list.List {
	return list.New()
}

func (tm *TestManager) BlockChain() *BlockChain {
	return tm.blockChain
}

func (tm *TestManager) TxPool() *TxPool {
	return tm.txPool
}

// func (tm *TestManager) StateManager() *StateManager {
// 	return tm.stateManager
// }

func (tm *TestManager) EventMux() *event.TypeMux {
	return tm.eventMux
}

// func (tm *TestManager) KeyManager() *crypto.KeyManager {
// 	return nil
// }

func (tm *TestManager) Db() mandb.Database {
	return tm.db
}

func NewTestManager() *TestManager {
	testManager := &TestManager{}
	testManager.eventMux = new(event.TypeMux)
	testManager.db = mandb.NewMemDatabase()
	// testManager.txPool = NewTxPool(testManager)
	// testManager.blockChain = NewBlockChain(testManager)
	// testManager.stateManager = NewStateManager(testManager)
	return testManager
}
