// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package common

type CommitContext struct {
	Version   string
	Submitter string
	Commit    []string
}

var (
	PutCommit = []CommitContext{
		CommitContext{
			Version:   "Gman_Alg_0.0.1",
			Submitter: "孙春风,胡源凯",
			Commit: []string{
				"修改委托交易下的vrf失败问题",
				"pos参数配置有误",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.2",
			Submitter: "孙春风",
			Commit: []string{
				"出块趋向时间由1改为6",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.3",
			Submitter: "孙春风",
			Commit: []string{
				"删除开发者模式 删除测试网模式 删除rinkeby模式",
				"禁用默认创世文件",
				"委托交易账户外部可见改为man账户",
			},
		},

	}
)
