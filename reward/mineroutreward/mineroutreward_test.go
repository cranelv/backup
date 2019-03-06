package mineroutreward

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/core/matrixstate"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
	"github.com/MatrixAINetwork/go-matrix/reward/util"

	"github.com/MatrixAINetwork/go-matrix/core/state"
	"github.com/MatrixAINetwork/go-matrix/mandb"
)

func TestSetPreMinerReward(t *testing.T) {
	chaindb := mandb.NewMemDatabase()
	state, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))

	matrixstate.SetVersionInfo(state, manparams.VersionBeta)
	_, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for i := 0; i < 100; i++ {
		SetPreMinerReward(state, big.NewInt(int64(i)), util.TxsReward, strconv.Itoa(i))
	}
	out2, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	if 100 != len(out2) {

	}
	for _, v := range out2 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
	for i := 100; i < 200; i++ {
		SetPreMinerReward(state, big.NewInt(int64(i)), util.TxsReward, strconv.Itoa(i))
	}
	out3, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for _, v := range out3 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
}
func TestModifyPreMinerReward(t *testing.T) {
	chaindb := mandb.NewMemDatabase()
	state, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))

	matrixstate.SetVersionInfo(state, manparams.VersionBeta)
	_, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for i := 0; i < 100; i++ {
		SetPreMinerReward(state, big.NewInt(int64(i)), util.TxsReward, strconv.Itoa(i))
	}
	out2, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	if 100 != len(out2) {

	}
	for _, v := range out2 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
	for i := 0; i < 200; i = i + 2 {
		SetPreMinerReward(state, big.NewInt(int64(i*1000)), util.TxsReward, strconv.Itoa(i))
	}
	out3, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for _, v := range out3 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
}

func TestModifyVersionReward(t *testing.T) {
	chaindb := mandb.NewMemDatabase()
	state, _ := state.NewStateDBManage(nil, chaindb, state.NewDatabase(chaindb))

	matrixstate.SetVersionInfo(state, manparams.VersionAlpha)
	_, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for i := 0; i < 100; i++ {
		SetPreMinerReward(state, big.NewInt(int64(i)), util.TxsReward, strconv.Itoa(i))
	}
	out2, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	if 1 != len(out2) {

	}

	for _, v := range out2 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
	matrixstate.SetVersionInfo(state, manparams.VersionBeta)
	for i := 0; i < 200; i = i + 2 {
		SetPreMinerReward(state, big.NewInt(int64(i*1000)), util.TxsReward, strconv.Itoa(i))
	}
	out3, err := util.GetPreMinerReward(state, util.TxsReward)
	if nil != err {
		t.Error("fail")
	}
	for _, v := range out3 {
		fmt.Println("币种", v.CoinType, "奖励", v.Reward.String())
	}
}
