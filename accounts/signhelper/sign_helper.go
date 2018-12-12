// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package signhelper

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/core/state"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/pkg/errors"

	"sync"

	"github.com/matrix/go-matrix/ca"
	"github.com/matrix/go-matrix/log"
)

type MatrixEth interface {
	BlockChain() *core.BlockChain
}

var (
	ErrNilAccountManager     = errors.New("account manager is nil")
	ErrEmptySignAddress      = errors.New("sign address is empty")
	ErrUnSetSignAccount      = errors.New("The sign account not set yet!")
	ErrBlockChain            = errors.New("err blockchain is nil")
	ErrHeaderIsNil           = errors.New("header is nil")
	ErrGetStateDB            = errors.New("error get state db")
	ErrGetAccountAndPassword = errors.New("get account and password  error")
)

type SignHelper struct {
	mu sync.RWMutex
	am *accounts.Manager
	//signWallet   accounts.Wallet
	bc *core.BlockChain

	testMode bool
	testKey  *ecdsa.PrivateKey
}

func NewSignHelper() *SignHelper {
	return &SignHelper{
		am: nil,
		//signWallet:  nil,
		//	signAccount: accounts.Account{},
		testMode: false,
	}
}

func (sh *SignHelper) SetBc(eth MatrixEth) error {
	if eth.BlockChain() == nil {
		return ErrBlockChain
	}
	sh.bc = eth.BlockChain()
	return nil
}

func (sh *SignHelper) SetAccountManager(am *accounts.Manager, signAddress common.Address, signPassword string) error {
	if am == nil {
		return ErrNilAccountManager
	}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.am = am
	//sh.signWallet = nil
	////sh.signAccount = accounts.Account{}
	//
	//if (signAddress != common.Address{}) {
	//	return sh.resetSignAccount(signAddress, signPassword)
	//}

	return nil
}

//func (sh *SignHelper) ResetSignAccount(signAddress common.Address, signPassword string) error {
//	if (signAddress == common.Address{}) {
//		return ErrEmptySignAddress
//	}
//
//	sh.mu.Lock()
//	defer sh.mu.Unlock()
//
//	if sh.am == nil {
//		return ErrNilAccountManager
//	}
//
//	return sh.resetSignAccount(signAddress, signPassword)
//}

func (sh *SignHelper) SetTestMode(prvKey *ecdsa.PrivateKey) {
	sh.testMode = true
	sh.testKey = prvKey
}

//func (sh *SignHelper) resetSignAccount(signAddress common.Address, signPassword string) error {
//	if signAddress == sh.signAccount.Address {
//		sh.signPassword = signPassword
//		return nil
//	}
//
//	sh.signAccount.Address = signAddress
//	sh.signWallet = nil
//	wallet, err := sh.am.Find(sh.signAccount)
//	if err != nil {
//		return err
//	}
//	sh.signWallet = wallet
//	sh.signPassword = signPassword
//	return nil
//}

func (sh *SignHelper) SignHashWithValidate(hash []byte, validate bool, blkHash common.Hash) (common.Signature, error) {
	if sh.testMode {
		sign, err := crypto.SignWithValidate(hash, validate, sh.testKey)
		if err != nil {
			return common.Signature{}, err
		}
		return common.BytesToSignature(sign), nil
	}

	sh.mu.RLock()
	defer sh.mu.RUnlock()

	signAccount, signPassword, err := sh.getSignAccountAndPassword(blkHash)
	log.ERROR("5555555 SignHashWithValidate", "signAccount", signAccount, "signPassword", signPassword, "err", err, "blkhash", blkHash)
	if err != nil {
		return common.Signature{}, ErrGetAccountAndPassword
	}
	wallet, err := sh.am.Find(signAccount)
	if err != nil {
		return common.Signature{}, err
	}

	sign, err := wallet.SignHashValidateWithPass(signAccount, signPassword, hash, validate)
	if err != nil {
		return common.Signature{}, err
	}
	return common.BytesToSignature(sign), nil
}

func (sh *SignHelper) SignTx(tx types.SelfTransaction, chainID *big.Int, blkHash common.Hash) (types.SelfTransaction, error) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	//if nil == sh.signWallet {
	//	return nil, ErrUnSetSignAccount
	//}

	// Sign the requested hash with the wallet
	signAccount, signPassword, err := sh.getSignAccountAndPassword(blkHash)
	log.ERROR("5555555 SignTx", "signAccount", signAccount, "signPassword", signPassword, "err", err, "blkhash", blkHash)
	if err != nil {
		return nil, ErrGetAccountAndPassword
	}
	wallet, err := sh.am.Find(signAccount)
	if err != nil {
		return nil, err
	}
	return wallet.SignTxWithPassphrase(signAccount, signPassword, tx, chainID)
}

func (sh *SignHelper) SignVrf(msg []byte, blkHash common.Hash) ([]byte, []byte, []byte, error) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	//if nil==sh.signWallet{
	//	return []byte{},[]byte{},[]byte{},ErrUnSetSignAccount
	//}
	signAccount, signPassword, err := sh.getSignAccountAndPassword(blkHash)
	log.ERROR("5555555 vrf", "signAccount", signAccount, "signPassword", signPassword, "err", err, "blkhash", blkHash)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, ErrGetAccountAndPassword
	}
	wallet, err := sh.am.Find(signAccount)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}
	return wallet.SignVrfWithPass(signAccount, signPassword, msg)
}

func (sh *SignHelper) GetStateDependBlkHash(blkHash common.Hash) (*state.StateDB, uint64, error) {
	header := sh.bc.GetHeaderByHash(blkHash)
	if header == nil {
		log.Error("签名助手", "根据hash获取区块头失败 hash", blkHash)
		return nil, 0, ErrHeaderIsNil
	}
	height := header.Number.Uint64()
	log.ERROR("asdasd", "height", height)
	stateDb, err := sh.bc.StateAt(header.Root)
	return stateDb, height, err
}

func (sh *SignHelper) getSignAccountAndPassword(blkHash common.Hash) (accounts.Account, string, error) {
	stateDb, height, err := sh.GetStateDependBlkHash(blkHash)
	log.ERROR("55555", "GetSignAccountAndPassword 高度", height, "err", err, "blkHash", blkHash)
	if err != nil {
		return accounts.Account{}, "", ErrGetStateDB
	}
	addr, password, err := sh.bc.GetEntrustSignInfo(height, ca.GetAddress(), stateDb)
	account := accounts.Account{}
	account.Address = addr
	log.ERROR("555555", "returnaddr", account.Address, "password", password, "err", err)
	return account, password, err
}

func (sh *SignHelper) VerifySignWithValidateDependHash(sighash []byte, sig []byte, blkHash common.Hash) (common.Address, bool, error) {
	addr, flag, err := crypto.VerifySignWithValidate(sighash, sig)
	stateDb, height, err := sh.GetStateDependBlkHash(blkHash)
	log.ERROR("55555-VerifySignWithValidateDependHash", "height", height, "err", err, "blkHash", blkHash)
	if err != nil {
		log.ERROR("签名助手", "根据区块hash获取stateDB失败", "err")
		return common.Address{}, false, ErrGetStateDB
	}
	authAddr, err := sh.bc.GetAuthAddr(addr, height, stateDb)
	log.ERROR("55555-VerifySignWithValidateDependHash", "addr", addr, "height", height, "err", err, "authAddr", authAddr)
	return authAddr, flag, err

}
func (sh *SignHelper) VerifySignWithValidateDependNumber(sighash []byte, sig []byte, number uint64) (common.Address, bool, error) {
	header := sh.bc.GetHeaderByNumber(number)
	if header == nil {
		log.ERROR("5555-VerifySignWithValidateDependNumber", "number", number)
		return common.Address{}, false, errors.New("header is nil")
	}
	return sh.VerifySignWithValidateDependHash(sighash, sig, header.Hash())

}
