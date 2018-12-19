// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php
package baseinterface

import (
	"math/big"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core"
	"github.com/matrix/go-matrix/event"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/mc"
	"github.com/matrix/go-matrix/params/manparams"
	"fmt"
)

const (
	ModuleRandom = "随机数接口服务"
)

var (
	mapReg = make(map[string]func(string, RandomChainSupport) (RandomSubService, error), 0)
)

func RegRandom(name string, fun func(string, RandomChainSupport) (RandomSubService, error)) {
	mapReg[name] = fun
}

type Random struct {
	roleUpdateCh  chan *mc.RoleUpdatedMsg
	roleUpdateSub event.Subscription
	quitChan      chan struct{}
	mapSubService map[string]RandomSubService
}

type RandomChainSupport interface {
	BlockChain() *core.BlockChain
}
type RandomSubService interface {
	Prepare(uint64) error
	CalcData(data common.Hash) (*big.Int, error)
}
func checkDataValidity(support interface{})bool{
	return common.IsNil(support)
}
func NewRandom(support RandomChainSupport) (*Random, error) {
	//if checkDataValidity(support)==false{
	//	log.Error(ModuleRandom,"创建随机数服务阶段,输入不合法","输入为空接口")
	//	return nil,errors.New("创建随机数服务阶段,输入不合法")
	//}
	random := &Random{
		roleUpdateCh:  make(chan *mc.RoleUpdatedMsg, 1),
		quitChan:      make(chan struct{}, 1),
		mapSubService: make(map[string]RandomSubService, 0),
	}
	for _, name := range manparams.RandomServiceName {
		Plug, needNewSubService := getSubServicePlug(name)
		if needNewSubService == false {
			log.Warn(ModuleRandom, "新建子服务阶段,子服务不需要被创建 名称", name)
			continue
		}
		if err := random.newSubServer(name, Plug, support); err != nil {
			log.Error(ModuleRandom, "新建子服务阶段,子服务创建失败 名称", name)
			return nil, err
		}
		log.Info(ModuleRandom, "新建子服务阶段,子服务创建成功 名称", name)
	}

	var err error
	random.roleUpdateSub, err= mc.SubscribeEvent(mc.CA_RoleUpdated, random.roleUpdateCh)
	if err != nil {
		log.Error(ModuleRandom, "订阅CA消息阶段,CA消息订阅失败 err", err)
		return nil, err
	}
	go random.update()
	return random, nil
}

func (self *Random) update() {
	defer self.roleUpdateSub.Unsubscribe()
	for {
		select {
		case RoleUpdateData := <-self.roleUpdateCh:
			go self.processRoleUpdateData(RoleUpdateData)
		case <-self.quitChan:
			return
		}
	}
}

func (self *Random) Stop() {
	close(self.quitChan)
}
func (self *Random) processRoleUpdateData(data *mc.RoleUpdatedMsg) {
	for _, v := range self.mapSubService {
		go v.Prepare(data.BlockNum)
	}
}

func (self *Random) newSubServer(name string, plugConfig string, support RandomChainSupport) error {
	var err error
	if _, ok := mapReg[name]; ok == false {
		log.Error(ModuleRandom, "新建子服务阶段,该子服务未注册", name)
		return fmt.Errorf("该子服务未注册 %v",name)
	}
	if self.mapSubService[name], err = mapReg[name](plugConfig, support); err != nil {
		log.Error(ModuleRandom, "新建子服务阶段,该子服务新建失败",name,"err",err)
	}
	log.Info(ModuleRandom, "新建子服务阶段,该子服务创建成功 index", name)
	return nil
}

func (self *Random) GetRandom(hash common.Hash, Type string) (*big.Int, error) {
	return self.mapSubService[Type].CalcData(hash)
}

func getSubServicePlug(name string) (string, bool) {
	plug, ok := manparams.RandomConfig[name]
	if ok == false {
		log.Warn(ModuleRandom, "获取插件阶段,配置中无该子服务,不需要开启", name)
		return "", false
	}

	plugs, ok := manparams.RandomServicePlugs[name]
	if ok == false {
		log.Error(ModuleRandom, "获取插件阶段 无该子服务 服务名称", name)
		return "", false
	}
	for _, v := range plugs {
		if v == plug {
			log.Info(ModuleRandom, "获取插件阶段", "", "插件列表中有该插件", plug)
			return v, true
		}
	}
	log.Warn(ModuleRandom, "获取插件阶段,配置中的插件不合法，使用默认插件 名称", manparams.RandomServiceDefaultPlugs[name])
	return manparams.RandomServiceDefaultPlugs[name], true
}
