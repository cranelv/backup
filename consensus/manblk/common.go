package manblk

import (
	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/reelection"
)

type MANBLK interface {
	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(interval *manparams.BCInterval) (*types.Header, error)
	ProcessState(header *types.Header) (*state.StateDB, []*types.Receipt, []types.SelfTransaction, error)
	Finalize(header *types.Header, state *state.StateDB, txs []types.SelfTransaction,
		uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error)
	VerifyBlock(header *types.Header) error
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
}

type MANBLKPlUGS interface {
	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(support BlKSupport, interval *manparams.BCInterval, num uint64) (*types.Header, error)
	ProcessState(support BlKSupport, header *types.Header) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, error)
	Finalize(support BlKSupport, header *types.Header, state *state.StateDB, txs []types.SelfTransaction,
		uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error)
	VerifyBlock(support BlKSupport, header *types.Header) error
}

type TopNodeService interface {
	GetConsensusOnlineResults() []*mc.HD_OnlineConsensusVoteResultMsg
}

type Reelection interface {
	GetTopoChange(hash common.Hash, offline []common.Address, online []common.Address) ([]mc.Alternative, error)
	GetNetTopologyAll(hash common.Hash) (*reelection.ElectReturnInfo, error)
	TransferToNetTopologyAllStu(info *reelection.ElectReturnInfo) *common.NetTopology
	GetElection(state *state.StateDB, hash common.Hash) (*reelection.ElectReturnInfo, error)
	TransferToNetTopologyChgStu(alterInfo []mc.Alternative) *common.NetTopology
	TransferToElectionStu(info *reelection.ElectReturnInfo) []common.Elect
}
type SignHelper interface {
	SignVrf(msg []byte, blkHash common.Hash) ([]byte, []byte, []byte, error)
}
type txPool interface {
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending() (map[common.Address]types.SelfTransactions, error)
}

type Mux interface {
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	EventMux() *event.TypeMux
}
type BlKSupport interface {
	ChainReader
	TopNodeService
	Reelection
	baseinterface.VrfInterface
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

func (bd *ManBlkDeal) RegisterManBLkPlugs(version string, plug MANBLKPlUGS) {
	bd.mapManBlkPlugs[bd.version] = plug
}

func (bd *ManBlkDeal) Prepare(interval *manparams.BCInterval) (*types.Header, error) {

	return bd.mapManBlkPlugs[bd.version].Prepare(bd.support, interval, bd.num)
}

func (bd *ManBlkDeal) ProcessState(header *types.Header) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, error) {

	return bd.mapManBlkPlugs[bd.version].ProcessState(bd.support, header)
}

func (bd *ManBlkDeal) Finalize(header *types.Header, state *state.StateDB, txs []types.SelfTransaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	return &types.Block{}, nil
}

func (bd *ManBlkDeal) VerifyBlock(header *types.Header) error {
	return nil
}
