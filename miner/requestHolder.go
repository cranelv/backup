package miner

import (
	"github.com/MatrixAINetwork/go-matrix/core/types"
	"github.com/MatrixAINetwork/go-matrix/common"
	"github.com/pkg/errors"
)

type MinerRequestInterface interface {
	RequstHeader()*types.Header
	IsMined()bool
	setMined()
	CreateWork() *Work
}
type RequestHolder struct{
	current MinerRequestInterface
	cache *RequestCache
	requestCreater func (header *types.Header, miner common.Address, isBroadcastReq bool,isfriend bool) MinerRequestInterface
	RequestHash func (header *types.Header) common.Hash
	RequestNumber func (header *types.Header) uint64
}
func (holder *RequestHolder) AddMineReq(header *types.Header, miner common.Address, isBroadcastReq bool,isfriend bool) (MinerRequestInterface, error) {
	if nil == header {
		return nil, errors.New("header为nil")
	}
	reqData := holder.requestCreater(header,miner,isBroadcastReq,isfriend)
	if reqData != nil {
		holder.cache.SetCacheItem(holder.RequestNumber(header),holder.RequestHash(header),reqData)
		return reqData, nil
	}
	return nil,nil
}
func (holder *RequestHolder) GetUnMinedReq() MinerRequestInterface {
	getMaxTimeItem := func (cm *cacheMap)interface{}{
		var foundItem interface{} = nil
		for _,value := range cm.cache {
			if foundItem == nil {
				foundItem = value
			}else{
				found := foundItem.(MinerRequestInterface)
				item := value.(MinerRequestInterface)
				if item.RequstHeader().Time.Cmp(found.RequstHeader().Time) > 0{
					foundItem = value
				}
			}
		}
		return foundItem
	}
	maxReq := holder.cache.ReadMaxBlockNumberCache(getMaxTimeItem)
	if maxReq == nil {
		return nil
	}
	return maxReq.(MinerRequestInterface)
}
func (holder *RequestHolder) beginMine() MinerRequestInterface {
	maxReq := holder.GetUnMinedReq()
	if maxReq == nil {
		return nil
	}
	curMineReq := holder.current
	if curMineReq != nil {
		if maxReq.RequstHeader().Time.Cmp(curMineReq.RequstHeader().Time) <= 0 {
			return nil
		}
	}
	holder.current = maxReq
	return maxReq
	//	self.StopMiner()
}
func (holder *RequestHolder) SetMiningResult(result *types.Header,hash common.Hash) (MinerRequestInterface, error) {
	if nil == result {
		return nil, errors.New("消息为nil")
	}
	headerHash := hash
	number := holder.RequestNumber(result)
//	log.Info("EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE","num",number,"hash",headerHash)
	req := holder.cache.GetCacheItem(number,headerHash)
	if req == nil {
		return nil,  errors.Errorf("Miner Request(%s) is not found", headerHash.TerminalString())
	}
	holder.cache.mu.Lock()
	defer holder.cache.mu.Unlock()
	data := req.(MinerRequestInterface)
	if data.IsMined() {
		return nil, errors.Errorf("请求(%s)已挖矿完成", headerHash.TerminalString())
	}
	data.setMined()
	data.RequstHeader().Nonce = result.Nonce
	data.RequstHeader().Sm3Nonce = result.Sm3Nonce
	data.RequstHeader().Coinbase = result.Coinbase
	data.RequstHeader().MixDigest = result.MixDigest
	data.RequstHeader().Signatures = result.Signatures

	return data,nil
}
