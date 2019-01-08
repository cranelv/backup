package matrixstate

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"testing"
)

type TestState struct {
	cache map[common.Hash][]byte
}

func newTestState() *TestState {
	return &TestState{
		cache: make(map[common.Hash][]byte),
	}
}

func (st *TestState) GetMatrixData(hash common.Hash) (val []byte) {
	data, exist := st.cache[hash]
	if exist {
		return data
	}
	return nil
}

func (st *TestState) SetMatrixData(hash common.Hash, val []byte) {
	st.cache[hash] = val
}

func Test_Manager(t *testing.T) {
	log.InitLog(3)

	st := newTestState()

	account1 := common.HexToAddress("0x12345")
	account2 := common.HexToAddress("0x543210")

	opt, _ := v1Manger.FindOperator(mc.OldMSKeyAccountBroadcast)
	opt.SetValue(st, account1)

	optV2, _ := v2Manger.FindOperator(mc.MSKeyAccountBroadcasts)
	optV2.SetValue(st, []common.Address{account2, account1})

	use_st(st)
}

func use_st(state *TestState) {
	opt, _ := v1Manger.FindOperator(mc.OldMSKeyAccountBroadcast)
	account, err := opt.GetValue(state)
	log.Info("old get", "account", account.(common.Address).Hex(), "err", err)

	optV2, _ := v2Manger.FindOperator(mc.MSKeyAccountBroadcasts)
	accounts, err := optV2.GetValue(state)
	log.Info("new get", "accounts", accounts.([]common.Address), "err", err)
}
