package manblk

import (
	"errors"

	"github.com/matrix/go-matrix/accounts/signhelper"
	"github.com/matrix/go-matrix/reelection"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/params/manparams"
)

type MANBLK interface {
	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(types string, version string, num uint64, interval *manparams.BCInterval, args ...interface{}) (*types.Header, interface{}, error)
	ProcessState(types string, version string, header *types.Header, args ...interface{}) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, interface{}, error)
	Finalize(types string, version string, header *types.Header, state *state.StateDB, txs []types.SelfTransaction, uncles []*types.Header, receipts []*types.Receipt, args ...interface{}) (*types.Block, interface{}, error)
	VerifyHeader(types string, version string, header *types.Header, args ...interface{}) (interface{}, error)
	VerifyTxsAndState(types string, version string, header *types.Header, Txs types.SelfTransactions, args ...interface{}) (interface{}, error)
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
	Prepare(support BlKSupport, interval *manparams.BCInterval, num uint64, args ...interface{}) (*types.Header, interface{}, error)
	ProcessState(support BlKSupport, header *types.Header, args ...interface{}) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, interface{}, error)
	Finalize(support BlKSupport, header *types.Header, state *state.StateDB, txs []types.SelfTransaction, uncles []*types.Header, receipts []*types.Receipt, args interface{}) (*types.Block, interface{}, error)
	VerifyHeader(support BlKSupport, header *types.Header, args ...interface{}) (interface{}, error)
	VerifyTxsAndState(support BlKSupport, header *types.Header, Txs types.SelfTransactions, args ...interface{}) (interface{}, error)
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
	BlockChain() *core.BlockChain
	TxPool() *core.TxPoolManager
	SignHelper() *signhelper.SignHelper
	EventMux() *event.TypeMux
	ReElection() *reelection.ReElection
}
type VrfMsg struct {
	VrfValue []byte
	VrfProof []byte
	Hash     common.Hash
}

var (
	ModuleManBlk = "区块生成验证"
	CommonBlk    = "common"
	BroadcastBlk = "broadcast"

	AVERSION = "1.0.0-stable"
)

type ManBlkDeal struct {
	support        BlKSupport
	mapManBlkPlugs map[string]MANBLKPlUGS
}

func New(support BlKSupport) (*ManBlkDeal, error) {
	obj := new(ManBlkDeal)
	obj.support = support

	obj.mapManBlkPlugs = make(map[string]MANBLKPlUGS)
	return obj, nil
}

func (bd *ManBlkDeal) RegisterManBLkPlugs(types string, version string, plug MANBLKPlUGS) {
	bd.mapManBlkPlugs[types+version] = plug
}

func (bd *ManBlkDeal) Prepare(types string, version string, num uint64, interval *manparams.BCInterval, args ...interface{}) (*types.Header, interface{}, error) {
	plug, ok := bd.mapManBlkPlugs[types+version]
	if !ok {
		log.ERROR(ModuleManBlk, "获取插件失败", "")
		return nil, nil, errors.New("获取插件失败")
	}
	return plug.Prepare(bd.support, interval, num, args)
}

func (bd *ManBlkDeal) ProcessState(types string, version string, header *types.Header, args ...interface{}) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, interface{}, error) {
	plug, ok := bd.mapManBlkPlugs[types+version]
	if !ok {
		log.ERROR(ModuleManBlk, "获取插件失败", "")
		return nil, nil, nil, nil, nil, nil, errors.New("获取插件失败")
	}
	return plug.ProcessState(bd.support, header, args)
}

func (bd *ManBlkDeal) Finalize(types string, version string, header *types.Header, state *state.StateDB, txs []types.SelfTransaction, uncles []*types.Header, receipts []*types.Receipt, args ...interface{}) (*types.Block, interface{}, error) {
	plug, ok := bd.mapManBlkPlugs[types+version]
	if !ok {
		log.ERROR(ModuleManBlk, "获取插件失败", "")
		return nil, nil, errors.New("获取插件失败")
	}
	return plug.Finalize(bd.support, header, state, txs, uncles, receipts, args)
}

func (bd *ManBlkDeal) VerifyHeader(types string, version string, header *types.Header, args ...interface{}) (interface{}, error) {
	plug, ok := bd.mapManBlkPlugs[types+version]
	if !ok {
		log.ERROR(ModuleManBlk, "获取插件失败", "")
		return nil, errors.New("获取插件失败")
	}
	return plug.VerifyHeader(bd.support, header)
}

func (bd *ManBlkDeal) VerifyTxsAndState(types string, version string, header *types.Header, Txs types.SelfTransactions, args ...interface{}) (interface{}, error) {
	plug, ok := bd.mapManBlkPlugs[types+version]
	if !ok {
		log.ERROR(ModuleManBlk, "获取插件失败", "")
		return nil, errors.New("获取插件失败")
	}
	return plug.VerifyTxsAndState(bd.support, header, Txs, args)
}
