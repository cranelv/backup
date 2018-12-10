// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package blkgenor

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/matrixwork"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/txpoolCache"
	"github.com/pkg/errors"
)

func (p *Process) processUpTime(work *matrixwork.Work, header *types.Header) error {

	if common.IsBroadcastNumber(header.Number.Uint64()-1) && header.Number.Uint64() > common.GetBroadcastInterval() {
		log.INFO("core", "区块插入验证", "完成创建work, 开始执行uptime")
		upTimeAccounts, err := work.GetUpTimeAccounts(header.Number.Uint64())
		if err != nil {
			log.ERROR("core", "获取所有抵押账户错误!", err, "高度", header.Number.Uint64())
			return err
		}
		calltherollMap, heatBeatUnmarshallMMap, err := work.GetUpTimeData(header.ParentHash)
		if err != nil {
			log.WARN("core", "获取心跳交易错误!", err, "高度", header.Number.Uint64())
		}

		err = work.HandleUpTime(work.State, upTimeAccounts, calltherollMap, heatBeatUnmarshallMMap, p.number, p.blockChain())
		if nil != err {
			log.ERROR("core", "处理uptime错误", err)
			return err
		}
	}

	return nil
}

func (p *Process) processHeaderGen() error {
	log.INFO(p.logExtraInfo(), "processHeaderGen", "start")
	defer log.INFO(p.logExtraInfo(), "processHeaderGen", "end")

	tstart := time.Now()
	parent, err := p.getParentBlock()
	if err != nil {
		return err
	}
	parentHash := parent.Hash()

	NetTopology, onlineConsensusResults := p.getNetTopology(p.number, parentHash)
	if nil == NetTopology {
		log.Error(p.logExtraInfo(), "获取网络拓扑图错误 ", "")
		NetTopology = &common.NetTopology{common.NetTopoTypeChange, nil}
	}
	if nil == onlineConsensusResults {
		onlineConsensusResults = make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)
	}

	log.Info(p.logExtraInfo(), "++++++++获取拓扑结果 ", NetTopology, "高度", p.number)

	tstamp := tstart.Unix()
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}
	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		log.Info("Mining too far in the future", "wait", common.PrettyDuration(wait))
		time.Sleep(wait)
	}
	account, vrfValue, vrfProof, err := p.getVrfValue(parent)
	if err != nil {
		log.INFO(p.logExtraInfo(), "区块生成阶段 获取vrfValue失败 err", err)
		return err
	}
	header := &types.Header{
		ParentHash:        parentHash,
		Leader:            ca.GetAddress(),
		Number:            new(big.Int).SetUint64(p.number),
		GasLimit:          core.CalcGasLimit(parent),
		Extra:             make([]byte, 0),
		Time:              big.NewInt(tstamp),
		NetTopology:       *NetTopology,
		Signatures:        make([]common.Signature, 0),
		Version:           parent.Header().Version, //param
		VersionSignatures: parent.Header().VersionSignatures,
		VrfValue:          common.GetHeaderVrf(account, vrfValue, vrfProof),
	}
	log.INFO("version-elect", "version", header.Version, "elect", header.Elect)
	log.INFO(p.logExtraInfo(), " vrf data headermsg", header.VrfValue, "账户户", account, "vrfValue", vrfValue, "vrfProff", vrfProof, "高度", header.Number.Uint64())
	if err := p.engine().Prepare(p.blockChain(), header); err != nil {
		log.ERROR(p.logExtraInfo(), "Failed to prepare header for mining", err)
		return err
	}

	tsBlock, txsCode, stateDB, receipts, err := p.genHeaderTxs(header)
	if err != nil {
		log.Error(p.logExtraInfo(), "运行交易失败", err)
		return err
	}

	err = p.blockChain().ProcessMatrixState(tsBlock, stateDB)
	if err != nil {
		log.Error(p.logExtraInfo(), "运行matrix状态树失败", err)
		return err
	}

	// 运行完状态树后，才能获取elect
	Elect := p.genElection(stateDB)
	if Elect == nil {
		return errors.New("生成elect信息错误!")
	}
	log.Info(p.logExtraInfo(), "++++++++获取选举结果 ", Elect, "高度", p.number)
	header = tsBlock.Header()
	header.Elect = Elect
	//运行完matrix状态树后，生成root
	txs := tsBlock.Transactions()
	block, err := p.engine().Finalize(p.blockChain(), header, stateDB, txs, nil, receipts)
	if err != nil {
		log.Error(p.logExtraInfo(), "最终finalize错误", err)
		return err
	}

	if block.IsBroadcastBlock() {
		header = block.Header()
		signHash := header.HashNoSignsAndNonce()
		sign, err := p.signHelper().SignHashWithValidate(signHash.Bytes(), true)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "广播区块生成，签名错误", err)
			return err
		}

		header.Signatures = make([]common.Signature, 0, 1)
		header.Signatures = append(header.Signatures, sign)
		sendMsg := &mc.BlockData{Header: header, Txs: txs}
		log.INFO(p.logExtraInfo(), "广播挖矿请求(本地), number", sendMsg.Header.Number, "root", header.Root.TerminalString(), "tx数量", sendMsg.Txs.Len())
		mc.PublishEvent(mc.HD_BroadcastMiningReq, &mc.BlockGenor_BroadcastMiningReqMsg{sendMsg})
	} else {
		header = block.Header()
		p2pBlock := &mc.HD_BlkConsensusReqMsg{
			Header:                 header,
			TxsCode:                txsCode,
			ConsensusTurn:          p.consensusTurn,
			OnlineConsensusResults: onlineConsensusResults,
			From: ca.GetAddress()}
		//send to local block verify module
		localBlock := &mc.LocalBlockVerifyConsensusReq{BlkVerifyConsensusReq: p2pBlock, Txs: txs, Receipts: receipts, State: stateDB}
		if len(txs[2:]) > 0 {
			txpoolCache.MakeStruck(txs[2:], header.HashNoSignsAndNonce(), p.number)
		}
		log.INFO(p.logExtraInfo(), "!!!!本地发送区块验证请求, root", p2pBlock.Header.Root.TerminalString(), "高度", p.number)
		mc.PublishEvent(mc.BlockGenor_HeaderVerifyReq, localBlock)
		p.startConsensusReqSender(p2pBlock)
	}

	return nil
}

func (p *Process) genHeaderTxs(header *types.Header) (*types.Block, []*common.RetCallTxN, *state.StateDB, []*types.Receipt, error) {
	//broadcast txs deal,remove no validators txs
	if common.IsBroadcastNumber(header.Number.Uint64()) {
		work, err := matrixwork.NewWork(p.blockChain().Config(), p.blockChain(), nil, header)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "NewWork!", err, "高度", p.number)
			return nil, nil, nil, nil, err
		}
		mapTxs := p.pm.matrix.TxPool().GetAllSpecialTxs()

		Txs := make([]types.SelfTransaction, 0)
		for _, txs := range mapTxs {
			for _, tx := range txs {
				log.INFO(p.logExtraInfo(), "交易数据 t", tx)
			}
			Txs = append(Txs, txs...)
		}
		// todo: add rewward and run
		//rewardList:=p.calcRewardAndSlash(work.State, header)

		work.ProcessBroadcastTransactions(p.pm.matrix.EventMux(), Txs, p.pm.bc)
		//work.ProcessBroadcastTransactions(p.pm.matrix.EventMux(), Txs, p.pm.bc)
		retTxs := work.GetTxs()
		for _, tx := range retTxs {
			log.INFO("==========", "Finalize:GasPrice", tx.GasPrice(), "amount", tx.Value())
		}

		block := types.NewBlock(header, retTxs, nil, work.Receipts)
		return block, nil, work.State, work.Receipts, nil

	} else {
		log.INFO(p.logExtraInfo(), "区块验证请求生成，交易部分", "开始创建work")
		work, err := matrixwork.NewWork(p.blockChain().Config(), p.blockChain(), nil, header)
		if err != nil {
			log.ERROR(p.logExtraInfo(), "NewWork!", err, "高度", p.number)
			return nil, nil, nil, nil, err
		}

		//work.commitTransactions(self.mux, Txs, self.chain)
		// todo： update uptime
		p.processUpTime(work, header)

		txsCode, Txs := work.ProcessTransactions(p.pm.matrix.EventMux(), p.pm.txPool, p.blockChain())
		//txsCode, Txs := work.ProcessTransactions(p.pm.matrix.EventMux(), p.pm.txPool, p.blockChain(),nil,nil)
		log.INFO("=========", "ProcessTransactions finish", len(txsCode))
		log.INFO(p.logExtraInfo(), "区块验证请求生成，交易部分", "完成执行交易, 开始finalize")
		block := types.NewBlock(header, Txs, nil, work.Receipts)
		log.INFO(p.logExtraInfo(), "区块验证请求生成，交易部分,完成 tx hash", block.TxHash())
		return block, txsCode, work.State, work.Receipts, nil
	}
}

func (p *Process) getParentBlock() (*types.Block, error) {
	if p.number == 1 { // 第一个块直接返回创世区块作为父区块
		return p.blockChain().Genesis(), nil
	}

	if (p.preBlockHash == common.Hash{}) {
		return nil, errors.Errorf("未知父区块hash[%s]", p.preBlockHash.TerminalString())
	}

	parent := p.blockChain().GetBlockByHash(p.preBlockHash)
	if nil == parent {
		return nil, errors.Errorf("未知的父区块[%s]", p.preBlockHash.TerminalString())
	}

	return parent, nil
}

func (p *Process) startConsensusReqSender(req *mc.HD_BlkConsensusReqMsg) {
	p.closeConsensusReqSender()
	sender, err := common.NewResendMsgCtrl(req, p.sendConsensusReqFunc, manparams.BlkPosReqSendInterval, manparams.BlkPosReqSendTimes)
	if err != nil {
		log.ERROR(p.logExtraInfo(), "创建POS完成的req发送器", "失败", "err", err)
		return
	}
	p.consensusReqSender = sender
}

func (p *Process) closeConsensusReqSender() {
	if p.consensusReqSender == nil {
		return
	}
	p.consensusReqSender.Close()
	p.consensusReqSender = nil
}

func (p *Process) sendConsensusReqFunc(data interface{}, times uint32) {
	req, OK := data.(*mc.HD_BlkConsensusReqMsg)
	if !OK {
		log.ERROR(p.logExtraInfo(), "发出区块共识req", "反射消息失败", "次数", times)
		return
	}
	log.INFO(p.logExtraInfo(), "!!!!网络发送区块验证请求, hash", req.Header.HashNoSignsAndNonce(), "tx数量", len(req.TxsCode), "次数", times)
	p.pm.hd.SendNodeMsg(mc.HD_BlkConsensusReq, req, common.RoleValidator, nil)
}

func (p *Process) getVrfValue(parent *types.Block) ([]byte, []byte, []byte, error) {
	_, preVrfValue, preVrfProof := common.GetVrfInfoFromHeader(parent.Header().VrfValue)
	parentMsg := VrfMsg{
		VrfProof: preVrfProof,
		VrfValue: preVrfValue,
		Hash:     parent.Hash(),
	}
	vrfmsg, err := json.Marshal(parentMsg)
	if err != nil {
		log.Error(p.logExtraInfo(), "生成vefmsg出错", err, "parentMsg", parentMsg)
		return []byte{}, []byte{}, []byte{}, errors.New("生成vrfmsg出错")
	} else {
		log.Error("生成vrfmsg成功")
	}

	log.Info("msgggggvrf_gen", "preVrfMsg", vrfmsg, "高度", p.number, "VrfProof", parentMsg.VrfProof, "VrfValue", parentMsg.VrfValue, "Hash", parentMsg.Hash)
	if err != nil {
		log.Error(p.logExtraInfo(), "生成vrfValue,vrfProof失败 err", err)
	} else {
		log.Error(p.logExtraInfo(), "生成vrfValue,vrfProof成功 err", err)
	}
	return p.signHelper().SignVrf(vrfmsg)
}
