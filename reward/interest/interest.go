package interest

import (
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/reward/util"
	"math/big"
	"sort"
)
const (
	PackageName = "利息奖励"
	Denominator = 10000000
)

type interest struct {
	chain util.ChainReader
}

type DepositInterestRate struct {
	Deposit *big.Int
	Interst *big.Rat
}

type DepositInterestRateList []*DepositInterestRate
func (p DepositInterestRateList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p DepositInterestRateList) Len() int           { return len(p) }
func (p DepositInterestRateList) Less(i, j int) bool { return p[i].Deposit.Cmp(p[j].Deposit)<0  }

func New(chain util.ChainReader) *interest {

	return &interest{
		chain: chain,
	}
}
func (tlr *interest)calcNodeInterest(deposit *big.Int,depositInterestRate []*DepositInterestRate)*big.Int{

	var blockInterest *big.Rat = nil
	for i,depositInteres:= range depositInterestRate{
		if deposit.Cmp(depositInteres.Deposit)<0{
			blockInterest = depositInterestRate[i-1].Interst
			break
		}
	}
	if blockInterest==nil{
		blockInterest = depositInterestRate[len(depositInterestRate)-1].Interst
	}
	interstReward,_:= new(big.Rat).Mul(new(big.Rat).SetInt(deposit), blockInterest).Float64()
	bigval := new(big.Float)
	bigval.SetFloat64(interstReward)
	result := new(big.Int)
	bigval.Int(result)
	//log.INFO(PackageName, "calc interest reward  all reward", interstReward, "reward", result.String())
	return  result
}

func (ic *interest) InterestCalc(state *state.StateDB,num uint64){
	//todo:状态树读取利息计算的周期、支付的周期、利率
	calcInterestPeriod:=100
	payInterestPeriod:=3600

	depositInterestRateList := DepositInterestRateList{
		&DepositInterestRate{big.NewInt(0),big.NewRat(5,Denominator)},
		&DepositInterestRate{new(big.Int).Exp(big.NewInt(10), big.NewInt(21), big.NewInt(0)),big.NewRat(10,Denominator)},
		&DepositInterestRate{new(big.Int).Exp(big.NewInt(10), big.NewInt(22), big.NewInt(0)),big.NewRat(15,Denominator)},
		&DepositInterestRate{new(big.Int).Exp(big.NewInt(10), big.NewInt(23), big.NewInt(0)),big.NewRat(20,Denominator)},
	}
	sort.Sort(depositInterestRateList)
	//sort.Search()
	if calcInterestPeriod==0||0==payInterestPeriod{
		log.ERROR(PackageName,"InterestPeriod is  error","")
		return
	}

	if calcInterestPeriod==1||0==(num+1)%uint64(calcInterestPeriod){
		depositNodes, _ := ca.GetElectedByHeight(new(big.Int).SetUint64(num-1))
		for _, v := range depositNodes {

			result:=ic.calcNodeInterest(v.Deposit,depositInterestRateList)
			depoistInfo.AddReward(state,v.Address,result)
			log.INFO(PackageName, "calc interest reward  all reward account", v.Address.String(), "deposit",v.Deposit.String(),  "reward", result.String())
		}
	}

	if payInterestPeriod==1||0==(num+1)%uint64(payInterestPeriod){
     //1.获取所有利息转到抵押账户 2.清除所有利息
		log.INFO(PackageName, "将利息转到抵押账户", "")
	}
}