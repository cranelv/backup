package election

import (
	"testing"

	"math/big"

	"fmt"

	"github.com/matrix/go-matrix/baseinterface"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	_ "github.com/matrix/go-matrix/election/layered"
	_ "github.com/matrix/go-matrix/election/stock"
	_ "github.com/matrix/go-matrix/election/nochoice"

	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/p2p/discover"
	"strconv"
	"github.com/matrix/go-matrix/log"
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
			temp.Deposit = new(big.Int).Mul(big.NewInt(100000), common.ManValue)
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

func MakeValidatorTopReq(num int, Seed uint64,vip1Num int,vip2Num int,white []common.Address,black []common.Address) *mc.MasterValidatorReElectionReqMsg {
	mList := GetDepositDetatil(num, vip1Num, vip2Num)

	ans := &mc.MasterValidatorReElectionReqMsg{
		SeqNum:        Seed,
		RandSeed:      big.NewInt(int64(Seed)),
		ValidatorList: mList,
		//	FoundationValidatoeList: []vm.DepositDetail{},
	}
	ans.ElectConfig = mc.ElectConfigInfo{
		ValidatorNum:  11,
		BackValidator: 5,
		WhiteList:white,
		BlackList:black,
	}
	ans.VIPList = []mc.VIPConfig{

		mc.VIPConfig{
			MinMoney:     10000000,
			InterestRate: 100,
			ElectUserNum: 5,
			StockScale:   1000,
		},
		mc.VIPConfig{
			MinMoney:     1000000,
			InterestRate: 100,
			ElectUserNum: 3,
			StockScale:   1000,
		},
	}
	return ans

}
func MakeMinerTopReq(num int, Seed uint64,vip1Num int,vip2Num int,white []common.Address,black []common.Address) *mc.MasterMinerReElectionReqMsg {
	mList := GetDepositDetatil(num, vip1Num, vip2Num)

	ans := &mc.MasterMinerReElectionReqMsg{
		SeqNum:        Seed,
		RandSeed:      big.NewInt(int64(Seed)),
	MinerList:mList,
	}
	ans.ElectConfig = mc.ElectConfigInfo{
		ValidatorNum:  11,
		BackValidator: 5,
		MinerNum:21,
		WhiteList:white,
		BlackList:black,
	}
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
	////矿工生成单元测试
	//
	//for Num := 20; Num <= 22; Num++ {
	//	for Key := 101; Key <= 105; Key++ {
	//		req := MakeMinerTopReq(Num, uint64(Key))
	//		fmt.Println("矿工备选列表个数", len(req.MinerList), "随机数", req.RandSeed)
	//		rspMiner := baseinterface.NewElect("layered").MinerTopGen(req)
	//		PrintMiner(rspMiner)
	//	}
	//}

}

func GOTestV(vip1Num int,vip2Num int,white []common.Address,black []common.Address,plug string){
	//验证者拓扑生成
	mapMaster:=make(map[common.Address]int,0)
	mapBackup:=make(map[common.Address]int,0)
	mapCand:=make(map[common.Address]int,0)
	//股权方案-（10-12）

	for Num := 50; Num <= 50; Num++ {
		for Key := 0; Key < 1; Key++ {
			req := MakeValidatorTopReq(Num, uint64(Key*2000 + 1),vip1Num,vip2Num,white,black)
			if Key==0{
				for _,v:=range req.ValidatorList{
					fmt.Println("账户",v.Address.String(),"NodeId",v.NodeID.String(),"抵押值",v.Deposit.String(),"在线时长",v.OnlineTime.String(),"withdraw",v.WithdrawH.String())
				}
			}

			rspValidator := baseinterface.NewElect(plug).ValidatorTopGen(req)
			for _,v:=range rspValidator.MasterValidator{
				mapMaster[v.Account]++
			}
			for _,v:=range rspValidator.BackUpValidator{
				mapBackup[v.Account]++
			}
			for _,v:=range rspValidator.CandidateValidator{
				mapCand[v.Account]++
			}
			//PrintValidator(rspValidator)
		}
	}

	ListAddr:=[]common.Address{}
	for i:=0;i<50;i++{
		ListAddr=append(ListAddr,common.BigToAddress(big.NewInt(int64(i))))
	}
	all:=0
	fmt.Println()
	for _,v:=range ListAddr{
		fmt.Println("账户",v.String(),"选择验证者次数",mapMaster[v],"选择备份验证者次数",mapBackup[v],"选择候选验证者次数",mapCand[v])
		all+=mapMaster[v]
		all+=mapBackup[v]
		all+=mapCand[v]
	}
	fmt.Println("所有节点被选中的总次数",all)
}


func GOTestM(vip1Num int,vip2Num int,white []common.Address,black []common.Address,plug string){
	//矿工拓扑生成
	mapMaster:=make(map[common.Address]int,0)


	for Num := 50; Num <= 50; Num++ {
		for Key := 0; Key < 1000; Key++ {
			req := MakeMinerTopReq(Num, uint64(Key*2000 + 1),vip1Num,vip2Num,white,black)
			//if Key==0{
			//	for _,v:=range req.ValidatorList{
			//		fmt.Println("账户",v.Address.String(),"NodeId",v.NodeID.String(),"抵押值",v.Deposit.String(),"在线时长",v.OnlineTime.String(),"withdraw",v.WithdrawH.String())
			//	}
			//}

			rspValidator := baseinterface.NewElect(plug).MinerTopGen(req)
			for _,v:=range rspValidator.MasterMiner{
				mapMaster[v.Account]++
			}

			//PrintValidator(rspValidator)
		}
	}

	ListAddr:=[]common.Address{}
	for i:=0;i<50;i++{
		ListAddr=append(ListAddr,common.BigToAddress(big.NewInt(int64(i))))
	}
	all:=0
	fmt.Println()
	for _,v:=range ListAddr{
		fmt.Println("账户",v.String(),"选择矿工次数",mapMaster[v])
		all+=mapMaster[v]

	}
	fmt.Println("所有节点被选中的总次数",all)
}


func TestUnit2(t *testing.T) {
	//不含黑白名单
	GOTestV(5,3,[]common.Address{},[]common.Address{},"layerd")
	GOTestV(4,4,[]common.Address{},[]common.Address{},"layerd")
	GOTestV(6,3,[]common.Address{},[]common.Address{},"layerd")
	GOTestV(0,0,[]common.Address{},[]common.Address{},"layerd")
}
func Test3(t *testing.T){
	//white:=[]common.Address{
	//	common.BigToAddress(big.NewInt(1)),
	//}
	//black:=[]common.Address{
	//	common.BigToAddress(big.NewInt(2)),
	//}
	//GOTest(0,0,white,black)


	//
	//white:=[]common.Address{
	//
	//}
	//black:=[]common.Address{
	//	common.BigToAddress(big.NewInt(2)),
	//}
	//GOTest(0,0,white,black)


	white:=[]common.Address{
		common.BigToAddress(big.NewInt(2)),
	}
	black:=[]common.Address{
		//common.BigToAddress(big.NewInt(2)),
	}
	GOTestV(0,0,white,black,"layerd")

}
func Test4(t *testing.T){
	GOTestM(0,0,[]common.Address{},[]common.Address{},"layerd")
}

func Test5(t *testing.T)  {
	log.InitLog(3)
	//GOTestV(0,0,[]common.Address{},[]common.Address{},"nochoice")
	//GOTestM(0,0,[]common.Address{},[]common.Address{},"nochoice")

	//GOTestV(0,0,[]common.Address{},[]common.Address{},"stock")
	//GOTestV(3,0,[]common.Address{},[]common.Address{},"stock")
	GOTestM(0,0,[]common.Address{},[]common.Address{},"stock")
}


