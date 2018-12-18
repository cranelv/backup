package txsreward

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/core/matrixstate"

	"bou.ke/monkey"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/reward/util"
	. "github.com/smartystreets/goconvey/convey"
)

type Chain struct {
}

func (chain *Chain) Config() *params.ChainConfig {
	return nil
}

// CurrentHeader retrieves the current header from the local chain.
func (chain *Chain) CurrentHeader() *types.Header {

	return nil
}

// GetHeader retrieves a block header from the database by hash and number.
func (chain *Chain) GetHeader(hash common.Hash, number uint64) *types.Header {
	return nil
}

// GetHeaderByHash retrieves a block header from the database by its hash.
func (chain *Chain) GetHeaderByHash(hash common.Hash) *types.Header {
	return nil
}

func (chain *Chain) GetBlockByNumber(number uint64) *types.Block {
	return nil
}

// GetBlock retrieves a block sfrom the database by hash and number.
func (chain *Chain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return nil
}
func (chain *Chain) StateAt(root common.Hash) (*state.StateDB, error) {

	return nil, nil
}
func (chain *Chain) State() (*state.StateDB, error) {

	return nil, nil
}
func (chain *Chain) NewTopologyGraph(header *types.Header) (*mc.TopologyGraph, error) {
	return nil, nil
}

func (chain *Chain) GetHeaderByNumber(number uint64) *types.Header {
	header := &types.Header{
		Coinbase: common.Address{},
	}
	//txs := make([]types.SelfTransaction, 0)

	return header
}

type State struct {
	balance int64
}

func (st *State) GetBalance(addr common.Address) common.BalanceType {
	return []common.BalanceSlice{{common.MainAccount, big.NewInt(st.balance)}}
}

func (st *State) CreateAccount(common.Address) {

}

func (st *State) SubBalance(uint32, common.Address, *big.Int) {}
func (st *State) AddBalance(uint32, common.Address, *big.Int) {}

func (st *State) GetNonce(common.Address) uint64  { return 0 }
func (st *State) SetNonce(common.Address, uint64) {}

func (st *State) GetCodeHash(common.Address) common.Hash { return common.Hash{} }
func (st *State) GetCode(common.Address) []byte          { return nil }
func (st *State) SetCode(common.Address, []byte)         {}
func (st *State) GetCodeSize(common.Address) int         { return 0 }

func (st *State) AddRefund(uint64)  {}
func (st *State) GetRefund() uint64 { return 0 }

func (st *State) GetState(common.Address, common.Hash) common.Hash  { return common.Hash{} }
func (st *State) SetState(common.Address, common.Hash, common.Hash) {}

func (st *State) Suicide(common.Address) bool     { return true }
func (st *State) HasSuicided(common.Address) bool { return true }

// Exist reports whether the given account exists in state.
// Notably this should also return true for suicided accounts.
func (st *State) Exist(common.Address) bool { return true }

// Empty returns whether the given account is empty. Empty
// is defined according to EIP161 (balance = nonce = code = 0).
func (st *State) Empty(common.Address) bool { return true }

func (st *State) RevertToSnapshot(int) {}
func (st *State) Snapshot() int        { return 0 }

func (st *State) AddLog(*types.Log)               {}
func (st *State) AddPreimage(common.Hash, []byte) {}

func (st *State) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) {}

func (st *State) GetMatrixData(hash common.Hash) (val []byte) {

	return nil
}
func (st *State) SetMatrixData(hash common.Hash, val []byte) {
	return
}
func TestNew1(t *testing.T) {
	Convey("计算交易费", t, func() {

		log.InitLog(3)
		eth := fakeEthNew(0)

		monkey.Patch(matrixstate.GetDataByState, func(key string, state matrixstate.StateDB) (interface{}, error) {
			fmt.Println("use monkey  ca.GetOnlineTime")
			cfg := mc.RewardRateCfg{MinerOutRate: 4000, ElectedMinerRate: 5000, FoundationMinerRate: 1000, LeaderRate: 4000, ElectedValidatorsRate: 5000, FoundationValidatorRate: 1000, OriginElectOfflineRate: 5000, BackupRewardRate: 5000}
			return &mc.TxsRewardCfgStruct{MinersRate: 5000, ValidatorsRate: 5000, RewardRate: cfg}, nil
		})
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		reward := New(Chain{}, &State{0})
		reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), 3)
	})
}

//func TestNew2(t *testing.T) {
//	Convey("计算交易费", t, func() {
//
//		log.InitLog(3)
//		eth := fakeEthNew(0)
//		header := eth.BlockChain().CurrentHeader()
//		newheader := types.CopyHeader(header)
//		newheader.Number = big.NewInt(1)
//		newheader.NetTopology.Type = common.NetTopoTypeAll
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
//		reward := New(eth.blockchain)
//		reward.CalcNodesRewards(big.NewInt(0), common.HexToAddress(testAddress), 1)
//	})
//}
//func TestNew3(t *testing.T) {
//	Convey("计算交易费", t, func() {
//
//		log.InitLog(3)
//		eth := fakeEthNew(0)
//		header := eth.BlockChain().CurrentHeader()
//		newheader := types.CopyHeader(header)
//		newheader.Number = big.NewInt(1)
//		newheader.NetTopology.Type = common.NetTopoTypeAll
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
//		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
//		reward := New(eth.blockchain)
//		reward.CalcNodesRewards(big.NewInt(-1), common.HexToAddress(testAddress), 1)
//	})
//}
