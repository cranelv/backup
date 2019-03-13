package blkreward

import (
	"reflect"
	"testing"

	"github.com/MatrixAINetwork/go-matrix/reward"
	"github.com/MatrixAINetwork/go-matrix/reward/util"
)

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
