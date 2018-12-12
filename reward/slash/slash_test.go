package slash

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/mandb"
	"github.com/matrix/go-matrix/params"

	"github.com/matrix/go-matrix/core/state"

	"github.com/matrix/go-matrix/depoistInfo"

	"bou.ke/monkey"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	. "github.com/smartystreets/goconvey/convey"
)

const account0 = "0x475baee143cf541ff3ee7b00c1c933129238d793"
const account1 = "0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"
const account2 = "0x519437b21e2a0b62788ab9235d0728dd7f1a7269"
const account3 = "0x29216818d3788c2505a593cbbb248907d47d9bce"

type Chain struct {
}

func (chain *Chain) Config() *params.ChainConfig {
	return nil
}

// CurrentHeader retrieves the current header from the local chain.
func (chain *Chain) CurrentHeader() *types.Header {

	return nil
}

// GetHeader retrieves a block header from the database by hash and number.
func (chain *Chain) GetHeader(hash common.Hash, number uint64) *types.Header {
	return nil
}

// GetHeaderByHash retrieves a block header from the database by its hash.
func (chain *Chain) GetHeaderByHash(hash common.Hash) *types.Header {
	return nil
}

func (chain *Chain) GetBlockByNumber(number uint64) *types.Block {
	return nil
}

// GetBlock retrieves a block sfrom the database by hash and number.
func (chain *Chain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return nil
}
func (chain *Chain) StateAt(root common.Hash) (*state.StateDB, error) {

	return nil, nil
}
func (chain *Chain) State() (*state.StateDB, error) {

	return nil, nil
}
func (chain *Chain) NewTopologyGraph(header *types.Header) (*mc.TopologyGraph, error) {
	return nil, nil
}

func (chain *Chain) GetHeaderByNumber(number uint64) *types.Header {
	header := &types.Header{
		Coinbase: common.Address{},
	}
	//txs := make([]types.SelfTransaction, 0)

	return header
}
func TestBlockSlash_CalcSlash(t *testing.T) {
	log.InitLog(3)

	slash := New(&Chain{})
	Convey("计算惩罚", t, func() {
		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
			return nil
		})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
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

		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		slash.CalcSlash(statedb, common.GetReElectionInterval()-1)
	})
}

func TestBlockSlash_CalcSlash22(t *testing.T) {
	log.InitLog(3)

	slash := New(&Chain{})
	Convey("计算惩罚", t, func() {
		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
			return nil
		})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
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
		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
			fmt.Println("use monkey  ca.GetTopologyByNumber")
			newGraph := &mc.TopologyGraph{
				Number:        number,
				NodeList:      make([]mc.TopologyNodeInfo, 0),
				CurNodeNumber: 0,
			}
			if common.RoleValidator == reqTypes&common.RoleValidator {
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
				newGraph.CurNodeNumber = 4
			}

			return newGraph, errors.New("拓扑非法")
		})
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		slash.CalcSlash(statedb, common.GetReElectionInterval()+1)
	})
}

func TestBlockSlash_CalcSlash33(t *testing.T) {
	log.InitLog(3)

	slash := New(&Chain{})
	Convey("计算惩罚", t, func() {
		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
			return nil
		})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
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
		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
			fmt.Println("use monkey  ca.GetTopologyByNumber")
			newGraph := &mc.TopologyGraph{
				Number:        number,
				NodeList:      make([]mc.TopologyNodeInfo, 0),
				CurNodeNumber: 0,
			}

			return newGraph, nil
		})
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		slash.CalcSlash(statedb, common.GetReElectionInterval()+1)
	})
}

func TestBlockSlash_CalcSlash44(t *testing.T) {
	log.InitLog(3)

	slash := New(&Chain{})
	Convey("计算惩罚", t, func() {
		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
			return nil
		})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
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
		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
			fmt.Println("use monkey  ca.GetTopologyByNumber")
			newGraph := &mc.TopologyGraph{
				Number:        number,
				NodeList:      make([]mc.TopologyNodeInfo, 0),
				CurNodeNumber: 0,
			}
			if common.RoleValidator == reqTypes&common.RoleValidator {
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
				newGraph.CurNodeNumber = 4
			}

			return newGraph, nil
		})
		monkey.Patch(depoistInfo.GetInterest, func(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
			fmt.Println("use monkey  ca.GetInterest")

			return nil, errors.New("利息非法")
		})
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		slash.CalcSlash(statedb, common.GetReElectionInterval()+1)
	})
}

func TestBlockSlash_CalcSlash55(t *testing.T) {
	log.InitLog(3)

	slash := New(&Chain{})
	Convey("计算惩罚", t, func() {
		monkey.Patch(depoistInfo.AddSlash, func(stateDB vm.StateDB, address common.Address, slash *big.Int) error {
			fmt.Println("use monkey  depoistInfo.AddSlash", "address", address.String(), "slash", slash.Uint64())
			return nil
		})
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(mandb.NewMemDatabase()))
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
		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
			fmt.Println("use monkey  ca.GetTopologyByNumber")
			newGraph := &mc.TopologyGraph{
				Number:        number,
				NodeList:      make([]mc.TopologyNodeInfo, 0),
				CurNodeNumber: 0,
			}
			if common.RoleValidator == reqTypes&common.RoleValidator {
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
				newGraph.CurNodeNumber = 4
			}

			return newGraph, nil
		})
		monkey.Patch(depoistInfo.GetInterest, func(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
			fmt.Println("use monkey  ca.GetInterest")

			return big.NewInt(100), nil
		})

		monkey.Patch(depoistInfo.GetOnlineTime, func(stateDB vm.StateDB, address common.Address) (*big.Int, error) {
			fmt.Println("use monkey  ca.GetInterest")

			return nil, errors.New("利息非法")
		})
		rewards := make(map[common.Address]*big.Int, 0)
		rewards[common.HexToAddress(account0)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account1)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account2)] = big.NewInt(4e+18)
		rewards[common.HexToAddress(account3)] = big.NewInt(4e+18)
		slash.CalcSlash(statedb, common.GetReElectionInterval()+1)
	})
}
