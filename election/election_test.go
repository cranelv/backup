// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package election

import (
	"math/big"
	"strconv"

	"testing"

	"fmt"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	//"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/baseinterface"
	_ "github.com/matrix/go-matrix/election/layered"
	_ "github.com/matrix/go-matrix/election/nochoice"
	_ "github.com/matrix/go-matrix/election/stock"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p/discover"
	//"encoding/json"
	"encoding/json"
	"github.com/matrix/go-matrix/log"
	"io/ioutil"
	"os"
)

func GetDepositDetatil(num int, m int, n int) []vm.DepositDetail {
	mList := []vm.DepositDetail{}
	for i := 0; i < num; i++ {
		temp := vm.DepositDetail{}
		temp.Address = common.BigToAddress(big.NewInt(int64(i)))

		if m > 0 {
			temp.Deposit = new(big.Int).Mul(big.NewInt(10000000), common.ManValue)
			m--
		} else if n > 0 {
			temp.Deposit = new(big.Int).Mul(big.NewInt(1000000), common.ManValue)
			n--
		} else {
			temp.Deposit = big.NewInt(int64(i))
		}

		temp.OnlineTime = big.NewInt(int64(i))
		temp.WithdrawH = big.NewInt(int64(i))

		tNodeID := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
		if i < 10 {
			tNodeID += "0"
		}
		tNodeID += strconv.Itoa(i)
		temp.NodeID, _ = discover.HexID(tNodeID)
		//fmt.Println("i", i, "err", err, len(tNodeID), "nodeId-string", tNodeID, "address-string", temp.Address.String())

		mList = append(mList, temp)

	}
	return mList
}
func MakeMinerTopReq(num int, Seed uint64) *mc.MasterMinerReElectionReqMsg {
	mList := GetDepositDetatil(num, 0, 0)

	ans := &mc.MasterMinerReElectionReqMsg{
		SeqNum:    Seed,
		RandSeed:  big.NewInt(int64(Seed)),
		MinerList: mList,
	}
	return ans
}

func MakeValidatorTopReq(num int, Seed uint64) *mc.MasterValidatorReElectionReqMsg {
	mList := GetDepositDetatil(num, 3, 2)

	ans := &mc.MasterValidatorReElectionReqMsg{
		SeqNum:        Seed,
		RandSeed:      big.NewInt(int64(Seed)),
		ValidatorList: mList,
		//	FoundationValidatoeList: []vm.DepositDetail{},
	}
	ans.ElectConfig = mc.ElectConfigInfo{
		ValidatorNum:  11,
		BackValidator: 5,
	}
	ans.VIPList = []mc.VIPConfig{

		mc.VIPConfig{
			MinMoney:     10000000,
			InterestRate: 100,
			ElectUserNum: 3,
			StockScale:   1000,
		},
		mc.VIPConfig{
			MinMoney:     1000000,
			InterestRate: 100,
			ElectUserNum: 3,
			StockScale:   1000,
		},
	}
	ans.ElectConfig.WhiteList = append(ans.ElectConfig.WhiteList, common.BigToAddress(big.NewInt(4)))
	ans.ElectConfig.BlackList = append(ans.ElectConfig.BlackList, common.BigToAddress(big.NewInt(5)))
	return ans

}
func GetFencengValidatorList(num int, Seed uint64, m int, n int) *mc.MasterValidatorReElectionReqMsg {
	mList := GetDepositDetatil(num, m, n)
	ans := &mc.MasterValidatorReElectionReqMsg{
		SeqNum:        Seed,
		RandSeed:      big.NewInt(int64(Seed)),
		ValidatorList: mList,
		//	FoundationValidatoeList: []vm.DepositDetail{},
	}
	return ans
}

func PrintMiner(miner *mc.MasterMinerReElectionRsp) {

	fmt.Println("MasterMiner")
	for _, v := range miner.MasterMiner {
		fmt.Println(v.Account, v.Position, v.Type, v.Stock)
	}
	fmt.Println("BackUpMiner")
	for _, v := range miner.BackUpMiner {
		fmt.Println(v.Account, v.Position, v.Type, v.Stock)
	}
	fmt.Println("\n\n\n\n")

}

func PrintValidator(validator *mc.MasterValidatorReElectionRsq) {

	fmt.Println("MasterValidator")
	for _, v := range validator.MasterValidator {
		fmt.Println(v.Account, v.Position, v.Type, v.Stock)
	}
	fmt.Println("BackupValidator")
	for _, v := range validator.BackUpValidator {
		fmt.Println(v.Account, v.Position, v.Type, v.Stock)
	}

	fmt.Println("CandidateValidator")
	for _, v := range validator.CandidateValidator {
		fmt.Println(v.Account, v.Position, v.Type, v.Stock)
	}
	fmt.Println("\n\n\n\n")
}

func TestUnit1(t *testing.T) {
	//矿工生成单元测试

	for Num := 20; Num <= 22; Num++ {
		for Key := 101; Key <= 105; Key++ {
			req := MakeMinerTopReq(Num, uint64(Key))
			fmt.Println("矿工备选列表个数", len(req.MinerList), "随机数", req.RandSeed)
			rspMiner := baseinterface.NewElect("layered").MinerTopGen(req)
			PrintMiner(rspMiner)
		}
	}

}

func TestUnit2(t *testing.T) {
	//验证者拓扑生成

	//股权方案-（10-12）
	log.InitLog(3)
	for Num := 20; Num <= 20; Num++ {
		for Key := 101; Key <= 101; Key++ {
			req := MakeValidatorTopReq(Num, uint64(Key))
			fmt.Println("验证者备选列表个数", len(req.ValidatorList), "随机数", req.RandSeed)

			rspValidator := baseinterface.NewElect("layerd").ValidatorTopGen(req)
			PrintValidator(rspValidator)
		}
	}
}

func TestUnit3(t *testing.T) {
	//验证者拓扑生成

	//股权方案-（15-18）
	for Num := 20; Num <= 20; Num++ {
		for Key := 101; Key <= 101; Key++ {
			req := MakeValidatorTopReq(Num, uint64(Key))
			fmt.Println("验证者备选列表个数", len(req.ValidatorList), "随机数", req.RandSeed)
			rspValidator := baseinterface.NewElect("layerd").ValidatorTopGen(req)
			PrintValidator(rspValidator)
		}
	}
}

//
//func TestUnit4(t *testing.T) {
//	//验证者拓扑生成
//	//不选方案-（10-12）
//	log.InitLog(3)
//	DefaultPlug = "nochoice"
//
//	for Num := 10; Num <= 12; Num++ {
//		for Key := 101; Key <= 105; Key++ {
//			req := MakeValidatorTopReq(Num, uint64(Key))
//			fmt.Println("验证者备选列表个数", len(req.ValidatorList), "随机数", req.RandSeed)
//			rspValidator := baseinterface.NewElect().ValidatorTopGen(req)
//			PrintValidator(rspValidator)
//		}
//	}
//}
//func TestUnit5(t *testing.T) {
//	//验证者拓扑生成
//	//不选方案-（15-18）
//	log.InitLog(3)
//	DefaultPlug = "nochoice"
//
//	for Num := 15; Num <= 18; Num++ {
//		for Key := 101; Key <= 105; Key++ {
//			req := MakeValidatorTopReq(Num, uint64(Key))
//			fmt.Println("验证者备选列表个数", len(req.ValidatorList), "随机数", req.RandSeed)
//			rspValidator := ValidatorTopGen(req)
//			PrintValidator(rspValidator)
//		}
//	}
//}
//
//func TestUnit6(t *testing.T) {
//	//验证者拓扑生成
//	//分层方案-（1000W 3个
//	// 			100W-1000W 3个）
//	log.InitLog(3)
//	DefaultPlug = "layered"
//	for Num := 10; Num <= 12; Num++ {
//		for key := 101; key <= 105; key++ {
//			req := GetFengcengValidatorList(Num, uint64(key), 3, 3)
//			fmt.Println("验证者备选列表个数", len(req.ValidatorList), "随机数", req.RandSeed, "1000W", 3, "100W", 3)
//			rsq := ValidatorTopGen(req)
//			PrintValidator(rsq)
//		}
//
//	}
//}
//
//func TestUnit7(t *testing.T) {
//	//验证者拓扑生成
//	//分层方案-（1000W 3个
//	// 			100W-1000W 3个）
//	log.InitLog(3)
//	DefaultPlug = "layered"
//	for Num := 15; Num <= 18; Num++ {
//		for key := 101; key <= 115; key++ {
//			req := GetFengcengValidatorList(Num, uint64(key), 3, 3)
//			fmt.Println("验证者备选列表个数", len(req.ValidatorList), "随机数", req.RandSeed, "1000W", 3, "100W", 3)
//			rsq := ValidatorTopGen(req)
//			PrintValidator(rsq)
//		}
//
//	}
//}
//
//func TestUnit8(t *testing.T) {
//	//验证者拓扑生成
//	//分层方案-（1000W 2个
//	// 			100W-1000W 2个）
//	log.InitLog(3)
//	DefaultPlug = "layered"
//	for Num := 10; Num <= 12; Num++ {
//		for key := 101; key <= 105; key++ {
//			req := GetFengcengValidatorList(Num, uint64(key), 2, 2)
//			fmt.Println("验证者备选列表个数", len(req.ValidatorList), "随机数", req.RandSeed, "1000W", 3, "100W", 3)
//			rsq := ValidatorTopGen(req)
//			PrintValidator(rsq)
//		}
//
//	}
//}
//
//func TestUnit9(t *testing.T) {
//	//验证者拓扑生成
//	//分层方案-（1000W 2个
//	// 			100W-1000W 2个）
//	log.InitLog(3)
//	DefaultPlug = "layered"
//	for Num := 15; Num <= 18; Num++ {
//		for key := 101; key <= 105; key++ {
//			req := GetFengcengValidatorList(Num, uint64(key), 2, 2)
//			fmt.Println("验证者备选列表个数", len(req.ValidatorList), "随机数", req.RandSeed, "1000W", 3, "100W", 3)
//			rsq := ValidatorTopGen(req)
//			PrintValidator(rsq)
//		}
//
//	}
//}
//*/
//func TestUnit10(t *testing.T) {
//	//验证者拓扑生成
//	//分层方案-（1000W 4个
//	// 			100W-1000W 4个）
//	tt := baseinterface.NewElect()
//	log.InitLog(3)
//	for Num := 13; Num <= 13; Num++ {
//		for key := 101; key <= 101; key++ {
//			req := GetFencengValidatorList(Num, uint64(key), 4, 4)
//			fmt.Println("验证者备选列表个数", len(req.ValidatorList), "随机数", req.RandSeed, "1000W", 4, "100W", 4)
//			for k,v:=range req.ValidatorList{
//				fmt.Println("k",k,"v",v)
//			}
//			rsq := tt.ValidatorTopGen(req)
//			PrintValidator(rsq)
//		}
//
//	}
//}
//
//
//func TestAAA(t *testing.T){
//	data,err:=json.Marshal(common.EchelonArrary)
//	fmt.Println("str data",string(data),"err",err)
//}

type MyDepos struct {
	Address    common.Address
	Deposit    *big.Int
	WithdrawH  *big.Int
	OnlineTime *big.Int
}
type JsonStruct struct {
	Type        string
	SeqNum      uint64
	RandSeed    *big.Int
	VM          []MyDepos
	ElectConfig mc.ElectConfigInfo
	VIPList     []mc.VIPConfig
}

func Test111(t *testing.T) {
	vipList := []mc.VIPConfig{
		mc.VIPConfig{
			MinMoney:     1000000,
			InterestRate: 1000, //(分母待定为1000w)
			ElectUserNum: 3,
			StockScale:   1000, //千分比
		},
		mc.VIPConfig{
			MinMoney:     5000000,
			InterestRate: 1700,
			ElectUserNum: 3,
			StockScale:   1700,
		},
	}
	vm := []MyDepos{
		MyDepos{
			Address:    common.BigToAddress(big.NewInt(1)),
			Deposit:    big.NewInt(1),
			WithdrawH:  big.NewInt(1),
			OnlineTime: big.NewInt(1),
		},
		MyDepos{
			Address:    common.BigToAddress(big.NewInt(2)),
			Deposit:    big.NewInt(2),
			WithdrawH:  big.NewInt(2),
			OnlineTime: big.NewInt(2),
		},
	}
	ans := JsonStruct{}
	ans.Type = "miner"
	ans.SeqNum = 1
	ans.RandSeed = big.NewInt(1)
	ans.VM = vm
	ans.ElectConfig = mc.ElectConfigInfo{
		MinerNum:      21,
		ValidatorNum:  11,
		BackValidator: 5,
		ElectPlug:     "layerd",
		WhiteList:     []common.Address{},
		BlackList:     []common.Address{},
	}
	ans.VIPList = vipList
	fmt.Println("ans", ans)
	data, err := json.Marshal(ans)
	fmt.Println("err", err, "data", data)
	fmt.Println("string(data)", string(data))

	filename := "./input.json"

	err = ioutil.WriteFile(filename, data, os.ModeAppend)
	if err != nil {
		log.Error("测试支持", "生成./input.json文件成功")
	}

}

type OutPut struct {
	Type      string
	SeqNum    uint64
	Master    []mc.ElectNodeInfo
	BackUp    []mc.ElectNodeInfo
	Candidate []mc.ElectNodeInfo
}

func TestOutPut(t *testing.T) {
	temp := []mc.ElectNodeInfo{
		mc.ElectNodeInfo{
			Account:  common.BigToAddress(big.NewInt(1)),
			Position: 0,
			Stock:    1,
			Type:     common.RoleMiner,
		},
		mc.ElectNodeInfo{
			Account:  common.BigToAddress(big.NewInt(2)),
			Position: 0,
			Stock:    1,
			Type:     common.RoleMiner,
		},
	}

	ans := OutPut{}
	ans.Type = "miner"
	ans.SeqNum = 1
	ans.Master = temp

	data, err := json.Marshal(ans)
	filename := "./output.json"

	err = ioutil.WriteFile(filename, data, os.ModeAppend)
	if err != nil {
		log.Error("测试支持", "生成./input.json文件成功")
	}

}
