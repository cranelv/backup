package miner

import (
	"testing"
	"github.com/MatrixAINetwork/go-matrix/consensus/manash"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/common"
	"math/big"
	"time"
)

func TestMinerSpeed(t *testing.T){
	agent := NewCpuAgent(manash.New(manash.Config{}))
	time1 := time.Now().UnixNano()
	for i:=0;i<1000;i++{
		header := &types.Header{ParentHash:common.BigToHash(big.NewInt(int64(i+10000))),
		Difficulty:big.NewInt(int64(15)),
		Number:big.NewInt(int64(i+10))}
		quitCurrentOp := make(chan struct{})
		agent.manhash.Seal(nil,header,quitCurrentOp,false)
		close(quitCurrentOp)
//		t.Log(i)
	}
	time2 := time.Now().UnixNano()
	t.Log(time2-time1)
}
