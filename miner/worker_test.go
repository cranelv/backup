package miner

import (
	"testing"
	"github.com/MatrixAINetwork/go-matrix/consensus/manash"
	"math/big"
	"github.com/MatrixAINetwork/go-matrix/params"
	"github.com/MatrixAINetwork/go-matrix/msgsend"
	"time"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/MatrixAINetwork/go-matrix/mc"
	_ "github.com/MatrixAINetwork/go-matrix/crypto/vrf"
	"github.com/MatrixAINetwork/go-matrix/consensus/amhash"
	"github.com/MatrixAINetwork/go-matrix/consensus/ai"
)
var ( testConfig = &params.ChainConfig{
		ChainId:             big.NewInt(1),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        big.NewInt(0),
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: nil,
		Manash:              new(params.ManashConfig),
	}
	testHD,_ = msgsend.NewHD()
)
func TestPowWorkerV1(t *testing.T) {
//	log.InitLog(3)
	ai.Init("/home/cranelv/work2/src/github.com/MatrixAINetwork/go-matrix/TestNet/data/picstore")
//	MsgCenter = ctx.MsgCenter
	engine := manash.New(manash.Config{
		CacheDir:       "./test",
		CachesInMem:    3,
		CachesOnDisk:   10,
		DatasetDir:     "./test",
		DatasetsInMem:  3,
		DatasetsOnDisk: 10,
	})
	aiMineEngine := amhash.New(amhash.Config{PowMode: amhash.ModeNormal,
		PictureStorePath: "/home/cranelv/work2/src/github.com/MatrixAINetwork/go-matrix/TestNet/data/picstore"})
	miner, err := New(engine,aiMineEngine, testConfig, nil, testHD)
	if err != nil {
		t.Fatal(err)
	}
	go miner.Start()

	for i:=int64(1);i<20000;i++{
		for j:=0;j<100;j++{
			testHeader := &types.Header{
				ParentHash: common.BigToHash(big.NewInt(100)),
				Difficulty: big.NewInt(int64(1000)),
				Number:     big.NewInt(i),
				Nonce:      types.EncodeNonce(8),
				Time:       big.NewInt(time.Now().UnixNano()),
				Coinbase:   common.BigToAddress(big.NewInt(123)),
				MixDigest:  common.BigToHash(big.NewInt(777)),
				Signatures: []common.Signature{common.BytesToSignature(common.BigToHash(big.NewInt(100)).Bytes())},
			}
			mc.PublishEvent(mc.HD_V2_MiningReq,&mc.HD_MiningReqMsg{From:common.Address{100,10},Header: testHeader})
			time.Sleep(10*time.Millisecond)
		}
		time.Sleep(100*time.Millisecond)
	}
	time.Sleep(10*time.Second)
}