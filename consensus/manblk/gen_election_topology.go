// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package manblk

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
)

func (p *ManBlkPlug1) genElection(support BlKSupport, state *state.StateDB) []common.Elect {
	info, err := support.GetElection(state, p.preBlockHash)
	if err != nil {
		log.Warn(ModuleManBlk, "获取选举信息错误", err)
		return nil
	}

	return support.TransferToElectionStu(info)
}

func (p *ManBlkPlug1) getNetTopology(support BlKSupport, num uint64, parentHash common.Hash, bcInterval *manparams.BCInterval) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg) {
	if bcInterval.IsReElectionNumber(num + 1) {
		return p.genAllNetTopology(support, parentHash)
	}

	return p.genChgNetTopology(support, num, parentHash)
}

func (p *ManBlkPlug1) genAllNetTopology(support BlKSupport, parentHash common.Hash) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg) {
	info, err := support.GetNetTopologyAll(parentHash)
	if err != nil {
		log.Warn(ModuleManBlk, "获取拓扑图错误", err)
		return nil, nil
	}

	return support.TransferToNetTopologyAllStu(info), nil
}

func (p *ManBlkPlug1) genChgNetTopology(support BlKSupport, num uint64, parentHash common.Hash) (*common.NetTopology, []*mc.HD_OnlineConsensusVoteResultMsg) {
	state, err := support.GetStateByHash(parentHash)
	if err != nil {
		log.Warn(ModuleManBlk, "生成拓扑变化", "获取父状态树失败", "err", err)
		return nil, nil
	}

	topoData, err := matrixstate.GetDataByState(mc.MSKeyTopologyGraph, state)
	if err != nil {
		log.Warn(ModuleManBlk, "生成拓扑变化", "状态树获取拓扑图失败", "err", err)
		return nil, nil
	}
	topology, OK := topoData.(*mc.TopologyGraph)
	if OK == false || topology == nil {
		log.Warn(ModuleManBlk, "生成拓扑变化", "拓扑图数据反射失败")
		return nil, nil
	}

	electStateData, err := matrixstate.GetDataByState(mc.MSKeyElectOnlineState, state)
	if err != nil {
		log.Warn(ModuleManBlk, "生成拓扑变化", "状态树获取elect在线状态失败", "err", err)
		return nil, nil
	}
	electState, OK := electStateData.(*mc.ElectOnlineStatus)
	if OK == false || topology == nil {
		log.Warn(ModuleManBlk, "生成拓扑变化", "elect在线状态数据反射失败")
		return nil, nil
	}

	onlineResults := support.GetConsensusOnlineResults()
	if len(onlineResults) == 0 {
		log.Info(ModuleManBlk, "生成拓扑变化信息", "无在线共识结果")
		return nil, nil
	}

	offlineNodes, onlineNods, consensusList := p.getOnlineStatus(onlineResults, topology, electState, num)

	// generate topology alter info
	alterInfo, err := support.GetTopoChange(parentHash, offlineNodes, onlineNods)
	if err != nil {
		log.Warn(ModuleManBlk, "获取拓扑变化信息错误", err)
		return nil, nil
	}
	for _, value := range alterInfo {
		log.Debug(ModuleManBlk, "获取拓扑变化地址", value.A, "位置", value.Position, "高度", num)
	}

	// generate self net topology
	ans := support.TransferToNetTopologyChgStu(alterInfo)
	return ans, consensusList
}

func (p *ManBlkPlug1) getOnlineStatus(onlineResults []*mc.HD_OnlineConsensusVoteResultMsg, topology *mc.TopologyGraph, electState *mc.ElectOnlineStatus, num uint64) ([]common.Address, []common.Address, []*mc.HD_OnlineConsensusVoteResultMsg) {
	offlineNodes := make([]common.Address, 0)
	onlineNods := make([]common.Address, 0)
	consensusList := make([]*mc.HD_OnlineConsensusVoteResultMsg, 0)
	// 筛选共识结果
	for i := 0; i < len(onlineResults); i++ {
		result := onlineResults[i]
		if result.Req == nil {
			log.Info(ModuleManBlk, "生成拓扑变化信息", "共识请求为空")
			continue
		}
		if result.IsValidity(num, manparams.OnlineConsensusValidityTime) == false {
			log.Info(ModuleManBlk, "生成拓扑变化信息", "高度 不合法", "请求高度", result.Req.Number, "当前高度", num)
			continue
		}

		node := result.Req.Node
		state := result.Req.OnlineState
		// 节点为当前拓扑图节点
		if topology.AccountIsInGraph(node) {
			if state == mc.OffLine {
				offlineNodes = append(offlineNodes, node)
				consensusList = append(consensusList, result)
			} else {
				log.Info(ModuleManBlk, "生成拓扑变化信息", "当前拓扑图中的节点，顶点共识状态错误", "状态", state)
			}
			continue
		}

		// 查看节点elect信息
		electInfo := electState.FindNodeElectOnlineState(node)
		if electInfo == nil {
			// 没有elect信息，表明节点非elect节点，不关心上下线信息
			continue
		}
		switch state {
		case mc.OnLine:
			if electInfo.Position == common.PosOffline {
				// 链上状态离线，当前共识结果在线，则需要上header
				onlineNods = append(onlineNods, node)
				consensusList = append(consensusList, result)
			}
		case mc.OffLine:
			if electInfo.Position == common.PosOnline {
				// 链上状态在线，当前共识结果离线，则需要上header
				offlineNodes = append(offlineNodes, node)
				consensusList = append(consensusList, result)
			}
		default:
			continue
		}
	}
	for i, value := range onlineNods {
		log.Debug(ModuleManBlk, "下线节点地址", value.String(), "序号", i)
	}
	for i, value := range offlineNodes {
		log.Debug(ModuleManBlk, "上线节点地址", value.String(), "序号", i)
	}
	return offlineNodes, onlineNods, consensusList
}
