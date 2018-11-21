package interest

import (
	"math/big"
	"testing"

	"github.com/matrix/go-matrix/reward/util"
)

func Test_interest_LotteryCalc(t *testing.T) {
	type fields struct {
		chain util.ChainReader
	}
	type args struct {
		num uint64
	}
	val :=10000000000.0
	bigval := new(big.Float)
	bigval.SetFloat64(val)
	// Set precision if required.
	// bigval.SetPrec(64)

	coin := new(big.Float)
	coin.SetInt(big.NewInt(1000000000000000000))

	bigval.Mul(bigval, coin)

	result := new(big.Int).SetUint64(1)
	bigval.Int(result) // store converted number in result

	println("result%v",result)


}
