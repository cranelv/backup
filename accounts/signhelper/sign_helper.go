// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package signhelper

import (
	"math/big"

	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
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

type AuthReader interface {
	GetEntrustSignInfo(authFrom common.Address, blockHash common.Hash) (common.Address, string, error)
	GetAuthAccount(signAccount common.Address, blockHash common.Hash) (common.Address, error)
}

var (
	ModeLog="签名助手"
	ErrNilAccountManager     = errors.New("account manager is nil")
	ErrEmptySignAddress      = errors.New("sign address is empty")
	ErrUnSetSignAccount      = errors.New("The sign account not set yet!")
	ErrReader                = errors.New("auth reader is nil")
	ErrHeaderIsNil           = errors.New("header is nil")
	ErrGetStateDB            = errors.New("error get state db")
	ErrGetAccountAndPassword = errors.New("get account and password  error")
)

type SignHelper struct {
	mu sync.RWMutex
	am *accounts.Manager
	//signWallet   accounts.Wallet
	authReader AuthReader
}

func NewSignHelper() *SignHelper {
	return &SignHelper{
		am: nil,
		//signWallet:  nil,
		//	signAccount: accounts.Account{},
	}
}

func (sh *SignHelper) SetAuthReader(reader AuthReader) error {
	if reader == nil {
		return ErrReader
	}
	sh.authReader = reader
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

func (sh *SignHelper) SignHashWithValidateByReader(reader AuthReader, hash []byte, validate bool, blkHash common.Hash) (common.Signature, error) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	signAccount, signPassword, err := sh.getSignAccountAndPassword(reader, blkHash)
	log.ERROR(ModeLog, "signAccount", signAccount, "signPassword", signPassword, "err", err, "blkhash", blkHash)
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

func (sh *SignHelper) SignHashWithValidate(hash []byte, validate bool, blkHash common.Hash) (common.Signature, error) {
	return sh.SignHashWithValidateByReader(sh.authReader, hash, validate, blkHash)
}

func (sh *SignHelper) SignTx(tx types.SelfTransaction, chainID *big.Int, blkHash common.Hash) (types.SelfTransaction, error) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	//if nil == sh.signWallet {
	//	return nil, ErrUnSetSignAccount
	//}

	// Sign the requested hash with the wallet
	signAccount, signPassword, err := sh.getSignAccountAndPassword(sh.authReader, blkHash)
	log.ERROR(ModeLog, "signAccount", signAccount, "signPassword", signPassword, "err", err, "blkhash", blkHash)
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
	signAccount, signPassword, err := sh.getSignAccountAndPassword(sh.authReader, blkHash)
	log.ERROR(ModeLog, "signAccount", signAccount, "signPassword", signPassword, "err", err, "blkhash", blkHash)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, ErrGetAccountAndPassword
	}
	wallet, err := sh.am.Find(signAccount)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}
	return wallet.SignVrfWithPass(signAccount, signPassword, msg)
}

func (sh *SignHelper) getSignAccountAndPassword(reader AuthReader, blkHash common.Hash) (accounts.Account, string, error) {
	addr, password, err := reader.GetEntrustSignInfo(ca.GetAddress(), blkHash)
	account := accounts.Account{}
	account.Address = addr
	log.ERROR(ModeLog, "returnaddr", account.Address, "password", password, "err", err)
	return account, password, err
}

func (sh *SignHelper) VerifySignWithValidateDependHash(signHash []byte, sig []byte, blkHash common.Hash) (common.Address, bool, error) {
	addr, flag, err := crypto.VerifySignWithValidate(signHash, sig)

	authAddr, err := sh.authReader.GetAuthAccount(addr, blkHash)
	log.ERROR(ModeLog, "addr", addr, "height", blkHash.TerminalString(), "err", err, "authAddr", authAddr)
	return authAddr, flag, err
}

func (sh *SignHelper) VerifySignWithValidateByReader(reader AuthReader, signHash []byte, sig []byte, blkHash common.Hash) (common.Address, bool, error) {
	if reader == nil {
		return common.Address{}, false, ErrReader
	}
	addr, flag, err := crypto.VerifySignWithValidate(signHash, sig)

	authAddr, err := reader.GetAuthAccount(addr, blkHash)
	log.ERROR(ModeLog, "addr", addr, "height", blkHash.TerminalString(), "err", err, "authAddr", authAddr)
	return authAddr, flag, err
}
