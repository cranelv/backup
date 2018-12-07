// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package matrixstate

import (
	"encoding/json"
	"github.com/hashicorp/golang-lru"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

const (
	topologyCacheLimit = 512
	electCacheLimit    = 512
)

type GraphStore struct {
	stateKey      string
	reader        stateReader
	topologyCache *lru.Cache
	electCache    *lru.Cache
}

func NewGraphStore(reader stateReader) *GraphStore {
	topologyCache, _ := lru.New(topologyCacheLimit)
	electCache, _ := lru.New(electCacheLimit)

	return &GraphStore{
		reader:        reader,
		topologyCache: topologyCache,
		electCache:    electCache,
	}
}

func (gs GraphStore) ProduceTopologyStateData(block *types.Block, readFn PreStateReadFn) ([]byte, error) {
	header := block.Header()
	number := header.Number.Uint64()

	preData, err := readFn(MSPTopologyGraph)
	if err != nil {
		return nil, errors.Errorf("read pre data err(%v)", err)
	}
	preGraph := new(mc.TopologyGraph)
	if err := json.Unmarshal(preData, &preGraph); err != nil {
		return nil, errors.Errorf("Invalid preGraph(number = %d) json data: %v", number-1, err)
	}
	newGraph, err := preGraph.Transfer2NextGraph(number, &header.NetTopology)
	if err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(newGraph)
	if err != nil {
		return nil, errors.Errorf("Failed to encode topology graph: %v", err)
	}
	return bytes, nil
}

func (gs GraphStore) GetHashByNumber(number uint64) common.Hash {
	return gs.reader.GetHashByNumber(number)
}

func (gs GraphStore) GetTopologyGraphByHash(blockHash common.Hash) (*mc.TopologyGraph, error) {
	if graph, ok := gs.topologyCache.Get(blockHash); ok {
		return graph.(*mc.TopologyGraph), nil
	}

	state, err := gs.getStateByBlock(blockHash)
	if err != nil {
		return nil, err
	}

	graph, err := gs.GetTopologyGraphByState(state)
	if err != nil {
		return nil, err
	}
	gs.topologyCache.Add(blockHash, graph)
	return graph, nil
}

func (gs GraphStore) GetTopologyGraphByState(state *state.StateDB) (*mc.TopologyGraph, error) {
	data, err := gs.reader.GetMatrixStateData(MSPTopologyGraph, state)
	if err != nil {
		return nil, errors.Errorf("get topology state data err(%s)", err)
	}

	graph := new(mc.TopologyGraph)
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, errors.Errorf("Invalid topology graph json data: %v", err)
	}

	if graph == nil {
		return nil, errors.New("topology graph is nil")
	}

	return graph, nil
}

func (gs GraphStore) GetElectGraphByHash(blockHash common.Hash) (*mc.ElectGraph, error) {
	if elect, ok := gs.electCache.Get(blockHash); ok {
		return elect.(*mc.ElectGraph), nil
	}

	state, err := gs.getStateByBlock(blockHash)
	if err != nil {
		return nil, err
	}

	elect, err := gs.GetElectGraphByState(state)
	if err != nil {
		return nil, err
	}

	gs.electCache.Add(blockHash, elect)
	return elect, nil
}

func (gs GraphStore) GetElectGraphByState(state *state.StateDB) (*mc.ElectGraph, error) {
	data, err := gs.reader.GetMatrixStateData(MSPElectGraph, state)
	if err != nil {
		return nil, errors.Errorf("get elect state data err(%s)", err)
	}

	elect := new(mc.ElectGraph)
	if err := json.Unmarshal(data, &elect); err != nil {
		return nil, errors.Errorf("Invalid elect graph json data: %v", err)
	}

	if elect == nil {
		return nil, errors.New("elect graph is nil")
	}

	return elect, nil
}

func (gs *GraphStore) GetOriginalElectByHash(blockHash common.Hash) ([]common.Elect, error) {
	elect, err := gs.GetElectGraphByHash(blockHash)
	if err != nil {
		return nil, err
	}

	if elect == nil {
		return nil, errors.New("elect data is illegal")
	}

	return elect.TransferElect2CommonElect(), nil
}

func (gs *GraphStore) GetNextElectByHash(blockHash common.Hash) ([]common.Elect, error) {
	elect, err := gs.GetElectGraphByHash(blockHash)
	if err != nil {
		return nil, err
	}

	if elect == nil {
		return nil, errors.New("elect data is illegal")
	}

	return elect.TransferNextElect2CommonElect(), nil
}

func (gs *GraphStore) NewTopologyGraph(header *types.Header) (*mc.TopologyGraph, error) {
	return nil, nil
}

func (gs *GraphStore) getStateByBlock(blockHash common.Hash) (*state.StateDB, error) {
	header := gs.reader.GetHeaderByHash(blockHash)
	if header == nil {
		return nil, errors.Errorf("can't find header by hash(%s)", blockHash.Hex())
	}
	state, err := gs.reader.StateAt(header.Root)
	if err != nil {
		return nil, errors.Errorf("can't find state by root(%s): %v", header.Root.TerminalString(), err)
	}
	if state == nil {
		return nil, errors.Errorf("state of root(%s) is nil", header.Root.TerminalString())
	}
	return state, nil
}
