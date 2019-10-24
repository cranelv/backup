package manash

import (
	"github.com/MatrixAINetwork/go-matrix/core/types"
)
type mineInfo struct {
	abort chan struct{}
	found chan *types.Header
	header *types.Header
}
type SealThread struct {
	id int
	seed uint64
	mineCh chan mineInfo
	manHash *Manash
	scratchPad []uint64
}
func (st* SealThread) waitSeal(){
	for{
		select{
		case mineInfo := <-st.mineCh:
			st.manHash.mine(mineInfo.header, st.id, st.seed, mineInfo.abort, mineInfo.found, st.scratchPad)
		}
	}
}