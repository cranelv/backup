// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package manparams

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/p2p/discover"
	"github.com/matrix/go-matrix/params"
)

const (
	//	VerifyNetChangeUpTime = 6
	//MinerNetChangeUpTime  = 4
	//
	//VerifyTopologyGenerateUpTime = 8
	//MinerTopologyGenerateUpTime  = 8

	//	RandomVoteTime = 5

	LRSParentMiningTime = int64(20)
	LRSPOSOutTime       = int64(20)
	LRSReelectOutTime   = int64(40)
	LRSReelectInterval  = 5

	VotePoolTimeout    = 55 * 1000
	VotePoolCountLimit = 5

	BlkPosReqSendInterval   = 5
	BlkPosReqSendTimes      = 6
	BlkVoteSendInterval     = 3
	BlkVoteSendTimes        = 8
	MinerReqSendInterval    = 3
	PosedReqSendInterval    = 10
	MinerResultSendInterval = 3

	MinerPickTimeout = 20

	OnlineConsensusValidityTime = 5
)


var (
	//随机数相关
	RandomConfig              = make(map[string]string, 0)   //man.json配置中读的
	RandomServiceName         = []string{}                   //子服务的名字
	RandomServicePlugs        = make(map[string][]string, 0) //子服务对应的插件名
	RandomServiceDefaultPlugs = make(map[string]string, 0)

	//选举相关
	//ElectPlugs string
)

func init() {
	RandomServiceName = []string{"electionseed", "everyblockseed", "everybroadcastseed"}
	//RandomServiceName = []string{"electionseed", "everyblockseed"}
	RandomServicePlugs[RandomServiceName[0]] = []string{"Minhash&Key"}
	RandomServicePlugs[RandomServiceName[1]] = []string{"Nonce&Address&Coinbase"}
	RandomServicePlugs[RandomServiceName[2]] = []string{"MaxNonce&Key"}

	RandomServiceDefaultPlugs[RandomServiceName[0]] = RandomServicePlugs[RandomServiceName[0]][0]
	RandomServiceDefaultPlugs[RandomServiceName[1]] = RandomServicePlugs[RandomServiceName[1]][0]
	RandomServiceDefaultPlugs[RandomServiceName[2]] = RandomServicePlugs[RandomServiceName[2]][0]
}

type NodeInfo struct {
	NodeID  discover.NodeID
	Address common.Address
}

var SuperVersionNodes = []NodeInfo{}
var SuperRollbackNodes = []NodeInfo{}

func Config_Init(Config_PATH string) {
	log.INFO("Config_Init 函数", "Config_PATH", Config_PATH)

	JsonParse := NewJsonStruct()
	v := Config{}
	JsonParse.Load(Config_PATH, &v)

	params.MainnetBootnodes = v.BootNode
	if len(params.MainnetBootnodes) <= 0 {
		fmt.Println("无bootnode节点")
		os.Exit(-1)
	}
	log.INFO("MainBootNode", "data", params.MainnetBootnodes)

	SuperVersionNodes = v.SuperVersion
	if len(SuperVersionNodes) <= 0 {
		fmt.Println("无版本超级节点")
		os.Exit(-1)
	}

	SuperRollbackNodes = v.SuperRollback
	if len(SuperRollbackNodes) <= 0 {
		fmt.Println("无回滚超级节点")
		os.Exit(-1)
	}
	RandomConfig = v.RandomConfig
	log.INFO("RandomConfig", "data", RandomConfig)
	//ElectPlugs = v.ElectPlugs
	//log.INFO("ElectPlugs", "data", ElectPlugs)
	//fmt.Println("echeloc",v.Echelon)
	//if len(v.Echelon) > 0 {
	//
	//	common.EchelonArrary = v.Echelon
	//}
	//log.INFO("EchelonArrary", "EchelonArrary", common.EchelonArrary)
}

type Config struct {
	BootNode       []string
	BroadNode      []NodeInfo
	InnerMinerNode []NodeInfo
	FoundationNode []NodeInfo
	SuperVersion   []NodeInfo
	SuperRollback  []NodeInfo
	RandomConfig   map[string]string
	ElectPlugs     string
	Echelon        []common.Echelon
}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("读取通用配置文件失败 err", err, "file", filename)
		os.Exit(-1)
		return
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		fmt.Println("通用配置文件数据获取失败 err", err)
		os.Exit(-1)
		return
	}
}
