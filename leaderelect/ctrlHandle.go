// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package leaderelect

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"time"
)

func (self *controller) handleMsg(data interface{}) {
	if nil == data {
		log.WARN(self.logInfo, "消息处理", "收到nil消息")
		return
	}

	switch data.(type) {
	case *startControllerMsg:
		msg, _ := data.(*startControllerMsg)
		self.handleStartMsg(msg)

	case *mc.BlockPOSFinishedNotify:
		msg, _ := data.(*mc.BlockPOSFinishedNotify)
		self.handleBlockPOSFinishedNotify(msg)

	case *mc.HD_ReelectInquiryReqMsg:
		msg, _ := data.(*mc.HD_ReelectInquiryReqMsg)
		self.handleInquiryReq(msg)

	case *mc.HD_ReelectInquiryRspMsg:
		msg, _ := data.(*mc.HD_ReelectInquiryRspMsg)
		self.handleInquiryRsp(msg)

	case *mc.HD_ReelectLeaderReqMsg:
		msg, _ := data.(*mc.HD_ReelectLeaderReqMsg)
		self.handleRLReq(msg)

	case *mc.HD_ConsensusVote:
		msg, _ := data.(*mc.HD_ConsensusVote)
		self.handleRLVote(msg)

	case *mc.HD_ReelectResultBroadcastMsg:
		msg, _ := data.(*mc.HD_ReelectResultBroadcastMsg)
		self.handleResultBroadcastMsg(msg)

	case *mc.HD_ReelectResultRspMsg:
		msg, _ := data.(*mc.HD_ReelectResultRspMsg)
		self.handleResultRsp(msg)

	default:
		log.WARN(self.logInfo, "消息处理", "未知消息类型")
	}
}

func (self *controller) handleStartMsg(msg *startControllerMsg) {
	if nil == msg || nil == msg.parentHeader {
		log.WARN(self.logInfo, "处理开始消息错误", "nil")
		return
	}

	log.INFO(self.logInfo, "处理开始消息", "start", "高度", self.dc.number, "preLeader", msg.parentHeader.Leader, "header time", msg.parentHeader.Time.Int64())
	preIsSupper := msg.parentHeader.IsSuperHeader()
	if err := self.dc.AnalysisState(msg.parentHeader.Hash(), preIsSupper, msg.parentHeader.Leader, msg.parentStateDB); err != nil {
		log.ERROR(self.logInfo, "处理开始消息", "分析状态树信息错误", "err", err)
		return
	}

	root2, _ := msg.parentStateDB.Commit(self.dc.chain.Config().IsEIP158(msg.parentHeader.Number))
	if root2 != msg.parentHeader.Root {
		log.Error("hyk_miss_trie_5", "root", msg.parentHeader.Root.TerminalString(), "state root", root2.TerminalString())
	}

	if self.dc.role != common.RoleValidator {
		log.INFO(self.logInfo, "处理开始消息", "身份错误, 不是验证者", "role", self.dc.role, "高度", self.dc.number)
		self.mp.SaveParentHeader(msg.parentHeader)
		return
	}

	if self.dc.bcInterval.IsBroadcastNumber(self.dc.number) {
		log.INFO(self.logInfo, "处理开始消息", "区块为广播区块，不开启定时器")
		self.dc.state = stIdle
		self.sendLeaderMsg()
		self.mp.SaveParentHeader(msg.parentHeader)
		return
	}

	if self.dc.turnTime.SetBeginTime(mc.ConsensusTurnInfo{}, msg.parentHeader.Time.Int64()) {
		log.INFO(self.logInfo, "处理开始消息", "更新轮次时间成功", "高度", self.dc.number)
		self.mp.SaveParentHeader(msg.parentHeader)
		if isFirstConsensusTurn(self.ConsensusTurn()) {
			curTime := time.Now().Unix()
			st, remainTime, reelectTurn := self.dc.turnTime.CalState(mc.ConsensusTurnInfo{}, curTime)
			log.INFO(self.logInfo, "处理开始消息", "计算状态结果", "状态", st, "剩余时间", remainTime, "重选轮次", reelectTurn)
			self.dc.state = st
			self.dc.curReelectTurn = 0
			self.setTimer(remainTime, self.timer)
			if st == stPos {
				self.processPOSState()
			} else if st == stReelect {
				self.startReelect(reelectTurn)
			}
		}
	}

	//公布leader身份
	self.sendLeaderMsg()
}

func (self *controller) handleBlockPOSFinishedNotify(msg *mc.BlockPOSFinishedNotify) {
	if nil == msg {
		log.WARN(self.logInfo, "处理POS完成通知消息错误", "nil")
		return
	}
	self.mp.SavePOSNotifyMsg(msg)
	self.processPOSState()
}

func (self *controller) timeOutHandle() {
	curTime := time.Now().Unix()
	st, remainTime, reelectTurn := self.dc.turnTime.CalState(self.dc.curConsensusTurn, curTime)
	switch self.State() {
	case stPos:
		log.INFO(self.logInfo, "超时事件", "POS未完成", "轮次", self.curTurnInfo(), "高度", self.Number(),
			"计算状态结果", st.String(), "下次超时时间", remainTime, "计算的重选轮次", reelectTurn,
			"轮次开始时间", self.dc.turnTime.GetBeginTime(self.ConsensusTurn()), "leader", self.dc.GetConsensusLeader().Hex())
	case stReelect:
		log.INFO(self.logInfo, "超时事件", "重选未完成", "轮次", self.curTurnInfo(), "高度", self.Number(),
			"计算状态结果", st.String(), "下次超时时间", remainTime, "计算的重选轮次", reelectTurn,
			"轮次开始时间", self.dc.turnTime.GetBeginTime(self.ConsensusTurn()), "master", self.dc.GetReelectMaster().Hex())
	default:
		log.ERROR(self.logInfo, "超时事件", "当前状态错误", self.State(), "轮次", self.curTurnInfo(), "高度", self.Number(),
			"轮次开始时间", self.dc.turnTime.GetBeginTime(self.ConsensusTurn()), "当前时间", curTime)
		return
	}

	self.setTimer(remainTime, self.timer)
	self.dc.state = st
	self.startReelect(reelectTurn)
}

func (self *controller) processPOSState() {
	if self.State() != stPos {
		log.INFO(self.logInfo, "执行检查POS状态", "状态不正常,不执行", "当前状态", self.State().String())
		return
	}

	if _, err := self.mp.GetPOSNotifyMsg(self.dc.GetConsensusLeader(), self.dc.curConsensusTurn); err != nil {
		log.INFO(self.logInfo, "执行检查POS状态", "获取POS完成消息失败", "err", err)
		return
	}

	self.dc.state = stMining
}
