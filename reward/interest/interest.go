package interest

import (
	"math/big"
	"sort"

	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/reward/util"

	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/log"
)

const (
	PackageName = "利息奖励"
	Denominator = 10000000
)

type interest struct {
	VIPConfig    []mc.VIPConfig
	CalcInterval uint64
	PayInterval  uint64
}

type DepositInterestRate struct {
	Deposit *big.Int
	Interst *big.Rat
}

type DepositInterestRateList []*DepositInterestRate

func (p DepositInterestRateList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p DepositInterestRateList) Len() int           { return len(p) }
func (p DepositInterestRateList) Less(i, j int) bool { return p[i].Deposit.Cmp(p[j].Deposit) < 0 }

func New(st util.StateDB) *interest {
	StateCfg, err := matrixstate.GetDataByState(mc.MSKeyInterestCfg, st)
	if nil != err {
		log.ERROR(PackageName, "获取状态树配置错误", "")
		return nil
	}

	if StateCfg.(*mc.InterestCfgStruct).PayInterval < StateCfg.(*mc.InterestCfgStruct).CalcInterval {
		log.ERROR(PackageName, "配置的发放周期小于计息周期", "")
		return nil
	}

	VipCfg, err := matrixstate.GetDataByState(mc.MSKeyVIPConfig, st)
	if nil != err {
		log.ERROR(PackageName, "获取状态树配置错误", "")
		return nil
	}
	Vip := VipCfg.([]mc.VIPConfig)
	if 0 == len(Vip) {
		log.ERROR(PackageName, "利率表为空", "")
		return nil
	}
	return &interest{Vip, StateCfg.(*mc.InterestCfgStruct).CalcInterval, StateCfg.(*mc.InterestCfgStruct).PayInterval}
}
func (tlr *interest) calcNodeInterest(deposit *big.Int, depositInterestRate []*DepositInterestRate) *big.Int {

	var blockInterest *big.Rat = nil
	for i, depositIntere := range depositInterestRate {
		if deposit.Cmp(big.NewInt(0)) <= 0 {
			log.ERROR(PackageName, "抵押获取错误", deposit)
			return big.NewInt(0)
		}
		if deposit.Cmp(depositIntere.Deposit) < 0 {
			blockInterest = depositInterestRate[i-1].Interst
			break
		}
	}
	if blockInterest == nil {
		blockInterest = depositInterestRate[len(depositInterestRate)-1].Interst
	}
	interstReward, _ := new(big.Rat).Mul(new(big.Rat).SetInt(deposit), blockInterest).Float64()
	bigval := new(big.Float)
	bigval.SetFloat64(interstReward)
	result := new(big.Int)
	bigval.Int(result)
	//log.INFO(PackageName, "calc interest reward  all reward", interstReward, "reward", result.String())
	return result
}

func (ic *interest) InterestCalc(state vm.StateDB, num uint64) {
	//todo:状态树读取利息计算的周期、支付的周期、利率
	if num == 1 {
		matrixstate.SetNumByState(mc.MSInterestCalcNum, state, num)
		matrixstate.SetNumByState(mc.MSInterestPayNum, state, num)
		log.INFO(PackageName, "初始化利息状态树高度", num)
		return
	}
	if nil == state {
		log.ERROR(PackageName, "状态树是空", state)
		return
	}
	calcInterestPeriod := ic.CalcInterval
	payInterestPeriod := ic.PayInterval

	if calcInterestPeriod == 0 || 0 == payInterestPeriod {
		log.ERROR(PackageName, "InterestPeriod is  error", "")
		return
	}
	ic.calcInterest(calcInterestPeriod, num, state)

	ic.payInterest(payInterestPeriod, num, state)
}

func (ic *interest) payInterest(payInterestPeriod uint64, num uint64, state vm.StateDB) {
	latestNum, err := matrixstate.GetNumByState(mc.MSInterestPayNum, state)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一计算利息高度错误", err)
		return
	}

	if latestNum >= ic.getLastInterestNumber(num-1, payInterestPeriod+1) {
		log.Info(PackageName, "当前惩罚已处理无须再处理", "")
		return
	}
	matrixstate.SetNumByState(mc.MSInterestPayNum, state, num)

	//1.获取所有利息转到抵押账户 2.清除所有利息
	log.INFO(PackageName, "将利息转到合约抵押账户,高度", num)

	AllInterestMap := depoistInfo.GetAllInterest(state)
	Deposit := depoistInfo.GetDeposit(state)
	if Deposit.Cmp(big.NewInt(0)) < 0 {
		log.WARN(PackageName, "利息合约账户余额非法", Deposit)
	}

	log.INFO(PackageName, "设置利息前的合约抵押账户余额", Deposit)
	for account, interest := range AllInterestMap {
		log.INFO(PackageName, "账户", account, "利息", interest.String())
		if interest.Cmp(big.NewInt(0)) <= 0 {
			log.ERROR(PackageName, "获取的利息非法", interest)
			continue
		}
		Deposit = new(big.Int).Add(Deposit, interest)
		depoistInfo.ResetInterest(state, account)
	}
	depoistInfo.SetDeposit(state, Deposit)
	readDeposit := depoistInfo.GetDeposit(state)
	log.INFO(PackageName, "设置利息后的合约抵押账户余额", readDeposit)

}

func (ic *interest) getLastInterestNumber(number uint64, InterestInterval uint64) uint64 {
	if number%InterestInterval == 0 {
		return number
	}
	ans := (number / InterestInterval) * InterestInterval
	return ans
}

func (ic *interest) calcInterest(calcInterestInterval uint64, num uint64, state vm.StateDB) {
	latestNum, err := matrixstate.GetNumByState(mc.MSInterestCalcNum, state)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一计算利息高度错误", err)
		return
	}

	if latestNum >= ic.getLastInterestNumber(num-1, calcInterestInterval+1) {
		log.Info(PackageName, "当前惩罚已处理无须再处理", "")
		return
	}
	matrixstate.SetNumByState(mc.MSInterestCalcNum, state, num)

	depositInterestRateList := make(DepositInterestRateList, 0)
	for _, v := range ic.VIPConfig {
		if v.MinMoney < 0 {
			log.ERROR(PackageName, "最小金额设置非法", "")
			return
		}
		deposit := new(big.Int).Mul(new(big.Int).SetUint64(v.MinMoney), util.ManPrice)
		depositInterestRateList = append(depositInterestRateList, &DepositInterestRate{deposit, big.NewRat(int64(v.InterestRate), Denominator)})
	}
	sort.Sort(depositInterestRateList)

	depositNodes, err := ca.GetElectedByHeight(new(big.Int).SetUint64(num - 1))
	if nil != err {
		log.ERROR(PackageName, "获取的抵押列表错误", err)
		return
	}
	if 0 == len(depositNodes) {
		log.ERROR(PackageName, "获取的抵押列表为空", "")
		return
	}
	log.INFO(PackageName, "计算利息,高度", num)
	for _, v := range depositNodes {

		result := ic.calcNodeInterest(v.Deposit, depositInterestRateList)
		if result.Cmp(big.NewInt(0)) <= 0 {
			log.ERROR(PackageName, "计算的利息非法", result)
			continue
		}
		depoistInfo.AddInterest(state, v.Address, result)
		log.INFO(PackageName, "账户", v.Address.String(), "deposit", v.Deposit.String(), "利息", result.String())
	}
	log.INFO(PackageName, "计算利息后", "")

}
