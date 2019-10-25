package miner

import (
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/consensus/manash"
	"github.com/MatrixAINetwork/go-matrix/consensus/amhash"
)
const (
	PowOld = 0
	PowX11 = 1
	PowSm3 = 2
)
type mineInfo struct {
	abort chan struct{}
	found chan *types.Header
	header *types.Header
	powType int
}
type SealThread struct {
	id int
	seed uint64
	mineCh chan mineInfo
	manHash *manash.Manash
	amHash *amhash.Amhash
	scratchPad []uint64
}
func (st* SealThread) waitSeal(){
	for{
		select{
		case mineInfo := <-st.mineCh:
			switch mineInfo.powType {
			case PowOld:
				st.manHash.Mine(mineInfo.header, st.id, st.seed, mineInfo.abort, mineInfo.found,st.scratchPad)
			case PowX11:
				st.amHash.MineX11(mineInfo.header, st.id, st.seed, mineInfo.abort, mineInfo.found)
			case PowSm3:
				st.amHash.Sm3Mine(mineInfo.header, st.id, st.seed, mineInfo.abort, mineInfo.found)
			}
		}
	}
}