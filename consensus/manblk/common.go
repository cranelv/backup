package manblk

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/params/manparams"
)

type MANBLK interface {
	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(interval *manparams.BCInterval, args ...interface{}) (*types.Header, error)
	ProcessState(header *types.Header, args ...interface{}) (*state.StateDB, []*types.Receipt, []types.SelfTransaction, error)
	Finalize(header *types.Header, state *state.StateDB, txs []types.SelfTransaction, uncles []*types.Header, receipts []*types.Receipt, args ...interface{}) (*types.Block, error)
	VerifyHeader(header *types.Header, args ...interface{}) error
	VerifyTxsAndState(header *types.Header, Txs types.SelfTransactions, args ...interface{}) error
}

type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *params.ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *types.Header

	// GetHeader retrieves a block header from the database by hash and number.
	GetHeader(hash common.Hash, number uint64) *types.Header

	// GetHeaderByNumber retrieves a block header from the database by number.
	GetHeaderByNumber(number uint64) *types.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash common.Hash) *types.Header

	// GetBlock retrieves a block from the database by hash and number.
	GetBlock(hash common.Hash, number uint64) *types.Block

	Genesis() *types.Block

	GetBlockByHash(hash common.Hash) *types.Block

	GetCurrentHash() common.Hash
	GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error)
	GetBroadcastAccount(blockHash common.Hash) (common.Address, error)
	GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBroadcastInterval(blockHash common.Hash) (*mc.BCIntervalInfo, error)
	GetAuthAccount(addr common.Address, hash common.Hash) (common.Address, error)
	GetStateByHash(hash common.Hash) (*state.StateDB, error)

	ProcessUpTime(state *state.StateDB, header *types.Header) (map[common.Address]uint64, error)
	StateAt(root common.Hash) (*state.StateDB, error)
	GetMatrixStateDataByNumber(key string, number uint64) (interface{}, error)
	ProcessMatrixState(block *types.Block, state *state.StateDB) error
	Engine() consensus.Engine
	DPOSEngine() consensus.DPOSEngine
	Processor() core.Processor
	VerifyHeader(header *types.Header) error
}

type MANBLKPlUGS interface {
	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(support BlKSupport, interval *manparams.BCInterval, num uint64, args ...interface{}) (*types.Header, error)
	ProcessState(support BlKSupport, header *types.Header, args ...interface{}) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, error)
	Finalize(support BlKSupport, header *types.Header, state *state.StateDB, txs []types.SelfTransaction,
		uncles []*types.Header, receipts []*types.Receipt, args ...interface{}) (*types.Block, error)
	VerifyHeader(support BlKSupport, header *types.Header, args ...interface{}) error
	VerifyTxsAndState(support BlKSupport, header *types.Header, Txs types.SelfTransactions, args ...interface{}) error
}

type TopNodeService interface {
	GetConsensusOnlineResults() []*mc.HD_OnlineConsensusVoteResultMsg
}

type Reelection interface {
	VerifyNetTopology(header *types.Header, onlineConsensusResults []*mc.HD_OnlineConsensusVoteResultMsg) error
	VerifyElection(header *types.Header, state *state.StateDB) error
	GetNetTopology(num uint64, parentHash common.Hash, bcInterval *manparams.BCInterval) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg)
	GenElection(state *state.StateDB, preBlockHash common.Hash) []common.Elect
	VerifyVrf(header *types.Header) error
}
type SignHelper interface {
	SignVrf(msg []byte, blkHash common.Hash) ([]byte, []byte, []byte, error)
}
type txPool interface {
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[common.Address]types.SelfTransactions, error)
	GetAllSpecialTxs() (reqVal map[common.Address][]types.SelfTransaction)
}

type Mux interface {
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	EventMux() *event.TypeMux
}

type BlKSupport interface {
	ChainReader
	Reelection
	SignHelper
	txPool
	Mux
}
type VrfMsg struct {
	VrfValue []byte
	VrfProof []byte
	Hash     common.Hash
}

var (
	ModuleManBlk = "区块生成验证"
)

type ManBlkDeal struct {
	num            uint64
	version        string
	support        BlKSupport
	mapManBlkPlugs map[string]MANBLKPlUGS
}

func New(version string, support BlKSupport, num uint64) (*ManBlkDeal, error) {
	obj := new(ManBlkDeal)
	obj.version = version
	obj.support = support
	obj.num = num
	obj.mapManBlkPlugs = make(map[string]MANBLKPlUGS)
	return obj, nil
}

func (bd *ManBlkDeal) RegisterManBLkPlugs(types string, version string, plug MANBLKPlUGS) {
	bd.mapManBlkPlugs[types+bd.version] = plug
}

func (bd *ManBlkDeal) Prepare(interval *manparams.BCInterval, args ...interface{}) (*types.Header, error) {

	return bd.mapManBlkPlugs[bd.version].Prepare(bd.support, interval, bd.num, args)
}

func (bd *ManBlkDeal) ProcessState(header *types.Header, args ...interface{}) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, error) {

	return bd.mapManBlkPlugs[bd.version].ProcessState(bd.support, header, args)
}

func (bd *ManBlkDeal) Finalize(header *types.Header, state *state.StateDB, txs []types.SelfTransaction,
	uncles []*types.Header, receipts []*types.Receipt, args ...interface{}) (*types.Block, error) {
	return bd.mapManBlkPlugs[bd.version].Finalize(bd.support, header, state, txs, uncles, receipts, args)
}

func (bd *ManBlkDeal) VerifyHeader(header *types.Header, args ...interface{}) error {
	return bd.mapManBlkPlugs[bd.version].VerifyHeader(bd.support, header)
}

func (bd *ManBlkDeal) VerifyTxsAndState(header *types.Header, Txs types.SelfTransactions, args ...interface{}) error {
	return bd.mapManBlkPlugs[bd.version].VerifyTxsAndState(bd.support, header, Txs, args)
}
