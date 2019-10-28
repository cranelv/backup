package miner

import (
	"github.com/MatrixAINetwork/go-matrix/common"
	"sync"
)
type cacheMap struct {
	blockNumber uint64
	cache map[common.Hash]interface{}
}
func (cm *cacheMap) IsCorrectCache(blocknumber uint64)bool{
	return cm.blockNumber == blocknumber
}
func (cm *cacheMap) SetBlockNumber(blocknumber uint64){
	if cm.blockNumber != blocknumber{
		cm.blockNumber = blocknumber
		cm.cache = make(map[common.Hash]interface{})
	}
}
func (cm *cacheMap) GetCacheItem(hash common.Hash)interface{}{
	return cm.cache[hash]
}
type RequestCache struct{
	mu sync.RWMutex
	requestCache []*cacheMap
}
func NewRequestCache(cacheNum int)*RequestCache{
	cache := RequestCache{
		requestCache:make([]*cacheMap,cacheNum,cacheNum),
	}
	for i:=0;i< len(cache.requestCache);i++{
		cache.requestCache[i] = &cacheMap{
			cache:make(map[common.Hash]interface{}),
		}
	}
	return &cache
}
func (rc *RequestCache) getCache(blockNumber uint64)*cacheMap{
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	cache := rc.requestCache[blockNumber%uint64(len(rc.requestCache))]
	if cache.IsCorrectCache(blockNumber){
		return cache
	}
	return nil
}
func (rc *RequestCache)GetCacheItem(blockNumber uint64,hash common.Hash)interface{}{
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	cache := rc.getCache(blockNumber)
	if cache != nil {
		return cache.GetCacheItem(hash)
	}
	return nil
}

func (rc *RequestCache)SetCacheItem(blockNumber uint64,hash common.Hash,value interface{}){
	rc.mu.Lock()
	defer rc.mu.Unlock()
	cache := rc.requestCache[blockNumber%uint64(len(rc.requestCache))]
	cache.SetBlockNumber(blockNumber)
	if _,exist := cache.cache[hash];!exist{
		cache.cache[hash] = value
	}
}
func (rc *RequestCache) getMaxBlockNumberCache()*cacheMap{
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	maxCache := rc.requestCache[0]
	for i:=1;i< len(rc.requestCache);i++{
		if maxCache.blockNumber < rc.requestCache[i].blockNumber {
			maxCache = rc.requestCache[i]
		}
	}
	return maxCache
}
func (rc *RequestCache) ReadCache(blockNumber uint64,reader func(cacheMap2 *cacheMap) interface{})interface{}{
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	cache := rc.requestCache[blockNumber%uint64(len(rc.requestCache))]
	if cache.IsCorrectCache(blockNumber){
		return reader(cache)
	}
	return nil
}
func (rc *RequestCache) ReadMaxBlockNumberCache(reader func(cacheMap2 *cacheMap) interface{})interface{}{
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	maxCache := rc.requestCache[0]
	for i:=1;i< len(rc.requestCache);i++{
		if maxCache.blockNumber < rc.requestCache[i].blockNumber {
			maxCache = rc.requestCache[i]
		}
	}
	return reader(maxCache)
}

