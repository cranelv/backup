package blkreward

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/core/state"

	"github.com/MatrixAINetwork/go-matrix/params"

	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/reward"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

type Chain struct {
	blockCache map[uint64]*types.Block
}

func (chain *Chain) GetHeaderByHash(hash common.Hash) *types.Header {
	header := &types.Header{}
	return header
}
func (chain *Chain) StateAt(root []common.CoinRoot) (*state.StateDBManage, error) {

}
func (chain *Chain) Config() *params.ChainConfig {
	return &params.ChainConfig{ChainId: big.NewInt(1), EIP155Block: big.NewInt(2), HomesteadBlock: new(big.Int)}
}
func TestNew(t *testing.T) {
	type args struct {
		chain  util.ChainReader
		st     util.StateDB
		preSt  util.StateDB
		ppreSt util.StateDB
	}
	tests := []struct {
		name string
		args args
		want reward.Reward
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.chain, tt.args.st, tt.args.preSt, tt.args.ppreSt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
