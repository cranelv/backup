package rewardexec

import (
	"fmt"
	"github.com/matrix/go-matrix/consensus/manash"
	"github.com/matrix/go-matrix/reward/cfg"
	"math/big"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/matrix/go-matrix/reward/util"

	"bou.ke/monkey"
	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/core/vm"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testAddress = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
)

var myNodeId string = "4b2f638f46c7ae5b1564ca7015d716621848a0d9be66f1d1e91d566d2a70eedc2f11e92b743acb8d97dec3fb412c1b2f66afd7fbb9399d4fb2423619eaa51411"

type FakeEth struct {
	blockchain *core.BlockChain
	once       *sync.Once
}

func (s *FakeEth) BlockChain() *core.BlockChain { return s.blockchain }

func fakeEthNew(n int) *FakeEth {
	eth := &FakeEth{once: new(sync.Once)}
	eth.once.Do(func() {
		_, blockchain, err := core.NewCanonical(manash.NewFaker(), n, true)
		if err != nil {
			fmt.Println("failed to create pristine chain: ", err)
			return
		}
		defer blockchain.Stop()
		eth.blockchain = blockchain

	})
	return eth
}

type InnerSeed struct {
}

func (s *InnerSeed) GetSeed(num uint64) *big.Int {
	random := rand.New(rand.NewSource(0))
	return new(big.Int).SetUint64(random.Uint64())
}

func TestBlockReward_setLeaderRewards(t *testing.T) {

	log.InitLog(3)
	eth := fakeEthNew(0)
	rewardCfg := cfg.New(nil, nil)
	rewardobject := New(eth.blockchain, rewardCfg)
	Convey("Leader测试0", t, func() {
		rewards := make(map[common.Address]*big.Int, 0)
		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(0), rewards, common.HexToAddress(testAddress), new(big.Int).SetUint64(uint64(1)))
	})

	Convey("Leader测试1", t, func() {
		rewards := make(map[common.Address]*big.Int, 0)
		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(2), rewards, common.Address{}, new(big.Int).SetUint64(uint64(1)))
	})

	Convey("Leader测试2", t, func() {
		rewards := make(map[common.Address]*big.Int, 0)
		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(2), rewards, common.HexToAddress(testAddress), new(big.Int).SetUint64(uint64(2)))
	})

	Convey("Leader测试3", t, func() {
		rewards := make(map[common.Address]*big.Int, 0)
		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(-1), rewards, common.HexToAddress(testAddress), new(big.Int).SetUint64(uint64(2)))
	})

	Convey("Leader测试4", t, func() {
		rewards := make(map[common.Address]*big.Int, 0)
		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(100), rewards, common.HexToAddress(testAddress), new(big.Int).SetUint64(uint64(0)))
	})

	Convey("Leader测试5", t, func() {
		rewards := make(map[common.Address]*big.Int, 0)
		rewardobject.rewardCfg.SetReward.SetLeaderRewards(big.NewInt(100), rewards, common.HexToAddress(testAddress), new(big.Int).SetUint64(uint64(100)))
	})
}

type State struct {
	balance int64
}

func (st *State) GetBalance(common.Address) *big.Int {
	return big.NewInt(st.balance)
}

type Chain struct {
	bc *core.BlockChain
}

func (chain *Chain) Config() *params.ChainConfig {
	return chain.bc.Config()
}

// CurrentHeader retrieves the current header from the local chain.
func (chain *Chain) CurrentHeader() *types.Header {

	return chain.bc.CurrentHeader()
}

// GetHeader retrieves a block header from the database by hash and number.
func (chain *Chain) GetHeader(hash common.Hash, number uint64) *types.Header {
	return chain.bc.GetHeader(hash, number)
}

// GetHeaderByHash retrieves a block header from the database by its hash.
func (chain *Chain) GetHeaderByHash(hash common.Hash) *types.Header {
	return chain.bc.GetHeaderByHash(hash)
}

func (chain *Chain) GetBlockByNumber(number uint64) *types.Block {
	return chain.bc.GetBlockByNumber(number)
}

// GetBlock retrieves a block sfrom the database by hash and number.
func (chain *Chain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return chain.bc.GetBlock(hash, number)
}
func (chain *Chain) StateAt(root common.Hash) (*state.StateDB, error) {

	return chain.bc.StateAt(root)
}
func (chain *Chain) State() (*state.StateDB, error) {

	return chain.bc.State()
}
func (chain *Chain) NewTopologyGraph(header *types.Header) (*mc.TopologyGraph, error) {
	return chain.bc.NewTopologyGraph(header)
}

func (chain *Chain) GetHeaderByNumber(number uint64) *types.Header {
	header := &types.Header{
		Coinbase: common.Address{},
	}
	//txs := make([]types.SelfTransaction, 0)

	return header
}
func TestBlockReward_setMinerOut(t *testing.T) {

	log.InitLog(3)
	eth := fakeEthNew(0)
	rewardCfg := cfg.New(nil, nil)
	rewardobject := New(eth.blockchain, rewardCfg)
	Convey("挖矿测试0", t, func() {
		rewards := make(map[common.Address]*big.Int, 0)
		rewardobject.rewardCfg.SetReward.SetMinerOutRewards(big.NewInt(0), eth.blockchain, new(big.Int).SetUint64(uint64(2)), rewards)
	})

	Convey("挖矿测试1", t, func() {
		rewards := make(map[common.Address]*big.Int, 0)
		rewardobject.rewardCfg.SetReward.SetMinerOutRewards(big.NewInt(2), eth.blockchain, new(big.Int).SetUint64(uint64(1)), rewards)
	})

	Convey("挖矿测试高度错误", t, func() {

		rewards := make(map[common.Address]*big.Int, 0)
		rewardobject.rewardCfg.SetReward.SetMinerOutRewards(big.NewInt(2), eth.blockchain, new(big.Int).SetUint64(uint64(100)), rewards)
	})
	Convey("挖矿账户nil", t, func() {

		rewards := make(map[common.Address]*big.Int, 0)
		rewardobject.rewardCfg.SetReward.SetMinerOutRewards(big.NewInt(2), &Chain{eth.blockchain}, new(big.Int).SetUint64(uint64(2)), rewards)
	})
}

func TestBlockReward_setSelectedBlockRewards(t *testing.T) {
	type args struct {
		chain util.ChainReader
	}
	log.InitLog(3)
	eth := fakeEthNew(0)
	rewardCfg := cfg.New(nil, nil)
	reward := New(eth.blockchain, rewardCfg)
	depositProportion := big.NewRat(1, 11)
	oneNodeReward, err := new(big.Rat).Mul(new(big.Rat).SetInt(big.NewInt(9e+18)), depositProportion).Float64()
	oneNodeRewardString := new(big.Rat).Mul(new(big.Rat).SetInt(big.NewInt(9e+18)), depositProportion).FloatString(10)
	v2, _ := strconv.ParseFloat(oneNodeRewardString, 64)
	log.INFO(PackageName, "奖励", oneNodeReward, "err", err, "oneNodeRewardString", oneNodeRewardString, "v2", v2)

	oneNodeRewardInt := new(big.Int).Div(big.NewInt(9e+17), big.NewInt(11)).Uint64()
	log.INFO(PackageName, "奖励", oneNodeRewardInt)
	SkipConvey("选中无节点变化测试", t, func() {

		rewards := make(map[common.Address]*big.Int, 0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		//newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		reward.rewardCfg.SetReward.SetSelectedRewards(big.NewInt(11e+17), eth.blockchain, rewards, common.RoleValidator|common.RoleBackupValidator, newheader, reward.rewardCfg.RewardMount.BackupRewardRate)

		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})

	SkipConvey("选中有节点变化测试", t, func() {

		rewards := make(map[common.Address]*big.Int, 0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Position: 8195})
		reward.rewardCfg.SetReward.SetSelectedRewards(util.ValidatorsBlockReward, eth.blockchain, rewards, common.RoleValidator|common.RoleBackupValidator, newheader, reward.rewardCfg.RewardMount.BackupRewardRate)
		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})

	Convey("奖励金额0", t, func() {

		rewards := make(map[common.Address]*big.Int, 0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Position: 8195})
		reward.rewardCfg.SetReward.SetSelectedRewards(big.NewInt(0), eth.blockchain, rewards, common.RoleValidator|common.RoleBackupValidator, newheader, reward.rewardCfg.RewardMount.BackupRewardRate)
		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})

	Convey("奖励金额小于0", t, func() {

		rewards := make(map[common.Address]*big.Int, 0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Position: 8195})
		reward.rewardCfg.SetReward.SetSelectedRewards(big.NewInt(-1), eth.blockchain, rewards, common.RoleValidator|common.RoleBackupValidator, newheader, reward.rewardCfg.RewardMount.BackupRewardRate)
		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})

	Convey("抵押列表为空", t, func() {

		rewards := make(map[common.Address]*big.Int, 0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Position: 8195})
		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
			fmt.Println("use monkey  ca.GetTopologyByNumber")
			newGraph := &mc.TopologyGraph{
				Number:        number,
				NodeList:      make([]mc.TopologyNodeInfo, 0),
				CurNodeNumber: 0,
			}
			if common.RoleValidator == reqTypes&common.RoleValidator {
				//newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
				newGraph.CurNodeNumber = 4
			}

			return newGraph, nil
		})
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")

			return nil, nil
		})
		reward.rewardCfg.SetReward.SetSelectedRewards(big.NewInt(100), eth.blockchain, rewards, common.RoleValidator|common.RoleBackupValidator, newheader, reward.rewardCfg.RewardMount.BackupRewardRate)
		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})

	Convey("原始拓扑图为空", t, func() {

		rewards := make(map[common.Address]*big.Int, 0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Position: 8195})
		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
			fmt.Println("use monkey  ca.GetTopologyByNumber")
			newGraph := &mc.TopologyGraph{
				Number:        number,
				NodeList:      make([]mc.TopologyNodeInfo, 0),
				CurNodeNumber: 0,
			}
			if common.RoleValidator == reqTypes&common.RoleValidator {
				//newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
				newGraph.CurNodeNumber = 4
			}

			return nil, errors.New("获取拓扑为nil")
		})
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")

			return nil, nil
		})
		reward.rewardCfg.SetReward.SetSelectedRewards(big.NewInt(100), eth.blockchain, rewards, common.RoleValidator|common.RoleBackupValidator, newheader, reward.rewardCfg.RewardMount.BackupRewardRate)
		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})

	Convey("原始拓扑图没有节点信息", t, func() {

		rewards := make(map[common.Address]*big.Int, 0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Position: 8195})
		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
			fmt.Println("use monkey  ca.GetTopologyByNumber")
			newGraph := &mc.TopologyGraph{
				Number:        number,
				NodeList:      make([]mc.TopologyNodeInfo, 0),
				CurNodeNumber: 0,
			}
			//if common.RoleValidator == reqTypes&common.RoleValidator {
			//	//newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
			//	newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
			//	newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
			//	newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
			//	newGraph.CurNodeNumber = 4
			//}

			return newGraph, nil
		})
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")

			return nil, nil
		})
		reward.rewardCfg.SetReward.SetSelectedRewards(big.NewInt(100), eth.blockchain, rewards, common.RoleValidator|common.RoleBackupValidator, newheader, reward.rewardCfg.RewardMount.BackupRewardRate)
		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})

	Convey("抵押列表节点抵押非法", t, func() {

		rewards := make(map[common.Address]*big.Int, 0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Position: 8195})
		monkey.Patch(ca.GetTopologyByNumber, func(reqTypes common.RoleType, number uint64) (*mc.TopologyGraph, error) {
			fmt.Println("use monkey  ca.GetTopologyByNumber")
			newGraph := &mc.TopologyGraph{
				Number:        number,
				NodeList:      make([]mc.TopologyNodeInfo, 0),
				CurNodeNumber: 0,
			}
			if common.RoleValidator == reqTypes&common.RoleValidator {
				//newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
				newGraph.NodeList = append(newGraph.NodeList, mc.TopologyNodeInfo{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
				newGraph.CurNodeNumber = 4
			}

			return newGraph, nil
		})
		monkey.Patch(ca.GetElectedByHeightAndRole, func(height *big.Int, roleType common.RoleType) ([]vm.DepositDetail, error) {
			fmt.Println("use monkey  ca.GetElectedByHeightAndRole")
			Deposit := make([]vm.DepositDetail, 0)
			if common.RoleValidator == roleType&common.RoleValidator {
				//Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Deposit: big.NewInt(1e+18)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Deposit: big.NewInt(-1)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Deposit: big.NewInt(3e+18)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Deposit: big.NewInt(4e+18)})
				Deposit = append(Deposit, vm.DepositDetail{Address: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bcf"), Deposit: big.NewInt(2e+18)})

			}

			return Deposit, nil
		})
		reward.rewardCfg.SetReward.SetSelectedRewards(big.NewInt(100), eth.blockchain, rewards, common.RoleValidator|common.RoleBackupValidator, newheader, reward.rewardCfg.RewardMount.BackupRewardRate)
		//So(rewards[common.HexToAddress(testAddress)], ShouldEqual, reward.leaderBlkReward)
	})
}

func TestBlockReward_calcTxsFees(t *testing.T) {
	Convey("计算交易费", t, func() {

		log.InitLog(3)
		eth := fakeEthNew(0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		rewardCfg := cfg.New(nil, nil)
		reward := New(eth.blockchain, rewardCfg)
		reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
	})
}

func TestBlockReward_CalcRewardMountByNumber(t *testing.T) {
	Convey("初始奖励金额小于0", t, func() {

		log.InitLog(3)
		eth := fakeEthNew(0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		rewardCfg := cfg.New(nil, nil)
		reward := New(eth.blockchain, rewardCfg)
		//reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
		reward.CalcRewardMountByNumber(&State{100}, 2, big.NewInt(-1), 100, common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"))
	})

	Convey("发放余额等于0", t, func() {

		log.InitLog(3)
		eth := fakeEthNew(0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		rewardCfg := cfg.New(nil, nil)
		reward := New(eth.blockchain, rewardCfg)
		//reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
		state00, _ := eth.blockchain.State()
		reward.CalcRewardMountByNumber(state00, 2, big.NewInt(0), 100, common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"))
	})

	Convey("状态树为nil", t, func() {

		log.InitLog(3)
		eth := fakeEthNew(0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		rewardCfg := cfg.New(nil, nil)
		reward := New(eth.blockchain, rewardCfg)
		//reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
		reward.CalcRewardMountByNumber(nil, 2, big.NewInt(100), 100, common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"))
	})

	Convey("账户余额为0", t, func() {

		log.InitLog(3)
		eth := fakeEthNew(0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		rewardCfg := cfg.New(nil, nil)
		reward := New(eth.blockchain, rewardCfg)
		//reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
		reward.CalcRewardMountByNumber(&State{0}, 2, big.NewInt(100), 100, common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"))
	})

	Convey("账户余额为负值", t, func() {

		log.InitLog(3)
		eth := fakeEthNew(0)
		header := eth.BlockChain().CurrentHeader()
		newheader := types.CopyHeader(header)
		newheader.Number = big.NewInt(1)
		newheader.NetTopology.Type = common.NetTopoTypeAll
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"), Position: 8192})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x82799145a60b4d1e88d5a895601508f2b7f4ee9b"), Position: 8193})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x519437b21e2a0b62788ab9235d0728dd7f1a7269"), Position: 8194})
		newheader.NetTopology.NetTopologyData = append(newheader.NetTopology.NetTopologyData, common.NetTopologyData{Account: common.HexToAddress("0x29216818d3788c2505a593cbbb248907d47d9bce"), Position: 8195})
		rewardCfg := cfg.New(nil, nil)
		reward := New(eth.blockchain, rewardCfg)
		//reward.CalcNodesRewards(util.ByzantiumTxsRewardDen, common.HexToAddress(testAddress), header)
		reward.CalcRewardMountByNumber(&State{-1000}, 2, big.NewInt(100), 100, common.HexToAddress("0x475baee143cf541ff3ee7b00c1c933129238d793"))
	})
}
