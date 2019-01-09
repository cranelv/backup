package interest

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
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
	Deposit  *big.Int
	Interest *big.Rat
}

type DepositInterestRateList []*DepositInterestRate

func (p DepositInterestRateList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p DepositInterestRateList) Len() int           { return len(p) }
func (p DepositInterestRateList) Less(i, j int) bool { return p[i].Deposit.Cmp(p[j].Deposit) < 0 }

func New(st util.StateDB) *interest {
	IC, err := matrixstate.GetInterestCfg(st)
	if nil != err {
		log.ERROR(PackageName, "获取利息状态树配置错误", "")
		return nil
	}
	if IC == nil {
		log.ERROR(PackageName, "利息配置", "配置为nil")
		return nil
	}
	if IC.InterestCalc == util.Stop {
		log.ERROR(PackageName, "停止发放", PackageName)
		return nil
	}
	if IC.PayInterval == 0 || 0 == IC.CalcInterval {
		log.ERROR(PackageName, "利息周期配置错误，支付周期", IC.PayInterval, "计算周期", IC.CalcInterval)
		return nil
	}
	if IC.PayInterval < IC.CalcInterval {
		log.ERROR(PackageName, "配置的发放周期小于计息周期，支付周期", IC.PayInterval, "计算周期", IC.CalcInterval)
		return nil
	}

	VipCfg, err := matrixstate.GetVIPConfig(st)
	if nil != err {
		log.ERROR(PackageName, "获取VIP状态树配置错误", "")
		return nil
	}
	if 0 == len(VipCfg) {
		log.ERROR(PackageName, "利率表为空", "")
		return nil
	}
	return &interest{VipCfg, IC.CalcInterval, IC.PayInterval}
}
func (tlr *interest) calcNodeInterest(deposit *big.Int, blockInterest *big.Rat) *big.Int {

	if deposit.Cmp(big.NewInt(0)) <= 0 {
		log.ERROR(PackageName, "抵押获取错误", deposit)
		return big.NewInt(0)
	}

	interstReward, _ := new(big.Rat).Mul(new(big.Rat).SetInt(deposit), blockInterest).Float64()
	bigval := new(big.Float)
	bigval.SetFloat64(interstReward)
	result := new(big.Int)
	bigval.Int(result)
	return result
}

func (ic *interest) PayInterest(state vm.StateDB, num uint64) map[common.Address]*big.Int {
	if !ic.canPayInterst(state, num, ic.PayInterval) {
		return nil
	}

	//1.获取所有利息转到抵押账户 2.清除所有利息
	log.Debug(PackageName, "发放利息,高度", num)

	AllInterestMap := depoistInfo.GetAllInterest(state)
	Deposit := big.NewInt(0)

	for account, originInterest := range AllInterestMap {
		if originInterest.Cmp(big.NewInt(0)) <= 0 {
			log.ERROR(PackageName, "获取的利息非法", originInterest)
			continue
		}
		slash, _ := depoistInfo.GetSlash(state, account)
		if slash.Cmp(big.NewInt(0)) < 0 {
			log.ERROR(PackageName, "获取的惩罚非法", originInterest)
			continue
		}

		finalInterest := new(big.Int).Sub(originInterest, slash)
		if finalInterest.Cmp(big.NewInt(0)) <= 0 {
			log.ERROR(PackageName, "支付的的利息非法", finalInterest)
			continue
		}
		log.Debug(PackageName, "账户", account, "原始利息", originInterest.String(), "惩罚利息", slash.String(), "剩余利息", finalInterest.String())
		AllInterestMap[account] = finalInterest
		Deposit = new(big.Int).Add(Deposit, finalInterest)
		depoistInfo.ResetSlash(state, account)
	}
	balance := state.GetBalance(common.InterestRewardAddress)
	log.Debug(PackageName, "设置利息前的账户余额", balance[common.MainAccount].Balance.String())
	if balance[common.MainAccount].Balance.Cmp(Deposit) < 0 {
		log.ERROR(PackageName, "利息账户余额不足，余额为", balance[common.MainAccount].Balance.String())
		return nil
	}
	AllInterestMap[common.ContractAddress] = Deposit
	return AllInterestMap
}

func (ic *interest) canPayInterst(state vm.StateDB, num uint64, payInterestPeriod uint64) bool {
	latestNum, err := matrixstate.GetInterestPayNum(state)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一计算利息高度错误", err)
		return false
	}
	if latestNum >= ic.getLastInterestNumber(num-1, payInterestPeriod)+1 {
		log.Debug(PackageName, "当前周期利息已支付无须再处理", "")
		return false
	}
	matrixstate.SetInterestPayNum(state, num)
	return true
}

func (ic *interest) getLastInterestNumber(number uint64, InterestInterval uint64) uint64 {
	if number%InterestInterval == 0 {
		return number
	}
	ans := (number / InterestInterval) * InterestInterval
	return ans
}

func (ic *interest) CalcInterest(state vm.StateDB, num uint64) map[common.Address]*big.Int {
	if !ic.canCalcInterest(state, num, ic.CalcInterval) {
		return nil
	}

	depositInterestRateList := make(DepositInterestRateList, 0)
	for _, v := range ic.VIPConfig {
		if v.MinMoney < 0 {
			log.ERROR(PackageName, "最小金额设置非法", "")
			return nil
		}
		deposit := new(big.Int).Mul(new(big.Int).SetUint64(v.MinMoney), util.ManPrice)
		depositInterestRateList = append(depositInterestRateList, &DepositInterestRate{deposit, big.NewRat(int64(v.InterestRate), Denominator)})
	}
	//sort.Sort(depositInterestRateList
	nonVipCfg := ic.VIPConfig[0]
	nonVipInterestRate := big.NewRat(int64(nonVipCfg.InterestRate), Denominator)
	depositNodes, err := ca.GetElectedByHeight(new(big.Int).SetUint64(num - 1))
	if nil != err {
		log.ERROR(PackageName, "获取的抵押列表错误", err)
		return nil
	}
	if 0 == len(depositNodes) {
		log.ERROR(PackageName, "获取的抵押列表为空", "")
		return nil
	}
	originElectNodes, err := matrixstate.GetElectGraph(state)
	if err != nil {
		log.Error(PackageName, "获取初选拓扑图错误", err)
		return nil
	}
	if originElectNodes == nil {
		log.Error(PackageName, "获取初选拓扑图", "结构为nil")
		return nil
	}
	if 0 == len(originElectNodes.ElectList) {
		log.Error(PackageName, "get获取初选列表为空", "")
		return nil
	}
	log.Debug(PackageName, "计算利息,高度", num)
	InterestMap := make(map[common.Address]*big.Int)
	for _, dv := range depositNodes {
		var interestRate *big.Rat = nil
		for _, ev := range originElectNodes.ElectList {
			if ev.Account.Equal(dv.Address) {
				interestRate = depositInterestRateList[ev.VIPLevel].Interest
			}
		}
		if nil == interestRate {
			interestRate = nonVipInterestRate
		}
		result := ic.calcNodeInterest(dv.Deposit, interestRate)
		if result.Cmp(big.NewInt(0)) <= 0 {
			log.ERROR(PackageName, "计算的利息非法", result)
			continue
		}
		depoistInfo.AddInterest(state, dv.Address, result)
		InterestMap[dv.Address] = result
		log.Debug(PackageName, "账户", dv.Address.String(), "deposit", dv.Deposit.String(), "利息", result.String())
	}
	return InterestMap
}

func (ic *interest) canCalcInterest(state vm.StateDB, num uint64, calcInterestInterval uint64) bool {
	latestNum, err := matrixstate.GetInterestCalcNum(state)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一计算利息高度错误", err)
		return false
	}
	if latestNum >= ic.getLastInterestNumber(num-1, calcInterestInterval)+1 {
		log.Info(PackageName, "当前利息已计算无须再处理", "")
		return false
	}
	matrixstate.SetInterestCalcNum(state, num)
	return true
}
