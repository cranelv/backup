package manblk

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/matrixwork"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/pkg/errors"
)

type ManBCBlkPlug struct {
	baseInterface *ManBlkBasePlug
}

func NewBCBlkPlug(preBlockHash common.Hash) (*ManBCBlkPlug, error) {
	obj := new(ManBCBlkPlug)
	obj.baseInterface, _ = NewBlkBasePlug()
	return obj, nil
}

func (bd *ManBCBlkPlug) Prepare(support BlKSupport, interval *manparams.BCInterval, num uint64, args ...interface{}) (*types.Header, error) {

	return bd.baseInterface.Prepare(support, interval, num, args)
}

func (bd *ManBCBlkPlug) ProcessState(support BlKSupport, header *types.Header, args ...interface{}) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, error) {

	work, err := matrixwork.NewWork(support.Config(), support, nil, header)
	if err != nil {
		log.ERROR(ModuleManBlk, "NewWork!", err, "高度", header.Number.Uint64())
		return nil, nil, nil, nil, nil, err
	}

	mapTxs := support.GetAllSpecialTxs()
	Txs := make([]types.SelfTransaction, 0)
	for _, txs := range mapTxs {
		for _, tx := range txs {
			log.Trace(ModuleManBlk, "交易数据", tx)
		}
		Txs = append(Txs, txs...)
	}
	work.ProcessBroadcastTransactions(support.EventMux(), Txs)

	return nil, work.State, work.Receipts, Txs, Txs, nil
}

func (bd *ManBCBlkPlug) Finalize(support BlKSupport, header *types.Header, state *state.StateDB, txs []types.SelfTransaction, uncles []*types.Header, receipts []*types.Receipt, args interface{}) (*types.Block, error) {

	block, err := bd.baseInterface.Finalize(support, header, state, txs, uncles, receipts, nil)
	if err != nil {
		log.Error(ModuleManBlk, "最终finalize错误", err)
		return nil, err
	}
	return block, nil
}

func (bd *ManBCBlkPlug) VerifyHeader(support BlKSupport, header *types.Header, args ...interface{}) error {
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

func (bd *ManBCBlkPlug) VerifyTxsAndState(support BlKSupport, verifyHeader *types.Header, verifyTxs types.SelfTransactions, args ...interface{}) error {
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
