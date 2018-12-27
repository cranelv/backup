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
		CommitContext{
			Version:   "Gman_Alg_0.0.4",
			Submitter: "孙春风",
			Commit: []string{
				"换届服务漏合并的代码",
				"顶点在线修改可能panic的问题",
			},
		},
		{
			Version:   "Gman_Alg_0.0.5",
			Submitter: "Ryan",
			Commit: []string{
				"merge nodeId fixed version, modify bucket limit from two to four and modify broadcast block sender",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.6",
			Submitter: "孙春风",
			Commit: []string{
				"提供创世文件默认配置,(用户可选择性的填写创世文件,也可不填)",
			},
		},
		CommitContext{
			Version:   "Gman_Alg_0.0.7",
			Submitter: "yeying",
			Commit: []string{
				"修复发送定时交易或者24小时可撤销交易后重启节点导致区块root不一致的问题",
				"修复24小时可撤销交易正常执行完毕后在撤销该笔交易出现崩溃的问题",
				"修复同时发送定时交易和24小时可撤销交易，撤销其中的一笔交易后，转账金额没有减少的问题",
				"修复dump崩溃问题",
				"修改log",
				"deposit bug fixed",
			},
		},
	}
)
