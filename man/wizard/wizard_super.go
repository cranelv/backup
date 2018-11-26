// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package wizard

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/params/manparams"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/matrix/go-matrix/core/types"

	"github.com/matrix/go-matrix/core/state"

	"github.com/matrix/go-matrix/common"

	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/log"
)

func MakeWizard(network string) *wizard {
	return &wizard{
		network: network,
		conf: config{
			Servers: make(map[string][]byte),
		},
		services: make(map[string][]string),
		in:       bufio.NewReader(os.Stdin),
	}
}

// makeGenesis creates a new genesis struct based on some user input.
func (w *wizard) MakeSuperGenesis(bc *core.BlockChain, db mandb.Database, num uint64) {
	// Construct a default genesis block
	var header, curheader *types.Header
	if num > 1 {
		header = bc.GetBlockByNumber(num - 1).Header()
		curheader = bc.GetBlockByNumber(num).Header()
	} else if num == 0 {
		header = bc.Genesis().Header()
		curheader = header
	} else if num == 1 {
		header = bc.Genesis().Header()
		curheader = bc.GetBlockByNumber(num).Header()
	}

	genesis := &core.Genesis{
		ParentHash:        header.Hash(),
		Leader:            curheader.Leader,
		Elect:             curheader.Elect,
		NetTopology:       curheader.NetTopology,
		Mixhash:           header.MixDigest,
		Coinbase:          manparams.InnerMinerNodes[0].Address,
		Signatures:        make([]common.Signature, 0),
		Timestamp:         uint64(time.Now().Unix()),
		GasLimit:          header.GasLimit,
		Difficulty:        header.Difficulty,
		Alloc:             make(core.GenesisAlloc),
		ExtraData:         make([]byte, 0),
		Version:           string(header.Version),
		VersionSignatures: header.VersionSignatures,
		Nonce:             header.Nonce.Uint64(),
		Number:            num,
		GasUsed:           header.GasUsed,
	}

	// Figure out which consensus engine to choose
	fmt.Println()

	genesis.Alloc[common.BlkRewardAddress] = core.GenesisAccount{Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(200), big.NewInt(0))}
	genesis.Alloc[common.TxGasRewardAddress] = core.GenesisAccount{Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(200), big.NewInt(0))}
	genesis.Alloc[common.HexToAddress("0x8000000000000000000000000000000000000002")] = core.GenesisAccount{Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(200), big.NewInt(0))}

	statedb, err := state.New(header.Root, state.NewDatabase(db))
	if err != nil {
		log.Error("state new is ", "err", err)
		return
	}
	for addr, account := range genesis.Alloc {
		statedb.AddBalance(common.MainAccount, addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}
	genesis.Root = statedb.IntermediateRoot(bc.Config().IsEIP158(new(big.Int).SetUint64(num)))
	// All done, store the genesis and flush to disk
	log.Info("Configured new genesis block")

	w.conf.Genesis = genesis
	//w.conf.flush()
	fmt.Printf("Which file to save the genesis into %s", w.network)
	out, _ := json.MarshalIndent(w.conf.Genesis, "", "  ")
	if err := ioutil.WriteFile(w.network, out, 0644); err != nil {
		log.Error("Failed to save genesis file", "err", err)
		return
	}
	err = json.Unmarshal(out, w.conf.Genesis)
	if err != nil {
		log.Error("Failed to save genesis file", "err", err)
		return
	}
	if err := ioutil.WriteFile(w.network, out, 0644); err != nil {
		log.Error("Failed to save genesis file", "err", err)
		return
	}
	log.Info("Exported existing genesis block")

}
