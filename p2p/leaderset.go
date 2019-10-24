package p2p

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"sync"
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/crypto"
)

func isBroadcastBlock(blockNum uint64)bool{
	return blockNum%300 == 0
}
/*
type RecentLeader struct {
	RecentBlock uint64
	address common.Address
}
// in descending order
func greater(key1,key2 uint64) bool {
	return key1 < key2
}
func less(key1,key2 uint64) bool {
	return key1 > key2
}
func find(key uint64, leaderSet []RecentLeader) (int, bool) {
	left, right, mid := 0, len(leaderSet)-1, 0
	if right < 0 {
		return 0, false
	}
	for {
		mid = (left + right) / 2
		if greater(leaderSet[mid].RecentBlock, key) {
			right = mid - 1
		} else if less(leaderSet[mid].RecentBlock, key) {
			left = mid + 1
		} else {
			return mid, true
		}
		if left > right {
			return left, false
		}
	}
	return mid, false
}

//binary insert
func insert(leaderSet *[]RecentLeader, index int, value RecentLeader) {
	*leaderSet = append(*leaderSet, value)
	end := len(*leaderSet) - 1
	for i := end; i > index; i-- {
		(*leaderSet)[i], (*leaderSet)[i-1] = (*leaderSet)[i-1], (*leaderSet)[i]
	}
}
*/
type leaderInfo struct {
	blockNum uint64
	state uint64
}
type RecentLeaderSet struct {
	mu sync.RWMutex
	blockLimit uint64
	leaderSet map[common.Address] leaderInfo
	owner *Server
}
func newRecentLeaderSet(limit uint64,owner *Server) *RecentLeaderSet{
	return &RecentLeaderSet{
		sync.RWMutex{},
		limit,
		make(map[common.Address] leaderInfo),
		owner,
	}
}
func (ls *RecentLeaderSet) insertLeader(header *types.Header){
	blockNum := header.Number.Uint64()
	if !isBroadcastBlock(blockNum){
		ls.mu.Lock()
		headerHash := header.HashNoSignsAndNonce()
		for _,sign := range header.Signatures{
			signAccount, _, err := crypto.VerifySignWithValidate(headerHash[:], sign[:])
			if err != nil {
				continue
			}
			info,exist := ls.leaderSet[signAccount]
			if !exist{
				ls.leaderSet[signAccount] = leaderInfo{blockNum,1}
				ls.owner.addstatic <- signAccount
			} else{
				if blockNum > info.blockNum{
					info.state +=blockNum-info.blockNum
					if info.state > 50 {
						info.state = 1
					}
//					log.INFO("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","state",info.state,"info.blockNum",info.blockNum)
					info.blockNum = blockNum
					ls.leaderSet[signAccount] = info
					if info.state == 1 {
						ls.owner.addstatic <- signAccount
					}
				}
			}
		}
		if (blockNum -5) % 500 == 0{
			ls.removeExpireLeader(blockNum)
		}
		ls.mu.Unlock()
	}
}
func (ls *RecentLeaderSet) findNode(leader common.Address)  {
	ls.mu.Lock()
	if info,exist := ls.leaderSet[leader];exist {
		ls.leaderSet[leader]= leaderInfo{info.blockNum,1}
	}
	ls.mu.Unlock()
}
func (ls *RecentLeaderSet)removeExpireLeader(blockNum uint64){
	limit := blockNum - ls.blockLimit
	for addr,num := range ls.leaderSet{
		if num.blockNum < limit {
			delete(ls.leaderSet,addr)
			ls.owner.removestatic <- addr
		}
	}
}