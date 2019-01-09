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

type ManBlkBasePlug struct {
	preBlockHash common.Hash
}

func NewBlkBasePlug(preBlockHash common.Hash) (*ManBlkBasePlug, error) {
	obj := new(ManBlkBasePlug)
	return obj, nil
}

func (p *ManBlkBasePlug) getVrfValue(support BlKSupport, parent *types.Block) ([]byte, []byte, []byte, error) {
	_, preVrfValue, preVrfProof := baseinterface.NewVrf().GetVrfInfoFromHeader(parent.Header().VrfValue)
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

func (p *ManBlkBasePlug) setVrf(support BlKSupport, parent *types.Block, header *types.Header) error {
	account, vrfValue, vrfProof, err := p.getVrfValue(support, parent)
	if err != nil {
		log.Error(ModuleManBlk, "区块生成阶段 获取vrfValue失败 错误", err)
		return err
	}
	header.VrfValue = baseinterface.NewVrf().GetHeaderVrf(account, vrfValue, vrfProof)
	return nil
}

func (bd *ManBlkBasePlug) setTopology(support BlKSupport, parentHash common.Hash, header *types.Header, interval *manparams.BCInterval, num uint64) ([]*mc.HD_OnlineConsensusVoteResultMsg, error) {
	NetTopology, onlineConsensusResults := support.GetNetTopology(num, parentHash, interval)
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

func (bd *ManBlkBasePlug) setTime(header *types.Header, tstamp int64) {
	header.Time = big.NewInt(tstamp)
}

func (bd *ManBlkBasePlug) setExtra(header *types.Header) {
	header.Extra = make([]byte, 0)
}

func (bd *ManBlkBasePlug) setGasLimit(header *types.Header, parent *types.Block) {
	header.GasLimit = core.CalcGasLimit(parent)
}

func (bd *ManBlkBasePlug) setNumber(header *types.Header, num uint64) {
	header.Number = new(big.Int).SetUint64(num)
}

func (bd *ManBlkBasePlug) setLeader(header *types.Header) {
	header.Leader = ca.GetAddress()
}
func (bd *ManBlkBasePlug) setTimeStamp(parent *types.Block, header *types.Header, num uint64) {
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

func (bd *ManBlkBasePlug) setParentHash(chain ChainReader, header *types.Header, num uint64) (*types.Block, error) {
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

func (bd *ManBlkBasePlug) Prepare(support BlKSupport, interval *manparams.BCInterval, num uint64, args ...interface{}) (*types.Header, error) {
	preBlockHash, ok := args[0].(common.Hash)
	if !ok {
		log.Error(ModuleManBlk, "反射失败", "")
		return nil, errors.New("反射失败")
	}
	bd.preBlockHash = preBlockHash
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

func (bd *ManBlkBasePlug) ProcessState(support BlKSupport, header *types.Header, args ...interface{}) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, error) {
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

func (bd *ManBlkBasePlug) Finalize(support BlKSupport, header *types.Header, state *state.StateDB, txs []types.SelfTransaction, uncles []*types.Header, receipts []*types.Receipt, args interface{}) (*types.Block, error) {

	block, err := support.Engine().Finalize(support, header, state, txs, uncles, receipts)
	if err != nil {
		log.Error(ModuleManBlk, "最终finalize错误", err)
		return nil, err
	}
	return block, nil
}

func (bd *ManBlkBasePlug) VerifyHeader(support BlKSupport, header *types.Header, args ...interface{}) error {
	if err := support.VerifyHeader(header); err != nil {
		log.ERROR(ModuleManBlk, "预验证头信息失败", err, "高度", header.Number.Uint64())
		return err
	}

	// verify net topology info
	onlineConsensusResults, ok := args[0].([]*mc.HD_OnlineConsensusVoteResultMsg)
	if !ok {
		log.ERROR(ModuleManBlk, "反射顶点配置失败", "")
	}
	if err := support.VerifyNetTopology(header, onlineConsensusResults); err != nil {
		log.ERROR(ModuleManBlk, "验证拓扑信息失败", err, "高度", header.Number.Uint64())
		return err
	}

	if err := support.DPOSEngine().VerifyVersion(support, header); err != nil {
		log.ERROR(ModuleManBlk, "验证版本号失败", err, "高度", header.Number.Uint64())
		return err
	}

	//verify vrf
	if err := support.VerifyVrf(header); err != nil {
		log.Error(ModuleManBlk, "验证vrf失败", err, "高度", header.Number.Uint64())
		return err
	}
	log.INFO(ModuleManBlk, "验证vrf成功 高度", header.Number.Uint64())

	return nil
}

func (bd *ManBlkBasePlug) VerifyTxsAndState(support BlKSupport, verifyHeader *types.Header, verifyTxs types.SelfTransactions, args ...interface{}) error {
	log.INFO(ModuleManBlk, "开始交易验证, 数量", len(verifyTxs), "高度", verifyHeader.Number.Uint64())

	//跑交易交易验证， Root TxHash ReceiptHash Bloom GasLimit GasUsed
	localHeader := types.CopyHeader(verifyHeader)
	localHeader.GasUsed = 0
	verifyHeaderHash := verifyHeader.HashNoSignsAndNonce()
	work, err := matrixwork.NewWork(support.Config(), support, nil, localHeader)
	if err != nil {
		log.ERROR(ModuleManBlk, "交易验证，创建work失败!", err, "高度", verifyHeader.Number.Uint64())
		return err
	}

	uptimeMap, err := support.ProcessUpTime(work.State, localHeader)
	if err != nil {
		log.Error(ModuleManBlk, "uptime处理错误", err)
		return err
	}
	err = work.ConsensusTransactions(support.EventMux(), verifyTxs, uptimeMap)
	if err != nil {
		log.ERROR(ModuleManBlk, "交易验证，共识执行交易出错!", err, "高度", verifyHeader.Number.Uint64())
		return err
	}
	finalTxs := work.GetTxs()
	localBlock := types.NewBlock(localHeader, finalTxs, nil, work.Receipts)
	// process matrix state
	err = support.ProcessMatrixState(localBlock, work.State)
	if err != nil {
		log.ERROR(ModuleManBlk, "matrix状态验证,错误", "运行matrix状态出错", "err", err)
		return err
	}

	// 运行完matrix state后，生成root
	localBlock, err = support.Engine().Finalize(support, localHeader, work.State, finalTxs, nil, work.Receipts)
	if err != nil {
		log.ERROR(ModuleManBlk, "matrix状态验证,错误", "Failed to finalize block for sealing", "err", err)
		return err
	}

	log.Info(ModuleManBlk, "共识后的交易本地hash", localBlock.TxHash(), "共识后的交易远程hash", verifyHeader.TxHash)
	log.Info("miss tree node debug", "finalize root", localBlock.Root().Hex(), "remote root", verifyHeader.Root.Hex())

	// verify election info
	if err := support.VerifyElection(verifyHeader, work.State); err != nil {
		log.ERROR(ModuleManBlk, "验证选举信息失败", err, "高度", verifyHeader.Number.Uint64())
		return err
	}

	//localBlock check
	localHeader = localBlock.Header()
	localHash := localHeader.HashNoSignsAndNonce()

	if localHash != verifyHeaderHash {
		log.ERROR(ModuleManBlk, "交易验证及状态，错误", "block hash不匹配",
			"local hash", localHash.TerminalString(), "remote hash", verifyHeaderHash.TerminalString(),
			"local root", localHeader.Root.TerminalString(), "remote root", verifyHeader.Root.TerminalString(),
			"local txHash", localHeader.TxHash.TerminalString(), "remote txHash", verifyHeader.TxHash.TerminalString(),
			"local ReceiptHash", localHeader.ReceiptHash.TerminalString(), "remote ReceiptHash", verifyHeader.ReceiptHash.TerminalString(),
			"local Bloom", localHeader.Bloom.Big(), "remote Bloom", verifyHeader.Bloom.Big(),
			"local GasLimit", localHeader.GasLimit, "remote GasLimit", verifyHeader.GasLimit,
			"local GasUsed", localHeader.GasUsed, "remote GasUsed", verifyHeader.GasUsed)
		return errors.New("hash 不一致")
	}
	return nil
}
