package blkmanage

import (
	"errors"
	"time"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/matrixwork"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
)

type ManBCBlkPlug struct {
	baseInterface *ManBlkBasePlug
}

func NewBCBlkPlug() (*ManBCBlkPlug, error) {
	obj := new(ManBCBlkPlug)
	obj.baseInterface, _ = NewBlkBasePlug()
	return obj, nil
}

func (bd *ManBCBlkPlug) Prepare(support BlKSupport, interval *manparams.BCInterval, num uint64, args interface{}) (*types.Header, interface{}, error) {

	return bd.baseInterface.Prepare(support, interval, num, args)
}

func (bd *ManBCBlkPlug) ProcessState(support BlKSupport, header *types.Header, args interface{}) ([]*common.RetCallTxN, *state.StateDB, []*types.Receipt, []types.SelfTransaction, []types.SelfTransaction, interface{}, error) {

	work, err := matrixwork.NewWork(support.BlockChain().Config(), support.BlockChain(), nil, header)
	if err != nil {
		log.ERROR(ModuleManBlk, "NewWork!", err, "高度", header.Number.Uint64())
		return nil, nil, nil, nil, nil, nil, err
	}

	mapTxs := support.TxPool().GetAllSpecialTxs()
	Txs := make([]types.SelfTransaction, 0)
	for _, txs := range mapTxs {
		for _, tx := range txs {
			log.Trace(ModuleManBlk, "交易数据", tx)
		}
		Txs = append(Txs, txs...)
	}
	work.ProcessBroadcastTransactions(support.EventMux(), Txs)
	log.Info(ModuleManBlk, "关键时间点", "开始执行MatrixState", "time", time.Now(), "块高", header.Number.Uint64())
	block := types.NewBlock(header, work.GetTxs(), nil, work.Receipts)
	err = support.BlockChain().ProcessMatrixState(block, work.State)
	if err != nil {
		log.Error(ModuleManBlk, "运行matrix状态树失败", err)
		return nil, nil, nil, nil, nil, nil, err
	}

	return nil, work.State, work.Receipts, Txs, work.GetTxs(), nil, nil
}

func (bd *ManBCBlkPlug) Finalize(support BlKSupport, header *types.Header, state *state.StateDB, txs []types.SelfTransaction, uncles []*types.Header, receipts []*types.Receipt, args interface{}) (*types.Block, interface{}, error) {

	block, _, err := bd.baseInterface.Finalize(support, header, state, txs, uncles, receipts, nil)
	if err != nil {
		log.Error(ModuleManBlk, "最终finalize错误", err)
		return nil, nil, err
	}
	return block, nil, nil
}

func (bd *ManBCBlkPlug) VerifyHeader(support BlKSupport, header *types.Header, args interface{}) (interface{}, error) {
	if err := support.BlockChain().VerifyHeader(header); err != nil {
		log.ERROR(ModuleManBlk, "预验证头信息失败", err, "高度", header.Number.Uint64())
		return nil, err
	}

	if err := support.BlockChain().DPOSEngine().VerifyBlock(support.BlockChain(), header); err != nil {
		log.WARN(ModuleManBlk, "验证广播挖矿结果", "结果异常", "err", err)
		return nil, err
	}
	onlineConsensusResults := make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)

	if err := support.ReElection().VerifyNetTopology(header, onlineConsensusResults); err != nil {
		log.ERROR(ModuleManBlk, "验证拓扑信息失败", err, "高度", header.Number.Uint64())
		return nil, err
	}

	if err := support.BlockChain().DPOSEngine().VerifyVersion(support.BlockChain(), header); err != nil {
		log.ERROR(ModuleManBlk, "验证版本号失败", err, "高度", header.Number.Uint64())
		return nil, err
	}

	//verify vrf
	if err := support.ReElection().VerifyVrf(header); err != nil {
		log.Error(ModuleManBlk, "验证vrf失败", err, "高度", header.Number.Uint64())
		return nil, err
	}
	log.INFO(ModuleManBlk, "验证vrf成功 高度", header.Number.Uint64())

	return nil, nil
}

func (bd *ManBCBlkPlug) VerifyTxsAndState(support BlKSupport, verifyHeader *types.Header, verifyTxs types.SelfTransactions, args interface{}) (*state.StateDB, types.SelfTransactions, []*types.Receipt, interface{}, error) {
	parent := support.BlockChain().GetBlockByHash(verifyHeader.ParentHash)
	if parent == nil {
		log.WARN(ModuleManBlk, "广播挖矿结果验证", "获取父区块错误!")
		return nil, nil, nil, nil, errors.New("广播挖矿结果验证,获取父区块错误!")
	}

	work, err := matrixwork.NewWork(support.BlockChain().Config(), support.BlockChain(), nil, verifyHeader)
	if err != nil {
		log.WARN(ModuleManBlk, "广播挖矿结果验证, 创建worker错误", err)
		return nil, nil, nil, nil, err
	}
	//执行交易
	work.ProcessBroadcastTransactions(support.EventMux(), verifyTxs)

	retTxs := work.GetTxs()
	// 运行matrix状态树
	block := types.NewBlock(verifyHeader, retTxs, nil, work.Receipts)
	if err := support.BlockChain().ProcessMatrixState(block, work.State); err != nil {
		log.ERROR(ModuleManBlk, "广播挖矿结果验证, matrix 状态树运行错误", err)
		return nil, nil, nil, nil, err
	}

	localBlock, err := support.BlockChain().Engine().Finalize(support.BlockChain(), block.Header(), work.State, retTxs, nil, work.Receipts)
	if err != nil {
		log.ERROR(ModuleManBlk, "Failed to finalize block for sealing", err)
		return nil, nil, nil, nil, err
	}

	if localBlock.Root() != verifyHeader.Root {
		log.ERROR(ModuleManBlk, "广播挖矿结果验证", "root验证错误, 不匹配", "localRoot", localBlock.Root().TerminalString(), "remote root", verifyHeader.Root.TerminalString())
		return nil, nil, nil, nil, errors.New("root不一致")
	}

	return work.State, retTxs, work.Receipts, nil, nil
}
