// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package miner

import (
	"github.com/MatrixAINetwork/go-matrix/mc"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/consensus"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/pkg/errors"
	"math/big"
)


func newMinReqHolder() *RequestHolder{
	return &RequestHolder{
		cache:NewRequestCache(5),
		requestCreater : func(header *types.Header, miner common.Address, isBroadcastReq bool, isfriend bool) MinerRequestInterface {
			headerHash := header.HashNoSignsAndNonce()
			reqData := newMineReqData(headerHash, header, nil, isBroadcastReq,isfriend)
			reqData.coinbase = miner
			return reqData
		},
		RequestHash : func(header *types.Header) common.Hash {
			return header.HashNoSignsAndNonce()
		},
		RequestNumber : func (header *types.Header) uint64 {
			return header.Number.Uint64()
		},
	}
}
func newMinX11Holder() *RequestHolder{
	return &RequestHolder{
		cache:NewRequestCache(5),
		requestCreater : func(header *types.Header, miner common.Address, isBroadcastReq bool, isfriend bool) MinerRequestInterface {
			reqData := CreateMinePowerTask(header)
			if reqData != nil{
				reqData.coinbase = miner
				reqData.isFriend = isfriend
				return reqData
			}
			return nil
		},
		RequestHash : func(header *types.Header) common.Hash {
			return header.HashNoSignsAndNonce()
		},
		RequestNumber : func (header *types.Header) uint64 {
			return header.Number.Uint64()
		},
	}
}
type mineReqCtrl struct {
	curSuperSeq     uint64
	curNumber       uint64
	currentMineReq  *mineReqData
	role            common.RoleType
	bcInterval      *mc.BCIntervalInfo
	bc              ChainReader
	validatorReader consensus.StateReader
	reqHolder        *RequestHolder
	x11Holder        *RequestHolder
	X11ReqCache        *RequestCache
}

func newMinReqCtrl(bc ChainReader) *mineReqCtrl {
	return &mineReqCtrl{
		curSuperSeq:     0,
		curNumber:       0,
		currentMineReq:  nil,
		role:            common.RoleNil,
		bcInterval:      nil,
		validatorReader: bc,
		bc:              bc,
		reqHolder:        newMinReqHolder(),
		x11Holder:		 newMinX11Holder(),
		X11ReqCache:     NewRequestCache(5),
	}
}
func (ctrl *mineReqCtrl) getCurrentHolder() *RequestHolder{
	return ctrl.x11Holder
}
func (ctrl *mineReqCtrl) AddMineReq(header *types.Header, miner common.Address, isBroadcastReq bool,isfriend bool) (MinerRequestInterface, error) {

	return ctrl.getCurrentHolder().AddMineReq(header,miner,isBroadcastReq,isfriend)
}

func (ctrl *mineReqCtrl) getBeginMine() MinerRequestInterface {
	return ctrl.getCurrentHolder().beginMine()
}
func (ctrl *mineReqCtrl) SetMiningResult(result *types.Header,hash common.Hash) (MinerRequestInterface, error) {
	if nil == result {
		return nil, errors.New("消息为nil")
	}
	result.Number.Sub(result.Number,big.NewInt(2))
	return ctrl.getCurrentHolder().SetMiningResult(result,hash)
}

func (ctrl *mineReqCtrl) checkMineReq(header *types.Header) error {
	return nil
	if header.Difficulty.Uint64() == 0 {
		return difficultyIsZero
	}

	err := ctrl.bc.DPOSEngine(header.Version).VerifyBlock(ctrl.validatorReader, header)
	if err != nil {
		return errors.Errorf("挖矿请求POS验证失败(%v)", err)
	}
	return nil
}

func (ctrl *mineReqCtrl) roleCanMine(role common.RoleType, number uint64) bool {
	if ctrl.bcInterval == nil {
		return false
	}

	if ctrl.bcInterval.IsBroadcastNumber(number) {
		return role == common.RoleBroadcast
	} else {
		return role == common.RoleMiner || role == common.RoleInnerMiner
	}
}

