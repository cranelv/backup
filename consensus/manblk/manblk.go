package manblk

import (
	"math/big"
	"time"

	"github.com/matrix/go-matrix/params/manparams"

	"github.com/pkg/errors"

	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/reelection"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/consensus"
	"github.com/matrix/go-matrix/core"
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

	Genesis() *types.Block

	GetBlockByHash(hash common.Hash) *types.Block
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

type MANBLKPlUGS interface {
	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(chain ChainReader, interval *manparams.BCInterval, num uint64) (*types.Header, error)
	ProcessState(chain ChainReader, header *types.Header) error
	Finalize(chain ChainReader, header *types.Header, state *state.StateDB, txs []types.SelfTransaction,
		uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error)
	VerifyBlock(reader StateReader, header *types.Header) error
}
type TopNodeService interface {
	GetConsensusOnlineResults() []*mc.HD_OnlineConsensusVoteResultMsg
}

type Reelection interface {
	GetTopoChange(hash common.Hash, offline []common.Address, online []common.Address) ([]mc.Alternative, error)
	GetNetTopologyAll(hash common.Hash) (*reelection.ElectReturnInfo, error)
	TransferToNetTopologyAllStu(info *reelection.ElectReturnInfo) *common.NetTopology
}

var (
	ModuleManBlk   = "区块生成验证"
	mapManBlkPlugs = make(map[string]MANBLKPlUGS)
)

type BlKSupport interface {
	ChainReader
	consensus.PoW
	consensus.DPOSEngine
	TopNodeService
	Reelection
	VrfInterface
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
	obj.num = num
	return obj, nil
}

func (bd *ManBlkDeal) RegisterManBLkPlugs(version string, plug MANBLKPlUGS) {
	mapManBlkPlugs[bd.version] = plug
}

func (bd *ManBlkDeal) Prepare(chain ChainReader, header *types.Header, interval *manparams.BCInterval) error {

	mapManBlkPlugs[bd.version].Prepare(chain, interval, bd.num)
	return nil
}

func (bd *ManBlkDeal) ProcessState(chain ChainReader, header *types.Header) error {
	mapManBlkPlugs[bd.version].ProcessState(chain, header)
	return nil
}

func (bd *ManBlkDeal) Finalize(chain ChainReader, header *types.Header, state *state.StateDB, txs []types.SelfTransaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	return &types.Block{}, nil
}

func (bd *ManBlkDeal) VerifyBlock(reader StateReader, header *types.Header) error {
	return nil
}

type ManBlkPlug1 struct {
	preBlockHash common.Hash
}

func NewBlkPlug1(preBlockHash common.Hash) (*ManBlkPlug1, error) {
	obj := new(ManBlkPlug1)
	obj.preBlockHash = preBlockHash
	return obj, nil
}
func (bd *ManBlkPlug1) Prepare(chain ChainReader, interval *manparams.BCInterval, num uint64) (*types.Header, error) {
	originHeader := new(types.Header)
	parent, err := bd.setParentHash(chain, originHeader, num)
	if nil != err {
		log.ERROR(ModuleManBlk, "区块生成阶段", "获取父区块失败")
		return nil, err
	}

	bd.setTimeStamp(parent, originHeader, num)
	bd.setLeader(originHeader)
	bd.setNumber(originHeader, num)
	bd.setGasLimit(originHeader, parent)
	bd.setExtra(originHeader)
	bd.setTopology(parent.Hash(), originHeader, interval, num)
	return originHeader, nil
}

func (bd *ManBlkPlug1) setTopology(parentHash common.Hash, header *types.Header, interval *manparams.BCInterval, num uint64) ([]*mc.HD_OnlineConsensusVoteResultMsg, error) {
	NetTopology, onlineConsensusResults := bd.getNetTopology(num, parentHash, p.bcInterval)
	if nil == NetTopology {
		log.Error(ModuleManBlk, "获取网络拓扑图错误 ", "")
		NetTopology = &common.NetTopology{common.NetTopoTypeChange, nil}
	}
	if nil == onlineConsensusResults {
		onlineConsensusResults = make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)
	}
	log.Debug(ModuleManBlk, "获取拓扑结果 ", NetTopology, "高度", num)
	header.NetTopology = *NetTopology
	return onlineConsensusResults
}

func (bd *ManBlkPlug1) setTime(header *types.Header, tstamp int64) {
	header.Time = big.NewInt(tstamp)
}

func (bd *ManBlkPlug1) setExtra(header *types.Header) {
	header.Extra = make([]byte, 0)
}

func (bd *ManBlkPlug1) setGasLimit(header *types.Header, parent *types.Block) {
	header.GasLimit = core.CalcGasLimit(parent)
}

func (bd *ManBlkPlug1) setNumber(header *types.Header, num uint64) {
	header.Number = new(big.Int).SetUint64(num)
}

func (bd *ManBlkPlug1) setLeader(header *types.Header) {
	header.Leader = ca.GetAddress()
}
func (bd *ManBlkPlug1) setTimeStamp(parent *types.Block, header *types.Header, num uint64) {
	tstart := time.Now()
	log.Info(ModuleManBlk, "关键时间点", "区块头开始生成", "time", tstart, "块高", num)
	tstamp := tstart.Unix()
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}
	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		log.Info(ModuleManBlk, "等待时间同步", common.PrettyDuration(wait))
		time.Sleep(wait)
	}
	bd.setTime(header, tstamp)
}

func (bd *ManBlkPlug1) setParentHash(chain ChainReader, header *types.Header, num uint64) (*types.Block, error) {
	if num == 1 { // 第一个块直接返回创世区块作为父区块
		return chain.Genesis(), nil
	}

	if (bd.preBlockHash == common.Hash{}) {
		return nil, errors.Errorf("未知父区块hash[%s]", bd.preBlockHash.TerminalString())
	}

	parent := chain.GetBlockByHash(bd.preBlockHash)
	if nil == parent {
		return nil, errors.Errorf("未知的父区块[%s]", bd.preBlockHash.TerminalString())
	}

	parentHash := parent.Hash()
	header.ParentHash = parentHash
	return parent, nil
}
