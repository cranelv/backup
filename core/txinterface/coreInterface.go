package txinterface

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"math/big"
)

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	//FromFrontier() (common.Address, error)
	To() *common.Address
	GasFrom() common.Address
	AmontFrom() common.Address
	GasPrice() *big.Int
	Gas() uint64
	Value() *big.Int
	Hash() common.Hash
	Nonce() uint64
	CheckNonce() bool
	Data() []byte
	GetMatrixType() byte
	GetMatrix_EX() []types.Matrix_Extra //YYY  注释 Extra() 方法 改用此方法
	TxType() byte
	IsEntrustTx() bool
	GetCreateTime() uint32
	GetTxCurrency() string
}

type StateTransitioner interface {
	//InitStateTransition(evm *vm.EVM, msg Message, gp uint64)
	TransitionDb() (ret []byte, usedGas uint64, failed bool,shardings []uint, err error)
	To() common.Address
	UseGas(amount uint64) error
	BuyGas() error
	PreCheck() error
	RefundGas()
	GasUsed() uint64
	//CreateTransition(evm *vm.EVM, msg Message, gp uint64)StateTransitioner
}
