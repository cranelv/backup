package miner

import (
	"math/big"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/mc"
	"github.com/pkg/errors"
	"sync"
	"github.com/MatrixAINetwork/go-matrix/params/manparams"
)
type mineReqData struct {
	mu                 sync.Mutex
	coinbase		   common.Address
	mined              bool
	headerHash         common.Hash
	header             *types.Header
	isBroadcastReq     bool
	isFriend		   bool
	txs                []types.CoinSelfTransaction
	mineDiff           *big.Int
	mineResultSendTime int64
}

func newMineReqData(headerHash common.Hash, header *types.Header, txs []types.CoinSelfTransaction, isBroadcastReq bool,isfriend bool) *mineReqData {
	return &mineReqData{
		mined:              false,
		headerHash:         headerHash,
		header:             header,
		isBroadcastReq:     isBroadcastReq,
		isFriend:           isfriend,
		txs:                txs,
		mineDiff:           nil,
		mineResultSendTime: 0,
	}
}
func (self *mineReqData) RequstHeader()*types.Header {
	return self.header
}
func (self *mineReqData)IsMined()bool{
	return self.mined
}
func (self *mineReqData)setMined(){
	self.mined = true
}
func (self *mineReqData)CreateWork() *Work{
	work := &Work{
		header:          types.CopyHeader(self.header),
		isBroadcastNode: self.isBroadcastReq,
	}
	work.mineType = mineTaskTypePow
	work.header.Coinbase = self.coinbase
	return work
}
func (self *mineReqData) ResendMineResult(curTime int64) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if curTime-self.mineResultSendTime < manparams.MinerResultSendInterval {
		return errors.Errorf("挖矿发送间隔尚未到, 上次发送时间(%d), 当前时间(%d)", self.mineResultSendTime, curTime)
	}
	self.mineResultSendTime = curTime
	return nil
}
type powMineTask struct {
	mineHash            common.Hash
	mineHeader          *types.Header
	bcInterval          *mc.BCIntervalInfo
	minedBasePow        bool
	minedPow            bool
	powMiningNumber     uint64
	powMiningDifficulty *big.Int
	powMiner            common.Address
	mixDigest           common.Hash
	nonce               types.BlockNonce
	sm3Nonce            types.BlockNonce
	coinbase		   common.Address
	isFriend			bool
}
func newPowMineTask(mineHash common.Hash, mineHeader *types.Header, powMiningNumber uint64, bcInterval *mc.BCIntervalInfo, difficulty *big.Int) *powMineTask {
	return &powMineTask{
		mineHash:            mineHash,
		mineHeader:          mineHeader,
		bcInterval:          bcInterval,
		minedBasePow:        false,
		minedPow:            false,
		powMiningNumber:     powMiningNumber,
		powMiningDifficulty: difficulty,
		powMiner:            common.Address{},
		mixDigest:           common.Hash{},
		nonce:               types.BlockNonce{},
		sm3Nonce:            types.BlockNonce{},
	}
}
func (self *powMineTask) RequstHeader()*types.Header {
	return self.mineHeader
}
func (self *powMineTask)IsMined()bool{
	return self.minedPow
}
func (self *powMineTask)setMined(){
	self.minedPow = true
}
func (self *powMineTask)CreateWork() *Work{
	work := &Work{
		header:  &types.Header{
		Number:     big.NewInt(int64(self.powMiningNumber)),
		ParentHash: self.mineHash,
		Difficulty: self.powMiningDifficulty,
		VrfValue:   self.mineHeader.VrfValue,
		Version:    self.mineHeader.Version,
		Coinbase:   self.coinbase,
		AICoinbase: self.coinbase,
	},
		isBroadcastNode: false,
	}
	work.mineType = mineTaskTypeX11
	work.header.Coinbase = self.coinbase
	return work
}

type aiMineTask struct {
	mineHash       common.Hash
	mineHeader     *types.Header
	bcInterval     *mc.BCIntervalInfo
	minedAI        bool
	aiMiningNumber uint64
	aiMiner        common.Address
	aiHash         common.Hash
}

func newAIMineTask(mineHash common.Hash, mineHeader *types.Header, aiMiningNumber uint64, bcInterval *mc.BCIntervalInfo) *aiMineTask {
	return &aiMineTask{
		mineHash:       mineHash,
		mineHeader:     types.CopyHeader(mineHeader),
		bcInterval:     bcInterval,
		minedAI:        false,
		aiMiningNumber: aiMiningNumber,
		aiMiner:        common.Address{},
		aiHash:         common.Hash{},
	}
}
const PowBlockPeriod = 3
func IsAIBlock(number uint64, broadcastInterval uint64) bool {
	remainder := number % broadcastInterval
	return remainder%PowBlockPeriod == 1
}

func GetCurAIBlockNumber(number uint64, broadcastInterval uint64) uint64 {
	remainder := number % broadcastInterval
	if remainder == 0 {
		return number - PowBlockPeriod
	}
	return number - (remainder+2)%PowBlockPeriod
}

func GetNextAIBlockNumber(number uint64, broadcastInterval uint64) uint64 {
	curAINumber := GetCurAIBlockNumber(number, broadcastInterval)
	nextAINumber := curAINumber + PowBlockPeriod
	if nextAINumber%broadcastInterval == 0 {
		nextAINumber += 1
	}
	return nextAINumber
}