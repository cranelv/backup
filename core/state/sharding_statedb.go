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
	"github.com/matrix/go-matrix/mandb"
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
	mdb		mandb.Database
	shardings	[]*CoinManage
	coinRoot    []common.CoinRoot
	retcoinRoot []common.CoinRoot
}

// Create a new state from a given trie.
func NewStateDBManage(roots []common.CoinRoot, mdb mandb.Database,db Database) (*StateDBManage, error) {

	if len(roots) == 0{
		roots = append(roots,common.CoinRoot{Cointyp:params.MAN_COIN,Root:common.Hash{}})
		roots = append(roots,common.CoinRoot{Cointyp:params.BTC_COIN,Root:common.Hash{}})//YYYYYYYYYYYYYYYYYYYYYYYY
	}
	stm := &StateDBManage{
		mdb:			   mdb,
		db:                db,
		shardings:         make([]*CoinManage,0),
		coinRoot:          make([]common.CoinRoot,len(roots)),
		retcoinRoot:       make([]common.CoinRoot,len(roots)),
	}
	copy(stm.coinRoot,roots)
	copy(stm.retcoinRoot,roots)
	stm.MakeStatedb(params.MAN_COIN)
	stm.MakeStatedb(params.BTC_COIN)//YYYYYYYYYYYYYYYYYYYYYYYY
	return stm, nil
}
func (shard *StateDBManage) MakeStatedb(cointyp string) {
	//没有对应币种或byte分区的时候，才创建
	for _,cm := range shard.shardings{
		if cm.Cointyp == cointyp{
			return
		}
	}
	isex := false
	for _,cr := range shard.coinRoot {
		if cr.Cointyp == cointyp {
			isex = true
		}
	}
	var types  []string
	if !isex && cointyp!=params.MAN_COIN{

		for _,cm:=range shard.shardings{
			if cm.Cointyp==params.MAN_COIN {
				v:=cm.Rmanage[0].State.trie.GetKey([]byte(params.COIN_NAME))
				err:=json.Unmarshal(v,&types)
				if err!=nil {
					log.Error("")
				}
				flag:=true
				for _,coinName:= range types  {
					if coinName==cointyp{
						flag=false
						 break
					}
				}
				if flag {
					return
				}
				break
			}
		}

		shard.coinRoot = append(shard.coinRoot,common.CoinRoot{Cointyp:cointyp,Root:common.Hash{}})
		shard.retcoinRoot = append(shard.retcoinRoot,common.CoinRoot{Cointyp:cointyp,Root:common.Hash{}})
	}
	shard.addShardings(cointyp)
}
func (shard *StateDBManage) addShardings(cointyp string){
	//获取指定的币种root
	for _,cr := range shard.coinRoot{
		if cr.Cointyp == cointyp{
			rms := make([]*RangeManage,0)
			var hashs []common.Hash
			Roots,err:=shard.mdb.Get(cr.Root[:])
			if err!=nil {
				log.Error("file sharding_statedb", "func addShardings:Get", err)
			}else {
				err=rlp.DecodeBytes(Roots,&hashs)
				if err!=nil{
					log.Error("file sharding_statedb", "func addShardings:DecodeBytes", err)
				}
			}
			if len(hashs) <=0{
				for idx:=0;idx<params.RANGE_MOUNTS;idx++  {
					hashs = append(hashs,common.Hash{})
				}
			}
			for i,hash:=range hashs{
				stdb,err := newStatedb(hash,shard.db)
				if err != nil{
					log.Info("file sharding_statedb","func addShardings:newStatedb:err",err)
				}
				rms = append(rms,&RangeManage{Range:byte(i),State:stdb})
			}
			cmg := &CoinManage{Cointyp:cointyp,Rmanage:rms}
			shard.shardings = append(shard.shardings,cmg)
			break
		}
	}
}
func (shard *StateDBManage) Reset(roots []common.CoinRoot) error {

	for _,cr := range roots{
		//var hashs []common.Hash
		//json.Unmarshal(cr.Root,&hashs)
		for _,cm := range shard.shardings{
			if cm.Cointyp == cr.Cointyp{
				for _,rm := range cm.Rmanage{
					rm.State.Reset(rm.State.trie.Hash())
				}
				break
			}
		}
	}
	return nil
}

func (shard *StateDBManage)GetStateDb(cointyp string,address common.Address) *StateDB {

	if cointyp == ""{
		cointyp = params.MAN_COIN
	}
	cms := shard.shardings
	for _, cm := range cms {
		if cm.Cointyp == cointyp {
			rms := cm.Rmanage
			for _, rm := range rms {
				if rm.Range == address[0] {
					return rm.State
				}
			}
			break
		}
	}
	return nil
}

func (shard *StateDBManage) setError(cointype string,addr common.Address, err error) {

	self:=shard.GetStateDb(cointype,addr)
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (shard *StateDBManage) Error() error {

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

func (shard *StateDBManage) Logs() []*types.Log {

	cms:=shard.shardings
	var logs []*types.Log
	for _,cm:=range cms {
		//if cm.Cointyp==cointyp {
			rms:=cm.Rmanage
			for _,rm :=range rms{
				log:=rm.State.logs
				for _,l:=range log  {
					logs=append(logs,l...)
				}}
			break
		} //}
	return logs
}

// AddPreimage records a SHA3 preimage seen by the VM.
func (shard *StateDBManage) AddPreimage(cointype string,addr common.Address,hash common.Hash, preimage []byte) {

	state:=shard.GetStateDb(cointype,addr)
	state.AddPreimage(hash,preimage)
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (shard *StateDBManage) Preimages() map[string]map[common.Hash][]byte {

	var mm =make(map[string](map[common.Hash][]byte),0)
	for _,cm:=range shard.shardings{
		for _,rm:=range cm.Rmanage{
			mm[cm.Cointyp]=rm.State.preimages
		}
	}
	return mm
}

func (shard *StateDBManage) AddRefund(cointyp string,address common.Address,gas uint64) {

	shard.GetStateDb(cointyp,address).AddRefund(gas)
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
			prevbalance: stateObject.Balance(),
		})
		stateObject.markSuicided()
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
				prevbalance: stateObject.Balance(),
			})
			stateObject.markSuicided()
			stateObject.data.Balance = make(common.BalanceType, 0)
		}
	}	}
	return true
}

//
// Setting, updating & deleting state object methods.
//

// updateStateObject writes the given object to the trie.

func (shard *StateDBManage) updateStateObject(cointyp string,a common.Address,stateObject *stateObject) {

	self:=shard.GetStateDb(cointyp,a)
	addr := stateObject.Address()
	data, err := rlp.EncodeToBytes(stateObject)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	self.setError(self.trie.TryUpdate(addr[:], data))
}

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
	if self == nil{
		shard.MakeStatedb(cointyp)
		self=shard.GetStateDb(cointyp,addr)
		if self == nil{
			log.Error("file","func getStateObject:err","Unknown Coin Type")
			return nil
		}
	}
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

func (shard *StateDBManage) Copy() *StateDBManage {

	state := &StateDBManage{
		db:                shard.db,
		mdb:               shard.mdb,
		shardings:		   make([]*CoinManage,0),
		coinRoot:		   make([]common.CoinRoot,0),
	}
	for _,cm:=range shard.shardings  {
		var rms	[]*RangeManage
		for _,rm:=range cm.Rmanage  {
			sd:=rm.State.Copy()
			rms=append(rms,&RangeManage{
				Range:	rm.Range,
				State:	sd,
			})
		}
		state.shardings=append(state.shardings,&CoinManage{
			Cointyp:	cm.Cointyp,
			Rmanage:	rms,
		})
	}
	for _,root:=range shard.coinRoot  {
		state.coinRoot=append(state.coinRoot,common.CoinRoot{
			Root:		root.Root,
			Cointyp:	root.Cointyp,
		})
	}
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
			break
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
			break
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
			break
		}
	}
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.

func (shard *StateDBManage) IntermediateRoot(deleteEmptyObjects bool) ([]common.CoinRoot,[]common.Coinbyte){

	coinbytes :=make([]common.Coinbyte,0)
	for _,cm:=range shard.shardings  {
		var bshash	common.Hash
		root256 :=make([]common.Hash,0)
		for _,rm:=range cm.Rmanage{
			root:=rm.State.IntermediateRoot(deleteEmptyObjects)
			root256=append(root256,root)
		}
		bshash=types.RlpHash(root256)
		bs,_ := rlp.EncodeToBytes(root256)
		err := shard.mdb.Put(bshash[:],bs)
		if err!=nil {
			log.Error("file:sharding_statedb.go","func:IntermediateRoot",err)
			panic(err)
		}
		isex := false
		//log.Info("YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY11111","coin",cm.Cointyp,"256Hash",bshash)
		for i,croot := range shard.retcoinRoot{
			if croot.Cointyp == cm.Cointyp{
				//log.Info("YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY22222","coin",cm.Cointyp,"256Hash",bshash,"shard.retcoinRoot[i].Root",shard.retcoinRoot[i].Root)
				shard.retcoinRoot[i].Root = bshash
				isex = true
				break
			}
		}
		if !isex{
			shard.retcoinRoot = append(shard.retcoinRoot,common.CoinRoot{Cointyp:cm.Cointyp,Root:bshash})
		}
		coinbytes=append(coinbytes,common.Coinbyte{Root:bshash,Byte256:root256})
	}
	return shard.retcoinRoot,coinbytes
}

func (shard *StateDBManage) IntermediateRootByCointype(cointype string,deleteEmptyObjects bool) common.Hash {

	var root256 []common.Hash
	for _,cm:=range shard.shardings {
		if cointype == cm.Cointyp {
		for _, rm := range cm.Rmanage {
			root := rm.State.IntermediateRoot(deleteEmptyObjects)
			root256 = append(root256, root)
		}
		bshash := types.RlpHash(root256)
		bs, _ := rlp.EncodeToBytes(root256)
		err := shard.mdb.Put(bshash[:], bs)
		if err != nil {
			log.Error("file:sharding_statedb.go", "func:IntermediateRoot", err)
			panic(err)
		}
		isex := false
		for i, croot := range shard.retcoinRoot {
			if croot.Cointyp == cm.Cointyp {
				shard.retcoinRoot[i].Root = bshash
				isex = true
				break
			}
		}
		if !isex {
			shard.retcoinRoot = append(shard.retcoinRoot, common.CoinRoot{Cointyp: cm.Cointyp, Root: bshash})
		}
	}
	}
	return types.RlpHash(root256)
}
// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.

func (shard *StateDBManage) Prepare(thash, bhash common.Hash, ti int) {

	for _,cm:=range shard.shardings{
		for _,rm:=range cm.Rmanage  {
			rm.State.Prepare(thash,bhash,ti)
		}
	}
}

func (shard *StateDBManage) clearJournalAndRefund() {

	for _,cm:=range shard.shardings  {
		for _,rm:=range cm.Rmanage  {
			rm.State.clearJournalAndRefund()
		}
	}
}

// Commit writes the state to the underlying in-memory trie database.
func (shard *StateDBManage) Commit(deleteEmptyObjects bool) ([]common.CoinRoot,[]common.Coinbyte, error) {

	var coinbytes =make([] common.Coinbyte,0)
	for _,cm:=range shard.shardings  {
		var roots =make([]common.Hash,0)
		for _,rm:=range cm.Rmanage{
			root,err:=rm.State.Commit(deleteEmptyObjects)
			if err!=nil {
				log.Error("file:sharding_statedb.go","func:Commit",err)
				panic(err)
			}
			roots=append(roots,root)
		}
		bshash:=types.RlpHash(roots)
		bs,err:=rlp.EncodeToBytes(roots)
		err=shard.mdb.Put(bshash[:],bs)
		if err!=nil {
			log.Error("file:sharding_statedb.go","func:Commit",err)
			panic(err)
		}
		isex := false
		for i,croot := range shard.retcoinRoot{
			if croot.Cointyp == cm.Cointyp{
				shard.retcoinRoot[i].Root = bshash
				isex = true
				break
			}
		}
		if !isex{
			shard.retcoinRoot = append(shard.retcoinRoot,common.CoinRoot{Cointyp:cm.Cointyp,Root:bshash})
		}
		coinbytes=append(coinbytes,common.Coinbyte{Root:bshash,Byte256:roots})
	}
	return shard.retcoinRoot,coinbytes,nil
}

func (self *StateDBManage) CommitSaveTx(cointyp string,addr common.Address) {

	for _,cm:=range self.shardings{
		if cm.Cointyp==cointyp {
		for _,rm:=range cm.Rmanage{
			if rm.Range==addr[0]{
				rm.State.CommitSaveTx()
				break
			}
		}
		break
		}
	}

}

func (self *StateDBManage) NewBTrie(cointyp string,addr common.Address,typ byte) {

	for _,cm:=range self.shardings{
		if cm.Cointyp==cointyp {
			for _,rm:=range cm.Rmanage{
				if rm.Range==addr[0]{
					rm.State.NewBTrie(typ)
					break
				}
			}
			break
		}
	}
}
//isdel:true 表示需要从map中删除hash，false 表示不需要删除
func (self *StateDBManage) GetSaveTx(cointyp string,addr common.Address,typ byte, key uint32, hashlist []common.Hash, isdel bool) {

	for _,cm:=range self.shardings{
		if cm.Cointyp==cointyp {
			for _,rm:=range cm.Rmanage{
				if rm.Range==addr[0]{
					rm.State.GetSaveTx(typ,key,hashlist,isdel)
					break
				}
			}
			break
		}
	}
}
func (self *StateDBManage) SaveTx(cointyp string,addr common.Address,typ byte, key uint32, data map[common.Hash][]byte) {

	for _,cm:=range self.shardings{
		if cm.Cointyp==cointyp {
			for _,rm:=range cm.Rmanage{
				if rm.Range==addr[0]{
					rm.State.SaveTx(typ,key,data)
					break
				}
			}
			break
		}
	}
}


//SetMatrixData，GetMatrixData，DeleteMxData都是针对man币种 分区[0]
func (self *StateDBManage) SetMatrixData(hash common.Hash, val []byte) {

	for _,cm:=range self.shardings {
		if cm.Cointyp==params.MAN_COIN{
			cm.Rmanage[0].State.SetMatrixData(hash,val)
			break
		}
	}
}

func (self *StateDBManage) GetMatrixData(hash common.Hash) (val []byte) {

	for _,cm:=range self.shardings {
		if cm.Cointyp==params.MAN_COIN{
			return cm.Rmanage[0].State.GetMatrixData(hash)
			break
		}
	}
	return
}

func (self *StateDBManage) DeleteMxData(hash common.Hash, val []byte) {

	for _,cm:=range self.shardings {
		if cm.Cointyp==params.MAN_COIN{
			cm.Rmanage[0].State.deleteMatrixData(hash,val)
			break
		}
	}
}

func (self *StateDBManage) UpdateTxForBtree(key uint32){

	for _,cm:=range self.shardings{
		for _,rm:=range cm.Rmanage  {
			rm.State.UpdateTxForBtree(key)
		}
	}
}
func (self *StateDBManage) UpdateTxForBtreeBytime(key uint32){

	for _,cm:=range self.shardings{
		for _,rm:=range cm.Rmanage  {
			rm.State.UpdateTxForBtreeBytime(key)
		}
	}
}
//根据委托人from和时间获取授权人的from,返回授权人地址(内部调用,仅适用委托gas)
func (self *StateDBManage) GetGasAuthFromByTime(cointyp string,entrustFrom common.Address, time uint64) common.Address {

	AuthMarsha1Data := self.GetStateByteArray(cointyp,entrustFrom, common.BytesToHash(entrustFrom[:]))
	if len(AuthMarsha1Data) == 0 {
		return common.Address{}
	}
	AuthDataList := make([]common.AuthType, 0) //授权数据是结构体切片
	err := json.Unmarshal(AuthMarsha1Data, &AuthDataList)
	if err != nil {
		return common.Address{}
	}
	for _, AuthData := range AuthDataList {
		if AuthData.EnstrustSetType == params.EntrustByTime && AuthData.IsEntrustGas == true && AuthData.StartTime <= time && AuthData.EndTime >= time {
			return AuthData.AuthAddres
		}
	}
	return common.Address{}
}

//根据委托人from和时间获取授权人的from,返回授权人地址(内部调用,仅适用委托gas)
func (self *StateDBManage) GetGasAuthFrom(cointyp string,entrustFrom common.Address, height uint64) common.Address {

	return self.GetStateDb(cointyp,entrustFrom).GetGasAuthFrom(entrustFrom,height)
}
func (self *StateDBManage) GetAuthFrom(cointyp string,entrustFrom common.Address, height uint64) common.Address {

	return self.GetStateDb(cointyp,entrustFrom).GetAuthFrom(entrustFrom,height)
}
//根据授权人from和高度获取委托人的from列表,返回委托人地址列表(算法组调用,仅适用委托签名)
func (self *StateDBManage) GetEntrustFrom(cointyp string,authFrom common.Address, height uint64) []common.Address {

	return self.GetStateDb(cointyp,authFrom).GetEntrustFrom(authFrom,height)
}
//根据授权人获取所有委托签名列表,(该方法用于取消委托时调用)
func (self *StateDBManage) GetAllEntrustSignFrom(cointyp string,authFrom common.Address) []common.Address {

	return self.GetStateDb(cointyp,authFrom).GetAllEntrustSignFrom(authFrom)
}

func (self *StateDBManage) GetAllEntrustGasFrom(cointyp string,authFrom common.Address) []common.Address{

	return self.GetStateDb(cointyp,authFrom).GetAllEntrustGasFrom(authFrom)
}

func (self *StateDBManage) GetEntrustFromByTime(cointyp string,authFrom common.Address, time uint64) []common.Address {


	return self.GetStateDb(cointyp,authFrom).GetEntrustFromByTime(authFrom,time)
}

//判断根据时间委托是否满足条件，用于执行按时间委托的交易(跑交易),此处time应该为header里的时间戳
func (self *StateDBManage) GetIsEntrustByTime(cointyp string,entrustFrom common.Address, time uint64) bool {

	return self.GetStateDb(cointyp,entrustFrom).GetIsEntrustByTime(entrustFrom,time)
}

//钱包调用显示
func (self *StateDBManage) GetAllEntrustList(cointyp string,authFrom common.Address) []common.EntrustType {

	return  self.GetStateDb(cointyp,authFrom).GetAllEntrustList(authFrom)
}

func (self *StateDBManage)RawDump(cointype string,address common.Address)Dump {

	for _,cm:=range self.shardings  {
		if cointype==cm.Cointyp {
		for _,rm:=range cm.Rmanage{
			if rm.Range==address[0] {
				return 	rm.State.RawDump()
			}
		}
		break
		}
	}
	return Dump{}

}

func (self *StateDBManage)Dump(cointype string,address common.Address) []byte {

	json, err := json.MarshalIndent(self.RawDump(cointype,address), "", "    ")
	if err != nil {
		fmt.Println("dump err", err)
	}
	return json

}
func (self *StateDBManage)RawDumpAcccount(cointype string,address common.Address) Dump {

	var dump Dump
	for _,cm:=range self.shardings  {
		if cm.Cointyp==cointype {
			for _,rm:=range cm.Rmanage  {
				if rm.Range==address[0] {
					dump=rm.State.RawDumpAcccount(address)
					break
				}
			}
		}
	}
	return dump

}