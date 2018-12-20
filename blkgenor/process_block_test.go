package blkgenor

//
//import (
//	"fmt"
//	"math/big"
//	"testing"
//
//	"bou.ke/monkey"
//	"github.com/matrix/go-matrix/ca"
//
//	"github.com/matrix/go-matrix/core/types"
//	"github.com/matrix/go-matrix/core/vm"
//
//	"github.com/matrix/go-matrix/common"
//	"github.com/matrix/go-matrix/log"
//	"github.com/matrix/go-matrix/mc"
//	. "github.com/smartystreets/goconvey/convey"
//)
//
//func TestProcess_LeaderInsertAndBcBlock(t *testing.T) {
//	log.InitLog(3)
//	eth := fakeEthNew(0)
//	if nil == eth {
//		fmt.Println("failed to create eth")
//	}
//	Convey("可否插入区块测试", t, func() {
//		Convey("当前高度0，leader可否生成高度1插入区块函数测试", func() {
//
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			header := types.CopyHeader(eth.BlockChain().CurrentHeader())
//			header.ParentHash = header.Hash()
//			header.Number = big.NewInt(1)
//			state, _ := eth.BlockChain().State()
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(0), common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			p := blockgen.pm.GetCurrentProcess()
//			p.genBlockData = &mc.BlockVerifyConsensusOK{header, common.Hash{}, nil, nil, state}
//			var guard *monkey.PatchGuard
//			guard = monkey.Patch(ca.GetElectedByHeightWithdraw, func(height *big.Int) ([]vm.DepositDetail, error) {
//				fmt.Println("use my GetElectedByHeightWithdraw")
//				guard.Unpatch()
//				defer guard.Restore()
//				return nil, nil
//			})
//
//			hash, err := p.insertAndBcBlock(true, nil)
//			So(err, ShouldEqual, nil)
//			So(hash, ShouldNotEqual, common.Hash{})
//			So(header.Hash(), ShouldEqual, eth.BlockChain().CurrentHeader().Hash())
//			So(eth.BlockChain().CurrentHeader().Number.Int64(), ShouldEqual, int64(1))
//
//		})
//	})
//}
//
//func TestProcess_FowllerInsertAndBcBlock(t *testing.T) {
//	log.InitLog(3)
//	eth := fakeEthNew(0)
//	if nil == eth {
//		fmt.Println("failed to create eth")
//	}
//
//	Convey("可否插入区块测试", t, func() {
//		Convey("当前高度0，follower可否生成高度1插入区块函数测试", func() {
//
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			header := types.CopyHeader(eth.BlockChain().CurrentHeader())
//			header.ParentHash = header.Hash()
//			header.Number = big.NewInt(2)
//			state, _ := eth.BlockChain().State()
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(1), common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			p := blockgen.pm.GetCurrentProcess()
//			p.genBlockData = &mc.BlockVerifyConsensusOK{header, common.Hash{}, nil, nil, state}
//			var guard *monkey.PatchGuard
//			monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
//				guard.Unpatch()
//				defer guard.Restore()
//				return nil, nil
//			})
//			var guard1 *monkey.PatchGuard
//			guard1 = monkey.Patch(ca.GetElectedByHeightWithdraw, func(height *big.Int) ([]vm.DepositDetail, error) {
//				guard1.Unpatch()
//				defer guard1.Restore()
//				return nil, nil
//			})
//			hash, err := p.insertAndBcBlock(false, header)
//			So(err, ShouldEqual, nil)
//			So(hash, ShouldNotEqual, common.Hash{})
//			So(header.Hash(), ShouldEqual, eth.BlockChain().CurrentHeader().Hash())
//			So(eth.BlockChain().CurrentHeader().Number.Int64(), ShouldEqual, int64(2))
//		})
//	})
//}
//
//func TestProcess_processSendBlock(t *testing.T) {
//	log.InitLog(3)
//	eth := fakeEthNew(0)
//	if nil == eth {
//		fmt.Println("failed to create eth")
//	}
//
//	Convey("可否插入区块测试", t, func() {
//		Convey("当前高度0，follower可否生成高度1插入区块函数测试", func() {
//
//			blockgen, err := New(eth)
//			if err != nil {
//			}
//			header := types.CopyHeader(eth.BlockChain().CurrentHeader())
//			header.ParentHash = header.Hash()
//			header.Number = big.NewInt(1)
//			state, _ := eth.BlockChain().State()
//			roleMsg := &mc.RoleUpdatedMsg{common.RoleValidator, uint64(0), common.HexToAddress(testAddress)}
//			blockgen.roleUpdatedMsgHandle(roleMsg)
//			p := blockgen.pm.GetCurrentProcess()
//			p.genBlockData = &mc.BlockVerifyConsensusOK{header, common.Hash{}, nil, nil, state}
//			p.processBlockInsert()
//			So(eth.BlockChain().CurrentHeader().Number.Int64(), ShouldEqual, int64(0))
//		})
//	})
//}
