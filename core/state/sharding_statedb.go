package state

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/trie"
	"github.com/matrix/go-matrix/rlp"
	"math/big"
	"encoding/json"
	_"github.com/matrix/go-matrix/base58"
	"fmt"
)
type RangeManage struct {
	Range   byte
	State 	*StateDB
}
type CoinManage struct {
	Cointyp string
	Rmanage  []*RangeManage
}
type StateDBManage struct {
	db          Database
	trie        Trie
	shardings	[]*CoinManage
	coinRoot    []common.CoinRoot
}

// Create a new state from a given trie.
func NewStateDBManage(roots []common.CoinRoot, db Database) (*StateDBManage, error) {
	if len(roots) == 0{
		roots = append(roots,common.CoinRoot{Cointyp:params.MAN_COIN,Root:common.Hash{}})
	}
	return &StateDBManage{
		db:                db,
		shardings:         make([]*CoinManage,0),
		coinRoot:          roots,
	}, nil
}
func (shard *StateDBManage) MakeStatedb(cointyp string,b byte) {
	//没有对应币种或byte分区的时候，才创建
	for _,sh := range shard.shardings{
		if sh.Cointyp == cointyp{
			for _,st := range sh.Rmanage{
				if st.Range == b{
					return
				}
			}
		}
	}
	//获取指定的币种root
	for _,cr := range shard.coinRoot{
		if cr.Cointyp == cointyp{
			tr,err := shard.db.OpenTrie(cr.Root)
			if err != nil{
				log.Error("file sharding_statedb", "func MakeStatedb:Unmarshal:root", err)
				panic(err)
			}
			shard.trie = tr
		}
	}
	bs,err:=json.Marshal(b)
	if err != nil{
		log.Error("file sharding_statedb", "func MakeStatedb:Marshal", err)
		panic(err)
	}
	root,err := shard.trie.TryGet(bs)
	stdb,_ := newStatedb(common.BytesToHash(root),shard.db)
	rms := make([]*RangeManage,0)
	rms = append(rms,&RangeManage{Range:b,State:stdb})
	cm := &CoinManage{Cointyp:cointyp,Rmanage:rms}
	shard.shardings = append(shard.shardings,cm)
}

func (shard *StateDBManage) Reset(roots []common.CoinRoot) error {
	for _,cr := range roots{
		tr,err := shard.db.OpenTrie(cr.Root)
		if err != nil{
			log.Error("file sharding_statedb", "func MakeStatedb:Unmarshal:root", err)
			panic(err)
		}
		shard.trie = tr
		for _,cm := range shard.shardings{
			if cm.Cointyp == cr.Cointyp{
				for _,rm := range cm.Rmanage{
					bs,err:=json.Marshal(rm.Range)
					if err != nil{
						log.Error("file sharding_statedb", "func MakeStatedb:Marshal", err)
						panic(err)
					}
					root,err := shard.trie.TryGet(bs)
					rm.State.Reset(common.BytesToHash(root))
				}
				break
			}
		}
	}
	return nil
}

func (shard *StateDBManage)GetStateDb(cointyp string,address common.Address) *StateDB {
	cms := shard.shardings
	for _, cm := range cms {
		if cm.Cointyp == cointyp {
			rms := cm.Rmanage
			for _, rm := range rms {
				if rm.Range == address[1] {
					return rm.State
				}
			}
		}
	}
	return nil
}

func (shard *StateDBManage) AddLog(cointyp string,address common.Address,log *types.Log) {
	self:=shard.GetStateDb(cointyp,address)
	self.journal.append(addLogChange{txhash: self.thash})

	log.TxHash = self.thash
	log.BlockHash = self.bhash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}

func (shard *StateDBManage) GetLogs(cointyp string,address common.Address,hash common.Hash) []*types.Log {
	sd:=shard.GetStateDb(cointyp,address)
	return sd.logs[hash]

}

func (shard *StateDBManage) Logs(cointyp string,roots []common.Hash) []*types.Log {
	cms:=shard.shardings
	var logs []*types.Log
	for _,cm:=range cms {
		if cm.Cointyp==cointyp {
			rms:=cm.Rmanage
			for _,rm :=range rms{
				log:=rm.State.logs
				for _,l:=range log  {
					logs=append(logs,l...)
				}

			}
		}
	}
	//self:=shard.sharding
	//
	//for i:=0;i<sharding_MOUNTS ;i++  {
	//	for _, lgs := range self[i].logs {
	//		logs = append(logs, lgs...)
	//	}
	//}
	return logs //logs
}

// AddPreimage records a SHA3 preimage seen by the VM.
//TODO	只区分币种？还要区分256么
func (shard *StateDBManage) AddPreimage(hash common.Hash, preimage []byte) {

}

// Preimages returns a list of SHA3 preimages that have been submitted.
//TODO	取所有的，还是某个statedb的
func (shard *StateDBManage) Preimages() map[common.Hash][]byte {

	return nil//shard.sharding[idx].preimages
}
//退款是某一个staetdb的？还是要在shardingdb上加上退款成员变量
func (shard *StateDBManage) AddRefund(cointyp string,address common.Address,gas uint64) {
	for _,cm:=range shard.shardings{
		if cm.Cointyp==cointyp {
			for _,rm:=range cm.Rmanage  {
				if rm.Range==address[1] {
					rm.State.AddRefund(gas)
				}
			}
		}
	}
}
func (self *StateDBManage) GetRefund(cointyp string,address common.Address) uint64 {
	sd:=self.GetStateDb(cointyp,address)
	return sd.refund
}
// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (shard *StateDBManage) Exist(cointyp string,addr common.Address) bool {
	return shard.getStateObject(cointyp,addr)!=nil
}

// Empty returns whether the state object is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (shard *StateDBManage) Empty(cointyp string,addr common.Address) bool {
	so := shard.getStateObject(cointyp,addr)
	return so == nil || so.empty()
}

// Retrieve the balance from the given address or 0 if object not found
func (shard *StateDBManage) GetBalance(cointyp string,addr common.Address) common.BalanceType {

	stateObject := shard.getStateObject(cointyp,addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return nil
}
func (shard *StateDBManage)GetBalanceAll(common.Address) common.BalanceType{
	return nil
}
func (self *StateDBManage)GetBalanceByType(cointyp string,addr common.Address, accType uint32) *big.Int{
	stateObject := self.getStateObject(cointyp,addr)
	if stateObject != nil {
		for _, tAccount := range stateObject.data.Balance {
			if tAccount.AccountType == accType {
				return tAccount.Balance
			}
		}
	}

	return big.NewInt(0)
}
func (shard *StateDBManage) GetNonce(cointyp string,addr common.Address) uint64 {
	stateObject := shard.getStateObject(cointyp,addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0 | params.NonceAddOne //YY
}


func (shard *StateDBManage) GetCode(cointyp string,addr common.Address) []byte{

	stateObject := shard.getStateObject(cointyp,addr)
	if stateObject != nil {
		return stateObject.Code(shard.GetStateDb(cointyp,addr).db)
	}
	return nil
}
//TODO
func (shard *StateDBManage) GetCodeSize(cointyp string,addr common.Address) int {

	stateObject := shard.getStateObject(cointyp,addr)
	if stateObject == nil {
		return 0
	}
	if stateObject.code != nil {
		return len(stateObject.code)
	}
	size, err := shard.db.ContractCodeSize(stateObject.addrHash, common.BytesToHash(stateObject.CodeHash()))
	if err != nil {
		shard.GetStateDb(cointyp,addr).setError(err)
	}
	return size
}

func (shard *StateDBManage) GetCodeHash(cointyp string,addr common.Address) common.Hash {
	stateObject := shard.getStateObject(cointyp,addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

func (shard *StateDBManage) GetState(cointyp string,addr common.Address, bhash common.Hash) common.Hash {
	stateObject := shard.getStateObject(cointyp,addr)
	if stateObject != nil {
		return stateObject.GetState(shard.GetStateDb(cointyp,addr).db, bhash)
	}
	return common.Hash{}
}

func (shard *StateDBManage) GetStateByteArray(cointyp string,a common.Address, b common.Hash) []byte {
	stateObject := shard.getStateObject(cointyp,a)
	if stateObject != nil {
		return stateObject.GetStateByteArray(shard.GetStateDb(cointyp,a).db, b)
	}
	return nil
}

// Database retrieves the low level database supporting the lower level trie ops.
func (shard *StateDBManage) Database() Database {
	return shard.db
}

// StorageTrie returns the storage trie of an account.
// The return value is a copy and is nil for non-existent accounts.
func (shard *StateDBManage) StorageTrie(cointyp string,addr common.Address) Trie {

	stateObject := shard.getStateObject(cointyp,addr)
	if stateObject == nil {
		return nil
	}
	cpy := stateObject.deepCopy(shard.GetStateDb(cointyp,addr))
	return cpy.updateTrie(shard.db)
}

func (shard *StateDBManage) HasSuicided(cointyp string,addr common.Address) bool {
	stateObject := shard.getStateObject(cointyp,addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

/*
 * SETTERS
 */

// AddBalance adds amount to the account associated with addr.
func (shard *StateDBManage) AddBalance(cointyp string,accountType uint32, addr common.Address, amount *big.Int) {
	stateObject := shard.GetOrNewStateObject(cointyp,addr)
	if stateObject != nil {
		stateObject.AddBalance(accountType, amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (shard *StateDBManage) SubBalance(cointyp string,accountType uint32,addr common.Address,amount *big.Int) {
	stateObject := shard.GetOrNewStateObject(cointyp,addr)
	if stateObject != nil {
		stateObject.SubBalance(accountType, amount)
	}
}

func (shard *StateDBManage) SetBalance(cointyp string,accountType uint32, addr common.Address, amount *big.Int) {
	stateObject := shard.GetOrNewStateObject(cointyp,addr)
	if stateObject != nil {
		stateObject.SetBalance(accountType, amount)
	}
}

func (shard *StateDBManage) SetNonce(cointyp string,addr common.Address, nonce uint64) {
	stateObject := shard.GetOrNewStateObject(cointyp,addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce | params.NonceAddOne) //YY
	}
}

func (shard *StateDBManage) SetCode(cointyp string,addr common.Address, code []byte) {
	stateObject := shard.GetOrNewStateObject(cointyp,addr)
	if stateObject != nil {
		stateObject.SetCode(crypto.Keccak256Hash(code), code)
	}
}

func (shard *StateDBManage) SetState(cointyp string,addr common.Address, key, value common.Hash) {

	stateObject := shard.GetOrNewStateObject(cointyp,addr)
	if stateObject != nil {
		stateObject.SetState(shard.db, key, value)
	}
}

func (shard *StateDBManage) SetStateByteArray(cointyp string,addr common.Address, key common.Hash, value []byte) {

	stateObject := shard.GetOrNewStateObject(cointyp,addr)
	if stateObject != nil {
		stateObject.SetStateByteArray(shard.db, key, value)
	}
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (shard *StateDBManage) Suicide(cointyp string,addr common.Address) bool {
	var self *StateDB
	if cointyp!="" {
		self=shard.GetStateDb(cointyp,addr)
		stateObject := self.getStateObject(addr)
		if stateObject == nil {
			return false
		}
		self.journal.append(suicideChange{
			account: &addr,
			prev:    stateObject.suicided,
			//prevbalance: new(big.Int).Set(stateObject.Balance()),
			prevbalance: stateObject.Balance(),
		})
		stateObject.markSuicided()
		//stateObject.data.Balance = new(big.Int)
		stateObject.data.Balance = make(common.BalanceType, 0)
	}else {
	for _,cm:=range shard.shardings {
		for _,rm:=range cm.Rmanage{
			self:=rm.State
			stateObject := self.getStateObject(addr)
			if stateObject == nil {
				return false
			}
			self.journal.append(suicideChange{
				account: &addr,
				prev:    stateObject.suicided,
				//prevbalance: new(big.Int).Set(stateObject.Balance()),
				prevbalance: stateObject.Balance(),
			})
			stateObject.markSuicided()
			//stateObject.data.Balance = new(big.Int)
			stateObject.data.Balance = make(common.BalanceType, 0)
		}
	}	}
	return true
}

//
// Setting, updating & deleting state object methods.
//

// updateStateObject writes the given object to the trie.
//TODO	=======================================================================
func (shard *StateDBManage) updateStateObject(cointyp string,a common.Address,stateObject *stateObject) {
	self:=shard.GetStateDb(cointyp,a)
	addr := stateObject.Address()
	data, err := rlp.EncodeToBytes(stateObject)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	self.setError(self.trie.TryUpdate(addr[:], data))
}
//TODO	=======================================================================
// deleteStateObject removes the given object from the state trie.
func (shard *StateDBManage) deleteStateObject(cointyp string,a common.Address,stateObject *stateObject) {
	self:=shard.GetStateDb(cointyp,a)
	stateObject.deleted = true
	addr := stateObject.Address()
	self.setError(self.trie.TryDelete(addr[:]))
}

// Retrieve a state object given by the address. Returns nil if not found.
func (shard StateDBManage) getStateObject(cointyp string,addr common.Address) (stateObject *stateObject) {
	self:=shard.GetStateDb(cointyp,addr)
	// Prefer 'live' objects.
	if obj := self.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}
		return obj
	}

	// Load the object from the database.
	enc, err := self.trie.TryGet(addr[:])
	if len(enc) == 0 {
		self.setError(err)
		return nil
	}
	var data Account
	if err := rlp.DecodeBytes(enc, &data); err != nil {
		log.Error("Failed to decode state object", "addr", addr, "err", err)
		return nil
	}
	// Insert into the live set.
	obj := newObject(self, addr, data)
	self.setStateObject(obj)
	return obj
}
//TODO	=======================================================================
func (shard *StateDBManage) setStateObject(cointyp string,a common.Address,object *stateObject) {
	shard.GetStateDb(cointyp,a).stateObjects[object.Address()] = object
}

// Retrieve a state object or create a new state object if nil.
func (shard *StateDBManage) GetOrNewStateObject(cointyp string,addr common.Address) *stateObject {
	stateObject := shard.getStateObject(cointyp,addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = shard.createObject(cointyp,addr)
	}
	return stateObject
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (shard *StateDBManage) createObject(cointyp string,addr common.Address) (newobj, prev *stateObject) {
	self:=shard.GetStateDb(cointyp,addr)
	prev = self.getStateObject(addr)
	newobj = newObject(self, addr, Account{})
	newobj.setNonce(0 | params.NonceAddOne) // sets the object to dirty    //YY
	if prev == nil {
		self.journal.append(createObjectChange{account: &addr})
	} else {
		self.journal.append(resetObjectChange{prev: prev})
	}
	self.setStateObject(newobj)
	return newobj, prev
}

// CreateAccount explicitly creates a state object. If a state object with the address
// already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might arise that
// a contract does the following:
//
//   1. sends funds to sha(account ++ (nonce + 1))
//   2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Maner doesn't disappear.
func (shard *StateDBManage) CreateAccount(cointyp string,addr common.Address) {
	new, prev := shard.createObject(cointyp,addr)
	if prev != nil {
		//new.setBalance(prev.data.Balance)
		for _, tAccount := range prev.data.Balance {
			new.setBalance(tAccount.AccountType, tAccount.Balance)
		}
	}
}

func (shard *StateDBManage) ForEachStorage(cointyp string,addr common.Address, cb func(key, value common.Hash) bool) {
	so := shard.getStateObject(cointyp,addr)
	if so == nil {
		return
	}

	// When iterating over the storage check the cache first
	for h, value := range so.cachedStorage {
		cb(h, value)
	}

	it := trie.NewIterator(so.getTrie(shard.db).NodeIterator(nil))
	for it.Next() {
		// ignore cached values
		key := common.BytesToHash(shard.GetStateDb(cointyp,addr).trie.GetKey(it.Key))
		if _, ok := so.cachedStorage[key]; !ok {
			cb(key, common.BytesToHash(it.Value))
		}
	}
}

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
//TODO	这个函数被改写过
func (shard *StateDBManage) Copy(cointyp string,addr common.Address,) *StateDB {
	self:=shard.GetStateDb(cointyp,addr)
	self.lock.Lock()
	defer self.lock.Unlock()

	// Copy all the basic fields, initialize the memory ones
	state := &StateDB{
		db:                self.db,
		trie:              self.db.CopyTrie(self.trie),
		stateObjects:      make(map[common.Address]*stateObject, len(self.journal.dirties)),
		stateObjectsDirty: make(map[common.Address]struct{}, len(self.journal.dirties)),
		btreeMap:          make([]BtreeDietyStruct, 0),
		btreeMapDirty:     make([]BtreeDietyStruct, 0),
		matrixData:        make(map[common.Hash][]byte),
		matrixDataDirty:   make(map[common.Hash][]byte),
		refund:            self.refund,
		logs:              make(map[common.Hash][]*types.Log, len(self.logs)),
		logSize:           self.logSize,
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}
	// Copy the dirty states, logs, and preimages
	for addr := range self.journal.dirties {
		// As documented [here](https://github.com/matrix/go-matrix/pull/16485#issuecomment-380438527),
		// and in the Finalise-method, there is a case where an object is in the journal but not
		// in the stateObjects: OOG after touch on ripeMD prior to Byzantium. Thus, we need to check for
		// nil
		if object, exist := self.stateObjects[addr]; exist {
			state.stateObjects[addr] = object.deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}
	// Above, we don't copy the actual journal. This means that if the copy is copied, the
	// loop above will be a no-op, since the copy's journal is empty.
	// Thus, here we iterate over stateObjects, to enable copies of copies
	for addr := range self.stateObjectsDirty {
		if _, exist := state.stateObjects[addr]; !exist {
			state.stateObjects[addr] = self.stateObjects[addr].deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}

	for hash, logs := range self.logs {
		state.logs[hash] = make([]*types.Log, len(logs))
		copy(state.logs[hash], logs)
	}
	for hash, preimage := range self.preimages {
		state.preimages[hash] = preimage
	}

	//for hash := range self.matrixDataDirty {
	//	if _, exist := state.matrixData[hash]; !exist {
	//		state.stateObjects[addr] = self.matrixData[hash].deepCopy(state)
	//		state.stateObjectsDirty[addr] = struct{}{}
	//	}
	//}
	for hash, mandata := range self.matrixData {
		state.matrixData[hash] = mandata
		state.matrixDataDirty[hash] = mandata
	}

	state.btreeMap = self.btreeMap
	state.btreeMapDirty = self.btreeMapDirty

	return state
}

// Snapshot returns an identifier for the current revision of the state.
func (shard *StateDBManage) Snapshot(cointyp string) map[byte]int {
	ss:=make(map[byte]int,0)
	for _,cm:=range shard.shardings{
		if cm.Cointyp==cointyp{
			for _,rm:=range cm.Rmanage  {
				self:=rm.State
				id:=self.Snapshot()
				ss[rm.Range]=id
			}
		}
	}

	return ss
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (shard *StateDBManage) RevertToSnapshot(cointyp string,ss map[byte]int) {
	// Find the snapshot in the stack of valid snapshots.
	for _,cm:=range shard.shardings{
		if cm.Cointyp==cointyp {
			for _,rm:=range cm.Rmanage {
				id:=ss[rm.Range]
				rm.State.RevertToSnapshot(id)
			}
		}

	}

}

// Finalise finalises the state by removing the self destructed objects
// and clears the journal as well as the refunds.
func (shard *StateDBManage) Finalise(cointyp string,deleteEmptyObjects bool) {
	for _,cm:=range shard.shardings  {
		if cm.Cointyp==cointyp {
			for _,cm:=range cm.Rmanage{
				cm.State.Finalise(deleteEmptyObjects)
			}
		}
	}
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (shard *StateDBManage) IntermediateRoot(deleteEmptyObjects bool) []common.CoinRoot {
	var cr []common.CoinRoot
	var Roots []common.Hash
	for _,cm:=range shard.shardings  {
			for _,rm:=range cm.Rmanage{
				root:=rm.State.IntermediateRoot(deleteEmptyObjects)
				Roots=append(Roots,root)
			}
		bs,err:=json.Marshal(Roots)
		if err!=nil {
			log.Error("file:sharding_statedb.go","func:IntermediateRoot",err)
			panic(err)
		}
		bh:=common.BytesToHash(bs)
		cr=append(cr,common.CoinRoot{
			Cointyp:cm.Cointyp,
			Root:bh,
		})
	}
	return cr
}

// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.
//TODO	=============================================================
func (shard *StateDBManage) Prepare(thash, bhash common.Hash, ti int) {
	for _,cm:=range shard.shardings{
		for _,rm:=range cm.Rmanage  {
			rm.State.Prepare(thash,bhash,ti)
		}
	}

}

func (shard *StateDBManage) clearJournalAndRefund(idx int) {
	//s:=shard.sharding[idx]
	//s.journal = newJournal()
	//s.validRevisions = s.validRevisions[:0]
	//s.refund = 0
}
//TODO	===============================
// Commit writes the state to the underlying in-memory trie database.
func (shard *StateDBManage) Commit(deleteEmptyObjects bool) (cr []common.CoinRoot, err error) {
	var Roots []common.Hash
	for _,cm:=range shard.shardings  {
		for _,rm:=range cm.Rmanage{
			root,err:=rm.State.Commit(deleteEmptyObjects)
			if err!=nil {
				log.Error("file:sharding_statedb.go","func:Commit",err)
				panic(err)
			}
			Roots=append(Roots,root)
		}
		bs,err:=json.Marshal(Roots)
		if err!=nil {
			log.Error("file:sharding_statedb.go","func:Commit",err)
			panic(err)
		}
		bh:=common.BytesToHash(bs)
		cr=append(cr,common.CoinRoot{
			Cointyp:cm.Cointyp,
			Root:bh,
		})
	}
	return cr,nil
}

func (self *StateDBManage) CommitSaveTx() {
	//var typ byte
	//for _, btree := range self.btreeMap {
	//	var hash common.Hash
	//	//var btrie *trie.BTree
	//	log.Info("file statedb", "func CommitSaveTx:Key", btree.Key, "mapData", btree.Data)
	//	switch btree.Typ {
	//	case common.StateDBRevocableBtree:
	//		if len(btree.Data) > 0 {
	//			self.revocablebtrie.ReplaceOrInsert(trie.SpcialTxData{btree.Key, btree.Data})
	//		}
	//		tmproot := self.revocablebtrie.Root()
	//		hash = trie.BtreeSaveHash(tmproot, self.db.TrieDB(), common.ExtraRevocable)
	//		b := []byte(common.StateDBRevocableBtree)
	//		err := self.trie.TryUpdate(b, hash.Bytes())
	//		if err != nil {
	//			log.Error("file statedb", "func CommitSaveTx:err2", err)
	//		}
	//	case common.StateDBTimeBtree:
	//		if len(btree.Data) > 0 {
	//			self.timebtrie.ReplaceOrInsert(trie.SpcialTxData{btree.Key, btree.Data})
	//		}
	//		tmproot := self.timebtrie.Root()
	//		hash = trie.BtreeSaveHash(tmproot, self.db.TrieDB(), common.ExtraTimeTxType)
	//		b := []byte(common.StateDBTimeBtree)
	//		err := self.trie.TryUpdate(b, hash.Bytes())
	//		if err != nil {
	//			log.Error("file statedb", "func CommitSaveTx:err2", err)
	//		}
	//	default:
	//
	//	}
	//}
	//self.btreeMap = make([]BtreeDietyStruct, 0)
	//self.btreeMapDirty = make([]BtreeDietyStruct, 0)
}

func (self *StateDBManage) NewBTrie(typ byte) {
	//switch typ {
	//case common.ExtraRevocable:
	//	self.revocablebtrie = *trie.NewBtree(2, self.db.TrieDB())
	//case common.ExtraTimeTxType:
	//	self.timebtrie = *trie.NewBtree(2, self.db.TrieDB())
	//}
}
//isdel:true 表示需要从map中删除hash，false 表示不需要删除
func (self *StateDBManage) GetSaveTx(typ byte, key uint32, hashlist []common.Hash, isdel bool) {
	//var str string
	//data := make(map[common.Hash][]byte)
	//
	//switch typ {
	//case common.ExtraRevocable:
	//	log.Info("file statedb", "func GetSaveTx:ExtraRevocable", key)
	//	item := self.revocablebtrie.Get(trie.SpcialTxData{key, nil})
	//	std, ok := item.(trie.SpcialTxData)
	//	if !ok {
	//		log.Info("file statedb", "func GetSaveTx:ExtraRevocable", "item is nil")
	//		return
	//	}
	//	self.revocablebtrie.Root().Printree(2)
	//	delitem := self.revocablebtrie.Delete(item)
	//	self.revocablebtrie.Root().Printree(2)
	//
	//	log.Info("file statedb", "revocablebtrie func GetSaveTx:del item key", delitem.(trie.SpcialTxData).Key_Time, "len(delitem.(trie.SpcialTxData).Value_Tx)", len(delitem.(trie.SpcialTxData).Value_Tx))
	//	log.Info("file statedb", "revocablebtrie func GetSaveTx:del item key", std.Key_Time)
	//	if isdel {
	//		log.Info("file statedb", "revocablebtrie func GetSaveTx:del item val:begin", len(std.Value_Tx))
	//		for _, hash := range hashlist {
	//			delete(std.Value_Tx, hash)
	//		}
	//		data = std.Value_Tx
	//		log.Info("file statedb", "revocablebtrie func GetSaveTx:del item val:end", len(std.Value_Tx))
	//	}
	//	str = common.StateDBRevocableBtree
	//case common.ExtraTimeTxType:
	//	log.Info("file statedb", "func GetSaveTx:ExtraTimeTxType:Key", key)
	//	item := self.timebtrie.Get(trie.SpcialTxData{key, nil})
	//	std, ok := item.(trie.SpcialTxData)
	//	if !ok {
	//		log.Info("file statedb", "func GetSaveTx:ExtraTimeTxType", "item is nil")
	//		return
	//	}
	//	self.timebtrie.Root().Printree(2)
	//	delitem := self.timebtrie.Delete(item)
	//	self.timebtrie.Root().Printree(2)
	//
	//	log.Info("file statedb", "timebtrie func GetSaveTx:del item key", delitem.(trie.SpcialTxData).Key_Time, "len(delitem.(trie.SpcialTxData).Value_Tx)", len(delitem.(trie.SpcialTxData).Value_Tx))
	//	log.Info("file statedb", "timebtrie func GetSaveTx:del item key", std.Key_Time)
	//	if isdel {
	//		log.Info("file statedb", "timebtrie func GetSaveTx:del item val:begin", len(std.Value_Tx))
	//		for _, hash := range hashlist {
	//			delete(std.Value_Tx, hash)
	//		}
	//		data = std.Value_Tx
	//		log.Info("file statedb", "timebtrie func GetSaveTx:del item val:end", len(std.Value_Tx))
	//	}
	//	str = common.StateDBTimeBtree
	//default:
	//
	//}
	//var tmpB BtreeDietyStruct
	//tmpB.Typ = str
	//tmpB.Key = key
	//tmpB.Data = data
	//self.btreeMap = append(self.btreeMap, tmpB)
	//var tmpBD BtreeDietyStruct
	//tmpBD.Typ = str
	//tmpBD.Key = key
	//tmpBD.Data = data
	//self.btreeMapDirty = append(self.btreeMapDirty, tmpBD)
	//self.journal.append(addBtreeChange{typ: str, key: key})

	self.CommitSaveTx()
	return
}
func (self *StateDBManage) SaveTx(typ byte, key uint32, data map[common.Hash][]byte) {
	//var str string
	//switch typ {
	//case common.ExtraRevocable:
	//	str = common.StateDBRevocableBtree
	//case common.ExtraTimeTxType:
	//	str = common.StateDBTimeBtree
	//default:
	//
	//}
	//var tmpB BtreeDietyStruct
	//tmpB.Typ = str
	//tmpB.Key = key
	//tmpB.Data = data
	//self.btreeMap = append(self.btreeMap, tmpB)
	//var tmpBD BtreeDietyStruct
	//tmpBD.Typ = str
	//tmpBD.Key = key
	//tmpBD.Data = data
	//self.btreeMapDirty = append(self.btreeMapDirty, tmpBD)
	//self.journal.append(addBtreeChange{typ: str, key: key})
}

func (self *StateDBManage) SetMatrixData(hash common.Hash, val []byte) {
	//self.lock.Lock()
	//defer self.lock.Unlock()
	//self.journal.append(addMatrixDataChange{hash: hash})
	//self.matrixData[hash] = val
	//self.matrixDataDirty[hash] = val
}

func (self *StateDBManage) GetMatrixData(hash common.Hash) (val []byte) {
	//self.lock.Lock()
	//defer self.lock.Unlock()
	//if val = self.matrixData[hash]; val != nil {
	//	return val
	//}
	//
	//// Load the data from the database.
	//val, err := self.trie.TryGet(hash[:])
	//if len(val) == 0 {
	//	self.setError(err)
	//	return nil
	//}
	return
}
func (self *StateDBManage) DeleteMxData(hash common.Hash, val []byte) {
	//self.deleteMatrixData(hash, val)
}

func (self *StateDBManage) UpdateTxForBtree(cointyp string,key uint32){

}
func (self *StateDBManage) UpdateTxForBtreeBytime(cointyp string,key uint32){

}


//根据委托人from和时间获取授权人的from,返回授权人地址(内部调用,仅适用委托gas)
func (self *StateDBManage) GetGasAuthFrom(cointyp string,entrustFrom common.Address, height uint64) common.Address {
	//AuthMarsha1Data := self.GetStateByteArray(entrustFrom, common.BytesToHash(entrustFrom[:]))
	//if len(AuthMarsha1Data) == 0 {
	//	return common.Address{}
	//}
	//AuthDataList := make([]common.AuthType, 0) //授权数据是结构体切片
	//err := json.Unmarshal(AuthMarsha1Data, &AuthDataList)
	//if err != nil {
	//	return common.Address{}
	//}
	//for _, AuthData := range AuthDataList {
	//	if AuthData.EnstrustSetType == params.EntrustByTime && AuthData.IsEntrustGas == true && AuthData.StartTime <= time && AuthData.EndTime >= time {
	//		return AuthData.AuthAddres
	//	}
	//}
	return common.Address{}
}
func (self *StateDBManage) GetAuthFrom(cointyp string,entrustFrom common.Address, height uint64) common.Address {
	//AuthMarsha1Data := self.GetStateByteArray(entrustFrom, common.BytesToHash(entrustFrom[:]))
	//if len(AuthMarsha1Data) == 0 {
	//	return common.Address{}
	//}
	//AuthDataList := make([]common.AuthType, 0) //授权数据是结构体切片
	//err := json.Unmarshal(AuthMarsha1Data, &AuthDataList)
	//if err != nil {
	//	return common.Address{}
	//}
	//for _, AuthData := range AuthDataList {
	//	if AuthData.EnstrustSetType == params.EntrustByHeight && AuthData.IsEntrustSign == true && AuthData.StartHeight <= height && AuthData.EndHeight >= height {
	//		return AuthData.AuthAddres
	//	}
	//}
	return common.Address{}
}
//根据授权人from和高度获取委托人的from列表,返回委托人地址列表(算法组调用,仅适用委托签名)
func (self *StateDBManage) GetEntrustFrom(cointyp string,authFrom common.Address, height uint64) []common.Address {
	//EntrustMarsha1Data := self.GetStateByteArray(authFrom, common.BytesToHash(authFrom[:]))
	//if len(EntrustMarsha1Data) == 0 {
	//	return nil
	//}
	//entrustDataList := make([]common.EntrustType, 0)
	//err := json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	//if err != nil {
	//	return nil
	//}
	//addressList := make([]common.Address, 0)
	//for _, entrustData := range entrustDataList {
	//	if entrustData.EnstrustSetType == params.EntrustByHeight && entrustData.IsEntrustSign == true && entrustData.StartHeight <= height && entrustData.EndHeight >= height {
	//		entrustFrom := base58.Base58DecodeToAddress(entrustData.EntrustAddres) //string地址转0x地址
	//		addressList = append(addressList, entrustFrom)
	//	}
	//} addressList
	return nil
}
//根据授权人获取所有委托签名列表,(该方法用于取消委托时调用)
func (self *StateDBManage) GetAllEntrustSignFrom(cointyp string,authFrom common.Address) []common.Address {
	//EntrustMarsha1Data := self.GetStateByteArray(authFrom, common.BytesToHash(authFrom[:]))
	//entrustDataList := make([]common.EntrustType, 0)
	//err := json.Unmarshal(EntrustMarsha1Data, &entrustDataList)
	//if err != nil {
	//	return nil
	//}
	//addressList := make([]common.Address, 0)
	//for _, entrustData := range entrustDataList {
	//	if entrustData.IsEntrustSign == true {
	//		entrustFrom := base58.Base58DecodeToAddress(entrustData.EntrustAddres) //string地址转0x地址
	//		addressList = append(addressList, entrustFrom)
	//	}
	//}
	return nil //addressList
}

func (self *StateDBManage) GetAllEntrustGasFrom(cointyp string,authFrom common.Address) []common.Address{
	return nil
}