package manblk

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params"
)

// ChainReader defines a small collection of methods needed to access the local
// blockchain during header and/or uncle verification.
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
}

type StateReader interface {
	GetCurrentHash() common.Hash
	GetGraphByHash(hash common.Hash) (*mc.TopologyGraph, *mc.ElectGraph, error)
	GetBroadcastAccount(blockHash common.Hash) (common.Address, error)
	GetVersionSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBlockSuperAccounts(blockHash common.Hash) ([]common.Address, error)
	GetBroadcastInterval(blockHash common.Hash) (*mc.BCIntervalInfo, error)
	GetAuthAccount(addr common.Address, hash common.Hash) (common.Address, error)
}

type MANBLK interface {
	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(chain ChainReader, header *types.Header) error
	ProcessState(chain ChainReader, header *types.Header) error
	Finalize(chain ChainReader, header *types.Header, state *state.StateDB, txs []types.SelfTransaction,
		uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error)
	VerifyBlock(reader StateReader, header *types.Header) error
}

var (
	ModuleManBlk   = "区块生成验证"
	mapManBlkPlugs = make(map[string]MANBLK)
)

type BlKSupport interface {
	ChainReader
	consensus.PoW
	consensus.DPOSEngine
}

type ManBlkDeal struct {
	num     uint64
	version string
	support BlKSupport
}

func New(version string, support BlKSupport, num uint64) (MANBLK, error) {
	obj := new(ManBlkDeal)
	obj.version = version
	obj.support = support
	return obj, nil
}

func RegisterManBLkPlugs(version string, plug MANBLK) {
	mapManBlkPlugs[version] = plug
}

func (bd *ManBlkDeal) Prepare(chain ChainReader, header *types.Header) error {
	map[]
	return nil
}

func (bd *ManBlkDeal) ProcessState(chain ChainReader, header *types.Header) error {
	return nil
}

func (bd *ManBlkDeal) Finalize(chain ChainReader, header *types.Header, state *state.StateDB, txs []types.SelfTransaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	return &types.Block{}, nil
}
func (bd *ManBlkDeal) VerifyBlock(reader StateReader, header *types.Header) error {
	return nil
}
