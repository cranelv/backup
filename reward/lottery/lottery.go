package lottery

import (
	"github.com/matrix/go-matrix/common/mt19937"
	"github.com/matrix/go-matrix/params"
	"math/big"

	"github.com/matrix/go-matrix/core/matrixstate"
	"github.com/matrix/go-matrix/mc"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params/manparams"
	"github.com/matrix/go-matrix/reward/util"
)

const (
	N           = 6
	FIRST       = 1 //一等奖数目
	SECOND      = 0 //二等奖数目
	THIRD       = 0 //三等奖数目
	PackageName = "彩票奖励"
)

var (
	FIRSTPRIZE   *big.Int = big.NewInt(6e+18) //一等奖金额  5man
	SENCONDPRIZE *big.Int = big.NewInt(3e+18) //二等奖金额 2man
	THIRDPRIZE   *big.Int = big.NewInt(1e+18) //三等奖金额 1man
)

type TxCmpResult struct {
	Tx        types.SelfTransaction
	CmpResult uint64
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type TxCmpResultList []TxCmpResult

func (p TxCmpResultList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p TxCmpResultList) Len() int           { return len(p) }
func (p TxCmpResultList) Less(i, j int) bool { return p[i].CmpResult < p[j].CmpResult }

type ChainReader interface {
	GetBlockByNumber(number uint64) *types.Block
	Config() *params.ChainConfig
}

type TxsLottery struct {
	chain      ChainReader
	seed       LotterySeed
	state      util.StateDB
	lotteryCfg *mc.LotteryCfgStruct
	bcInterval *manparams.BCInterval
}

type LotterySeed interface {
	GetRandom(hash common.Hash, Type string) (*big.Int, error)
}

func New(chain ChainReader, st util.StateDB, seed LotterySeed) *TxsLottery {
	intervalData, err := matrixstate.GetDataByState(mc.MSKeyBroadcastInterval, st)
	if err != nil {
		log.ERROR(PackageName, "获取广播周期失败", err)
		return nil
	}
	bcInterval, err := manparams.NewBCIntervalWithInterval(intervalData)
	if err != nil {
		log.ERROR(PackageName, "创建广播周期数据结构失败", err)
		return nil
	}

	lotteryCfg, err := matrixstate.GetDataByState(mc.MSKeyLotteryCfg, st)
	if nil != err {
		log.ERROR(PackageName, "获取状态树配置错误", "")
		return nil
	}

	cfg := lotteryCfg.(*mc.LotteryCfgStruct)
	if cfg.LotteryCalc == util.Stop {
		log.ERROR(PackageName, "停止发放彩票奖励", "")
		return nil
	}

	if len(cfg.LotteryInfo) == 0 {
		log.ERROR(PackageName, "没有配置彩票名额", "")
		return nil
	}
	tlr := &TxsLottery{
		chain:      chain,
		seed:       seed,
		state:      st,
		lotteryCfg: cfg,
		bcInterval: bcInterval,
	}

	return tlr
}
func abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

func (tlr *TxsLottery) LotteryCalc(parentHash common.Hash, num uint64) map[common.Address]*big.Int {
	//选举周期的最后时刻分配

	if !tlr.canChooseLottery(num) {
		return nil
	}

	txsCmpResultList := tlr.getLotteryList(parentHash, num, len(tlr.lotteryCfg.LotteryInfo))
	if 0 == len(txsCmpResultList) {
		log.ERROR(PackageName, "本周期没有交易不抽奖", "")
		return nil
	}

	LotteryAccount := make(map[common.Address]*big.Int, 0)
	tlr.lotteryChoose(txsCmpResultList, LotteryAccount)

	if 0 == len(LotteryAccount) {
		log.ERROR(PackageName, "抽奖结果为nil", "")
		return nil
	}
	return LotteryAccount
}

func (tlr *TxsLottery) canChooseLottery(num uint64) bool {
	if !tlr.ProcessMatrixState(num) {
		return false
	}

	balance := tlr.state.GetBalance(common.LotteryRewardAddress)
	if len(balance) == 0 {
		log.ERROR(PackageName, "状态树获取彩票账户余额错误", "")
		//return false
	}
	var allPrice uint64
	for _, v := range tlr.lotteryCfg.LotteryInfo {
		if v.PrizeMoney < 0 {
			log.ERROR(PackageName, "彩票奖励配置错误，金额", v.PrizeMoney, "奖项", v.PrizeLevel)
			return false
		}
		allPrice = allPrice + v.PrizeMoney*v.PrizeNum
	}
	if allPrice <= 0 {
		log.ERROR(PackageName, "总奖励不合法", allPrice)
		return false
	}
	if balance[common.MainAccount].Balance.Cmp(new(big.Int).Mul(new(big.Int).SetUint64(allPrice), util.ManPrice)) < 0 {
		log.ERROR(PackageName, "彩票账户余额不足，余额为", balance[common.MainAccount].Balance.String(), "总奖励", util.ManPrice)
		return false
	}
	return true
}

func (tlr *TxsLottery) ProcessMatrixState(num uint64) bool {
	if tlr.bcInterval.IsBroadcastNumber(num) {
		log.WARN(PackageName, "广播周期不处理", "")
		return false
	}
	latestNum, err := matrixstate.GetNumByState(mc.MSKEYLotteryNum, tlr.state)
	if nil != err {
		log.ERROR(PackageName, "状态树获取前一发放彩票高度错误", err)
		return false
	}
	if latestNum > tlr.bcInterval.GetLastReElectionNumber() {
		log.Info(PackageName, "当前彩票奖励已发放无须补发", "")
		return false
	}
	matrixstate.SetNumByState(mc.MSKEYLotteryNum, tlr.state, num)

	return true
}

func (tlr *TxsLottery) getLotteryList(parentHash common.Hash, num uint64, lotteryNum int) TxCmpResultList {
	originBlockNum := tlr.bcInterval.GetLastReElectionNumber() - tlr.bcInterval.GetReElectionInterval()

	if num < tlr.bcInterval.GetReElectionInterval() {
		originBlockNum = 0
	}
	randSeed, err := tlr.seed.GetRandom(parentHash, manparams.ElectionSeed)
	if nil != err {
		log.Error(PackageName, "获取随机数错误", err)
		return nil
	}

	log.INFO(PackageName, "随机数种子", randSeed.Int64())
	rand := mt19937.RandUniformInit(randSeed.Int64())
	txsCmpResultList := make(TxCmpResultList, 0)
	for originBlockNum < num {
		txs := tlr.chain.GetBlockByNumber(originBlockNum).Transactions()
		log.INFO(PackageName, "交易获取高度", originBlockNum)
		for _, tx := range txs {
			log.INFO(PackageName, "交易类型", tx.GetMatrixType())
			if tx.GetMatrixType() == common.ExtraNormalTxType {
				txCmpResult := TxCmpResult{tx, tx.Hash().Big().Uint64()}
				txsCmpResultList = append(txsCmpResultList, txCmpResult)
			}

		}
		originBlockNum++
	}
	if 0 == len(txsCmpResultList) {
		return nil
	}
	//sort.Sort(txsCmpResultList)
	chooseResultList := make(TxCmpResultList, 0)
	log.INFO(PackageName, "交易数目", len(txsCmpResultList))
	for i := 0; i < lotteryNum && i < len(txsCmpResultList); i++ {
		randomData := uint64(rand.Uniform(0, float64(^uint64(0))))
		log.INFO(PackageName, "随机数", randomData)
		index := randomData % (uint64(len(txsCmpResultList)))
		log.INFO(PackageName, "交易序号", index)
		chooseResultList = append(chooseResultList, txsCmpResultList[index])
	}

	return chooseResultList
}

func (tlr *TxsLottery) lotteryChoose(txsCmpResultList TxCmpResultList, LotteryMap map[common.Address]*big.Int) {

	sig := types.NewEIP155Signer(tlr.chain.Config().ChainId)
	RecordMap := make(map[uint8]uint64)
	for i := 0; i < len(tlr.lotteryCfg.LotteryInfo); i++ {
		RecordMap[uint8(i)] = 0
	}

	for _, txs := range txsCmpResultList {
		from, err := types.Sender(sig, txs.Tx)
		if nil != err {
			log.Error(PackageName, "解析交易from出错", err)
			continue
		}
		//抽取一等奖
		log.Info(PackageName, "解析交易from", from)
		for i := 0; i < len(tlr.lotteryCfg.LotteryInfo); i++ {
			prizeLevel := tlr.lotteryCfg.LotteryInfo[i].PrizeLevel
			prizeNum := tlr.lotteryCfg.LotteryInfo[i].PrizeNum
			prizeMoney := tlr.lotteryCfg.LotteryInfo[i].PrizeMoney
			if RecordMap[prizeLevel] < prizeNum {
				util.SetAccountRewards(LotteryMap, from, new(big.Int).Mul(new(big.Int).SetUint64(prizeMoney), util.ManPrice))
				RecordMap[prizeLevel]++
				break
			}
		}
		break
	}

}
