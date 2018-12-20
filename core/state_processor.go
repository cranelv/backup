// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package core

import (
	"errors"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	_"github.com/matrix/go-matrix/consensus/misc"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
	"runtime"
	"sync"
	"encoding/json"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Matrix rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDBManage, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error) {
	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		header   = block.Header()
		allLogs  []*types.Log
		gp       = new(GasPool).AddGas(block.GasLimit())
	)
	// Mutate the the block and state according to any hard-fork specs
	//if p.config.DAOForkSupport && p.config.DAOForkBlock != nil && p.config.DAOForkBlock.Cmp(block.Number()) == 0 {
	//	misc.ApplyDAOHardFork(statedb)
	//}
	// Iterate over and process the individual transactions
	statedb.UpdateTxForBtree(uint32(block.Time().Uint64()))
	statedb.UpdateTxForBtreeBytime(uint32(block.Time().Uint64()))
	stxs := make([]types.SelfTransaction, 0)
	var txcount int
	txs := block.Transactions()
	var waitG = &sync.WaitGroup{}
	maxProcs := runtime.NumCPU() //获取cpu个数
	if maxProcs >= 2 {
		runtime.GOMAXPROCS(maxProcs - 1) //限制同时运行的goroutines数量
	}
	normalTxindex := 0
	for _, tx := range txs {
		if tx.GetMatrixType() == common.ExtraUnGasTxType {
			tmpstxs := make([]types.SelfTransaction, 0)
			tmpstxs = append(tmpstxs, tx)
			tmpstxs = append(tmpstxs, stxs...)
			stxs = tmpstxs
			normalTxindex++
			continue
		}
		sig := types.NewEIP155Signer(tx.ChainId())
		waitG.Add(1)
		ttx := tx
		go types.Sender_self(sig, ttx, waitG)
	}
	waitG.Wait()
	for i, tx := range txs[normalTxindex:] {
		if tx.GetMatrixType() == common.ExtraUnGasTxType {
			tmpstxs := make([]types.SelfTransaction, 0)
			tmpstxs = append(tmpstxs, tx)
			tmpstxs = append(tmpstxs, stxs...)
			stxs = tmpstxs
			continue
		}

		statedb.Prepare(tx.Hash(), block.Hash(), i)
		receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			return nil, nil, 0, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
		txcount = i
	}
	for _, tx := range stxs {
		statedb.Prepare(tx.Hash(), block.Hash(), txcount+1)
		receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			return nil, nil, 0, err
		}
		tmpr := make(types.Receipts, 0)
		tmpr = append(tmpr, receipt)
		tmpr = append(tmpr, receipts...)
		receipts = tmpr
		tmpl := make([]*types.Log, 0)
		tmpl = append(tmpl, receipt.Logs...)
		tmpl = append(tmpl, allLogs...)
		allLogs = tmpl
	}

	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Transactions(), block.Uncles(), receipts)

	return receipts, allLogs, *usedGas, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc *BlockChain, author *common.Address, gp *GasPool, statedb *state.StateDBManage, header *types.Header, tx types.SelfTransaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	// Create a new context to be used in the EVM environment
	from, err := tx.GetTxFrom()
	if err != nil {
		from, err = types.Sender(types.NewEIP155Signer(config.ChainId), tx)
	}
	context := NewEVMContext(from, tx.GasPrice(), header, bc, author)

	vmenv := vm.NewEVM(context, statedb, config, cfg,tx.GetTxCurrency())
	// Apply the transaction to the current state (included in the env)
	var gas uint64
	var failed bool
	if tx.TxType() == types.BroadCastTxIndex {
		if extx := tx.GetMatrix_EX(); (extx != nil) && len(extx) > 0 && extx[0].TxType == 1 {
			gas = uint64(0)
			failed = true
		}
	} else {
		_, gas, failed, err = ApplyMessage(vmenv, tx, gp)
		if err != nil {
			return nil, 0, err
		}
	}
	//如果是委托gas并且是按时间委托
	if tx.GetIsEntrustGas() && tx.GetIsEntrustByTime() {
		//from = base58.Base58DecodeToAddress("MAN.3oW6eUV7MmQcHiD4WGQcRnsN8ho1aFTWPaYADwnqu2wW3WcJzbEfZNw2") //******测试用，要删除
		if !statedb.GetIsEntrustByTime(tx.GetTxCurrency(),from, header.Time.Uint64()) {
			log.Error("按时间委托gas的交易失效")
			return nil, 0, errors.New("entrustTx is invalid")
		}
	}
	// Update the state with pending changes
	var root []byte
	if config.IsByzantium(header.Number) {
		statedb.Finalise(tx.GetTxCurrency(),true)
	} else {
		root,_ = json.Marshal(statedb.IntermediateRoot(config.IsEIP158(header.Number)))
	}
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if tx.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.GetTxCurrency(),tx.From(),tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, gas, err
}
