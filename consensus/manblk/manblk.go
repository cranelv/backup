package manblk

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/matrix/go-matrix/matrixwork"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/params/manparams"

	"github.com/pkg/errors"

	"github.com/matrix/go-matrix/log"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mc"
)

type ManBlkPlug1 struct {
	preBlockHash common.Hash
}

func NewBlkPlug1(preBlockHash common.Hash) (*ManBlkPlug1, error) {
	obj := new(ManBlkPlug1)
	obj.preBlockHash = preBlockHash
	return obj, nil
}

func (p *ManBlkPlug1) getVrfValue(support BlKSupport, parent *types.Block) ([]byte, []byte, []byte, error) {
	_, preVrfValue, preVrfProof := support.GetVrfInfoFromHeader(parent.Header().VrfValue)
	parentMsg := VrfMsg{
		VrfProof: preVrfProof,
		VrfValue: preVrfValue,
		Hash:     parent.Hash(),
	}
	vrfmsg, err := json.Marshal(parentMsg)
	if err != nil {
		log.Error(ModuleManBlk, "生成vrfmsg出错", err, "parentMsg", parentMsg)
		return []byte{}, []byte{}, []byte{}, errors.New("生成vrfmsg出错")
	}
	return support.SignVrf(vrfmsg, p.preBlockHash)
}

func (p *ManBlkPlug1) setVrf(support BlKSupport, parent *types.Block, header *types.Header) error {
	account, vrfValue, vrfProof, err := p.getVrfValue(support, parent)
	if err != nil {
		log.Error(ModuleManBlk, "区块生成阶段 获取vrfValue失败 错误", err)
		return err
	}
	header.VrfValue = baseinterface.NewVrf().GetHeaderVrf(account, vrfValue, vrfProof)
	return nil
}

func (bd *ManBlkPlug1) setTopology(support BlKSupport, parentHash common.Hash, header *types.Header, interval *manparams.BCInterval, num uint64) ([]*mc.HD_OnlineConsensusVoteResultMsg, error) {
	NetTopology, onlineConsensusResults := bd.getNetTopology(support, num, parentHash, interval)
	if nil == NetTopology {
		log.Error(ModuleManBlk, "获取网络拓扑图错误 ", "")
		NetTopology = &common.NetTopology{common.NetTopoTypeChange, nil}
		return nil, errors.New("获取网络拓扑图错误 ")
	}
	if nil == onlineConsensusResults {
		onlineConsensusResults = make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)
	}
	log.Debug(ModuleManBlk, "获取拓扑结果 ", NetTopology, "高度", num)
	header.NetTopology = *NetTopology
	return onlineConsensusResults, nil
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

func (bd *ManBlkPlug1) Prepare(support BlKSupport, interval *manparams.BCInterval, num uint64) (*types.Header, error) {
	originHeader := new(types.Header)
	parent, err := bd.setParentHash(support, originHeader, num)
	if nil != err {
		log.ERROR(ModuleManBlk, "区块生成阶段", "获取父区块失败")
		return nil, err
	}

	bd.setTimeStamp(parent, originHeader, num)
	bd.setLeader(originHeader)
	bd.setNumber(originHeader, num)
	bd.setGasLimit(originHeader, parent)
	bd.setExtra(originHeader)
	bd.setTopology(support, parent.Hash(), originHeader, interval, num)
	err = bd.setVrf(support, parent, originHeader)
	if nil != err {
		return nil, err
	}
	if err := support.Engine().Prepare(support, originHeader); err != nil {
		log.ERROR(ModuleManBlk, "Failed to prepare header for mining", err)
		return nil, err
	}
	return originHeader, nil
}

func (bd *ManBlkPlug1) ProcessState(support BlKSupport, header *types.Header) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, error) {
	work, err := matrixwork.NewWork(support.Config(), support, nil, header)
	upTimeMap, err := support.ProcessUpTime(work.State, header)
	if err != nil {
		log.ERROR(ModuleManBlk, "执行uptime错误", err, "高度", header.Number)
		return nil, nil, nil, nil, nil, err
	}
	txsCode, originalTxs, finalTxs := work.ProcessTransactions(support.EventMux(), support, upTimeMap)
	block := types.NewBlock(header, finalTxs, nil, work.Receipts)
	log.Debug(ModuleManBlk, "区块验证请求生成，交易部分,完成 tx hash", block.TxHash())

	err = support.ProcessMatrixState(block, work.State)
	if err != nil {
		log.Error(ModuleManBlk, "运行matrix状态树失败", err)
		return nil, nil, nil, nil, nil, err
	}
	return txsCode, work.State, work.Receipts, originalTxs, finalTxs, nil
}

func (bd *ManBlkPlug1) Finalize(support BlKSupport, header *types.Header, state *state.StateDB, txs []types.SelfTransaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {

	block, err := support.Engine().Finalize(support, header, state, txs, uncles, receipts)
	if err != nil {
		log.Error(ModuleManBlk, "最终finalize错误", err)
		return nil, err
	}
	return block, nil
}

func VerifyBlock(support BlKSupport, header *types.Header) error {
	return nil
}
