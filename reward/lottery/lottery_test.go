package lottery

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/bouk/monkey"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
)

type Chain struct {
}

type randSeed struct {
}

func (r *randSeed) GetSeed(num uint64) *big.Int {

	return big.NewInt(1000)
}

type State struct {
	balance int64
}

func (st *State) GetBalance(addr common.Address) common.BalanceType {
	return []common.BalanceSlice{{common.MainAccount, big.NewInt(st.balance)}}
}

func (st *State) GetMatrixData(hash common.Hash) (val []byte) {
	return nil

}
func (st *State) SetMatrixData(hash common.Hash, val []byte) {
	return
}

func (chain *Chain) GetBlockByNumber(num uint64) *types.Block {
	header := &types.Header{}
	txs := make([]types.SelfTransaction, 0)
	if num == 298 {
		for i := 0; i < 3; i++ {

			tx := types.NewTransactions(uint64(i), common.Address{}, big.NewInt(100), 100, big.NewInt(int64(100)), nil, nil, 0, common.ExtraNormalTxType)
			addr := common.Address{}
			addr.SetString(strconv.Itoa(i))
			tx.SetFromLoad(addr)
			txs = append(txs, tx)

		}
	}

	return types.NewBlockWithTxs(header, txs)
}
func (chain *Chain) Config() *params.ChainConfig {
	return &params.ChainConfig{big.NewInt(1), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, nil}
}

func TestTxsLottery_LotteryCalc(t *testing.T) {
	log.InitLog(3)
	monkey.Patch(depoistInfo.GetOnlineTime, func(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
		fmt.Println("use monkey  ca.GetOnlineTime")
		onlineTime := big.NewInt(291)
		if stateDB == statedb {
			switch {
			case address.Equal(common.HexToAddress(account0)):
				onlineTime = big.NewInt(291 * 2) //100%
			case address.Equal(common.HexToAddress(account1)):
				onlineTime = big.NewInt(291) //0%
			case address.Equal(common.HexToAddress(account2)):
				onlineTime = big.NewInt(291 + 291/2) //50%
			case address.Equal(common.HexToAddress(account3)):
				onlineTime = big.NewInt(291 + 291/4) //25%

			}

		}

		return onlineTime, nil
	})

	lotterytest := New(&Chain{}, &State{0}, &randSeed{})
	lotterytest.LotteryCalc(299)
}

func TestTxsLottery_LotteryCalc2(t *testing.T) {
	log.InitLog(3)
	lotterytest := New(&Chain{}, &randSeed{})
	lotterytest.LotteryCalc(&State{-1}, 299)
}

func TestTxsLottery_LotteryCalc3(t *testing.T) {
	log.InitLog(3)
	lotterytest := New(&Chain{}, &randSeed{})
	lotterytest.LotteryCalc(&State{3e18}, 299)
}

func TestTxsLottery_LotteryCalc4(t *testing.T) {
	log.InitLog(3)
	lotterytest := New(&Chain{}, &randSeed{})
	lotterytest.LotteryCalc(&State{6e18}, 299)
}
