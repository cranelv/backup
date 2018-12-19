package state

import (
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/params"
	"github.com/matrix/go-matrix/trie"
	"math/big"
	"encoding/json"
	"github.com/matrix/go-matrix/base58"
)
type RangeManage struct {
	Range   byte
	State 	*StateDB
}
type CoinManage struct {
	Cointyp string
	Rmanage  []*RangeManage
}
type ShardingStateDB struct {
	db          Database
	trie        Trie
	shardings	[]*CoinManage
	coinRoot    []common.CoinRoot
}

// Create a new state from a given trie.
func NewSharding(roots []common.CoinRoot, db Database) (*ShardingStateDB, error) {
	return &ShardingStateDB{
		db:                db,
		shardings:         make([]*CoinManage,0),
		coinRoot:          roots,
	}, nil
}
func (shard *ShardingStateDB) MakeStatedb(cointyp string,b byte) {
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
	stdb,_ := New(common.BytesToHash(root),shard.db)
	rms := make([]*RangeManage,0)
	rms = append(rms,&RangeManage{Range:b,State:stdb})
	cm := &CoinManage{Cointyp:cointyp,Rmanage:rms}
	shard.shardings = append(shard.shardings,cm)
}

func (shard *ShardingStateDB) Reset(roots []common.CoinRoot) error {
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

func (shard *ShardingStateDB)GetStateDb(cointyp string,address common.Address) *StateDB {
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
}

func (shard *ShardingStateDB) AddLog(cointyp string,address common.Address,log *types.Log) {
	self:=shard.GetStateDb(cointyp,address)
	self.journal.append(addLogChange{txhash: self.thash})

	log.TxHash = self.thash
	log.BlockHash = self.bhash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}

func (shard *ShardingStateDB) GetLogs(cointyp string,address common.Address,hash common.Hash) []*types.Log {
	sd:=shard.GetStateDb(cointyp,address)
	return sd.logs[hash]

}

func (shard *ShardingStateDB) Logs(cointyp string,roots []common.Hash) []*types.Log {
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
func (shard *ShardingStateDB) AddPreimage(cointyp string,hash common.Hash, preimage []byte) {
	//self:=shard.sharding[idx]
	//if _, ok := self.preimages[hash]; !ok {
	//	self.journal.append(addPreimageChange{hash: hash})
	//	pi := make([]byte, len(preimage))
	//	copy(pi, preimage)
	//	self.preimages[hash] = pi
	//}
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (shard *ShardingStateDB) Preimages() map[common.Hash][]byte {
	return nil//shard.sharding[idx].preimages
}

func (shard *ShardingStateDB) AddRefund(gas uint64) {
	//self:=shard.sharding[idx]
	//self.journal.append(refundChange{prev: self.refund})
	//self.refund += gas
}
func (self *ShardingStateDB) GetRefund() uint64 {
	return uint64(0)//self.refund
}
// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (shard *ShardingStateDB) Exist(cointyp string,addr common.Address) bool {
	return shard.getStateObject(addr) != nil
}



// Empty returns whether the state object is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (shard *ShardingStateDB) Empty(cointyp string,addr common.Address) bool {
	so := shard.getStateObject(addr)
	return so == nil || so.empty()
}

// Retrieve the balance from the given address or 0 if object not found
func (shard *ShardingStateDB) GetBalance(cointyp string,addr common.Address) common.BalanceType {

	stateObject := shard.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return nil
}
func (shard *ShardingStateDB)GetBalanceAll(common.Address) common.BalanceType{
	return nil
}
func (self *ShardingStateDB)GetBalanceByType(cointyp string,addr common.Address, accType uint32) *big.Int{
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		for _, tAccount := range stateObject.data.Balance {
			if tAccount.AccountType == accType {
				return tAccount.Balance
			}
		}
	}

	return big.NewInt(0)
}
func (shard *ShardingStateDB) GetNonce(cointyp string,addr common.Address) uint64 {
	stateObject := shard.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0 | params.NonceAddOne //YY
}


func (shard *ShardingStateDB) GetCode(cointyp string,addr common.Address) []byte{

	//self:=shard.sharding[idx]
	//stateObject := shard.getStateObject(addr)
	//if stateObject != nil {
	//	return stateObject.Code(self.db)
	//}
	return nil
}

func (shard *ShardingStateDB) GetCodeSize(cointyp string,addr common.Address) int {

	//self:=shard.sharding[idx]
	//stateObject := shard.getStateObject(addr)
	//if stateObject == nil {
	//	return 0
	//}
	//if stateObject.code != nil {
	//	return len(stateObject.code)
	//}
	//size, err := self.db.ContractCodeSize(stateObject.addrHash, common.BytesToHash(stateObject.CodeHash()))
	//if err != nil {
	//	self.setError(err)
	//}
	return 1 //size
}

func (shard *ShardingStateDB) GetCodeHash(cointyp string,addr common.Address) common.Hash {
	stateObject := shard.getStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

func (shard *ShardingStateDB) GetState(cointyp string,addr common.Address, bhash common.Hash) common.Hash {
	//stateObject := shard.getStateObject(addr)
	//if stateObject != nil {
	//	return stateObject.GetState(shard.sharding[idx].db, bhash)
	//}
	return common.Hash{}
}

func (shard *ShardingStateDB) GetStateByteArray(cointyp string,a common.Address, b common.Hash) []byte {
	//stateObject := shard.getStateObject(a)
	//if stateObject != nil {
	//	return stateObject.GetStateByteArray(shard.sharding[idx].db, b)
	//}
	return nil
}

// Database retrieves the low level database supporting the lower level trie ops.
func (shard *ShardingStateDB) Database(idx int) Database {
	return nil //shard.sharding[idx].db
}

// StorageTrie returns the storage trie of an account.
// The return value is a copy and is nil for non-existent accounts.
func (shard *ShardingStateDB) StorageTrie(addr common.Address) Trie {
	//self:=shard.sharding[idx]
	//stateObject := self.getStateObject(addr)
	//if stateObject == nil {
	//	return nil
	//}
	//cpy := stateObject.deepCopy(self)
	return nil//cpy.updateTrie(self.db)
}

func (shard *ShardingStateDB) HasSuicided(addr common.Address) bool {
	//stateObject := shard.sharding[idx].getStateObject(addr)
	//if stateObject != nil {
	//	return stateObject.suicided
	//}
	return false
}

/*
 * SETTERS
 */

// AddBalance adds amount to the account associated with addr.
func (shard *ShardingStateDB) AddBalance(cointyp string,accountType uint32, addr common.Address, amount *big.Int) {
	stateObject := shard.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(accountType, amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (shard *ShardingStateDB) SubBalance(cointyp string,idx uint32,addr common.Address,am *big.Int) {
	stateObject := shard.GetOrNewStateObject(addr)
	if stateObject != nil {
		//stateObject.SubBalance(accountType, amount)
	}
}

func (shard *ShardingStateDB) SetBalance(cointyp string,idx uint32, addr common.Address, am *big.Int) {
	stateObject := shard.GetOrNewStateObject(addr)
	if stateObject != nil {
		//stateObject.SetBalance(accountType, amount)
	}
}

func (shard *ShardingStateDB) SetNonce(cointyp string,addr common.Address, nonce uint64) {
	stateObject := shard.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce | params.NonceAddOne) //YY
	}
}

func (shard *ShardingStateDB) SetCode(cointyp string,addr common.Address, code []byte) {
	stateObject := shard.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(crypto.Keccak256Hash(code), code)
	}
}

func (shard *ShardingStateDB) SetState(cointyp string,addr common.Address, key, value common.Hash) {

	//stateObject := shard.GetOrNewStateObject(addr)
	//if stateObject != nil {
	//	stateObject.SetState(shard.sharding[idx].db, key, value)
	//}
}

func (shard *ShardingStateDB) SetStateByteArray(cointyp string,addr common.Address, key common.Hash, value []byte) {
	//
	//stateObject := shard.GetOrNewStateObject(addr)
	//if stateObject != nil {
	//	stateObject.SetStateByteArray(shard.sharding[idx].db, key, value)
	//}
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (shard *ShardingStateDB) Suicide(addr common.Address) bool {
	//self:=shard.sharding[idx]
	//stateObject := self.getStateObject(addr)
	//if stateObject == nil {
	//	return false
	//}
	//self.journal.append(suicideChange{
	//	account: &addr,
	//	prev:    stateObject.suicided,
	//	//prevbalance: new(big.Int).Set(stateObject.Balance()),
	//	prevbalance: stateObject.Balance(),
	//})
	//stateObject.markSuicided()
	////stateObject.data.Balance = new(big.Int)
	//stateObject.data.Balance = make(common.BalanceType, 0)

	return true
}

//
// Setting, updating & deleting state object methods.
//

// updateStateObject writes the given object to the trie.
func (shard *ShardingStateDB) updateStateObject(stateObject *stateObject) {
	//self:=shard.sharding[idx]
	//addr := stateObject.Address()
	//data, err := rlp.EncodeToBytes(stateObject)
	//if err != nil {
	//	panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	//}
	//self.setError(self.trie.TryUpdate(addr[:], data))
}

// deleteStateObject removes the given object from the state trie.
func (shard *ShardingStateDB) deleteStateObject(stateObject *stateObject) {
	//self:=shard.sharding[idx]
	//stateObject.deleted = true
	//addr := stateObject.Address()
	//self.setError(self.trie.TryDelete(addr[:]))
}

// Retrieve a state object given by the address. Returns nil if not found.
func (shard ShardingStateDB) getStateObject(addr common.Address) (stateObject *stateObject) {
	//self:=shard.sharding[idx]
	//// Prefer 'live' objects.
	//if obj := self.stateObjects[addr]; obj != nil {
	//	if obj.deleted {
	//		return nil
	//	}
	//	return obj
	//}
	//
	//// Load the object from the database.
	//enc, err := self.trie.TryGet(addr[:])
	//if len(enc) == 0 {
	//	self.setError(err)
	//	return nil
	//}
	//var data Account
	//if err := rlp.DecodeBytes(enc, &data); err != nil {
	//	log.Error("Failed to decode state object", "addr", addr, "err", err)
	//	return nil
	//}
	//// Insert into the live set.
	//obj := newObject(self, addr, data)
	//self.setStateObject(obj)
	return nil
}

func (shard *ShardingStateDB) setStateObject(object *stateObject) {
	//shard.sharding[idx].stateObjects[object.Address()] = object
}

// Retrieve a state object or create a new state object if nil.
func (shard *ShardingStateDB) GetOrNewStateObject(addr common.Address) *stateObject {
	//self:=shard.sharding[idx]
	//stateObject := self.getStateObject(addr)
	//if stateObject == nil || stateObject.deleted {
	//	stateObject, _ = self.createObject(addr)
	//}
	return nil
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (shard *ShardingStateDB) createObject(addr common.Address) (newobj, prev *stateObject) {
	//self:=shard.sharding[idx]
	//prev = self.getStateObject(addr)
	//newobj = newObject(self, addr, Account{})
	//newobj.setNonce(0 | params.NonceAddOne) // sets the object to dirty    //YY
	//if prev == nil {
	//	self.journal.append(createObjectChange{account: &addr})
	//} else {
	//	self.journal.append(resetObjectChange{prev: prev})
	//}
	//self.setStateObject(newobj)
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
func (shard *ShardingStateDB) CreateAccount(cointyp string,addr common.Address) {
	//self:=shard.sharding[idx]
	//new, prev := self.createObject(addr)
	//if prev != nil {
	//	//new.setBalance(prev.data.Balance)
	//	for _, tAccount := range prev.data.Balance {
	//		new.setBalance(tAccount.AccountType, tAccount.Balance)
	//	}
	//}
}

func (shard *ShardingStateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) {
	//db:=shard.sharding[idx]
	//so := db.getStateObject(addr)
	//if so == nil {
	//	return
	//}
	//
	//// When iterating over the storage check the cache first
	//for h, value := range so.cachedStorage {
	//	cb(h, value)
	//}
	//
	//it := trie.NewIterator(so.getTrie(db.db).NodeIterator(nil))
	//for it.Next() {
	//	// ignore cached values
	//	key := common.BytesToHash(db.trie.GetKey(it.Key))
	//	if _, ok := so.cachedStorage[key]; !ok {
	//		cb(key, common.BytesToHash(it.Value))
	//	}
	//}
}

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
func (shard *ShardingStateDB) Copy(idx int) *StateDB {
	//self:=shard.sharding[idx]
	//self.lock.Lock()
	//defer self.lock.Unlock()
	//
	//// Copy all the basic fields, initialize the memory ones
	//state := &StateDB{
	//	db:                self.db,
	//	trie:              self.db.CopyTrie(self.trie),
	//	stateObjects:      make(map[common.Address]*stateObject, len(self.journal.dirties)),
	//	stateObjectsDirty: make(map[common.Address]struct{}, len(self.journal.dirties)),
	//	refund:            self.refund,
	//	logs:              make(map[common.Hash][]*types.Log, len(self.logs)),
	//	logSize:           self.logSize,
	//	preimages:         make(map[common.Hash][]byte),
	//	journal:           newJournal(),
	//}
	//// Copy the dirty states, logs, and preimages
	//for addr := range self.journal.dirties {
	//	// As documented [here](https://github.com/matrix/go-matrix/pull/16485#issuecomment-380438527),
	//	// and in the Finalise-method, there is a case where an object is in the journal but not
	//	// in the stateObjects: OOG after touch on ripeMD prior to Byzantium. Thus, we need to check for
	//	// nil
	//	if object, exist := self.stateObjects[addr]; exist {
	//		state.stateObjects[addr] = object.deepCopy(state)
	//		state.stateObjectsDirty[addr] = struct{}{}
	//	}
	//}
	//// Above, we don't copy the actual journal. This means that if the copy is copied, the
	//// loop above will be a no-op, since the copy's journal is empty.
	//// Thus, here we iterate over stateObjects, to enable copies of copies
	//for addr := range self.stateObjectsDirty {
	//	if _, exist := state.stateObjects[addr]; !exist {
	//		state.stateObjects[addr] = self.stateObjects[addr].deepCopy(state)
	//		state.stateObjectsDirty[addr] = struct{}{}
	//	}
	//}
	//
	//for hash, logs := range self.logs {
	//	state.logs[hash] = make([]*types.Log, len(logs))
	//	copy(state.logs[hash], logs)
	//}
	//for hash, preimage := range self.preimages {
	//	state.preimages[hash] = preimage
	//}
	return nil
}

// Snapshot returns an identifier for the current revision of the state.
func (shard *ShardingStateDB) Snapshot(cointyp string) int {
	//self:=shard.sharding[idx]
	//id := self.nextRevisionId
	//self.nextRevisionId++
	//self.validRevisions = append(self.validRevisions, revision{id, self.journal.length()})
	return 1
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (shard *ShardingStateDB) RevertToSnapshot(cointyp string,revid int) {
	// Find the snapshot in the stack of valid snapshots.
	//self:=shard.sharding[idx]
	//idx := sort.Search(len(self.validRevisions), func(i int) bool {
	//	return self.validRevisions[i].id >= revid
	//})
	//if idx == len(self.validRevisions) || self.validRevisions[idx].id != revid {
	//	panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	//}
	//snapshot := self.validRevisions[idx].journalIndex
	//
	//// Replay the journal to undo changes and remove invalidated snapshots
	//self.journal.revert(self, snapshot)
	//self.validRevisions = self.validRevisions[:idx]
}


// Finalise finalises the state by removing the self destructed objects
// and clears the journal as well as the refunds.
func (shard *ShardingStateDB) Finalise(deleteEmptyObjects bool) {
	//s:=shard.sharding[idx]
	//for addr := range s.journal.dirties {
	//	stateObject, exist := s.stateObjects[addr]
	//	if !exist {
	//		// ripeMD is 'touched' at block 1714175, in tx 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
	//		// That tx goes out of gas, and although the notion of 'touched' does not exist there, the
	//		// touch-event will still be recorded in the journal. Since ripeMD is a special snowflake,
	//		// it will persist in the journal even though the journal is reverted. In this special circumstance,
	//		// it may exist in `s.journal.dirties` but not in `s.stateObjects`.
	//		// Thus, we can safely ignore it here
	//		continue
	//	}
	//
	//	if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
	//		s.deleteStateObject(stateObject)
	//	} else {
	//		stateObject.updateRoot(s.db)
	//		s.updateStateObject(stateObject)
	//	}
	//	s.stateObjectsDirty[addr] = struct{}{}
	//}
	//// Invalidate journal because reverting across transactions is not allowed.
	//s.clearJournalAndRefund()
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (shard *ShardingStateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	//s:=shard.sharding[idx]
	//s.Finalise(deleteEmptyObjects)
	return common.Hash{}
}

// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.
func (shard *ShardingStateDB) Prepare(cointyp string,thash, bhash common.Hash, ti int) {
	//self:=shard.sharding[idx]
	//self.thash = thash
	//self.bhash = bhash
	//self.txIndex = ti
}

func (shard *ShardingStateDB) clearJournalAndRefund(idx int) {
	//s:=shard.sharding[idx]
	//s.journal = newJournal()
	//s.validRevisions = s.validRevisions[:0]
	//s.refund = 0
}

// Commit writes the state to the underlying in-memory trie database.
func (shard *ShardingStateDB) Commit(deleteEmptyObjects bool) (root common.Hash, err error) {
	//s:=shard.sharding[idx]
	//defer s.clearJournalAndRefund()
	//
	//for addr := range s.journal.dirties {
	//	s.stateObjectsDirty[addr] = struct{}{}
	//}
	//// Commit objects to the trie.
	//for addr, stateObject := range s.stateObjects {
	//	_, isDirty := s.stateObjectsDirty[addr]
	//	switch {
	//	case stateObject.suicided || (isDirty && deleteEmptyObjects && stateObject.empty()):
	//		// If the object has been removed, don't bother syncing it
	//		// and just mark it for deletion in the trie.
	//		s.deleteStateObject(stateObject)
	//	case isDirty:
	//		// Write any contract code associated with the state object
	//		if stateObject.code != nil && stateObject.dirtyCode {
	//			s.db.TrieDB().Insert(common.BytesToHash(stateObject.CodeHash()), stateObject.code)
	//			stateObject.dirtyCode = false
	//		}
	//		// Write any storage changes in the state object to its storage trie.
	//		if err := stateObject.CommitTrie(s.db); err != nil {
	//			return common.Hash{}, err
	//		}
	//		// Update the object in the main account trie.
	//		s.updateStateObject(stateObject)
	//	}
	//	delete(s.stateObjectsDirty, addr)
	//}
	//// Write trie changes.
	//root, err = s.trie.Commit(func(leaf []byte, parent common.Hash) error {
	//	var account Account
	//	if err := rlp.DecodeBytes(leaf, &account); err != nil {
	//		return nil
	//	}
	//	if account.Root != emptyState {
	//		s.db.TrieDB().Reference(account.Root, parent)
	//	}
	//	code := common.BytesToHash(account.CodeHash)
	//	if code != emptyCode {
	//		s.db.TrieDB().Reference(code, parent)
	//	}
	//	return nil
	//})
	log.Debug("Trie cache stats after commit", "misses", trie.CacheMisses(), "unloads", trie.CacheUnloads())
	return
}

func (self *ShardingStateDB) CommitSaveTx() {
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

func (self *ShardingStateDB) NewBTrie(typ byte) {
	//switch typ {
	//case common.ExtraRevocable:
	//	self.revocablebtrie = *trie.NewBtree(2, self.db.TrieDB())
	//case common.ExtraTimeTxType:
	//	self.timebtrie = *trie.NewBtree(2, self.db.TrieDB())
	//}
}
//isdel:true 表示需要从map中删除hash，false 表示不需要删除
func (self *ShardingStateDB) GetSaveTx(typ byte, key uint32, hashlist []common.Hash, isdel bool) {
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
func (self *ShardingStateDB) SaveTx(typ byte, key uint32, data map[common.Hash][]byte) {
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

func (self *ShardingStateDB) SetMatrixData_sh(hash common.Hash, val []byte) {
	//self.lock.Lock()
	//defer self.lock.Unlock()
	//self.journal.append(addMatrixDataChange{hash: hash})
	//self.matrixData[hash] = val
	//self.matrixDataDirty[hash] = val
}

func (self *ShardingStateDB) GetMatrixData_sh(hash common.Hash) (val []byte) {
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
func (self *ShardingStateDB) DeleteMxData_sh(hash common.Hash, val []byte) {
	//self.deleteMatrixData(hash, val)
}
//根据委托人from和时间获取授权人的from,返回授权人地址(内部调用,仅适用委托gas)
func (self *ShardingStateDB) GetGasAuthFrom(cointyp string,entrustFrom common.Address, height uint64) common.Address {
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
func (self *ShardingStateDB) GetAuthFrom(cointyp string,entrustFrom common.Address, height uint64) common.Address {
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
func (self *ShardingStateDB) GetEntrustFrom(cointyp string,authFrom common.Address, height uint64) []common.Address {
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
func (self *ShardingStateDB) GetAllEntrustSignFrom(cointyp string,authFrom common.Address) []common.Address {
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

func (self *ShardingStateDB) GetAllEntrustGasFrom(cointyp string,authFrom common.Address) []common.Address{

}