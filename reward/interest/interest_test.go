package interest

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/consensus/manash"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/depoistInfo"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"math/big"
	"sync"
	"testing"

	"github.com/matrix/go-matrix/common"
	. "github.com/smartystreets/goconvey/convey"
)

type State struct {
	balance int64
}

func (st *State) GetBalance(addr common.Address) *big.Int {
	return big.NewInt(st.balance)
}

func (st *State) CreateAccount(common.Address) {

}

func (st *State) SubBalance(common.Address, *big.Int) {}
func (st *State) AddBalance(common.Address, *big.Int) {}

func (st *State) GetNonce(common.Address) uint64  { return 0 }
func (st *State) SetNonce(common.Address, uint64) {}

func (st *State) GetCodeHash(common.Address) common.Hash { return common.Hash{} }
func (st *State) GetCode(common.Address) []byte          { return nil }
func (st *State) SetCode(common.Address, []byte)         {}
func (st *State) GetCodeSize(common.Address) int         { return 0 }

func (st *State) AddRefund(uint64)  {}
func (st *State) GetRefund() uint64 { return 0 }

func (st *State) GetState(common.Address, common.Hash) common.Hash  { return common.Hash{} }
func (st *State) SetState(common.Address, common.Hash, common.Hash) {}

func (st *State) Suicide(common.Address) bool     { return true }
func (st *State) HasSuicided(common.Address) bool { return true }

// Exist reports whether the given account exists in state.
// Notably this should also return true for suicided accounts.
func (st *State) Exist(common.Address) bool { return true }

// Empty returns whether the given account is empty. Empty
// is defined according to EIP161 (balance = nonce = code = 0).
func (st *State) Empty(common.Address) bool { return true }

func (st *State) RevertToSnapshot(int) {}
func (st *State) Snapshot() int        { return 0 }

func (st *State) AddLog(*types.Log)               {}
func (st *State) AddPreimage(common.Hash, []byte) {}

func (st *State) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) {}
func Test_interest_Calc(t *testing.T) {
	Convey("利息测试计算利息", t, func() {
		log.InitLog(3)

		interestTest := New()

		interestTest.InterestCalc(&State{0}, 99)

	})

}

func Test_interest_Send(t *testing.T) {
	Convey("利息测试计算利息", t, func() {
		log.InitLog(3)

		interestTest := New()

		interestTest.InterestCalc(&State{0}, 3599)

	})

}

func Test_interest_number(t *testing.T) {
	Convey("利息测试计算利息", t, func() {
		log.InitLog(3)

		interestTest := New()

		interestTest.InterestCalc(&State{0}, 1)

	})

}

func Test_interest_number2(t *testing.T) {
	Convey("利息测试计算利息", t, func() {
		log.InitLog(3)

		interestTest := New()

		interestTest.InterestCalc(nil, 101)

	})

}

func Test_interest3(t *testing.T) {
	Convey("利息测试计算利息", t, func() {
		log.InitLog(3)
		monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)

			return Deposit, errors.New("获取抵押错误")
		})
		interestTest := New()

		interestTest.InterestCalc(&State{0}, 101)

	})

}
func Test_interest4(t *testing.T) {
	Convey("抵押列表长度为0", t, func() {
		log.InitLog(3)
		monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)

			return Deposit, nil
		})
		interestTest := New()

		interestTest.InterestCalc(&State{0}, 101)

	})

}

func Test_interest5(t *testing.T) {
	Convey("利息错误", t, func() {
		log.InitLog(3)
		monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: big.NewInt(-1)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: big.NewInt(3e+18)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: big.NewInt(4e+18)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: big.NewInt(2e+18)})
			return Deposit, nil
		})

		monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
			return nil
		})

		interestTest := New()

		interestTest.InterestCalc(&State{0}, 101)

	})

}

func Test_interest6(t *testing.T) {
	Convey("利息错误", t, func() {
		log.InitLog(3)
		monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: big.NewInt(-1)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: big.NewInt(3e+18)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: big.NewInt(4e+18)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: big.NewInt(2e+18)})
			return Deposit, nil
		})

		monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
			return nil
		})

		interestTest := New()

		interestTest.InterestCalc(&State{0}, 3601)

	})

}

func Test_interest6(t *testing.T) {
	Convey("利息错误", t, func() {
		log.InitLog(3)
		monkey.Patch(ca.GetElectedByHeight, func(height *big.Int) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: big.NewInt(-1)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: big.NewInt(3e+18)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: big.NewInt(4e+18)})
			Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: big.NewInt(2e+18)})
			return Deposit, nil
		})

		monkey.Patch(depoistInfo.AddInterest, func(stateDB vm.StateDB, address common.Address, reward *big.Int) error {
			return nil
		})

		interestTest := New()

		interestTest.InterestCalc(&State{0}, 3601)

	})

}

