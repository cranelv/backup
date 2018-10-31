package txinterface

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"math/big"
	"github.com/matrix/go-matrix/core/types"
)
// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	//FromFrontier() (common.Address, error)
	To() *common.Address

	GasPrice() *big.Int
	Gas() uint64
	Value() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte
	//Extra() types.Matrix_Extra //YY
	GetMatrix_EX() []types.Matrix_Extra //YYY  注释 Extra() 方法 改用此方法
	TxType() types.TxTypeInt
}

type StateTransitioner interface {
	InitStateTransition(evm *vm.EVM, msg Message, gp uint64)
	TransitionDb() (ret []byte, usedGas uint64, failed bool, err error)
	To() common.Address
	UseGas(amount uint64) error
	BuyGas() error
	PreCheck() error
	RefundGas()
	GasUsed() uint64
}