// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or or http://www.opensource.org/licenses/mit-license.php

package mandb

/*
levelDB官方网站介绍的特点

**特点**：

- key和value都是任意长度的字节数组；
- entry（即一条K-V记录）默认是按照key的字典顺序存储的，当然开发者也可以重载这个排序函数；
- 提供的基本操作接口：Put()、Delete()、Get()、Batch()；
- 支持批量操作以原子操作进行；
- 可以创建数据全景的snapshot(快照)，并允许在快照中查找数据；
- 可以通过前向（或后向）迭代器遍历数据（迭代器会隐含的创建一个snapshot）；
- 自动使用Snappy压缩数据；
- 可移植性；

**限制**：

- 非关系型数据模型（NoSQL），不支持sql语句，也不支持索引；
- 一次只允许一个进程访问一个特定的数据库；
- 没有内置的C/S架构，但开发者可以使用LevelDB库自己封装一个server；


源码所在的目录在ethereum/ethdb目录。代码比较简单， 分为下面三个文件

- database.go                  levelDB的封装代码
- memory_database.go           供测试用的基于内存的数据库，不会持久化为文件，仅供测试
- interface.go                 定义了数据库的接口
- database_test.go             测试案例
---------------------
作者：尹成
来源：CSDN
原文：https://blog.csdn.net/itcastcpp/article/details/80305372
版权声明：本文为博主原创文章，转载请附上博文链接！
*/

/*不仅关键数据库参数有差异，java版本和go版本调用LevelDB后数据存放的目录也是不同的。以太坊的存储层存储着两类相对独立但又有联系的数据：区块链数据库（chainDB）和账户状态数据库（stateDB）。
其中，go版本将区块链数据库和账户状态数据库都存放在.ethereum目录下；而java版本将两者分开存放，分别放在block目录下和state目录下。


区块链数据库是一个区块编号和区块内容对应关系的数据库；而账户状态数据库是一个维护链中所有账户地址和其状态对应关系的数据库 ,
 以账户地址为key，以账户状态（包含nonce，余额，storageRoot，codeHash，见黄皮书4.1）为value。
 账户状态维护的是账户余额变动历史和合约账户执行历史,每次余额变动或合约代码被执行,都会生成一条记录,并被记录。
 所有账户状态数据库的查询，以账户地址为查询输入，而所有区块链上的查询，以区块编号等作为查询输入。
 技术上，可以理解为按照模块垂直划分成2个数据库实例。业务上，可以理解为交易流水账一个数据库实例，账户分户账一个数据库实例。

另外，以太坊还维护了一个节点信息的数据库，go版本的该数据库在nodes目录下，java版本的数据库在peers目录下，该部分是动态组网时所用，不是区块链本身内容。
---------------------
作者：郑泽洲
来源：CSDN
原文：https://blog.csdn.net/wxid2798226/article/details/83689579
版权声明：本文为博主原创文章，转载请附上博文链接！

*/
//看下面的代码，基本上定义了KeyValue数据库的基本操作， Put， Get， Has， Delete等基本操作，levelDB是不支持SQL的，基本可以理解为数据结构里面的Map。
// Code using batches should try to add this much data to the batch.
// The value was determined empirically.
const IdealBatchSize = 100 * 1024

// Putter wraps the database write operation supported by both batches and regular databases.
//Putter接口定义了批量操作和普通操作的写入接口
type Putter interface {
	Put(key []byte, value []byte) error
}

// Database wraps all database operations. All methods are safe for concurrent use.(并行使用)
//数据库接口定义了所有的数据库操作， 所有的方法都是多线程安全的。
type Database interface {
	Putter
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Close()
	NewBatch() Batch
}

// Batch is a write-only database that commits changes to its host database
// when Write is called. Batch cannot be used concurrently.
//批量操作接口，不能多线程同时使用，当Write方法被调用的时候，数据库会提交写入的更改。
type Batch interface {
	Putter
	ValueSize() int // amount of data in the batch
	Write() error
	// Reset resets the batch for reuse
	Reset()
}
