package matrixstate

import (
	"encoding/json"
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/pkg/errors"
)

/////////////////////////////////////////////////////////////////////////////////////////
// 区块奖励配置
type operatorBlkRewardCfg struct {
	key common.Hash
}

func newBlkRewardCfgOpt() *operatorBlkRewardCfg {
	return &operatorBlkRewardCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBlkRewardCfg),
	}
}

func (opt *operatorBlkRewardCfg) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBlkRewardCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "blkRewardCfg data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.BlkRewardCfg)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "blkRewardCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorBlkRewardCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.BlkRewardCfg)
	if !OK {
		log.Error(logInfo, "input param(blkRewardCfg) err", "reflect failed")
		return ErrParamReflect
	}
	if cfg == nil {
		log.Error(logInfo, "input param(blkRewardCfg) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "blkRewardCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 交易奖励配置
type operatorTxsRewardCfg struct {
	key common.Hash
}

func newTxsRewardCfgOpt() *operatorTxsRewardCfg {
	return &operatorTxsRewardCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyTxsRewardCfg),
	}
}

func (opt *operatorTxsRewardCfg) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorTxsRewardCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "txsRewardCfg data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.TxsRewardCfg)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "txsRewardCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorTxsRewardCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.TxsRewardCfg)
	if !OK {
		log.Error(logInfo, "input param(txsRewardCfg) err", "reflect failed")
		return ErrParamReflect
	}
	if cfg == nil {
		log.Error(logInfo, "input param(txsRewardCfg) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "txsRewardCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 利息配置
type operatorInterestCfg struct {
	key common.Hash
}

func newInterestCfgOpt() *operatorInterestCfg {
	return &operatorInterestCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyInterestCfg),
	}
}

func (opt *operatorInterestCfg) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorInterestCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "interestCfg data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.InterestCfg)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "interestCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorInterestCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.InterestCfg)
	if !OK {
		log.Error(logInfo, "input param(interestCfg) err", "reflect failed")
		return ErrParamReflect
	}
	if cfg == nil {
		log.Error(logInfo, "input param(interestCfg) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "interestCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 彩票配置
type operatorLotteryCfg struct {
	key common.Hash
}

func newLotteryCfgOpt() *operatorLotteryCfg {
	return &operatorLotteryCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyLotteryCfg),
	}
}

func (opt *operatorLotteryCfg) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorLotteryCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "lotteryCfg data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.LotteryCfg)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "lotteryCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorLotteryCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.LotteryCfg)
	if !OK {
		log.Error(logInfo, "input param(lotteryCfg) err", "reflect failed")
		return ErrParamReflect
	}
	if cfg == nil {
		log.Error(logInfo, "input param(lotteryCfg) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "lotteryCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 惩罚配置
type operatorSlashCfg struct {
	key common.Hash
}

func newSlashCfgOpt() *operatorSlashCfg {
	return &operatorSlashCfg{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeySlashCfg),
	}
}

func (opt *operatorSlashCfg) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorSlashCfg) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		log.Error(logInfo, "slashCfg data", "is empty")
		return nil, ErrDataEmpty
	}

	value := new(mc.SlashCfg)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "slashCfg unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorSlashCfg) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	cfg, OK := value.(*mc.SlashCfg)
	if !OK {
		log.Error(logInfo, "input param(slashCfg) err", "reflect failed")
		return ErrParamReflect
	}
	if cfg == nil {
		log.Error(logInfo, "input param(slashCfg) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Error(logInfo, "slashCfg marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 上一矿工区块奖励金额
type operatorPreMinerBlkReward struct {
	key common.Hash
}

func newPreMinerBlkRewardOpt() *operatorPreMinerBlkReward {
	return &operatorPreMinerBlkReward{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyPreMinerBlkReward),
	}
}

func (opt *operatorPreMinerBlkReward) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorPreMinerBlkReward) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.MinerOutReward{Reward: *big.NewInt(0)}, nil
	}

	value := new(mc.MinerOutReward)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "preMinerBlkReward unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorPreMinerBlkReward) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	reward, OK := value.(*mc.MinerOutReward)
	if !OK {
		log.Error(logInfo, "input param(preMinerBlkReward) err", "reflect failed")
		return ErrParamReflect
	}
	if reward == nil {
		log.Error(logInfo, "input param(preMinerBlkReward) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(reward)
	if err != nil {
		log.Error(logInfo, "preMinerBlkReward marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 上一矿工交易奖励金额
type operatorPreMinerTxsReward struct {
	key common.Hash
}

func newPreMinerTxsRewardOpt() *operatorPreMinerTxsReward {
	return &operatorPreMinerTxsReward{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyPreMinerTxsReward),
	}
}

func (opt *operatorPreMinerTxsReward) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorPreMinerTxsReward) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.MinerOutReward{Reward: *big.NewInt(0)}, nil
	}

	value := new(mc.MinerOutReward)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "preMinerTxsReward unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorPreMinerTxsReward) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	reward, OK := value.(*mc.MinerOutReward)
	if !OK {
		log.Error(logInfo, "input param(preMinerTxsReward) err", "reflect failed")
		return ErrParamReflect
	}
	if reward == nil {
		log.Error(logInfo, "input param(preMinerTxsReward) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(reward)
	if err != nil {
		log.Error(logInfo, "preMinerTxsReward marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// upTime状态
type operatorUpTimeNum struct {
	key common.Hash
}

func newUpTimeNumOpt() *operatorUpTimeNum {
	return &operatorUpTimeNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyUpTimeNum),
	}
}

func (opt *operatorUpTimeNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorUpTimeNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return uint64(0), nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "upTimeNum decode failed", err)
		return uint64(0), err
	}
	return num, nil
}

func (opt *operatorUpTimeNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	num, OK := value.(uint64)
	if !OK {
		log.Error(logInfo, "input param(upTimeNum) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, encodeUint64(num))
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 彩票状态
type operatorLotteryNum struct {
	key common.Hash
}

func newLotteryNumOpt() *operatorLotteryNum {
	return &operatorLotteryNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyLotteryNum),
	}
}

func (opt *operatorLotteryNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorLotteryNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return uint64(0), nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "lotteryNum decode failed", err)
		return uint64(0), err
	}
	return num, nil
}

func (opt *operatorLotteryNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	num, OK := value.(uint64)
	if !OK {
		log.Error(logInfo, "input param(lotteryNum) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, encodeUint64(num))
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 彩票候选账户
type operatorLotteryAccount struct {
	key common.Hash
}

func newLotteryAccountOpt() *operatorLotteryAccount {
	return &operatorLotteryAccount{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyLotteryAccount),
	}
}

func (opt *operatorLotteryAccount) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorLotteryAccount) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return &mc.LotteryFrom{From: make([]common.Address, 0)}, nil
	}

	value := new(mc.LotteryFrom)
	err := json.Unmarshal(data, &value)
	if err != nil {
		log.Error(logInfo, "lotteryAccount unmarshal failed", err)
		return nil, err
	}
	return value, nil
}

func (opt *operatorLotteryAccount) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	accounts, OK := value.(*mc.LotteryFrom)
	if !OK {
		log.Error(logInfo, "input param(lotteryAccount) err", "reflect failed")
		return ErrParamReflect
	}
	if accounts == nil {
		log.Error(logInfo, "input param(lotteryAccount) err", "cfg is nil")
		return ErrParamNil
	}
	data, err := json.Marshal(accounts)
	if err != nil {
		log.Error(logInfo, "lotteryAccount marshal failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 利息计算状态
type operatorInterestCalcNum struct {
	key common.Hash
}

func newInterestCalcNumOpt() *operatorInterestCalcNum {
	return &operatorInterestCalcNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyInterestCalcNum),
	}
}

func (opt *operatorInterestCalcNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorInterestCalcNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return uint64(0), nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "interestCalcNum decode failed", err)
		return uint64(0), err
	}
	return num, nil
}

func (opt *operatorInterestCalcNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	num, OK := value.(uint64)
	if !OK {
		log.Error(logInfo, "input param(interestCalcNum) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, encodeUint64(num))
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 利息支付状态
type operatorInterestPayNum struct {
	key common.Hash
}

func newInterestPayNumOpt() *operatorInterestPayNum {
	return &operatorInterestPayNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyInterestPayNum),
	}
}

func (opt *operatorInterestPayNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorInterestPayNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return uint64(0), nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "interestPayNum decode failed", err)
		return uint64(0), err
	}
	return num, nil
}

func (opt *operatorInterestPayNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	num, OK := value.(uint64)
	if !OK {
		log.Error(logInfo, "input param(interestPayNum) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, encodeUint64(num))
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 惩罚状态
type operatorSlashNum struct {
	key common.Hash
}

func newSlashNumOpt() *operatorSlashNum {
	return &operatorSlashNum{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeySlashNum),
	}
}

func (opt *operatorSlashNum) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorSlashNum) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return uint64(0), nil
	}
	num, err := decodeUint64(data)
	if err != nil {
		log.Error(logInfo, "slashNum decode failed", err)
		return uint64(0), err
	}
	return num, nil
}

func (opt *operatorSlashNum) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	num, OK := value.(uint64)
	if !OK {
		log.Error(logInfo, "input param(slashNum) err", "reflect failed")
		return ErrParamReflect
	}
	st.SetMatrixData(opt.key, encodeUint64(num))
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 固定区块算法配置
type operatorBlkCalc struct {
	key common.Hash
}

func newBlkCalcOpt() *operatorBlkCalc {
	return &operatorBlkCalc{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyBlkCalc),
	}
}

func (opt *operatorBlkCalc) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorBlkCalc) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}
	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "BlkCalc decode failed", err)
		return nil, err
	}
	return calc, nil
}

func (opt *operatorBlkCalc) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(BlkCalc) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "BlkCalc encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 交易费算法配置
type operatorTxsCalc struct {
	key common.Hash
}

func newTxsCalcOpt() *operatorTxsCalc {
	return &operatorTxsCalc{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyTxsCalc),
	}
}

func (opt *operatorTxsCalc) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorTxsCalc) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}
	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "TxsCalc decode failed", err)
		return nil, err
	}
	return calc, nil
}

func (opt *operatorTxsCalc) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(TxsCalc) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "TxsCalc encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 利息算法配置
type operatorInterestCalc struct {
	key common.Hash
}

func newInterestCalcOpt() *operatorInterestCalc {
	return &operatorInterestCalc{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyInterestCalc),
	}
}

func (opt *operatorInterestCalc) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorInterestCalc) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}
	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "InterestCalc decode failed", err)
		return nil, err
	}
	return calc, nil
}

func (opt *operatorInterestCalc) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(InterestCalc) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "InterestCalc encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 彩票算法配置
type operatorLotteryCalc struct {
	key common.Hash
}

func newLotteryCalcOpt() *operatorLotteryCalc {
	return &operatorLotteryCalc{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeyLotteryCalc),
	}
}

func (opt *operatorLotteryCalc) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorLotteryCalc) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}
	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "LotteryCalc decode failed", err)
		return nil, err
	}
	return calc, nil
}

func (opt *operatorLotteryCalc) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(LotteryCalc) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "LotteryCalc encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 惩罚算法配置
type operatorSlashCalc struct {
	key common.Hash
}

func newSlashCalcOpt() *operatorSlashCalc {
	return &operatorSlashCalc{
		key: types.RlpHash(matrixStatePrefix + mc.MSKeySlashCalc),
	}
}

func (opt *operatorSlashCalc) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorSlashCalc) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}
	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "SlashCalc decode failed", err)
		return nil, err
	}
	return calc, nil
}

func (opt *operatorSlashCalc) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(SlashCalc) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "SlashCalc encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//入池gas门限
type operatorTxpoolGasLimit struct {
	key common.Hash
}

func newTxpoolGasLimitOpt() *operatorTxpoolGasLimit {
	return &operatorTxpoolGasLimit{
		key: types.RlpHash(matrixStatePrefix + mc.MSTxpoolGasLimitCfg),
	}
}

func (opt *operatorTxpoolGasLimit) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorTxpoolGasLimit) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}

	msg := new(big.Int)
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.Errorf("json.Unmarshal failed: %s", err)
	}

	return msg, nil
}

func (opt *operatorTxpoolGasLimit) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(big.Int)
	if !OK {
		log.Error(logInfo, "input param(TxpoolGasLimit) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := json.Marshal(data)
	if err != nil {
		log.Error(logInfo, "TxpoolGasLimit encode failed", err)
		return err
	}

	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
//币种打包限制
type operatorCurrencyPack struct {
	key common.Hash
}

func newCurrencyPackOpt() *operatorCurrencyPack {
	return &operatorCurrencyPack{
		key: types.RlpHash(matrixStatePrefix + mc.MSCurrencyPack),
	}
}

func (opt *operatorCurrencyPack) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorCurrencyPack) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return uint64(0), err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return "0", nil
	}

	calc, err := decodeString(data)
	if err != nil {
		log.Error(logInfo, "CurrencyPack decode failed", err)
		return nil, err
	}

	return calc, nil
}

func (opt *operatorCurrencyPack) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	data, OK := value.(string)
	if !OK {
		log.Error(logInfo, "input param(CurrencyPack) err", "reflect failed")
		return ErrParamReflect
	}
	encodeData, err := encodeString(data)
	if err != nil {
		log.Error(logInfo, "CurrencyPack encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, encodeData)
	return nil
}

/////////////////////////////////////////////////////////////////////////////////////////
// 账户黑名单
type operatorAccountBlackList struct {
	key common.Hash
}

func newAccountBlackListOpt() *operatorAccountBlackList {
	return &operatorAccountBlackList{
		key: types.RlpHash(matrixStatePrefix + mc.MSAccountBlackList),
	}
}

func (opt *operatorAccountBlackList) KeyHash() common.Hash {
	return opt.key
}

func (opt *operatorAccountBlackList) GetValue(st StateDB) (interface{}, error) {
	if err := checkStateDB(st); err != nil {
		return nil, err
	}

	data := st.GetMatrixData(opt.key)
	if len(data) == 0 {
		return make([]common.Address, 0), nil
	}
	accounts, err := decodeAccounts(data)
	if err != nil {
		log.Error(logInfo, "AccountBlackList decode failed", err)
		return nil, err
	}
	return accounts, nil
}

func (opt *operatorAccountBlackList) SetValue(st StateDB, value interface{}) error {
	if err := checkStateDB(st); err != nil {
		return err
	}

	accounts, OK := value.([]common.Address)
	if !OK {
		log.Error(logInfo, "input param(AccountBlackList) err", "reflect failed")
		return ErrParamReflect
	}
	if len(accounts) == 0 {
		log.Error(logInfo, "input param(AccountBlackList) err", "accounts is empty")
		return ErrParamReflect
	}
	data, err := encodeAccounts(accounts)
	if err != nil {
		log.Error(logInfo, "AccountBlackList encode failed", err)
		return err
	}
	st.SetMatrixData(opt.key, data)
	return nil
}