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

	}
)
